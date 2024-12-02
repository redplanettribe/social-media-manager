-- add salt property to users table
ALTER TABLE users ADD COLUMN salt CHAR(32) NOT NULL;