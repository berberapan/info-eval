CREATE TABLE IF NOT EXISTS scenarios (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description text,
    difficulty SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
