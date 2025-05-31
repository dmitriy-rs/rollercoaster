package jsmanager_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
)

func TestParseJsWorkspace(t *testing.T) {
	tests := []struct {
		name           string
		testdataDir    string
		wantName       string
		wantNil        bool
		wantError      bool
		defaultManager string
	}{
		{
			name:           "npm with lock",
			testdataDir:    "npm-with-lock",
			wantName:       "npm",
			wantNil:        false,
			wantError:      false,
			defaultManager: "",
		},
		{
			name:           "yarn v1 with lock",
			testdataDir:    "yarn-v1-with-lock",
			wantName:       "yarn",
			wantNil:        false,
			wantError:      false,
			defaultManager: "",
		},
		{
			name:           "yarn v2+ with lock (unsupported)",
			testdataDir:    "yarn-v2-with-lock",
			wantName:       "",
			wantNil:        true,
			wantError:      false,
			defaultManager: "",
		},
		{
			name:           "pnpm lock v6",
			testdataDir:    "pnpm-lock-v6",
			wantName:       "pnpm",
			wantNil:        false,
			wantError:      false,
			defaultManager: "",
		},
		{
			name:           "pnpm lock v9",
			testdataDir:    "pnpm-lock-v9",
			wantName:       "pnpm",
			wantNil:        false,
			wantError:      false,
			defaultManager: "",
		},
		{
			name:           "pnpm with random lock (unsupported)",
			testdataDir:    "pnpm-random-lock",
			wantName:       "",
			wantNil:        true,
			wantError:      false,
			defaultManager: "",
		},
		{
			name:           "empty directory",
			testdataDir:    "empty",
			wantName:       "",
			wantNil:        true,
			wantError:      false,
			defaultManager: "",
		},
		{
			name:           "multiple locks (error)",
			testdataDir:    "multiple-locks",
			wantName:       "",
			wantNil:        true,
			wantError:      true,
			defaultManager: "",
		},
		{
			name:           "empty directory with default",
			testdataDir:    "empty",
			wantName:       "",
			wantNil:        true,
			wantError:      false,
			defaultManager: "npm",
		},
		{
			name:           "package.json without lock files with default",
			testdataDir:    "no-scripts",
			wantName:       "npm",
			wantNil:        false,
			wantError:      false,
			defaultManager: "npm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join("testdata", tt.testdataDir)

			workspace, err := jsmanager.ParseJsWorkspace(&testDir, tt.defaultManager)

			if tt.wantError {
				assert.Error(t, err, "ParseJsWorkspace() should return an error")
				return
			}

			require.NoError(t, err, "ParseJsWorkspace() should not return an error")

			if tt.wantNil {
				assert.Nil(t, workspace, "ParseJsWorkspace() should return nil workspace")
				return
			}

			require.NotNil(t, workspace, "ParseJsWorkspace() should return a workspace")
			assert.Contains(t, (*workspace).Name(), tt.wantName, "workspace name should match expected")
		})
	}
}
