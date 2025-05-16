CREATE TABLE IF NOT EXISTS exercise_question_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exercise_question_id UUID NOT NULL REFERENCES exercise_questions(id) ON DELETE CASCADE,
    option_text TEXT NOT NULL,
    is_correct BOOLEAN DEFAULT FALSE,
    feedback TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
