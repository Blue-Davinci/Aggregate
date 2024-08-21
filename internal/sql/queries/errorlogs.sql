-- name: CreateScraperErrorLog :one
INSERT INTO scraper_error_logs (
    error_type, message, feed_url, status_code, retry_attempts, admin_notified, resolved, resolution_notes, occurred_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
)
ON CONFLICT (error_type, feed_url)
DO UPDATE SET
    occurrence_count = scraper_error_logs.occurrence_count + 1,
    last_occurrence = NOW(),
    retry_attempts = scraper_error_logs.retry_attempts + 1,
    updated_at = NOW(),
    message = EXCLUDED.message, -- Optionally update message
    status_code = EXCLUDED.status_code -- Optionally update status code
RETURNING id, created_at, updated_at, occurrence_count, last_occurrence;



-- name: GetScraperErrorLogByID :one
SELECT 
    id, error_type, message, feed_url, occurred_at, status_code, retry_attempts, admin_notified, resolved, resolution_notes, created_at, updated_at, occurrence_count, last_occurrence
FROM 
    scraper_error_logs
WHERE 
    id = $1;

-- name: GetAllScraperErrorLogs :many
SELECT 
    COUNT(*) OVER() AS total_count,
    id, error_type, message, feed_url, occurred_at, status_code, retry_attempts, admin_notified, resolved, resolution_notes, created_at, updated_at, occurrence_count, last_occurrence
FROM 
    scraper_error_logs
ORDER BY 
    occurred_at DESC
LIMIT $1 OFFSET $2;



-- name: DeleteScraperErrorLogByID :exec
DELETE FROM scraper_error_logs
WHERE 
    id = $1;
