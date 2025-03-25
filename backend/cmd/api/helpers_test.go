package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/berberapan/info-eval/internal/assert"
)

func TestApplication_writeJSON(t *testing.T) {
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
