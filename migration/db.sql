-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNIQUE NOT NULL,
    ism VARCHAR(50),
    viloyat VARCHAR(30),
    shahar VARCHAR(30),
    telefon VARCHAR(14) UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS channels (
    name VARCHAR(50)
);

CREATE TABLE IF NOT EXISTS configs (
    bot_token VARCHAR(100)
);

-- Create admins table
CREATE TABLE IF NOT EXISTS admins (
    id BIGINT UNIQUE NOT NULL
);

