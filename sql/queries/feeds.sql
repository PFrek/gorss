-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id, last_fetched_at)
VALUES ($1, $2, $3, $4, $5, $6, NULL)
RETURNING *;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeed :one
SELECT * FROM feeds
WHERE id = $1;

-- name: GetNextFeedsToFetch :many
SELECT * FROM feeds
ORDER BY last_fetched_at DESC
LIMIT $1;
