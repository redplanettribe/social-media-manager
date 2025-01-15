CREATE TABLE IF NOT EXISTS media (
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    media_type VARCHAR(20) NOT NULL,
    format VARCHAR(10) NOT NULL,
    width INT,
    height INT,
    length INT,
    size INT,
    alt_text VARCHAR(255),
    added_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE,
    FOREIGN KEY (added_by) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS post_platform_media (
    media_id UUID NOT NULL,
    post_id UUID NOT NULL,
    platform_id VARCHAR(10) NOT NULL,
    PRIMARY KEY (
        media_id,
        post_id,
        platform_id
    ),
    FOREIGN KEY (media_id) REFERENCES media (id) ON DELETE CASCADE,
    FOREIGN KEY (post_id, platform_id) REFERENCES post_platforms (post_id, platform_id) ON DELETE CASCADE
);