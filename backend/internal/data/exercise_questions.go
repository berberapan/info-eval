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
	ID           uuid.UUID        `json:"id"`
	ExerciseType ExerciseType     `json:"type"`
	Question     string           `json:"question"`
	Options      []QuestionOption `json:"options"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type ExerciseQuestionModel struct {
	DB *sql.DB
}

func (em *ExerciseQuestionModel) GetByExerciseID(exerciseID uuid.UUID) ([]ExerciseQuestion, error) {
	query := `
	SELECT id, type, question, created_at, updated_at
	FROM exercise_questions
	WHERE exercise_id = $1
	`
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
		if err := rows.Scan(&q.ID, &q.ExerciseType, &q.Question, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, err
		}
		questionSlice = append(questionSlice, q)
	}
	return questionSlice, nil
}
