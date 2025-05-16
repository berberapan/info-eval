CREATE TABLE IF NOT EXISTS scenario_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token TEXT UNIQUE NOT NULL,
    scenario_id UUID REFERENCES scenarios(id) ON DELETE CASCADE,
    notes TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
