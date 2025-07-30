CREATE TABLE
    IF NOT EXISTS password_resets (
        token TEXT PRIMARY KEY,
        user_id INTEGER NOT NULL,
        expiration_at TIMESTAMP DEFAULT NOW (),
        FOREIGN KEY (user_id) REFERENCES users (id)
    );