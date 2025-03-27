package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/berberapan/info-eval/internal/assert"
)

func TestRecoverPanic(t *testing.T) {
	app := &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	tests := []struct {
		name           string
		handler        http.Handler
		expectedStatus int
		expectedHeader string
	}{
		{
			name: "No panic",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			expectedStatus: http.StatusOK,
			expectedHeader: "",
		},
		{
			name: "Panic",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(fmt.Errorf("Something gone wrong here"))
			}),
			expectedStatus: http.StatusInternalServerError,
			expectedHeader: "close",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest("GET", "/", nil)

			app.recoverPanic(tt.handler).ServeHTTP(recorder, request)
			assert.Equal(t, recorder.Code, tt.expectedStatus)
			assert.Equal(t, recorder.Header().Get("Connection"), tt.expectedHeader)
		})
	}
}
