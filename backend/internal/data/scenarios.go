package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/berberapan/info-eval/internal/validator"
	"github.com/google/uuid"
)

type Scenario struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Difficulty  int16      `json:"difficulty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Exercises   []Exercise `json:"exercises"`
}

type ScenarioModel struct {
	DB                *sql.DB
	Exercises         ExerciseStore
	ExerciseMedia     ExerciseMediaStore
	ExerciseQuestions ExerciseQuestionStore
	QuestionOptions   QuestionOptionStore
}

func ValidateScenario(v *validator.Validator, scenario *Scenario) {
	v.Check(scenario.Title != "", "title", "must be provided")
	v.Check(len(scenario.Title) <= 300, "title", "can't exceed 300 chars")

	v.Check(scenario.Difficulty <= 5, "difficulty", "must be 5 or smaller")
	v.Check(scenario.Difficulty >= 1, "difficulty", "must be 1 or bigger")
}

func (sm *ScenarioModel) Get(id uuid.UUID) (*Scenario, error) {
	query := `
	SELECT id, title, description, difficulty, created_at, updated_at
	FROM scenarios
	WHERE id = $1`
	var s Scenario
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := sm.DB.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.Title, &s.Description, &s.Difficulty, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	exercises, err := sm.Exercises.GetByScenarioID(id)
	if err != nil {
		return nil, err
	}
	for i := range exercises {
		exercise := &exercises[i]
		exercise.Media, err = sm.ExerciseMedia.GetByExerciseID(exercise.ID)
		if err != nil {
			return nil, err
		}
		exercise.Questions, err = sm.ExerciseQuestions.GetByExerciseID(exercise.ID)
		if err != nil {
			return nil, err
		}
		for j := range exercise.Questions {
			question := &exercise.Questions[j]
			question.Options, err = sm.QuestionOptions.GetByQuestionID(question.ID)
			if err != nil {
				return nil, err
			}
		}
	}
	s.Exercises = exercises
	return &s, nil
}

func (sm *ScenarioModel) GetAll() ([]*Scenario, error) {
	query := `
	SELECT id, title, description, difficulty, created_at, updated_at
	FROM scenarios`
	scenarios := []*Scenario{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := sm.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s Scenario
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.Difficulty, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		scenarios = append(scenarios, &s)
	}
	return scenarios, nil
}

func (sm *ScenarioModel) Insert(scenario *Scenario) error {
	query := `
	INSERT INTO scenarios (title, description, difficulty)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at`
	args := []any{scenario.Title, scenario.Description, scenario.Difficulty}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return sm.DB.QueryRowContext(ctx, query, args...).Scan(&scenario.ID, &scenario.CreatedAt, &scenario.UpdatedAt)
}
