package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type QuestionOption struct {
	ID         uuid.UUID `json:"id"`
	OptionText string    `json:"option_text"`
	IsCorrect  bool      `json:"is_correct"`
	Feedback   string    `json:"feedback"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type QuestionOptionModel struct {
	DB *sql.DB
}

func (qm *QuestionOptionModel) GetByQuestionID(questionID uuid.UUID) ([]QuestionOption, error) {
	query := `
	SELECT id, option_text, is_correct, feedback, created_at, updated_at
	FROM exercise_question_options
	WHERE exercise_question_id = $1`
	var questionOptionSlice []QuestionOption
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := qm.DB.QueryContext(ctx, query, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var option QuestionOption
		if err := rows.Scan(&option.ID, &option.OptionText, &option.IsCorrect, &option.Feedback, &option.CreatedAt, &option.UpdatedAt); err != nil {
			return nil, err
		}
		questionOptionSlice = append(questionOptionSlice, option)
	}
	return questionOptionSlice, nil
}
