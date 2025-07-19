-- name: AddFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListFeeds :many
SELECT f.name as name, f.url as url, u.name as user_name FROM
feeds as f JOIN users as u
ON u.id = f.user_id; 
