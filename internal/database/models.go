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

type ChallengedTransaction struct {
	ID                       int64
	UserID                   int64
	ReferencedSubscriptionID uuid.UUID
	AuthorizationUrl         string
	Reference                string
	CreatedAt                time.Time
	UpdatedAt                time.Time
	Status                   string
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

type FailedTransaction struct {
	ID                int64
	UserID            int64
	SubscriptionID    uuid.UUID
	AttemptDate       time.Time
	AuthorizationCode sql.NullString
	Reference         string
	Amount            string
	FailureReason     sql.NullString
	ErrorCode         sql.NullString
	CardLast4         sql.NullString
	CardExpMonth      sql.NullString
	CardExpYear       sql.NullString
	CardType          sql.NullString
	CreatedAt         time.Time
	UpdatedAt         time.Time
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

type PaymentPlan struct {
	ID          int32
	Name        string
	Image       string
	Description sql.NullString
	Duration    string
	Price       string
	Features    []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Status      string
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

type Subscription struct {
	ID                uuid.UUID
	UserID            int64
	PlanID            int32
	StartDate         time.Time
	EndDate           time.Time
	Price             string
	Status            string
	TransactionID     int64
	PaymentMethod     sql.NullString
	AuthorizationCode sql.NullString
	CardLast4         sql.NullString
	CardExpMonth      sql.NullString
	CardExpYear       sql.NullString
	CardType          sql.NullString
	Currency          sql.NullString
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type User struct {
	ID           int64
	CreatedAt    time.Time
	Name         string
	Email        string
	PasswordHash []byte
	Activated    bool
	Version      int32
	UserImg      string
}
