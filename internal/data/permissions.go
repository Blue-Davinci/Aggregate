package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/blue-davinci/aggregate/internal/validator"
)

var (
	ErrDuplicatePermission = errors.New("duplicate permission")
	ErrPermissionNotFound  = errors.New("permission not found")
)
var (
	PermissionAdminWrite = "admin:write"
	PermissionAdminRead  = "admin:read"
)

// Define the PermissionModel type.
type PermissionModel struct {
	DB *database.Queries
}

type UserPermission struct {
	PermissionID int64    `json:"permission_id"`
	UserID       int64    `json:"user_id"`
	Permissions  []string `json:"permissions"`
}

func ValidatePermissionsAddition(v *validator.Validator, permissions *UserPermission) {
	v.Check(len(permissions.Permissions) != 0, "permissions", "must be provided")
	//v.Check()
	v.Check(permissions.UserID != 0, "user_id", "must be provided")
}
func ValidatePermissionsDeletion(v *validator.Validator, userID int64, permissionCode string) {
	v.Check(permissionCode != "", "codes", "must be provided")
	//v.Check()
	v.Check(userID != 0, "user_id", "must be provided")
}

// Make a slice to hold the the permission codes (like
// "admin:read" and "admin:write") for an admin user.
type Permissions []string

// Add a helper method to check whether the Permissions slice contains a specific
// permission code.
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

// GetAllPermissionsForUser() is a method that retrieves all permissions for a specific user
// from the database. It expects the user's ID as input and returns a slice of permission codes.
func (m PermissionModel) GetAllPermissionsForUser(userID int64) (Permissions, error) {
	// set up context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// create our permissions
	var permissions Permissions
	// call the database method
	dbPermissions, err := m.DB.GetAllPermissionsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, permission := range dbPermissions {
		permissions = append(permissions, permission)
	}
	// return permissions
	return permissions, nil
}

// AddPermissionsForUser() is an admin method that adds permissions for a specific user
// in the database. It expects the user's ID and a slice of permission codes as input.
func (m PermissionModel) AddPermissionsForUser(userID int64, codes ...string) (*UserPermission, error) {
	// setup our context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// insert our permissions
	queryResult, err := m.DB.AddPermissionsForUser(ctx, database.AddPermissionsForUserParams{
		UserID:  userID,
		Column2: codes,
	})
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_permissions_pkey"`:
			return nil, ErrDuplicatePermission
		default:
			return nil, err
		}
	}
	// create our permissions
	userPermission := &UserPermission{
		PermissionID: queryResult.PermissionID,
		UserID:       queryResult.UserID,
		Permissions:  codes,
	}
	// return the permissions
	return userPermission, nil
}

func (m PermissionModel) DeletePermissionsForUser(userID int64, permissionCode string) (int64, error) {
	// Setup our context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the deletion query
	permissionID, err := m.DB.DeletePermissionsForUser(ctx, database.DeletePermissionsForUserParams{
		UserID: userID,
		Code:   permissionCode,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, ErrPermissionNotFound
		default:
			return 0, err
		}
	}

	return permissionID, nil
}
