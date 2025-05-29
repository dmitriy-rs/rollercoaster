package jsmanager

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
)

type JsWorkspace interface {
	Name() string
	Cmd() *exec.Cmd

	RunCmd() *exec.Cmd
	InstallCmd() *exec.Cmd
	AddCmd() *exec.Cmd
	RemoveCmd() *exec.Cmd
}

func ParseJsWorkspace(dir *string) (*JsWorkspace, error) {
	_, err := os.Stat(path.Join(*dir, packageJsonFilename))
	if err != nil {
		return nil, nil
	}

	workspaces := []JsWorkspace{}

	pnpmWorkspace, err := ParsePnpmWorkspace(dir)
	if err != nil {
		logger.Warning(err.Error())
	}
	if pnpmWorkspace != nil {
		workspaces = append(workspaces, pnpmWorkspace)
	}

	yarnWorkspace, err := ParseYarnWorkspace(dir)
	if err != nil {
		logger.Warning(err.Error())
	}
	if yarnWorkspace != nil {
		workspaces = append(workspaces, yarnWorkspace)
	}

	npmWorkspace, err := ParseNpmWorkspace(dir)
	if err != nil {
		logger.Warning(err.Error())
	}
	if npmWorkspace != nil {
		workspaces = append(workspaces, npmWorkspace)
	}

	if len(workspaces) == 0 {
		return nil, nil
	}

	if len(workspaces) > 1 {
		formattedWorkspaces := strings.Builder{}
		for _, workspace := range workspaces {
			formattedWorkspaces.WriteString(workspace.Name())
			formattedWorkspaces.WriteString(", ")
		}
		return nil, fmt.Errorf("multiple workspace package managers found: %s", formattedWorkspaces.String())
	}

	return &workspaces[0], nil
}
