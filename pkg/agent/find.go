package agent

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func findLogDir() (path string, err error) {

	username, err := getCurrentUser()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	baseDir := filepath.Join("C:\\Users", username)
	targetDir := `https_teams.microsoft.com_0.indexeddb.leveldb`

	err = filepath.WalkDir(baseDir, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsPermission(walkErr) {
				return nil // Skip directories where access is denied
			}
			return fmt.Errorf("error accessing directory %s: %w", currentPath, walkErr)
		}

		if d.IsDir() && d.Name() == targetDir {
			path = currentPath
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to walk the base directory %s: %w", baseDir, err)
	}
	if path == "" {
		return "", fmt.Errorf("log directory not found in directory %s", baseDir)
	}
	return path, nil
}

func findLogFiles(logDirPath string) ([]string, error) {
	var logFiles []string

	err := filepath.WalkDir(logDirPath, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsPermission(walkErr) {
				return nil // Skip files where access is denied (This should not happen there, but you never know)
			}
			return fmt.Errorf("error accessing file %s: %w", currentPath, walkErr)
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".log") {
			logFiles = append(logFiles, currentPath)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search for log files in directory %s: %w", logDirPath, err)
	}

	return logFiles, nil
}
