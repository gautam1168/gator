// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: feed_follows.sql

package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const createFeedFollow = `-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
	INSERT INTO feed_follows (id, user_id, feed_id, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, user_id, feed_id, created_at, updated_at
) 
SELECT inserted_feed_follow.id, inserted_feed_follow.user_id, inserted_feed_follow.feed_id, inserted_feed_follow.created_at, inserted_feed_follow.updated_at, feeds.name, users.name
FROM inserted_feed_follow 
INNER JOIN users ON users.id = inserted_feed_follow.user_id
INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id
`

type CreateFeedFollowParams struct {
	ID        uuid.UUID
	UserID    uuid.NullUUID
	FeedID    uuid.NullUUID
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
}

type CreateFeedFollowRow struct {
	ID        uuid.UUID
	UserID    uuid.NullUUID
	FeedID    uuid.NullUUID
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
	Name      sql.NullString
	Name_2    string
}

func (q *Queries) CreateFeedFollow(ctx context.Context, arg CreateFeedFollowParams) (CreateFeedFollowRow, error) {
	row := q.db.QueryRowContext(ctx, createFeedFollow,
		arg.ID,
		arg.UserID,
		arg.FeedID,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var i CreateFeedFollowRow
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.FeedID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Name_2,
	)
	return i, err
}

const getFeedFollowsForUser = `-- name: GetFeedFollowsForUser :many
WITH u AS (
	SELECT id, created_at, updated_at, name FROM users where users.name = $1
)
SELECT feed_follows.id, feed_follows.user_id, feed_follows.feed_id, feed_follows.created_at, feed_follows.updated_at, f.name as feed_name, u.name as user_name 
FROM feed_follows 
INNER JOIN u on u.id = feed_follows.user_id
INNER JOIN feeds as f on f.id = feed_follows.feed_id
WHERE feed_follows.user_id = u.id
`

type GetFeedFollowsForUserRow struct {
	ID        uuid.UUID
	UserID    uuid.NullUUID
	FeedID    uuid.NullUUID
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
	FeedName  sql.NullString
	UserName  string
}

func (q *Queries) GetFeedFollowsForUser(ctx context.Context, name string) ([]GetFeedFollowsForUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getFeedFollowsForUser, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeedFollowsForUserRow
	for rows.Next() {
		var i GetFeedFollowsForUserRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.FeedID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.FeedName,
			&i.UserName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const unfollow = `-- name: Unfollow :exec
DELETE FROM feed_follows
WHERE feed_follows.user_id = $1
AND feed_id = (SELECT id FROM feeds WHERE url = $2)
`

type UnfollowParams struct {
	UserID uuid.NullUUID
	Url    sql.NullString
}

func (q *Queries) Unfollow(ctx context.Context, arg UnfollowParams) error {
	_, err := q.db.ExecContext(ctx, unfollow, arg.UserID, arg.Url)
	return err
}
