package vcs

import (
	"runtime/debug"
	"testing"

	"github.com/berberapan/info-eval/internal/assert"
)

func TestVersion(t *testing.T) {
	original := readBuildInfo
	defer func() {
		readBuildInfo = original
	}()

	tests := []struct {
		name            string
		version         string
		expectedVersion string
	}{
		{
			name:            "Version with info",
			version:         "v1.89.1",
			expectedVersion: "v1.89.1",
		},
		{
			name:            "Version with no info",
			version:         "",
			expectedVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readBuildInfo = func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{
					Main: debug.Module{
						Version: tt.version,
					},
				}, true
			}
			actualVersion := Version()
			assert.Equal(t, actualVersion, tt.expectedVersion)
		})
	}
}
