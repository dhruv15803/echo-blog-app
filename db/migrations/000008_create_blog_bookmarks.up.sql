CREATE TABLE
    IF NOT EXISTS blog_bookmarks (
        bookmarked_by_id INTEGER NOT NULL,
        bookmarked_blog_id INTEGER NOT NULL,
        bookmarked_at TIMESTAMP DEFAULT NOW (),
        FOREIGN KEY (bookmarked_by_id) REFERENCES users (id) ON DELETE CASCADE,
        FOREIGN KEY (bookmarked_blog_id) REFERENCES blogs (id) ON DELETE CASCADE,
        UNIQUE (bookmarked_by_id, bookmarked_blog_id)
    );