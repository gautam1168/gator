-- +goose Up
CREATE TABLE feeds (
	id UUID PRIMARY KEY, 
	name TEXT, 
	created_at TIMESTAMP DEFAULT NOW(), 
	updated_at TIMESTAMP DEFAULT NOW(), 
	url TEXT UNIQUE, 
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;
