CREATE TABLE
    IF NOT EXISTS topics (
        id SERIAL PRIMARY KEY,
        topic_title TEXT NOT NULL,
        topic_created_at TIMESTAMP DEFAULT NOW (),
        topic_updated_at TIMESTAMP
    );