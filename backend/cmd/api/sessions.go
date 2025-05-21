package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/berberapan/info-eval/internal/data"
	"github.com/google/uuid"
)

func (app *application) createScenarioSessionHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ScenarioID            string `json:"scenario_id"`
		Notes                 string `json:"notes"`
		ValidityDurationHours int    `json:"validity_duration_hours"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	scenarioID, err := uuid.Parse(input.ScenarioID)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid scenario_id format"))
		return
	}
	tokenBytes := make([]byte, 16)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	tokenString := hex.EncodeToString(tokenBytes)

	validityDuration := 24 * time.Hour
	if input.ValidityDurationHours > 0 {
		validityDuration = time.Duration(input.ValidityDurationHours) * time.Hour
	}
	expiresAt := time.Now().Add(validityDuration)
	session := &data.ScenarioSession{
		ScenarioID: scenarioID,
		Token:      tokenString,
		Notes:      input.Notes,
		ExpiresAt:  expiresAt,
	}
	err = app.models.ScenarioSessions.Create(session)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/sessions/%s", session.ID))
	err = app.writeJSON(w, http.StatusCreated, jsonEnvelope{"scenario_session": session}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getScenarioSessionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	session, err := app.models.ScenarioSessions.Get(sessionID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if time.Now().After(session.ExpiresAt) {
		app.errorResponse(w, r, http.StatusGone, "this session has expired")
		return
	}
	err = app.writeJSON(w, http.StatusOK, jsonEnvelope{"scenario_session": session}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
