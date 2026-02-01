-- name: CreateFile :one
INSERT INTO files (user_id, filename, original_name, mime_type, size, storage_driver, storage_path, url)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetFileByID :one
SELECT * FROM files
WHERE id = $1;

-- name: ListFilesByUserID :many
SELECT * FROM files
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListFiles :many
SELECT * FROM files
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: DeleteFile :exec
DELETE FROM files
WHERE id = $1;

-- name: DeleteFilesByUserID :exec
DELETE FROM files
WHERE user_id = $1;

-- name: CountFilesByUserID :one
SELECT COUNT(*) FROM files
WHERE user_id = $1;

-- name: CountFiles :one
SELECT COUNT(*) FROM files;
