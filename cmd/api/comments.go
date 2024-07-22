package main

import (
	"errors"
	"net/http"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/google/uuid"
)

// createCommentHandler creates a new comment on a post
// Accepts a post_id, parent_comment_id and comment_text
func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the comment data
	var input struct {
		Post_ID           uuid.UUID     `json:"post_id"`
		Parent_Comment_ID uuid.NullUUID `json:"parent_comment_id"`
		Comment_Text      string        `json:"comment_text"`
	}
	// Read our data
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Check if parent comment id is nil, if so we use uuid.Nil
	// This signifies a parent comment rather than a child comment. Can also
	// Help in filtering in future updates
	if !input.Parent_Comment_ID.Valid {
		input.Parent_Comment_ID = uuid.NullUUID{UUID: uuid.Nil, Valid: false}
	}
	// Create a new comments
	comment := &data.Comment{
		ID:                uuid.New(),
		Post_ID:           input.Post_ID,
		User_ID:           app.contextGetUser(r).ID,
		Parent_Comment_ID: input.Parent_Comment_ID,
		Comment_Text:      input.Comment_Text,
	}
	// validate the Post ID and comment text
	v := validator.New()
	if data.ValidateComment(v, comment); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Create the comment
	err = app.models.Comments.CreateComment(comment)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return the comment with a 201 Created status code
	err = app.writeJSON(w, http.StatusCreated, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getCommentsForPostHandler retrieves all comments for a specific post based
// on the posts Post ID
func (app *application) getCommentsForPostHandler(w http.ResponseWriter, r *http.Request) {
	//  Read our post ID as a parameter from the URL
	postID, err := app.readIDParam(r, "postID")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// do a quick validate for the UUID
	_, isValid := data.ValidateUUID(postID.String())
	if !isValid {
		app.badRequestResponse(w, r, err)
		return
	}
	// Get all comments for the post
	comments, err := app.models.Comments.GetCommentsForPost(postID, app.contextGetUser(r).ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return the comments
	err = app.writeJSON(w, http.StatusOK, envelope{"comments": comments}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentID, err := app.readIDParam(r, "commentID")
	//check whether there's an error or the feedID is invalid
	if err != nil || commentID == uuid.Nil {
		app.notFoundResponse(w, r)
		return
	}
	// quick validation
	_, isValid := data.ValidateUUID(commentID.String())
	if !isValid {
		app.badRequestResponse(w, r, err)
		return
	}
	// Delete the comment
	err = app.models.Comments.DeleteComment(commentID, app.contextGetUser(r).ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrCommentNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return a 200 OK status code
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "comment deleted successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// updateUserCommentHandler updates a comment taking in the ID, comment_text and version
// We take in the version to ensure that the user is updating the correct comment and also
// safeguard against race conditions if it may ever be an issue in the future.
func (app *application) updateUserCommentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID           uuid.UUID `json:"id"`
		Comment_Text string    `json:"comment_text"`
		Version      int32     `json:"version"`
	}
	// Read our data
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Create a new update comment
	comment := &data.Comment{
		ID:           input.ID,
		Comment_Text: input.Comment_Text,
		Version:      input.Version,
	}
	// validate the Comment ID and Comment text
	v := validator.New()
	if data.ValidateUpdateComment(v, comment); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Check if the comment exists
	_, err = app.models.Comments.GetCommentByID(comment.ID, app.contextGetUser(r).ID)
	// if there is an error, we check if it is a not found error
	if err != nil {
		switch {
		case errors.Is(err, data.ErrCommentNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Update the comment
	version, err := app.models.Comments.UpdateUserComment(comment, app.contextGetUser(r).ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return the comment with a 200 OK status code
	err = app.writeJSON(w, http.StatusOK,
		envelope{
			"message": "comment updated successfully",
			"version": version,
		}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
