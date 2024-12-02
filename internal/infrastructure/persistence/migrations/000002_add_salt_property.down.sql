-- revert add salt property to users table
ALTER TABLE users DROP COLUMN salt;