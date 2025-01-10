CREATE TABLE IF NOT EXISTS posts (
    id uuid PRIMARY KEY,
    project_id uuid NOT NULL,
    title TEXT NOT NULL,
    text_content TEXT NOT NULL,
    is_idea BOOLEAN NOT NULL,
    status VARCHAR(20) NOT NULL,
    scheduled_at TIMESTAMP,
    created_by uuid NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects (id)
);

CREATE TABLE IF NOT EXISTS platforms (
    id VARCHAR(10) PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

INSERT INTO platforms (id, name)
VALUES ('linkedin', 'LinkedIn');


CREATE TABLE IF NOT EXISTS post_platforms (
    post_id uuid NOT NULL,
    platform_id VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL,
    PRIMARY KEY (post_id, platform_id),
    FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE,
    FOREIGN KEY (platform_id) REFERENCES platforms (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS project_platforms (
    project_id uuid NOT NULL,
    platform_id VARCHAR(10) NOT NULL,
    secrets TEXT DEFAULT '',
    PRIMARY KEY (project_id, platform_id),
    FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE,
    FOREIGN KEY (platform_id) REFERENCES platforms (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL,
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE
);