package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/scenario/:id", app.showScenarioHandler)
	router.HandlerFunc(http.MethodGet, "/v1/scenarios", app.showScenariosHandler)
	router.HandlerFunc(http.MethodPost, "/v1/scenarios", app.requireAuthenticatedUser(http.HandlerFunc(app.createScenarioHandler)))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.requireAuthenticatedUser(http.HandlerFunc(app.registerUserHandle)))
	router.HandlerFunc(http.MethodGet, "/v1/users/me", app.requireAuthenticatedUser(http.HandlerFunc(app.userProfileHandler)))

	router.HandlerFunc(http.MethodPost, "/v1/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/logout", app.requireAuthenticatedUser(http.HandlerFunc(app.removeAuthenticationTokenHandler)))

	router.HandlerFunc(http.MethodPost, "/v1/sessions", app.requireAuthenticatedUser(http.HandlerFunc(app.createScenarioSessionHandler)))
	router.HandlerFunc(http.MethodGet, "/v1/sessions/:id", app.getScenarioSessionHandler)
	router.HandlerFunc(http.MethodGet, "/v1/sessions/:id/scenario", app.getScenarioIDHandler)
	router.HandlerFunc(http.MethodPost, "/v1/sessions/:id/responses", app.createSessionResponseHandler)
	router.HandlerFunc(http.MethodGet, "/v1/sessions/:id/responses", app.listSessionResponsesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/session-responses/:id", app.getSessionResponseHandler)

	return app.recoverPanic(app.enableCORS(app.authenticate(router)))
}
