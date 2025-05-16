CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT UNIQUE NOT NULL,
    password_hash BYTEA NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(), 
    updated_at TIMESTAMPTZ DEFAULT now() 
);
