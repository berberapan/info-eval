CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    role_id bigint REFERENCES roles(id) ON DELETE SET NULL,
    activated bool NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    last_updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    deleted_at timestamp(0) with time zone
);
