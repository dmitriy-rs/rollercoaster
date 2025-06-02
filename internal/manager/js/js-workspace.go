package jsmanager

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager/cache"
)

// PackageManagerInfo holds information about detected package manager
type PackageManagerInfo struct {
	Type     string
	LockFile string
	Version  string
}

type JsWorkspace interface {
	Name() string
	ExecName() string
	Cmd() *exec.Cmd

	ExecuteCmd() *exec.Cmd
	InstallCmd() *exec.Cmd
	AddCmd() *exec.Cmd
	RemoveCmd() *exec.Cmd
}

func ParseJsWorkspace(dir *string, defaultJSManager string) (*JsWorkspace, error) {
	// Check if package.json exists first
	if !cache.DefaultFSCache.FileExists(filepath.Join(*dir, packageJsonFilename)) {
		return nil, nil
	}

	// Optimized package manager detection with single directory scan
	pmInfo, err := detectPackageManagerOptimized(*dir)
	if err != nil {
		return nil, err
	}

	// Pre-allocate slice with capacity 1 (typically only one workspace type)
	var workspace JsWorkspace

	if pmInfo != nil {
		switch pmInfo.Type {
		case "bun":
			bunWorkspace, err := ParseBunWorkspace(dir)
			if err != nil {
				logger.Warning(err.Error())
			}
			if bunWorkspace != nil {
				workspace = bunWorkspace
			}
		case "pnpm":
			pnpmWorkspace, err := ParsePnpmWorkspace(dir)
			if err != nil {
				logger.Warning(err.Error())
			}
			if pnpmWorkspace != nil {
				workspace = pnpmWorkspace
			}
		case "yarn":
			yarnWorkspace, err := ParseYarnWorkspace(dir)
			if err != nil {
				logger.Warning(err.Error())
			}
			if yarnWorkspace != nil {
				workspace = yarnWorkspace
			}
		case "npm":
			npmWorkspace, err := ParseNpmWorkspace(dir)
			if err != nil {
				logger.Warning(err.Error())
			}
			if npmWorkspace != nil {
				workspace = npmWorkspace
			}
		}
	}

	// Only use default if package.json exists but no workspace was detected
	if workspace == nil {
		defaultJSManager := parseDefaultJSManager(defaultJSManager)
		if defaultJSManager == nil {
			return nil, nil
		}
		return &defaultJSManager, nil
	}

	return &workspace, nil
}

// detectPackageManagerOptimized uses single directory scan to detect package managers
func detectPackageManagerOptimized(dir string) (*PackageManagerInfo, error) {
	// Batch check for all lock files in one operation
	lockFiles := []string{bunLockFilename, bunLockTextFilename, pnpmLockFilename, "yarn.lock", "package-lock.json"}
	lockFileExists := cache.DefaultFSCache.FindFilesInDirectory(dir, lockFiles)

	// Count existing lock files for early validation
	var detectedPMs []string
	var pmInfo *PackageManagerInfo

	if lockFileExists[bunLockFilename] || lockFileExists[bunLockTextFilename] {
		detectedPMs = append(detectedPMs, "bun")
		lockFile := bunLockFilename
		if lockFileExists[bunLockTextFilename] {
			lockFile = bunLockTextFilename
		}
		pmInfo = &PackageManagerInfo{
			Type:     "bun",
			LockFile: lockFile,
		}
	}
	if lockFileExists[pnpmLockFilename] {
		detectedPMs = append(detectedPMs, "pnpm")
		if pmInfo == nil {
			pmInfo = &PackageManagerInfo{
				Type:     "pnpm",
				LockFile: pnpmLockFilename,
			}
		}
	}
	if lockFileExists["yarn.lock"] {
		detectedPMs = append(detectedPMs, "yarn")
		if pmInfo == nil {
			pmInfo = &PackageManagerInfo{
				Type:     "yarn",
				LockFile: "yarn.lock",
			}
		}
	}
	if lockFileExists["package-lock.json"] {
		detectedPMs = append(detectedPMs, "npm")
		if pmInfo == nil {
			pmInfo = &PackageManagerInfo{
				Type:     "npm",
				LockFile: "package-lock.json",
			}
		}
	}

	// Early exit if multiple lock files detected
	if len(detectedPMs) > 1 {
		return nil, fmt.Errorf("multiple package manager lock files detected: %s", strings.Join(detectedPMs, ", "))
	}

	return pmInfo, nil
}

func parseDefaultJSManager(defaultJSManager string) JsWorkspace {
	switch defaultJSManager {
	case "bun":
		workspace := GetDefaultBunWorkspace()
		return &workspace
	case "npm":
		workspace := GetDefaultNpmWorkspace()
		return &workspace
	case "yarn":
		workspace := GetDefaultYarnWorkspace()
		return &workspace
	case "pnpm":
		workspace := GetDefaultPnpmWorkspace()
		return &workspace
	default:
		return nil
	}
}
