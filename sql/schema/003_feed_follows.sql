-- +goose Up
CREATE TABLE feed_follows (
	id UUID PRIMARY KEY,
	user_id UUID REFERENCES users(id) ON DELETE CASCADE,
	feed_id UUID REFERENCES feeds(id) ON DELETE CASCADE,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;
