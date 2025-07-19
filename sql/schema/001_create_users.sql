-- +goose Up
CREATE TABLE users (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NOW(),
	name TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE users;
