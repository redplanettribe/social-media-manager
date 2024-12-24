CREATE TABLE IF NOT EXISTS posts (
    id uuid PRIMARY KEY,
    project_id uuid NOT NULL,
    title TEXT NOT NULL,
    text_content TEXT NOT NULL,
	image_links TEXT[],
	video_links TEXT[],
    is_idea BOOLEAN NOT NULL,
    status VARCHAR(20) NOT NULL,
    scheduled_publish_date TIMESTAMP,
    created_by uuid NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS social_networks (
    id uuid PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    api_key VARCHAR(20),
    logo_link VARCHAR(50)
);

CREATE TABLE IF NOT EXISTS post_social_networks (
    post_id uuid NOT NULL,
    social_network_id uuid NOT NULL,
    status VARCHAR(20) NOT NULL,
    PRIMARY KEY (post_id, social_network_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (social_network_id) REFERENCES social_networks(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL,
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);