CREATE TABLE
    IF NOT EXISTS blogs (
        id SERIAL PRIMARY KEY,
        blog_title TEXT NOT NULL,
        blog_description TEXT,
        blog_content TEXT NOT NULL,
        blog_thumbnail TEXT,
        blog_author_id INTEGER NOT NULL,
        blog_created_at TIMESTAMP DEFAULT NOW (),
        blog_updated_at TIMESTAMP,
        FOREIGN KEY (blog_author_id) REFERENCES users (id) ON DELETE CASCADE
    );