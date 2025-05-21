package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ExerciseType string

const (
	FreeTextType       ExerciseType = "free_text"
	TrueFalseType      ExerciseType = "true_false"
	MultipleChoiceType ExerciseType = "multiple_choice"
)

type ExerciseQuestion struct {
	ID             uuid.UUID        `json:"id"`
	ExerciseID     uuid.UUID        `json:"exercise_id"`
	ExerciseType   ExerciseType     `json:"type"`
	Question       string           `json:"question"`
	Options        []QuestionOption `json:"options"`
	PromptGuidance sql.NullString   `json:"prompt_guidance"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type ExerciseQuestionModel struct {
	DB *sql.DB
}

func (em *ExerciseQuestionModel) GetByExerciseID(exerciseID uuid.UUID) ([]ExerciseQuestion, error) {
	query := `
	SELECT id, type, question, prompt_guidance, created_at, updated_at
	FROM exercise_questions
	WHERE exercise_id = $1`
	var questionSlice []ExerciseQuestion
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := em.DB.QueryContext(ctx, query, exerciseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var q ExerciseQuestion
		q.ExerciseID = exerciseID
		if err := rows.Scan(&q.ID, &q.ExerciseType, &q.Question, &q.PromptGuidance, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, err
		}
		questionSlice = append(questionSlice, q)
	}
	return questionSlice, nil
}

func (em *ExerciseQuestionModel) GetAllByScenarioIDAsMap(scenarioID uuid.UUID) (map[uuid.UUID]ExerciseQuestion, error) {
	query := `
	SELECT eq.id, eq.exercise_id, eq.type, eq.question, eq.prompt_guidance, eq.created_at, eq.updated_at
	FROM exercise_questions eq
	INNER JOIN exercises e ON eq.exercise_id = e.id
	WHERE e.scenario_id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := em.DB.QueryContext(ctx, query, scenarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	questionMap := make(map[uuid.UUID]ExerciseQuestion)
	for rows.Next() {
		var q ExerciseQuestion
		if err := rows.Scan(&q.ID, &q.ExerciseID, &q.ExerciseType, &q.Question, &q.PromptGuidance, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, err
		}
		questionMap[q.ID] = q
	}
	return questionMap, nil
}
