-- Create session table for storing session data

CREATE TABLE "sessions" (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    fingerprint VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX session_user_id_index ON sessions (user_id);

-- user_id is a foreign key to the user table
ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");