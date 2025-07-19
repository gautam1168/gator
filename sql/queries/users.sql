-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUserByName :one
SELECT * from users where name = $1;

-- name: ResetUsers :exec
DELETE from users;

-- name: GetUsers :many
SELECT * FROM users;
