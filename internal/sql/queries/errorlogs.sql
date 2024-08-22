-- name: CreateScraperErrorLog :one
INSERT INTO scraper_error_logs (
    error_type, message, feed_id, status_code, retry_attempts, admin_notified, resolved, resolution_notes, occurred_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
)
ON CONFLICT (error_type, feed_id)
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
    sel.id, sel.error_type, sel.message, sel.feed_id, f.name as feed_name, f.url as feed_url, f.img_url as feed_img_url, 
    sel.occurred_at, sel.status_code, sel.retry_attempts, sel.admin_notified, sel.resolved, 
    sel.resolution_notes, sel.created_at, sel.updated_at, sel.occurrence_count, sel.last_occurrence
FROM 
    scraper_error_logs sel
JOIN 
    feeds f ON sel.feed_id = f.id
WHERE 
    sel.id = $1;


-- name: GetAllScraperErrorLogs :many
SELECT 
    COUNT(*) OVER() AS total_count,
    sel.id, sel.error_type, sel.message, sel.feed_id, f.name as feed_name, f.url as feed_url, f.img_url as feed_img_url,
    sel.occurred_at, sel.status_code, sel.retry_attempts, sel.admin_notified, sel.resolved, 
    sel.resolution_notes, sel.created_at, sel.updated_at, sel.occurrence_count, sel.last_occurrence
FROM 
    scraper_error_logs sel
JOIN 
    feeds f ON sel.feed_id = f.id
ORDER BY 
    sel.occurred_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateScraperErrorLog :one
UPDATE scraper_error_logs
SET
    admin_notified = $1,
    resolved = $2,
    resolution_notes = $3,
    updated_at = NOW()
WHERE id = $4
RETURNING id, admin_notified, resolved, resolution_notes, updated_at;



-- name: DeleteScraperErrorLogByID :one
DELETE FROM scraper_error_logs
WHERE 
    id = $1
RETURNING id;