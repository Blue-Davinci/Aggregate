package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"time"

	"encoding/base32"

	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/blue-davinci/aggregate/internal/validator"
)

type ApiKeyModel struct {
	DB *database.Queries
}

// Define constants for the token scope.
const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
	ScopePasswordReset  = "password-reset"
	// Define the lengths of the API key and token verification strings.
	// The Key Length represents the initial size before encoding
	// The Verification Length represents the final size after encoding which
	// is 26 for the signup tokens and 32 for the api keys
	APIKeyLength            = 20
	APIVerificationLength   = 32
	TokenKeyLength          = 16
	TokenVerificationLength = 26
)

type ApiKey struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func (m ApiKeyModel) New(userID int64, ttl time.Duration, scope string, size int) (*ApiKey, error) {
	api_key, err := generateAPI(userID, ttl, scope, size)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("API Key: %v\n || User ID: %d", api_key, userID)
	// insert the api key into the database
	err = m.Insert(api_key)
	return api_key, err
}
func generateAPI(userID int64, ttl time.Duration, scope string, size int) (*ApiKey, error) {
	api_key := &ApiKey{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}
	// Initialize a 0 valued byte slice with len=20, len=16 giving us 32 & 20 chars
	randomBytes := make([]byte, size)
	// Use the Read() function from the crypto/rand package to fill the byte slice with random bytes
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	api_key.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// generate a sha256 hash of our plaintext token\apikey. We will store this in the db
	// and send back the original plaintext token\apikey to the user which should look like.
	//	Token:		"SHLO6UKE33F7BTCRMJPIKPTCKI"	   for example, all 26chars long!
	// 	Api Key:	"ZMRX2REGM66XT5QLGCVI25KT3Z7FJW63" for example, all 32chars long!
	hash := sha256.Sum256([]byte(api_key.Plaintext))
	api_key.Hash = hash[:]
	return api_key, nil
}
func (m ApiKeyModel) Insert(api_key *ApiKey) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := m.DB.InsertApiKey(ctx, database.InsertApiKeyParams{
		ApiKey: api_key.Hash,
		UserID: api_key.UserID,
		Expiry: api_key.Expiry,
		Scope:  api_key.Scope,
	})
	return err
}

// DeleteAllForUser() deletes all tokens for a specific user and scope.
func (m ApiKeyModel) DeleteAllForUser(scope string, userID int64) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := m.DB.DeletAllAPIKeysForUser(ctx, database.DeletAllAPIKeysForUserParams{
		UserID: userID,
		Scope:  scope,
	})
	return err
}

// Validation
// Check that the plaintext api key has been provided and is exactly 26 & 32 bytes long.
// for activation, we will use 26 bytes and for authentication, we will use 32 bytes
func ValidateAPIKeyPlaintext(v *validator.Validator, ApiKeyPlaintext string, size int) {
	info := fmt.Sprintf("must be %d bytes long", size)
	v.Check(ApiKeyPlaintext != "", "api_key", "must be provided")
	v.Check(len(ApiKeyPlaintext) == size, "api_key", info)
}
