// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ApiKey struct {
	ApiKey []byte
	UserID int64
	Expiry time.Time
	Scope  string
}

type Comment struct {
	ID              uuid.UUID
	PostID          uuid.UUID
	UserID          int64
	ParentCommentID uuid.NullUUID
	CommentText     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Version         int32
}

type CommentNotification struct {
	ID        int32
	CommentID uuid.UUID
	PostID    uuid.UUID
	UserID    int64
	CreatedAt time.Time
}

type Feed struct {
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Name            string
	Url             string
	Version         int32
	UserID          int64
	ImgUrl          string
	LastFetchedAt   sql.NullTime
	FeedType        string
	FeedDescription string
	IsHidden        bool
}

type FeedFollow struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    int64
	FeedID    uuid.UUID
}

type Notification struct {
	ID        int32
	FeedID    uuid.UUID
	FeedName  string
	PostCount int32
	CreatedAt time.Time
}

type Postfavorite struct {
	ID        int64
	PostID    uuid.UUID
	FeedID    uuid.UUID
	UserID    int64
	CreatedAt time.Time
}

type RssfeedPost struct {
	ID                 uuid.UUID
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Channeltitle       string
	Channelurl         sql.NullString
	Channeldescription sql.NullString
	Channellanguage    sql.NullString
	Itemtitle          string
	Itemdescription    sql.NullString
	ItempublishedAt    time.Time
	Itemurl            string
	ImgUrl             string
	FeedID             uuid.UUID
}

type User struct {
	ID           int64
	CreatedAt    time.Time
	Name         string
	Email        string
	PasswordHash []byte
	Activated    bool
	Version      int32
}
