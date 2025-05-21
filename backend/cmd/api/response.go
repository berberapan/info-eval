package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/berberapan/info-eval/internal/data"
)

type createSessionResponseInput struct {
	RawAnswers map[string]any `json:"raw_answers"`
}

func (app *application) createSessionResponseHandler(w http.ResponseWriter, r *http.Request) {
	scenarioSessionID, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var input createSessionResponseInput
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	rawAnswersBytes, err := json.Marshal(input.RawAnswers)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("failed to marshal raw_answers: %w", err))
		return
	}
	sessionResponse := data.SessionResponse{
		ScenarioSessionID: scenarioSessionID,
		RawAnswers:        rawAnswersBytes,
	}
	createdResponse, err := app.models.SessionResponses.Create(sessionResponse)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.triggerAIFeedbackGeneration(createdResponse.ID, createdResponse.ScenarioSessionID, rawAnswersBytes)
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/sessions/%s/responses/%s", scenarioSessionID, createdResponse.ID))
	err = app.writeJSON(w, http.StatusCreated, jsonEnvelope{"session_response": createdResponse}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getSessionResponseHandler(w http.ResponseWriter, r *http.Request) {
	responseID, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	sessionResponse, err := app.models.SessionResponses.Get(responseID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	output := data.SessionResponseOutput{
		ID:                sessionResponse.ID,
		ScenarioSessionID: sessionResponse.ScenarioSessionID,
		SubmittedAt:       sessionResponse.SubmittedAt,
	}
	if sessionResponse.RawAnswers != nil {
		if errUnmarshal := json.Unmarshal(sessionResponse.RawAnswers, &output.RawAnswers); errUnmarshal != nil {
			app.serverErrorResponse(w, r, err)
		}
	}
	if sessionResponse.AIFeedback != nil {
		if errUnmarshal := json.Unmarshal(sessionResponse.AIFeedback, &output.AIFeedback); errUnmarshal != nil {
			app.serverErrorResponse(w, r, err)
		}
	}
	err = app.writeJSON(w, http.StatusOK, jsonEnvelope{"session_response": output}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listSessionResponsesHandler(w http.ResponseWriter, r *http.Request) {
	scenarioSessionID, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid scenario session ID parameter"))
		return
	}
	responses, err := app.models.SessionResponses.GetAllByScenarioSessionID(scenarioSessionID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	outputResponses := make([]data.SessionResponseOutput, len(responses))
	for i, sr := range responses {
		output := data.SessionResponseOutput{
			ID:                sr.ID,
			ScenarioSessionID: sr.ScenarioSessionID,
			SubmittedAt:       sr.SubmittedAt,
		}
		if sr.RawAnswers != nil {
			if errUnmarshal := json.Unmarshal(sr.RawAnswers, &output.RawAnswers); errUnmarshal != nil {
				app.logger.Error("Failed to unmarshal raw_answers for list output", "response_id", sr.ID, "error", errUnmarshal)
			}
		}
		if sr.AIFeedback != nil {
			if errUnmarshal := json.Unmarshal(sr.AIFeedback, &output.AIFeedback); errUnmarshal != nil {
				app.logger.Error("Failed to unmarshal ai_feedback for list output", "response_id", sr.ID, "error", errUnmarshal)
			}
		}
		outputResponses[i] = output
	}
	err = app.writeJSON(w, http.StatusOK, jsonEnvelope{"session_responses": outputResponses}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
