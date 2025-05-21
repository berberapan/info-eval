package data

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type ExerciseStore interface {
	GetByScenarioID(scenarioID uuid.UUID) ([]Exercise, error)
}

type ExerciseMediaStore interface {
	GetByExerciseID(exerciseID uuid.UUID) ([]ExerciseMedia, error)
}

type ExerciseQuestionStore interface {
	GetByExerciseID(exerciseID uuid.UUID) ([]ExerciseQuestion, error)
}

type QuestionOptionStore interface {
	GetByQuestionID(questionID uuid.UUID) ([]QuestionOption, error)
}

type Models struct {
	Exercises         ExerciseModel
	Scenarios         ScenarioModel
	ExerciseMedia     ExerciseMediaModel
	ExerciseQuestions ExerciseQuestionModel
	QuestionOptions   QuestionOptionModel
	ScenarioSessions  ScenarioSessionModel
	SessionResponses  SessionResponseModel
	Users             UserModel
}

func NewModels(db *sql.DB) Models {
	models := Models{
		Exercises:         ExerciseModel{DB: db},
		ExerciseMedia:     ExerciseMediaModel{DB: db},
		ExerciseQuestions: ExerciseQuestionModel{DB: db},
		QuestionOptions:   QuestionOptionModel{DB: db},
		ScenarioSessions:  ScenarioSessionModel{DB: db},
		SessionResponses:  SessionResponseModel{DB: db},
		Users:             UserModel{DB: db},
	}
	models.Scenarios = ScenarioModel{
		DB:                db,
		Exercises:         &models.Exercises,
		ExerciseMedia:     &models.ExerciseMedia,
		ExerciseQuestions: &models.ExerciseQuestions,
		QuestionOptions:   &models.QuestionOptions,
	}
	return models
}
