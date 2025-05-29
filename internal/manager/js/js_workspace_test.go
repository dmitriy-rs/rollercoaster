package jsmanager_test

import (
	"path/filepath"
	"testing"

	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
)

func TestParseJsWorkspace(t *testing.T) {
	tests := []struct {
		name        string
		testdataDir string
		wantName    string
		wantNil     bool
		wantError   bool
	}{
		{
			name:        "npm with lock",
			testdataDir: "npm-with-lock",
			wantName:    "npm",
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "yarn v1 with lock",
			testdataDir: "yarn-v1-with-lock",
			wantName:    "yarn",
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "yarn v2+ with lock (unsupported)",
			testdataDir: "yarn-v2-with-lock",
			wantName:    "",
			wantNil:     true,
			wantError:   false,
		},
		{
			name:        "pnpm lock v6",
			testdataDir: "pnpm-lock-v6",
			wantName:    "pnpm",
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "pnpm lock v9",
			testdataDir: "pnpm-lock-v9",
			wantName:    "pnpm",
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "pnpm with random lock (unsupported)",
			testdataDir: "pnpm-random-lock",
			wantName:    "",
			wantNil:     true,
			wantError:   false,
		},
		{
			name:        "empty directory",
			testdataDir: "empty",
			wantName:    "",
			wantNil:     true,
			wantError:   false,
		},
		{
			name:        "multiple locks (error)",
			testdataDir: "multiple-locks",
			wantName:    "",
			wantNil:     true,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join("testdata", tt.testdataDir)
			workspace, err := jsmanager.ParseJsWorkspace(&testDir)

			if tt.wantError {
				if err == nil {
					t.Errorf("ParseJsWorkspace() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseJsWorkspace() unexpected error: %v", err)
				return
			}

			if tt.wantNil {
				if workspace != nil {
					t.Errorf("ParseJsWorkspace() expected nil, got %v", workspace)
				}
				return
			}

			if workspace == nil {
				t.Errorf("ParseJsWorkspace() expected workspace, got nil")
				return
			}

			if (*workspace).Name() != tt.wantName {
				t.Errorf("ParseJsWorkspace() name = %v, want %v", (*workspace).Name(), tt.wantName)
			}
		})
	}
}
