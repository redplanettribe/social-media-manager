-- Create roles table if it does not exist
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    role VARCHAR(255) NOT NULL
); 

-- Create user_roles table if it does not exist
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL,
    role_id SERIAL NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- Insert roles into roles table
INSERT INTO roles (role) VALUES ('manager'), ('member'), ('developer');