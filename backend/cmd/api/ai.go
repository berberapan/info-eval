package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/berberapan/info-eval/internal/data"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

func (app *application) triggerAIFeedbackGeneration(responseID uuid.UUID, scenarioSessionID uuid.UUID, rawAnswersJSON []byte) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		err := app.generateAndStoreAIFeedback(responseID, scenarioSessionID, rawAnswersJSON)
		if err != nil {
			app.logger.Error("Failed to generate or store AI feedback", "response_id", responseID.String(), "scenario_session_id", scenarioSessionID.String(), "error", err)
		}
	}()
}

func (app *application) generateAndStoreAIFeedback(responseID uuid.UUID, scenarioSessionID uuid.UUID, rawAnswersJSON []byte) error {
	var studentAnswers map[string]any
	if err := json.Unmarshal(rawAnswersJSON, &studentAnswers); err != nil {
		return fmt.Errorf("failed to unmarshal raw_answers for response %s: %w", responseID, err)
	}
	scenarioID, err := app.models.ScenarioSessions.GetScenarioByID(scenarioSessionID)
	if err != nil {
		if err == data.ErrRecordNotFound {
			return fmt.Errorf("scenario_session_id %s not found when trying to get scenario_id: %w", scenarioSessionID, err)
		}
		return fmt.Errorf("failed to get scenario_id for scenario_session_id %s: %w", scenarioSessionID, err)
	}
	questions, err := app.models.ExerciseQuestions.GetAllByScenarioIDAsMap(scenarioID)
	if err != nil {
		return fmt.Errorf("failed to fetch questions for scenario %s: %w", scenarioID, err)
	}
	ctx := context.Background()
	if app.config.ai.key == "" {
		app.logger.Warn("Gemini API key is not configured. Skipping AI feedback generation.", "response_id", responseID.String())
		return nil
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(app.config.ai.key))
	if err != nil {
		return fmt.Errorf("failed to create Gemini client for response %s: %w", responseID, err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash-latest")
	model.SetTemperature(0.1)

	aiFeedbackResults := make(map[string]string)

	for qIDStr, studentAnswerInterface := range studentAnswers {
		questionID, err := uuid.Parse(qIDStr)
		if err != nil {
			app.logger.Warn("Invalid UUID format for question ID in raw_answers", "question_id_str", qIDStr, "response_id", responseID.String())
			continue
		}
		questionDetails, ok := questions[questionID]
		if !ok {
			app.logger.Warn("Question details not found for ID in raw_answers", "question_id", questionID.String(), "response_id", responseID.String())
			continue
		}
		if questionDetails.ExerciseType == data.FreeTextType {
			studentAnswer, ok := studentAnswerInterface.(string)
			if !ok {
				app.logger.Warn("Free-text answer is not a string", "question_id", questionID.String(), "response_id", responseID.String(), "answer_type", fmt.Sprintf("%T", studentAnswerInterface))
				continue
			}
			var promptBuilder strings.Builder
			promptBuilder.WriteString(fmt.Sprintf("The student was asked the following question: \"%s\"\n", questionDetails.Question))
			if questionDetails.PromptGuidance.Valid && questionDetails.PromptGuidance.String != "" {
				promptBuilder.WriteString(fmt.Sprintf("Consider the following guidance when evaluating the answer: \"%s\"\n", questionDetails.PromptGuidance.String))
			}
			promptBuilder.WriteString(fmt.Sprintf("The student's answer was: \"%s\"\n\n", studentAnswer))
			promptBuilder.WriteString("You're an expert on information evaluation and sources. Provide concise, constructive feedback on the student's answer (1-3 sentences). Focus on clarity, accuracy, and completeness. The feedback should be encouraging and help the student understand how to improve or what they did well. The feedback should also be in Swedish")
			prompt := promptBuilder.String()

			apiCtx, cancelAPICall := context.WithTimeout(ctx, 60*time.Second)

			var geminiResp *genai.GenerateContentResponse
			var genErr error
			maxRetries := 2
			for i := 0; i < maxRetries; i++ {
				geminiResp, genErr = model.GenerateContent(apiCtx, genai.Text(prompt))
				if genErr == nil {
					break
				}
				app.logger.Error("Error generating feedback from Gemini", "attempt", i+1, "max_attempts", maxRetries, "response_id", responseID.String(), "question_id", questionID.String(), "error", genErr)
				if i < maxRetries-1 {
					select {
					case <-time.After(time.Duration(i+1) * 2 * time.Second):
					case <-apiCtx.Done():
						genErr = apiCtx.Err()
						break
					}
				}
				if genErr != nil && strings.Contains(genErr.Error(), "context deadline exceeded") {
					break
				}
			}
			cancelAPICall()
			if genErr != nil {
				app.logger.Error("Failed to generate feedback after retries", "response_id", responseID.String(), "question_id", questionID.String(), "error", genErr)
				aiFeedbackResults[qIDStr] = "Error: Could not generate feedback at this time."
				continue
			}
			feedbackText := extractTextFromGeminiResponse(geminiResp)
			if feedbackText == "" {
				aiFeedbackResults[qIDStr] = "No specific feedback generated."
			} else {
				aiFeedbackResults[qIDStr] = feedbackText
			}
		}
	}
	if len(aiFeedbackResults) > 0 {
		err = app.models.SessionResponses.AddFeedback(responseID, aiFeedbackResults)
		if err != nil {
			return fmt.Errorf("failed to store AI feedback for response %s: %w", responseID, err)
		}
	}
	return nil
}

func extractTextFromGeminiResponse(resp *genai.GenerateContentResponse) string {
	var feedback strings.Builder
	if resp != nil {
		for _, cand := range resp.Candidates {
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					if textPart, ok := part.(genai.Text); ok {
						feedback.WriteString(string(textPart))
					}
				}
			}
		}
	}
	return feedback.String()
}
