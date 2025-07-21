CREATE TABLE
    IF NOT EXISTS blog_topics (
        blog_id INTEGER NOT NULL,
        topic_id INTEGER NOT NULL,
        FOREIGN KEY (blog_id) REFERENCES blogs (id) ON DELETE CASCADE,
        FOREIGN KEY (topic_id) REFERENCES topics (id) ON DELETE CASCADE,
        UNIQUE (blog_id, topic_id)
    );