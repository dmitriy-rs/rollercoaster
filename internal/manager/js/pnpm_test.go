package jsmanager_test

import (
	"path/filepath"
	"testing"

	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePnpmWorkspace(t *testing.T) {
	tests := []struct {
		name        string
		testdataDir string
		wantVersion int
		wantNil     bool
		wantError   bool
	}{
		{
			name:        "pnpm lock v6",
			testdataDir: "pnpm-lock-v6",
			wantVersion: 9,
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "pnpm lock v9",
			testdataDir: "pnpm-lock-v9",
			wantVersion: 10,
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "pnpm with random lock (unsupported)",
			testdataDir: "pnpm-random-lock",
			wantVersion: 0,
			wantNil:     true,
			wantError:   true,
		},
		{
			name:        "empty directory (no pnpm-lock.yaml)",
			testdataDir: "empty",
			wantVersion: 0,
			wantNil:     true,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join("testdata", tt.testdataDir)
			workspace, err := jsmanager.ParsePnpmWorkspace(&testDir)

			if tt.wantError {
				assert.Error(t, err, "ParsePnpmWorkspace() should return error for %s", tt.testdataDir)
				assert.Nil(t, workspace, "ParsePnpmWorkspace() should return nil workspace when error occurs")
				return
			}

			assert.NoError(t, err, "ParsePnpmWorkspace() should not return error for %s", tt.testdataDir)

			if tt.wantNil {
				assert.Nil(t, workspace, "ParsePnpmWorkspace() should return nil for %s", tt.testdataDir)
				return
			}

			require.NotNil(t, workspace, "ParsePnpmWorkspace() should return workspace for %s", tt.testdataDir)
			assert.Contains(t, workspace.Name(), "pnpm", "ParsePnpmWorkspace() workspace name should be 'pnpm'")
		})
	}
}

func TestPnpmWorkspace_Name(t *testing.T) {
	workspace := &jsmanager.PnpmWorkspace{}
	got := workspace.Name()

	assert.Equal(t, "pnpm", got, "PnpmWorkspace.Name() should return 'pnpm'")
}

func TestPnpmWorkspace_Cmd(t *testing.T) {
	workspace := &jsmanager.PnpmWorkspace{}
	cmd := workspace.Cmd()

	expectedArgs := []string{"pnpm", "run"}
	assert.Equal(t, expectedArgs, cmd.Args, "PnpmWorkspace.Cmd() should return correct args")
}

func TestPnpmWorkspace_InstallCmd(t *testing.T) {
	workspace := &jsmanager.PnpmWorkspace{}
	cmd := workspace.InstallCmd()

	expectedArgs := []string{"pnpm", "install"}
	assert.Equal(t, expectedArgs, cmd.Args, "PnpmWorkspace.InstallCmd() should return correct args")
}

func TestPnpmWorkspace_ExecuteCmd(t *testing.T) {
	workspace := &jsmanager.PnpmWorkspace{}
	cmd := workspace.ExecuteCmd()

	expectedArgs := []string{"pnpx"}
	assert.Equal(t, expectedArgs, cmd.Args, "PnpmWorkspace.ExecuteCmd() should return correct args")
}

func TestPnpmWorkspace_AddCmd(t *testing.T) {
	workspace := &jsmanager.PnpmWorkspace{}
	cmd := workspace.AddCmd()

	expectedArgs := []string{"pnpm", "add"}
	assert.Equal(t, expectedArgs, cmd.Args, "PnpmWorkspace.AddCmd() should return correct args")
}

func TestPnpmWorkspace_RemoveCmd(t *testing.T) {
	workspace := &jsmanager.PnpmWorkspace{}
	cmd := workspace.RemoveCmd()

	expectedArgs := []string{"pnpm", "remove"}
	assert.Equal(t, expectedArgs, cmd.Args, "PnpmWorkspace.RemoveCmd() should return correct args")
}
