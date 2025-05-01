CREATE TYPE IF NOT EXISTS media_type AS ENUM ('image', 'video', 'audio');

CREATE TABLE IF NOT EXISTS exercise_media (
    id UUID PRIMARY KEY,
    exercise_id UUID REFERENCES exercises(id) ON DELETE CASCADE,
    media_url TEXT NOT NULL,
    media_type media_type,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
