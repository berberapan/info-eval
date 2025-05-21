package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ScenarioSession struct {
	ID         uuid.UUID `json:"id"`
	Token      string    `json:"token"`
	ScenarioID uuid.UUID `json:"scenario_id"`
	Notes      string    `json:"notes"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type ScenarioSessionModel struct {
	DB *sql.DB
}

func (sm *ScenarioSessionModel) Create(ss *ScenarioSession) error {
	query := `
	INSERT INTO scenario_sessions (token, scenario_id, notes, expires_at)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at`
	args := []any{ss.Token, ss.ScenarioID, ss.Notes, ss.ExpiresAt}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := sm.DB.QueryRowContext(ctx, query, args...).Scan(&ss.ID, &ss.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (m *ScenarioSessionModel) Get(id uuid.UUID) (*ScenarioSession, error) {
	query := `
	SELECT id, scenario_id, token, notes, expires_at, created_at
	FROM scenario_sessions
	WHERE id = $1`
	var s ScenarioSession
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.ScenarioID, &s.Token, &s.Notes, &s.ExpiresAt, &s.CreatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &s, nil
}

func (sm *ScenarioSessionModel) GetByToken(token string) (ScenarioSession, error) {
	query := `
	SELECT id, token, scenario_id, notes, expires_at, created_at
	FROM scenario_sessions
	WHERE token = $1`
	var s ScenarioSession
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := sm.DB.QueryRowContext(ctx, query, token).Scan(
		&s.ID, &s.Token, &s.ScenarioID, &s.Notes, &s.ExpiresAt, &s.CreatedAt)
	return s, err
}

func (sm *ScenarioSessionModel) GetScenarioByID(id uuid.UUID) (uuid.UUID, error) {
	query := `
	SELECT scenario_id
	FROM scenario_sessions
	WHERE id = $1`
	var s uuid.UUID
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := sm.DB.QueryRowContext(ctx, query, id).Scan(&s)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return uuid.Nil, ErrRecordNotFound
		default:
			return uuid.Nil, err
		}
	}
	return s, nil
}
