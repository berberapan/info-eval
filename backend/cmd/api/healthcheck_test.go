package main

import (
	"net/http"
	"testing"

	"github.com/berberapan/info-eval/internal/assert"
)

func TestHealthcheck(t *testing.T) {
	app := newTestApplication(t)
	s := newTestServer(t, app.routes())

	statusCode, _, body := s.get(t, "/v1/healthcheck")

	assert.Equal(t, statusCode, http.StatusOK)
	assert.StringContains(t, body, "available")
	assert.StringContains(t, body, "version")
}
