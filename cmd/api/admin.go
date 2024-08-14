package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/validator"
)

// adminGetAllUsersHandler() is the admin endpoint that returns all available
// users in our DB. This route supports a full text search for the user Name as well
// as pagination and sorting.
func (app *application) adminGetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	// prepare the input struct, ready to receive the query string data + filters
	var input struct {
		Name string
		data.Filters
	}
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	//get the page & pagesizes as ints and set to the embedded struct
	input.Name = app.readString(qs, "name", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// get the sort values falling back to "created_at" if it is not provided
	input.Filters.Sort = app.readString(qs, "sort", "created_at")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"created_at", "-created_at"}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	users, metadata, err := app.models.Admin.AdminGetAllUsers(input.Name, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// write the data and metadata
	err = app.writeJSON(w, http.StatusOK, envelope{"users": users, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// adminGetStatisticsHandler() is an admin endpoint that returns all the statistics,aggregated
// together for representation in the frontend. It's to be used in tandem with the debug
// and health endpoints.
func (app *application) adminGetStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	// get the statistics
	stats, err := app.models.Admin.AdminGetStatistics()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// get the healthcheck data
	env := app.returnEnvInfo()
	// write the data
	err = app.writeJSON(w, http.StatusOK, envelope{"statistics": stats, "health": env}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// addPermissionsForUserHandler() is the endpoint handler responsible for allowing an admin
// user to add permissions for a specific user. It expects a JSON request containing a
// permissions array and a user's ID.
func (app *application) addPermissionsForUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Permissions []string `json:"permissions"`
		UserID      int64    `json:"user_id"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Validate data
	userPermissions := &data.UserPermission{
		UserID:      input.UserID,
		Permissions: input.Permissions,
	}
	v := validator.New()
	if data.ValidatePermissionsAddition(v, userPermissions); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// insert permissions
	userPermission, err := app.models.Permissions.AddPermissionsForUser(userPermissions.UserID, userPermissions.Permissions...)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicatePermission):
			v.AddError("permissions", "user already has these permissions")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// write the data back showing a successful creation
	err = app.writeJSON(w, http.StatusCreated, envelope{"user_permission": userPermission}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deletePermissionsForUserHandler() is the endpoint handler responsible for allowing an admin
// user to delete permissions for a specific user. It expects a parameterized url taking
// in the permission code and user id i.e /v1/admin/:permissionCode/:userID
func (app *application) deletePermissionsForUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := app.readIDIntParam(r, "userID")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	app.logger.PrintInfo(fmt.Sprintf("user id: %d", userID), nil)
	permissionCode, err := app.readIDStrParam(r, "pCode")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// validate data
	v := validator.New()
	if data.ValidatePermissionsDeletion(v, userID, permissionCode); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// delete permissions
	_, err = app.models.Permissions.DeletePermissionsForUser(userID, permissionCode)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrPermissionNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// write the data back with a message on success
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "permission(s) deleted successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
