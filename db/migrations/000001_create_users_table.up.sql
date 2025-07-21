CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE
    IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        name TEXT,
        is_verified BOOLEAN DEFAULT FALSE,
        image_url TEXT,
        role user_role DEFAULT 'user',
        created_at TIMESTAMP DEFAULT NOW (),
        updated_at TIMESTAMP
    );