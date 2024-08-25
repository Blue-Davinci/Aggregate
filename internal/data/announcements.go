package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/blue-davinci/aggregate/internal/validator"
)

type AnnouncementModel struct {
	DB *database.Queries
}

var (
	ErrAnnouncementNotFound = errors.New("announcement not found")
)
var (
	UrgencyLow    = "low"
	UrgencyMedium = "medium"
	UrgencyHigh   = "high"
)

// title, message, created_at, expires_at, updated_at, created_by, is_active
type Announcement struct {
	ID        int32     `json:"id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy int64     `json:"created_by"`
	IsActive  bool      `json:"is_active"`
	Urgency   string    `json:"urgency"`
}
type AnnouncementRead struct {
	ID             int32     `json:"id"`
	AnnouncementID int32     `json:"announcement_id"`
	UserID         int64     `json:"user_id"`
	ReadAt         time.Time `json:"read_at"`
}

// ValidateAnnouncement
func ValidateAnnouncement(v *validator.Validator, announcement *Announcement) {
	v.Check(announcement.Title != "", "title", "must be provided")
	v.Check(announcement.Message != "", "message", "must be provided")
	v.Check(announcement.CreatedBy != 0, "created_by", "must be provided")
	v.Check(!announcement.ExpiresAt.IsZero(), "expires_at", "must be provided")
	v.Check(announcement.Urgency != "", "urgency", "must be provided")
	v.Check(announcement.Urgency == UrgencyLow ||
		announcement.Urgency == UrgencyMedium ||
		announcement.Urgency == UrgencyHigh, "urgency", "must be one of 'low', 'medium', or 'high'")
}

func ValidateReadAnnouncement(v *validator.Validator, announcementRead *AnnouncementRead) {
	v.Check(announcementRead.AnnouncementID != 0, "announcement_id", "must be provided")
}

// AdminCreateNewAnnouncement() method simply created a new announcement for the admin
// to display to the users. It takes in an announcement struct and returns an error if
// there was an issue with the database query.
func (m AnnouncementModel) AdminCreateNewAnnouncement(announcement *Announcement) error {
	// create our context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// insert our data
	queryResult, err := m.DB.AdminCreateNewAnnouncment(ctx, database.AdminCreateNewAnnouncmentParams{
		Title:     announcement.Title,
		Message:   announcement.Message,
		ExpiresAt: sql.NullTime{Time: announcement.ExpiresAt, Valid: true},
		CreatedBy: announcement.CreatedBy,
		IsActive:  sql.NullBool{Bool: announcement.IsActive, Valid: true},
		Urgency:   announcement.Urgency,
	})
	if err != nil {
		return err
	}
	// fill in the announcement with the data from the query
	announcement.ID = queryResult.ID
	announcement.CreatedAt = queryResult.CreatedAt.Time
	announcement.UpdatedAt = queryResult.UpdatedAt.Time
	return nil
}

// AdminGetAllAnnouncments() method gets all the announcements for the admin. It takes in
// filters for the pagination and returns a slice of announcements and an error if there was
// an issue with the database query.
func (m AnnouncementModel) AdminGetAllAnnouncments(filters Filters) ([]*Announcement, Metadata, error) {
	// create our context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// get our data
	queryResults, err := m.DB.AdminGetAllAnnouncments(ctx, database.AdminGetAllAnnouncmentsParams{
		Limit:  int32(filters.limit()),
		Offset: int32(filters.offset()),
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, Metadata{}, ErrAnnouncementNotFound
		default:
			return nil, Metadata{}, err
		}
	}
	// fill in the announcements with the data from the query
	var announcements []*Announcement
	totalRecords := 0
	for _, result := range queryResults {
		totalRecords = int(result.TotalRecords)
		announcements = append(announcements, &Announcement{
			ID:        result.ID,
			Title:     result.Title,
			Message:   result.Message,
			CreatedAt: result.CreatedAt.Time,
			ExpiresAt: result.ExpiresAt.Time,
			UpdatedAt: result.UpdatedAt.Time,
			CreatedBy: result.CreatedBy,
			IsActive:  result.IsActive.Bool,
			Urgency:   result.Urgency,
		})
	}
	// calculate metadata
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return announcements, metadata, nil
}

// GetAnnouncmentsForUser() method gets all the announcements for a user. It takes in a
// user ID and returns a slice of announcements and an error if there was an issue with the
// database query.
func (m AnnouncementModel) GetAnnouncmentsForUser(userID int64) ([]*Announcement, error) {
	// create our context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// get our data
	queryResults, err := m.DB.GetAnnouncmentsForUser(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrAnnouncementNotFound
		default:
			return nil, err
		}
	}
	// fill in the announcements with the data from the query
	var announcements []*Announcement
	for _, result := range queryResults {
		announcements = append(announcements, &Announcement{
			ID:        result.ID,
			Title:     result.Title,
			Message:   result.Message,
			CreatedAt: result.CreatedAt.Time,
			ExpiresAt: result.ExpiresAt.Time,
			UpdatedAt: result.UpdatedAt.Time,
			CreatedBy: result.CreatedBy,
			IsActive:  result.IsActive.Bool,
			Urgency:   result.Urgency,
		})
	}
	return announcements, nil
}

// MarkAnnouncmentAsReadByUser() method marks an announcement as read by a user. It takes in
// a user ID and an announcement ID and returns an announcement read struct and an error if
// there was an issue with the database query.
func (m AnnouncementModel) MarkAnnouncmentAsReadByUser(announcementRead *AnnouncementRead) error {
	// create our context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// insert our data
	queryResult, err := m.DB.MarkAnnouncmentAsReadByUser(ctx, database.MarkAnnouncmentAsReadByUserParams{
		UserID:         announcementRead.UserID,
		AnnouncementID: sql.NullInt32{Int32: int32(announcementRead.AnnouncementID), Valid: true},
	})
	if err != nil {
		return err
	}
	// fill in the announcement with the data from the query
	announcementRead.ID = queryResult.ID
	announcementRead.ReadAt = queryResult.ReadAt

	return nil
}
