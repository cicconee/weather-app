CREATE TABLE admins(
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    approved bool NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);