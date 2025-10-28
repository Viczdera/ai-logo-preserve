-- name: CreateLogo :one
INSERT INTO logos (
    job_id,
    bounding_box,
    confidence,
    logo_type,
    s3_key
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetLogosByJobID :many
SELECT * FROM logos WHERE job_id = $1 ORDER BY confidence DESC;