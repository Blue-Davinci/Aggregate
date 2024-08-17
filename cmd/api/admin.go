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

// adminGetPaymentPlansHandler() is the endpoint handler responsible for returning all the
// available payment plans regardless of their status, i.e active or inactive
func (app *application) adminGetPaymentPlansHandler(w http.ResponseWriter, r *http.Request) {
	plans, err := app.models.Admin.AdminGetPaymentPlans()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"plans": plans}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// adminGetAllSubscriptionsHandler() is the endpoint handler responsible for returning all the
// available subscriptions in the DB. It supports pagination and sorting.
func (app *application) adminGetAllSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.Filters
	}
	//validate if queries are provided
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	//get the pagesizes as ints and set to the embedded struct
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// get the sort values falling back to "id" if it is not provided
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"id", "-id"}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	subscriptions, metadata, err := app.models.Admin.AdminGetAllSubscriptions(input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"subscriptions": subscriptions, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// adminGetAllPermissionsHandler() is an admin endpoint that returns all the available
// permissions in the system. This is useful for the admin to see what permissions are
// available to assign to users.
func (app *application) adminGetAllPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	permissions, err := app.models.Permissions.GetAllPermissions()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"permissions": permissions}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// adminGetAllSuperUsersWithPermissionsHandler() returns all the users with active permissions
// in the system. If you add additional permissions even for regular users such as
// commern:write you'll need to filter those in here
func (app *application) adminGetAllSuperUsersWithPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	superUsers, err := app.models.Admin.AdminGetAllSuperUsersWithPermissions()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"super_users": superUsers}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) adminCreateNewPermissionHandler(w http.ResponseWriter, r *http.Request) {
	// we take in a permission in the format permissioncode:permission
	// e.g admin:read
	var input struct {
		Permission string `json:"permission"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// get the permission
	permission := input.Permission
	// validate the data
	v := validator.New()
	if data.ValidatePermission(v, permission); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// insert the permission
	newPermission, err := app.models.Admin.AdminCreateNewPermission(permission)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicatePermission):
			v.AddError("permission", "permission already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// write the data back
	err = app.writeJSON(w, http.StatusCreated, envelope{"permission": newPermission}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// addPermissionsForUserHandler() is the endpoint handler responsible for allowing an admin
// user to add permissions for a specific user. It expects a JSON request containing a
// permissions array and a user's ID.
func (app *application) adminAddPermissionsForUserHandler(w http.ResponseWriter, r *http.Request) {
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

// adminCreatePaymentPlansHandler() is an admin route that allows the admin
// to create a new payment/subscription plan. Any plan created and set to 'active'
// will be shown to all other users. To hide plans, the status should be set to = 'inactive'
func (app *application) adminCreatePaymentPlansHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string   `json:"name"`
		Image       string   `json:"image"`
		Description string   `json:"description"`
		Duration    string   `json:"duration"`
		Price       int64    `json:"price"`
		Features    []string `json:"features"`
		Status      string   `json:"status"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// create the payment plan
	paymentPlan := &data.Payment_Plan{
		Name:        input.Name,
		Image:       input.Image,
		Description: input.Description,
		Duration:    input.Duration,
		Price:       input.Price,
		Features:    input.Features,
		Status:      input.Status,
	}
	// validate the data
	v := validator.New()
	if data.ValidatePaymentPlan(v, paymentPlan); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// insert our data
	err = app.models.Admin.AdminCreatePaymentPlans(paymentPlan)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicatePaymentPlan):
			v.AddError("name", "a payment plan with this name already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// write the data back
	err = app.writeJSON(w, http.StatusCreated, envelope{"payment_plan": paymentPlan}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// adminUpdatePaymentPlanHandler() is the admin endpoint that is responsible for updating
// a payment plan in the DB. It expects a JSON request containing the updated fields.
// This supports partial updates in that, only the fields that are provided will be updated.
func (app *application) adminUpdatePaymentPlanHandler(w http.ResponseWriter, r *http.Request) {
	// we get the payment ID from the URL
	paymentID, err := app.readIDIntParam(r, "planID")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// quick check for the id
	if paymentID < 1 {
		app.notFoundResponse(w, r)
		return
	}
	// get our plan information from the DB
	paymentPlan, err := app.models.Admin.AdminGetPaymentPlanByID(int32(paymentID))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Have the input field that holds the incoming data
	var input struct {
		Name        *string  `json:"name"`
		Image       *string  `json:"image"`
		Description *string  `json:"description"`
		Duration    *string  `json:"duration"`
		Price       *int64   `json:"price"`
		Features    []string `json:"features"`
		Status      *string  `json:"status"`
	}
	// read the incoming data
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// check if the input the input fields are empty
	// if they are nil, we know they need updates
	if input.Name != nil {
		paymentPlan.Name = *input.Name
	}
	if input.Image != nil {
		paymentPlan.Image = *input.Image
	}
	if input.Description != nil {
		paymentPlan.Description = *input.Description
	}
	if input.Duration != nil {
		paymentPlan.Duration = *input.Duration
	}
	if input.Price != nil {
		paymentPlan.Price = *input.Price
	}
	if input.Features != nil {
		paymentPlan.Features = input.Features
	}
	if input.Status != nil {
		paymentPlan.Status = *input.Status
	}
	// initialize the payment plan validator
	v := validator.New()
	if data.ValidatePaymentPlan(v, paymentPlan); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// update the payment plan
	err = app.models.Admin.AdminUpdatePaymentPlan(paymentPlan)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// return a status OK and the updated payment plan
	err = app.writeJSON(w, http.StatusOK, envelope{"payment_plan": paymentPlan}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// adminUpdatePermissionCodeHandler() is the endpoint handler responsible for allowing an admin
// user to update the permission code for a specific permission. It expects a JSON request
// containing the new permission code and a URL param with the code.
func (app *application) adminUpdatePermissionCodeHandler(w http.ResponseWriter, r *http.Request) {
	// we get the ID from the URL
	permissionID, err := app.readIDIntParam(r, "pCode")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// a reciever struct to obtain the new permission code
	var input struct {
		Permission string `json:"permission"`
	}
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// create a ppermission
	adminPermission := data.AdminPermission{
		Permission_ID:   permissionID,
		Permission_Code: input.Permission,
	}
	// validate the data
	v := validator.New()
	if data.ValidatePermissionsDeletion(v, adminPermission.Permission_ID, adminPermission.Permission_Code); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// update the permission
	permission, err := app.models.Admin.AdminUpdatePermissionCode(adminPermission)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrPermissionNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// write the data back
	err = app.writeJSON(w, http.StatusOK, envelope{"permission": permission}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deletePermissionsForUserHandler() is the endpoint handler responsible for allowing an admin
// user to delete permissions for a specific user. It expects a parameterized url taking
// in the permission code and user id i.e /v1/admin/:permissionCode/:userID
func (app *application) adminDeletePermissionsForUserHandler(w http.ResponseWriter, r *http.Request) {
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

func (app *application) adminDeletePermissionHandler(w http.ResponseWriter, r *http.Request) {
	permissionID, err := app.readIDIntParam(r, "pCode")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// validate data
	if permissionID == 0 {
		app.notFoundResponse(w, r)
		return
	}
	// delete the permission
	err = app.models.Admin.AdminDeletePermission(permissionID)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "permission deleted successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
