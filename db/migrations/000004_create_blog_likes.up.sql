CREATE TABLE
    IF NOT EXISTS blog_likes (
        liked_by_id INTEGER NOT NULL,
        liked_blog_id INTEGER NOT NULL,
        liked_at TIMESTAMP DEFAULT NOW (),
        FOREIGN KEY (liked_by_id) REFERENCES users (id) ON DELETE CASCADE,
        FOREIGN KEY (liked_blog_id) REFERENCES blogs (id) ON DELETE CASCADE,
        UNIQUE (liked_by_id, liked_blog_id)
    );