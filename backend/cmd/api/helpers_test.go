package main

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/berberapan/info-eval/internal/assert"
)

func TestWriteJSON(t *testing.T) {
	app := &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	tests := []struct {
		name       string
		status     int
		data       jsonEnvelope
		headers    http.Header
		wantStatus int
		wantBody   string
		wantHeader string
		wantErr    bool
	}{
		{
			name:       "valid JSON",
			status:     http.StatusOK,
			data:       jsonEnvelope{"stockholm": 9, "dalarna": 1},
			headers:    http.Header{"X-Custom": []string{"1990"}},
			wantStatus: http.StatusOK,
			wantBody:   "{\n\t\"dalarna\": 1,\n\t\"stockholm\": 9\n}\n",
			wantHeader: "1990",
			wantErr:    false,
		},
		{
			name:       "empty data",
			status:     http.StatusAccepted,
			data:       jsonEnvelope{},
			headers:    http.Header{},
			wantStatus: http.StatusAccepted,
			wantBody:   "{}\n",
			wantHeader: "",
			wantErr:    false,
		},
		{
			name:       "invalid JSON data",
			status:     http.StatusOK,
			data:       jsonEnvelope{"invalid": func() {}},
			headers:    http.Header{},
			wantStatus: http.StatusOK,
			wantBody:   "",
			wantHeader: "",
			wantErr:    true,
		},
		{
			name:       "nil data",
			status:     http.StatusOK,
			data:       nil,
			headers:    http.Header{},
			wantStatus: http.StatusOK,
			wantBody:   "null\n",
			wantHeader: "",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := app.writeJSON(w, tt.status, tt.data, tt.headers)

			assert.Equal(t, (err != nil), tt.wantErr)
			assert.Equal(t, w.Code, tt.wantStatus)
			if tt.wantBody != "" {
				assert.Equal(t, w.Body.String(), tt.wantBody)
			}
			assert.Equal(t, len(w.Header().Get("X-Custom")), len(tt.wantHeader))
			assert.Equal(t, w.Header().Get("X-Custom"), tt.wantHeader)
		})
	}
}

func TestReadJSON(t *testing.T) {
	app := &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	tests := []struct {
		name          string
		data          any
		expectedErr   error
		expectedValue any
	}{
		{
			name:          "Valid JSON, decoding fine",
			data:          "{\"info\": \"eval\"}",
			expectedErr:   nil,
			expectedValue: "eval",
		},
		{
			name:          "Double JSON",
			data:          "{\"info\": \"eval\"}{\"info\": \"eval\"}",
			expectedErr:   errors.New("body only allowed to contain one JSON value"),
			expectedValue: "eval",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type ts struct {
				Info any `json:"info"`
			}
			var data ts
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(tt.data.(string)))

			err := app.readJSON(w, r, &data)
			if err != nil {
				assert.Equal(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
			assert.Equal(t, data.Info, tt.expectedValue)
		})
	}
}
