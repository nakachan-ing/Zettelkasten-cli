package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func CleanupBackups(backupDir string, retentionPeriod time.Duration) error {
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := filepath.Join(backupDir, file.Name())
		fileInfo, err := file.Info()
		if err != nil {
			return err
		}
		modTime := fileInfo.ModTime()

		if now.Sub(modTime) > retentionPeriod {
			err := os.Remove(filePath)
			if err != nil {
				fmt.Printf("Failed to remove backup file %s: %v\n", filePath, err)
			} else {
				fmt.Printf("Removed backup file: %s\n", filePath)
			}
		}
	}
	return nil
}
