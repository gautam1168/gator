-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
	INSERT INTO feed_follows (id, user_id, feed_id, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING *
) 
SELECT inserted_feed_follow.*, feeds.name, users.name
FROM inserted_feed_follow 
INNER JOIN users ON users.id = inserted_feed_follow.user_id
INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id;

-- name: GetFeedFollowsForUser :many
WITH u AS (
	SELECT * FROM users where users.name = $1
)
SELECT feed_follows.*, f.name as feed_name, u.name as user_name 
FROM feed_follows 
INNER JOIN u on u.id = feed_follows.user_id
INNER JOIN feeds as f on f.id = feed_follows.feed_id
WHERE feed_follows.user_id = u.id;
