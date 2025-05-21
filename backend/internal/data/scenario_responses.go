package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SessionResponse struct {
	ID                uuid.UUID `json:"id"`
	ScenarioSessionID uuid.UUID `json:"scenario_session_id"`
	SubmittedAt       time.Time `json:"submitted_at"`
	RawAnswers        []byte    `json:"raw_answers"`
	AIFeedback        []byte    `json:"ai_feedback"`
}

type SessionResponseOutput struct {
	ID                uuid.UUID         `json:"id"`
	ScenarioSessionID uuid.UUID         `json:"scenario_session_id"`
	SubmittedAt       time.Time         `json:"submitted_at"`
	RawAnswers        map[string]any    `json:"raw_answers,omitempty"`
	AIFeedback        map[string]string `json:"ai_feedback,omitempty"`
}

type SessionResponseModel struct {
	DB *sql.DB
}

func (sm *SessionResponseModel) Create(sr SessionResponse) (SessionResponse, error) {
	query := `
	INSERT INTO session_responses (scenario_session_id, raw_answers)
	VALUES ($1, $2)
	RETURNING id, submitted_at`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := sm.DB.QueryRowContext(ctx, query, sr.ScenarioSessionID, sr.RawAnswers).
		Scan(&sr.ID, &sr.SubmittedAt)
	return sr, err
}

func (sm *SessionResponseModel) AddFeedback(responseID uuid.UUID, feedback map[string]string) error {
	feedbackJSON, err := json.Marshal(feedback)
	if err != nil {
		return fmt.Errorf("failed to marshal feedback to JSON for response %s: %w", responseID, err)
	}
	query := `
	UPDATE session_responses
	SET ai_feedback = $1
	WHERE id = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := sm.DB.ExecContext(ctx, query, feedbackJSON, responseID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (sm *SessionResponseModel) Get(id uuid.UUID) (*SessionResponse, error) {
	query := `
	SELECT id, scenario_session_id, submitted_at, raw_answers, ai_feedback
	FROM session_responses
	WHERE id = $1`
	var sr SessionResponse
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := sm.DB.QueryRowContext(ctx, query, id).Scan(
		&sr.ID,
		&sr.ScenarioSessionID,
		&sr.SubmittedAt,
		&sr.RawAnswers,
		&sr.AIFeedback,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &sr, nil
}

func (sm *SessionResponseModel) GetAllByScenarioSessionID(scenarioSessionID uuid.UUID) ([]*SessionResponse, error) {
	query := `
	SELECT id, scenario_session_id, submitted_at, raw_answers, ai_feedback
	FROM session_responses
	WHERE scenario_session_id = $1
	ORDER BY submitted_at DESC`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := sm.DB.QueryContext(ctx, query, scenarioSessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var responses []*SessionResponse
	for rows.Next() {
		var sr SessionResponse
		err := rows.Scan(
			&sr.ID,
			&sr.ScenarioSessionID,
			&sr.SubmittedAt,
			&sr.RawAnswers,
			&sr.AIFeedback,
		)
		if err != nil {
			return nil, err
		}
		responses = append(responses, &sr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return responses, nil
}
