CREATE TABLE IF NOT EXISTS permissions (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    description text
);
