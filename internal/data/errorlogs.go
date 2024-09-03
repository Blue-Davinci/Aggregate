package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/google/uuid"
)

var (
	FeedTypeErrorType            = "FeedType"
	FeedContextExceededErrorType = "FeedContextExceeded"
)

var (
	ErrLogNotFound = errors.New("error log not found")
)

type ErrorLogsDataModel struct {
	DB *database.Queries
}

type ScraperErrorLog struct {
	ID              int32       `json:"id"`
	ErrorType       string      `json:"error_type"`
	Message         string      `json:"message,omitempty"`
	FeedID          uuid.UUID   `json:"feed_id"`
	FeedDetails     FeedDetails `json:"feed_details"`
	OccurredAt      time.Time   `json:"occurred_at"`
	StatusCode      int32       `json:"status_code,omitempty"`
	RetryAttempts   int32       `json:"retry_attempts"`
	AdminNotified   bool        `json:"admin_notified"`
	Resolved        bool        `json:"resolved"`
	ResolutionNotes string      `json:"resolution_notes,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	OccurenceCount  int32       `json:"occurence_count,omitempty"`
	LastOccurrence  time.Time   `json:"last_occurrence,omitempty"`
}

type FeedDetails struct {
	FeedName string `json:"feed_name"`
	FeedUrl  string `json:"feed_url"`
	FeedImg  string `json:"feed_img"`
}

func ValidateScraperErrorLog(v *validator.Validator, errorDetails *ScraperErrorLog) error {
	v.Check(errorDetails.ResolutionNotes != "", "resolution_notes", "must be provided")
	return nil
}

func (m ErrorLogsDataModel) InsertScraperErrorLog(errorDetails *ScraperErrorLog) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queryResult, err := m.DB.CreateScraperErrorLog(ctx, database.CreateScraperErrorLogParams{
		ErrorType:       errorDetails.ErrorType,
		Message:         sql.NullString{String: errorDetails.Message, Valid: true},
		FeedID:          errorDetails.FeedID,
		StatusCode:      sql.NullInt32{Int32: errorDetails.StatusCode, Valid: true},
		RetryAttempts:   sql.NullInt32{Int32: errorDetails.RetryAttempts, Valid: true},
		AdminNotified:   sql.NullBool{Bool: errorDetails.AdminNotified, Valid: true},
		Resolved:        sql.NullBool{Bool: errorDetails.Resolved, Valid: true},
		ResolutionNotes: sql.NullString{String: errorDetails.ResolutionNotes, Valid: true},
		OccurredAt:      sql.NullTime{Time: errorDetails.OccurredAt, Valid: true},
	})
	if err != nil {
		fmt.Println("Error inserting error log: ", err)
		return err
	}
	// fill in with returned data i.e id, created_at and updated_at
	errorDetails.ID = queryResult.ID
	errorDetails.CreatedAt = queryResult.CreatedAt.Time
	errorDetails.UpdatedAt = queryResult.UpdatedAt.Time
	errorDetails.OccurenceCount = queryResult.OccurrenceCount.Int32
	errorDetails.LastOccurrence = queryResult.LastOccurrence.Time
	// return nil if no error
	return err
}

func (m ErrorLogsDataModel) GetScraperErrorLogByID(errorID int32) (*ScraperErrorLog, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errorLog, err := m.DB.GetScraperErrorLogByID(ctx, errorID)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrLogNotFound
		default:
			return nil, err
		}
	}
	// return the error log
	return &ScraperErrorLog{
		ID:        errorLog.ID,
		ErrorType: errorLog.ErrorType,
		Message:   errorLog.Message.String,
		FeedID:    errorLog.FeedID,
		FeedDetails: FeedDetails{FeedName: errorLog.FeedName,
			FeedUrl: errorLog.FeedUrl,
			FeedImg: errorLog.FeedImgUrl},
		OccurredAt:      errorLog.OccurredAt.Time,
		StatusCode:      errorLog.StatusCode.Int32,
		RetryAttempts:   errorLog.RetryAttempts.Int32,
		AdminNotified:   errorLog.AdminNotified.Bool,
		Resolved:        errorLog.Resolved.Bool,
		ResolutionNotes: errorLog.ResolutionNotes.String,
		CreatedAt:       errorLog.CreatedAt.Time,
		UpdatedAt:       errorLog.UpdatedAt.Time,
		OccurenceCount:  errorLog.OccurrenceCount.Int32,
		LastOccurrence:  errorLog.LastOccurrence.Time,
	}, nil
}

// GetAllScraperErrorLogs() retrieves all the error logs from the db
// based on the filters provided
func (m ErrorLogsDataModel) GetAllScraperErrorLogs(filters Filters) (*[]ScraperErrorLog, Metadata, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errorLogs, err := m.DB.GetAllScraperErrorLogs(ctx, database.GetAllScraperErrorLogsParams{
		Limit:  int32(filters.limit()),
		Offset: int32(filters.offset()),
	})
	if err != nil {
		return nil, Metadata{}, err
	}
	// total count
	totalRecords := 0
	// create a slice of ScraperErrorLog
	var errorLogsSlice []ScraperErrorLog
	// loop through the errorLogs and append to the slice
	for _, errorLog := range errorLogs {
		totalRecords = int(errorLog.TotalCount)
		errorLogsSlice = append(errorLogsSlice, ScraperErrorLog{
			ID:        errorLog.ID,
			ErrorType: errorLog.ErrorType,
			Message:   errorLog.Message.String,
			FeedID:    errorLog.FeedID,
			FeedDetails: FeedDetails{FeedName: errorLog.FeedName,
				FeedUrl: errorLog.FeedUrl,
				FeedImg: errorLog.FeedImgUrl},
			OccurredAt:      errorLog.OccurredAt.Time,
			StatusCode:      errorLog.StatusCode.Int32,
			RetryAttempts:   errorLog.RetryAttempts.Int32,
			AdminNotified:   errorLog.AdminNotified.Bool,
			Resolved:        errorLog.Resolved.Bool,
			ResolutionNotes: errorLog.ResolutionNotes.String,
			CreatedAt:       errorLog.CreatedAt.Time,
			UpdatedAt:       errorLog.UpdatedAt.Time,
			OccurenceCount:  errorLog.OccurrenceCount.Int32,
			LastOccurrence:  errorLog.LastOccurrence.Time,
		})
	}
	// create metadata
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// return the slice
	return &errorLogsSlice, metadata, nil
}

func (m ErrorLogsDataModel) UpdateScraperErrorLogByID(errorDetails *ScraperErrorLog) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queryResult, err := m.DB.UpdateScraperErrorLog(ctx, database.UpdateScraperErrorLogParams{
		AdminNotified:   sql.NullBool{Bool: errorDetails.AdminNotified, Valid: true},
		Resolved:        sql.NullBool{Bool: errorDetails.Resolved, Valid: true},
		ResolutionNotes: sql.NullString{String: errorDetails.ResolutionNotes, Valid: errorDetails.ResolutionNotes != ""},
		ID:              errorDetails.ID,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	// fill in admin_notified, resolved, resolution_notes, updated_at;
	errorDetails.ID = queryResult.ID
	errorDetails.UpdatedAt = queryResult.UpdatedAt.Time
	errorDetails.ResolutionNotes = queryResult.ResolutionNotes.String
	errorDetails.AdminNotified = queryResult.AdminNotified.Bool
	errorDetails.Resolved = queryResult.Resolved.Bool
	// return nil if no error
	return err
}

// DeleteScraperErrorLogByID() simply deletes an error log from our db
// by the id provided
func (m ErrorLogsDataModel) DeleteScraperErrorLogByID(id int32) (int32, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deleted_id, err := m.DB.DeleteScraperErrorLogByID(ctx, id)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return 0, ErrLogNotFound
		default:
			return 0, err
		}
	}
	return deleted_id, nil
}
