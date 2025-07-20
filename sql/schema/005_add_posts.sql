-- +goose Up
CREATE TABLE posts (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NOW(),
	title TEXT,
	url TEXT,
	description TEXT,
	published_at TIMESTAMP,
	feed_id UUID REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;
