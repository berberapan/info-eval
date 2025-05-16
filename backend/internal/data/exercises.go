package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Exercise struct {
	ID        uuid.UUID          `json:"id"`
	Info      string             `json:"info"`
	Order     int16              `json:"order"`
	Media     []ExerciseMedia    `json:"media"`
	Questions []ExerciseQuestion `json:"questions"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type ExerciseModel struct {
	DB *sql.DB
}

func (em *ExerciseModel) GetByScenarioID(scenarioID uuid.UUID) ([]Exercise, error) {
	query := `
	SELECT id, info, "order", created_at, updated_at
	FROM exercises
	WHERE scenario_id = $1
	ORDER BY "order"
	`
	var exercisSlice []Exercise
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := em.DB.QueryContext(ctx, query, scenarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var e Exercise
		if err := rows.Scan(&e.ID, &e.Info, &e.Order, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		exercisSlice = append(exercisSlice, e)
	}
	return exercisSlice, nil
}
