package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/berberapan/info-eval/internal/data"
	"github.com/berberapan/info-eval/internal/validator"
)

func (app *application) showScenarioHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	scenario, err := app.models.Scenarios.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, jsonEnvelope{"scenario": scenario}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showScenariosHandler(w http.ResponseWriter, r *http.Request) {
	scenarios, err := app.models.Scenarios.GetAll()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, jsonEnvelope{"scenarios": scenarios}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createScenarioHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Difficulty  int16  `json:"difficulty"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	scenario := &data.Scenario{
		Title:       input.Title,
		Description: input.Description,
		Difficulty:  input.Difficulty,
	}
	v := validator.New()
	if data.ValidateScenario(v, scenario); !v.Valid() {
		app.failedValidateResponse(w, r, v.Errors)
		return
	}
	err = app.models.Scenarios.Insert(scenario)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/scenario/%s", scenario.ID))
	err = app.writeJSON(w, http.StatusCreated, jsonEnvelope{"scenario": scenario}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getScenarioIDHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	scenarioID, err := app.models.ScenarioSessions.GetScenarioByID(sessionID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, jsonEnvelope{"scenario_id": scenarioID}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
