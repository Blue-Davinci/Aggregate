package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/google/uuid"
)

var ErrCommentNotFound = errors.New("comment not found")

type CommentsModel struct {
	DB *database.Queries
}

// This PostComment struct represent the data we will send back to the user
// We include the username of the user who made the comment and the comment itself
type PostComment struct {
	Comment   Comment `json:"comment"`
	User_Name string  `json:"user_name"`
}

// The Comment struct represents what our what our comments look like
// We will recieve a comment from a user
// Not IsEditable is a field that will be used to determine if a comment is editable
// for a specific user. If this user sending this request is the owner of the comment
// then isEditable will be true otherwise it'll be false.
type Comment struct {
	ID                uuid.UUID     `json:"id"`
	Post_ID           uuid.UUID     `json:"post_id"`
	User_ID           int64         `json:"user_id"`
	Parent_Comment_ID uuid.NullUUID `json:"parent_comment_id"`
	Comment_Text      string        `json:"comment_text"`
	Created_At        time.Time     `json:"created_at"`
	Updated_At        time.Time     `json:"updated_at"`
	IsEditable        bool          `json:"is_editable"`
	Version           int32         `json:"version"`
}

func ValidateComment(v *validator.Validator, comment *Comment) {
	// Check that the post ID is provided
	v.Check(comment.Post_ID != uuid.Nil, "post_id", "must be provided")
	_, isvalid := ValidateUUID(comment.Post_ID.String())
	v.Check(isvalid, "post id", "must be a valid UUID")
	// Check that the comment text is provided and is not more than 500 bytes long
	v.Check(comment.Comment_Text != "", "comment_text", "must be provided")
	v.Check(len(comment.Comment_Text) <= 500, "comment_text", "must not be more than 500 bytes long")
}

func ValidateUpdateComment(v *validator.Validator, comment *Comment) {
	// Check that the comment ID is provided
	v.Check(comment.ID != uuid.Nil, "comment_id", "must be provided")
	_, isvalid := ValidateUUID(comment.ID.String())
	v.Check(isvalid, "comment id", "must be a valid UUID")
	// Check that the comment text is provided and is not more than 500 bytes long
	v.Check(comment.Comment_Text != "", "comment_text", "must be provided")
	v.Check(len(comment.Comment_Text) <= 500, "comment_text", "must not be more than 500 bytes long")
	// Check that the version is provided
	v.Check(comment.Version != 0, "version", "must be valid and provided")
}

// CreateComment() creates a new comment in the database, we get the
// post id, user id and parent comment id from our comment and save it.
func (m CommentsModel) CreateComment(comment *Comment) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Insert the comment into the database
	queryresult, err := m.DB.CreateComments(ctx, database.CreateCommentsParams{
		ID:              comment.ID,
		PostID:          comment.Post_ID,
		UserID:          comment.User_ID,
		ParentCommentID: comment.Parent_Comment_ID,
		CommentText:     comment.Comment_Text,
	})

	if err != nil {
		return err
	}
	// set additional details
	comment.Created_At = queryresult.CreatedAt
	comment.Updated_At = queryresult.UpdatedAt
	// Nowwe need to save the comment notification
	err = m.CreateCommentNotification(comment.User_ID, comment.ID, comment.Post_ID)
	if err != nil {
		return err
	}
	// we are good, hopefully no error.
	return nil
}

func (m CommentsModel) UpdateUserComment(comment *Comment, userID int64) (int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Update the comment in the database
	queryresult, err := m.DB.UpdateUserComment(ctx, database.UpdateUserCommentParams{
		ID:          comment.ID,
		CommentText: comment.Comment_Text,
		UserID:      userID,
		Version:     comment.Version,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, ErrEditConflict
		default:
			return 0, err
		}
	}
	return queryresult, nil
}

func (m CommentsModel) GetCommentByID(id uuid.UUID, userID int64) (*Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Get the comment from the database
	row, err := m.DB.GetCommentByID(ctx, database.GetCommentByIDParams{
		ID:     id,
		UserID: userID,
	})
	// If the comment is not found, return a specific error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrCommentNotFound
		default:
			return nil, err
		}
	}
	// Otherwise, return the comment
	comment := &Comment{
		ID:                row.ID,
		Post_ID:           row.PostID,
		User_ID:           row.UserID,
		Parent_Comment_ID: row.ParentCommentID,
		Comment_Text:      row.CommentText,
		Created_At:        row.CreatedAt,
		Version:           row.Version,
	}
	return comment, nil
}

// GetCommentsForPost() returns all comments for a specific post
func (m CommentsModel) GetCommentsForPost(id uuid.UUID, userID int64) ([]*PostComment, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Get the comments from the backend for a specific post
	rows, err := m.DB.GetCommentsForPost(ctx, database.GetCommentsForPostParams{
		PostID: id,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	comments := []*PostComment{}
	for _, row := range rows {
		var postComment PostComment
		comment := &Comment{
			ID:                row.ID,
			Post_ID:           row.PostID,
			User_ID:           row.UserID,
			Parent_Comment_ID: row.ParentCommentID,
			Comment_Text:      row.CommentText,
			Created_At:        row.CreatedAt,
			IsEditable:        row.Iseditable,
			Version:           row.Version,
		}
		username := row.UserName
		postComment = PostComment{Comment: *comment, User_Name: username}
		comments = append(comments, &postComment)
	}
	return comments, nil
}

// CreateCommentNotification() Creates a new comment notification in the database
// This notification will be included back in our getnotification function.
func (m CommentsModel) CreateCommentNotification(userID int64, commentID, postID uuid.UUID) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Insert the comment into the database
	_, err := m.DB.CreateCommentNotification(ctx, database.CreateCommentNotificationParams{
		UserID:    userID,
		PostID:    postID,
		CommentID: commentID,
	})
	if err != nil {
		return err
	}
	return nil
}
