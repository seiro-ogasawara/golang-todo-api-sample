CREATE TABLE users (
	user_id TEXT PRIMARY KEY,
	password TEXT NOT NULL
);

CREATE TABLE todos (
	id SERIAL PRIMARY KEY,
	user_id TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	description TEXT,
	status INT NOT NULL,
	priority INT NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
