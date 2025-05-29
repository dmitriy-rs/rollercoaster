package jsmanager_test

import (
	"path/filepath"
	"testing"

	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNpmWorkspace(t *testing.T) {
	tests := []struct {
		name        string
		testdataDir string
		wantNil     bool
		wantError   bool
	}{
		{
			name:        "npm with lock",
			testdataDir: "npm-with-lock",
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "empty directory (no package-lock.json)",
			testdataDir: "empty",
			wantNil:     true,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join("testdata", tt.testdataDir)
			workspace, err := jsmanager.ParseNpmWorkspace(&testDir)

			if tt.wantError {
				assert.Error(t, err, "ParseNpmWorkspace() should return error for %s", tt.testdataDir)
				assert.Nil(t, workspace, "ParseNpmWorkspace() should return nil workspace when error occurs")
				return
			}

			assert.NoError(t, err, "ParseNpmWorkspace() should not return error for %s", tt.testdataDir)

			if tt.wantNil {
				assert.Nil(t, workspace, "ParseNpmWorkspace() should return nil for %s", tt.testdataDir)
				return
			}

			require.NotNil(t, workspace, "ParseNpmWorkspace() should return workspace for %s", tt.testdataDir)
			assert.Equal(t, "npm", workspace.Name(), "ParseNpmWorkspace() workspace name should be 'npm'")
		})
	}
}

func TestNpmWorkspace_Name(t *testing.T) {
	workspace := &jsmanager.NpmWorkspace{}
	got := workspace.Name()

	assert.Equal(t, "npm", got, "NpmWorkspace.Name() should return 'npm'")
}

func TestNpmWorkspace_Cmd(t *testing.T) {
	workspace := &jsmanager.NpmWorkspace{}
	cmd := workspace.Cmd()

	expectedArgs := []string{"npm"}
	assert.Equal(t, expectedArgs, cmd.Args, "NpmWorkspace.Cmd() should return correct args")
}

func TestNpmWorkspace_InstallCmd(t *testing.T) {
	workspace := &jsmanager.NpmWorkspace{}
	cmd := workspace.InstallCmd()

	expectedArgs := []string{"npm", "install"}
	assert.Equal(t, expectedArgs, cmd.Args, "NpmWorkspace.InstallCmd() should return correct args")
}

func TestNpmWorkspace_RunCmd(t *testing.T) {
	workspace := &jsmanager.NpmWorkspace{}
	cmd := workspace.RunCmd()

	expectedArgs := []string{"npm", "run"}
	assert.Equal(t, expectedArgs, cmd.Args, "NpmWorkspace.RunCmd() should return correct args")
}

func TestNpmWorkspace_AddCmd(t *testing.T) {
	workspace := &jsmanager.NpmWorkspace{}
	cmd := workspace.AddCmd()

	expectedArgs := []string{"npm", "i"}
	assert.Equal(t, expectedArgs, cmd.Args, "NpmWorkspace.AddCmd() should return correct args")
}

func TestNpmWorkspace_RemoveCmd(t *testing.T) {
	workspace := &jsmanager.NpmWorkspace{}
	cmd := workspace.RemoveCmd()

	expectedArgs := []string{"npm", "uninstall"}
	assert.Equal(t, expectedArgs, cmd.Args, "NpmWorkspace.RemoveCmd() should return correct args")
}
