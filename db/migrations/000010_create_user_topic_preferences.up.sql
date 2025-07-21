CREATE TABLE
    IF NOT EXISTS user_topic_preferences (
        user_id INTEGER NOT NULL,
        topic_id INTEGER NOT NULL,
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
        FOREIGN KEY (topic_id) REFERENCES topics (id) ON DELETE CASCADE,
        UNIQUE (user_id, topic_id)
    );