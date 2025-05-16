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
	router.HandlerFunc(http.MethodPost, "/v1/scenarios", app.createScenarioHandler)

	return app.recoverPanic(app.enableCORS(router))
}
