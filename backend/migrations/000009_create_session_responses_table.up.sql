CREATE TABLE IF NOT EXISTS session_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_session_id UUID REFERENCES scenario_sessions(id) ON DELETE CASCADE,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    raw_answers JSONB,
    ai_feedback TEXT
);
