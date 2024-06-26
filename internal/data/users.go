package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/blue-davinci/aggregate/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *database.Queries
}

// Define a custom ErrDuplicateEmail error.
var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

// Declare a new AnonymousUser variable.
var AnonymousUser = &User{}

// Check if a User instance is the AnonymousUser.
func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// Create a custom password type which is a struct containing the plaintext and hashed
// versions of the password for a user.
type password struct {
	plaintext *string
	hash      []byte
}

// set() calculates the bcrypt hash of a plaintext password, and stores both
// the hash and the plaintext versions in the struct.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

// The Matches() method checks whether the provided plaintext password matches the
// hashed password stored in the struct, returning true if it matches and false
// otherwise.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		//fmt.Printf(">>>>> Plain text: %s\nHash: %v\n", plaintextPassword, p.hash)
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// The user struct represents a user account in our application. It contains fields for
// the user ID, created timestamp, name, email address, password hash, and activation data
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}
func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")
	// Call the standalone ValidateEmail() helper.
	ValidateEmail(v, user.Email)
	// If the plaintext password is not nil, call the standalone
	// ValidatePasswordPlaintext() helper.
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	// If the password hash is ever nil, this will be due to a logic error in our
	// codebase. So rather than adding an error to the validation map we
	// raise a panic instead.
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

// Insert() creates a new User and returns success on completion.
// The function will also check for the uniqueness of the user email.
// Note, this will only "Sign Up" our USER, not log them in.
func (m UserModel) Insert(user *User) error {
	// Create a new context with a 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	createduser, err := m.DB.CreateUser(ctx, database.CreateUserParams{
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.Password.hash,
		Activated:    user.Activated,
	})
	user.ID = createduser.ID
	user.CreatedAt = createduser.CreatedAt
	user.Version = int(createduser.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	// Calculate the SHA-256 hash of the plaintext token provided by the client.
	// Remember that this returns a byte *array* with length 32, not a slice.
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	queryresult, err := m.DB.GetForToken(ctx, database.GetForTokenParams{
		ApiKey: tokenHash[:],
		Scope:  tokenScope,
		Expiry: time.Now(),
	})
	// check for any error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Create a new password struct instance for the user.
	userPassword := password{
		hash: queryresult.PasswordHash,
	}
	//fmt.Printf(">>~~>>> User Password Hash: %v\n", userPassword.hash)
	// Now let us write to our user struct and return it
	user := User{
		ID:        queryresult.ID,
		CreatedAt: queryresult.CreatedAt,
		Name:      queryresult.Name,
		Email:     queryresult.Email,
		Password:  userPassword,
		Activated: queryresult.Activated,
		Version:   int(queryresult.Version),
	}
	return &user, nil
}

// GetByEmail() method Retrieves the User details from the database based on the user's email address.
// Because we have a UNIQUE constraint on the email column, this SQL query will only
// return one record (or none at all, in which case we return a ErrRecordNotFound error).
func (m UserModel) GetByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	queryresult, err := m.DB.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	userPassword := password{
		hash: queryresult.PasswordHash,
	}
	user := &User{
		ID:        queryresult.ID,
		CreatedAt: queryresult.CreatedAt,
		Name:      queryresult.Name,
		Email:     queryresult.Email,
		Password:  userPassword,
		Activated: queryresult.Activated,
		Version:   int(queryresult.Version),
	}
	return user, nil
}

// Update() method updates the details for a specific user. We check against the version
// field to help prevent any race conditions during the request cycle.
// We also check for a violation of the "users_email_key" constraint when performing the update.
func (m UserModel) Update(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	queryresult, err := m.DB.Update(ctx, database.UpdateParams{
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.Password.hash,
		Activated:    user.Activated,
		Version:      int32(user.Version),
		ID:           user.ID,
	})
	// check for any errors including duplicated emails
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	// update the user struct with the new version number
	user.Version = int(queryresult)
	return nil
}
