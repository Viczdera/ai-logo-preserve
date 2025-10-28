-- name: CreateJob :one
INSERT INTO jobs (
    id,
    status,
    s3_key,
    upload_url
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetJob :one
SELECT * FROM jobs WHERE id = $1;

-- name: UpdateJobStatus :one
UPDATE jobs 
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateJobCompleted :one
UPDATE jobs 
SET status = $2, 
    logos_found = $3,
    result_url = $4,
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateJobError :one
UPDATE jobs 
SET status = $2, 
    error_message = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListJobs :many
SELECT * FROM jobs 
ORDER BY created_at DESC 
LIMIT $1 OFFSET $2;



-- name: DeleteJob :exec
DELETE FROM jobs WHERE id = $1;
