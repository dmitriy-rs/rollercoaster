package jsmanager_test

import (
	"path/filepath"
	"testing"

	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBunWorkspace(t *testing.T) {
	tests := []struct {
		name        string
		testdataDir string
		wantNil     bool
		wantError   bool
	}{
		{
			name:        "bun with binary lock",
			testdataDir: "bun-with-binary-lock",
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "bun with text lock",
			testdataDir: "bun-with-text-lock",
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "empty directory (no bun lock files)",
			testdataDir: "empty",
			wantNil:     true,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join("testdata", tt.testdataDir)
			workspace, err := jsmanager.ParseBunWorkspace(&testDir)

			if tt.wantError {
				assert.Error(t, err, "ParseBunWorkspace() should return error for %s", tt.testdataDir)
				assert.Nil(t, workspace, "ParseBunWorkspace() should return nil workspace when error occurs")
				return
			}

			assert.NoError(t, err, "ParseBunWorkspace() should not return error for %s", tt.testdataDir)

			if tt.wantNil {
				assert.Nil(t, workspace, "ParseBunWorkspace() should return nil for %s", tt.testdataDir)
				return
			}

			require.NotNil(t, workspace, "ParseBunWorkspace() should return workspace for %s", tt.testdataDir)
			assert.Equal(t, "bun", workspace.Name(), "ParseBunWorkspace() workspace name should be 'bun'")
		})
	}
}

func TestBunWorkspace_Name(t *testing.T) {
	workspace := &jsmanager.BunWorkspace{}
	got := workspace.Name()

	assert.Equal(t, "bun", got, "BunWorkspace.Name() should return 'bun'")
}

func TestBunWorkspace_ExecName(t *testing.T) {
	workspace := &jsmanager.BunWorkspace{}
	got := workspace.ExecName()

	assert.Equal(t, "bunx", got, "BunWorkspace.ExecName() should return 'bunx'")
}

func TestBunWorkspace_Cmd(t *testing.T) {
	workspace := &jsmanager.BunWorkspace{}
	cmd := workspace.Cmd()

	expectedArgs := []string{"bun", "run"}
	assert.Equal(t, expectedArgs, cmd.Args, "BunWorkspace.Cmd() should return correct args")
}

func TestBunWorkspace_InstallCmd(t *testing.T) {
	workspace := &jsmanager.BunWorkspace{}
	cmd := workspace.InstallCmd()

	expectedArgs := []string{"bun", "install"}
	assert.Equal(t, expectedArgs, cmd.Args, "BunWorkspace.InstallCmd() should return correct args")
}

func TestBunWorkspace_ExecuteCmd(t *testing.T) {
	workspace := &jsmanager.BunWorkspace{}
	cmd := workspace.ExecuteCmd()

	expectedArgs := []string{"bunx"}
	assert.Equal(t, expectedArgs, cmd.Args, "BunWorkspace.ExecuteCmd() should return correct args")
}

func TestBunWorkspace_AddCmd(t *testing.T) {
	workspace := &jsmanager.BunWorkspace{}
	cmd := workspace.AddCmd()

	expectedArgs := []string{"bun", "add"}
	assert.Equal(t, expectedArgs, cmd.Args, "BunWorkspace.AddCmd() should return correct args")
}

func TestBunWorkspace_RemoveCmd(t *testing.T) {
	workspace := &jsmanager.BunWorkspace{}
	cmd := workspace.RemoveCmd()

	expectedArgs := []string{"bun", "remove"}
	assert.Equal(t, expectedArgs, cmd.Args, "BunWorkspace.RemoveCmd() should return correct args")
}
