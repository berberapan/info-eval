CREATE TABLE IF NOT EXISTS exercise_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exercise_id UUID REFERENCES exercises(id) ON DELETE CASCADE,
    "type" exercise_type NOT NULL,
    question TEXT NOT NULL,
    prompt_guidance TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
