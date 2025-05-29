package jsmanager_test

import (
	"path/filepath"
	"testing"

	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseYarnWorkspace(t *testing.T) {
	tests := []struct {
		name        string
		testdataDir string
		wantVersion int
		wantNil     bool
		wantError   bool
	}{
		{
			name:        "yarn v1 with lock",
			testdataDir: "yarn-v1-with-lock",
			wantVersion: 1,
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "yarn v2+ with lock (unsupported)",
			testdataDir: "yarn-v2-with-lock",
			wantVersion: 0,
			wantNil:     true,
			wantError:   true,
		},
		{
			name:        "empty directory (no yarn.lock)",
			testdataDir: "empty",
			wantVersion: 0,
			wantNil:     true,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join("testdata", tt.testdataDir)
			workspace, err := jsmanager.ParseYarnWorkspace(&testDir)

			if tt.wantError {
				assert.Error(t, err, "ParseYarnWorkspace() should return error for %s", tt.testdataDir)
				assert.Nil(t, workspace, "ParseYarnWorkspace() should return nil workspace when error occurs")
				return
			}

			assert.NoError(t, err, "ParseYarnWorkspace() should not return error for %s", tt.testdataDir)

			if tt.wantNil {
				assert.Nil(t, workspace, "ParseYarnWorkspace() should return nil for %s", tt.testdataDir)
				return
			}

			require.NotNil(t, workspace, "ParseYarnWorkspace() should return workspace for %s", tt.testdataDir)
			assert.Equal(t, "yarn", workspace.Name(), "ParseYarnWorkspace() workspace name should be 'yarn'")
		})
	}
}

func TestYarnWorkspace_Name(t *testing.T) {
	workspace := &jsmanager.YarnWorkspace{}
	got := workspace.Name()

	assert.Equal(t, "yarn", got, "YarnWorkspace.Name() should return 'yarn'")
}

func TestYarnWorkspace_Cmd(t *testing.T) {
	workspace := &jsmanager.YarnWorkspace{}
	cmd := workspace.Cmd()

	expectedArgs := []string{"yarn"}
	assert.Equal(t, expectedArgs, cmd.Args, "YarnWorkspace.Cmd() should return correct args")
}

func TestYarnWorkspace_InstallCmd(t *testing.T) {
	workspace := &jsmanager.YarnWorkspace{}
	cmd := workspace.InstallCmd()

	expectedArgs := []string{"yarn", "install"}
	assert.Equal(t, expectedArgs, cmd.Args, "YarnWorkspace.InstallCmd() should return correct args")
}

func TestYarnWorkspace_RunCmd(t *testing.T) {
	workspace := &jsmanager.YarnWorkspace{}
	cmd := workspace.RunCmd()

	expectedArgs := []string{"yarn", "run"}
	assert.Equal(t, expectedArgs, cmd.Args, "YarnWorkspace.RunCmd() should return correct args")
}

func TestYarnWorkspace_AddCmd(t *testing.T) {
	workspace := &jsmanager.YarnWorkspace{}
	cmd := workspace.AddCmd()

	expectedArgs := []string{"yarn", "add"}
	assert.Equal(t, expectedArgs, cmd.Args, "YarnWorkspace.AddCmd() should return correct args")
}

func TestYarnWorkspace_RemoveCmd(t *testing.T) {
	workspace := &jsmanager.YarnWorkspace{}
	cmd := workspace.RemoveCmd()

	expectedArgs := []string{"yarn", "remove"}
	assert.Equal(t, expectedArgs, cmd.Args, "YarnWorkspace.RemoveCmd() should return correct args")
}
