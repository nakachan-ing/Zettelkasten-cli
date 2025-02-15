package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Cleanup old files in the trash directory
func CleanupTrash(trashDir string, retentionPeriod time.Duration) error {
	files, err := os.ReadDir(trashDir)
	if err != nil {
		return fmt.Errorf("❌ Failed to read trash directory: %w", err)
	}
	now := time.Now()

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := filepath.Join(trashDir, file.Name())
		fileInfo, err := file.Info()
		if err != nil {
			log.Printf("⚠️ Failed to get file info: %s (%v)", filePath, err)
			continue
		}
		modTime := fileInfo.ModTime()

		if now.Sub(modTime) > retentionPeriod {
			if err := os.Remove(filePath); err != nil {
				log.Printf("❌ Failed to remove trash file: %s (%v)", filePath, err)
			} else {
				log.Printf("✅ Removed trash file: %s", filePath)
			}
		}
	}
	return nil
}

// Cleanup old backups
func CleanupBackups(backupDir string, retentionPeriod time.Duration) error {
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("❌ Failed to read backup directory: %w", err)
	}
	now := time.Now()

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := filepath.Join(backupDir, file.Name())
		fileInfo, err := file.Info()
		if err != nil {
			log.Printf("⚠️ Failed to get file info: %s (%v)", filePath, err)
			continue
		}
		modTime := fileInfo.ModTime()

		if now.Sub(modTime) > retentionPeriod {
			if err := os.Remove(filePath); err != nil {
				log.Printf("❌ Failed to remove backup file: %s (%v)", filePath, err)
			} else {
				log.Printf("✅ Removed backup file: %s", filePath)
			}
		}
	}
	return nil
}

// Load Zettels from JSON file
func LoadJson(config Config) ([]Zettel, error) {
	var zettels []Zettel
	if _, err := os.Stat(config.ZettelJson); err == nil {
		jsonBytes, err := os.ReadFile(config.ZettelJson)
		if err != nil {
			return nil, fmt.Errorf("❌ Failed to read JSON file: %w", err)
		}
		if len(jsonBytes) > 0 {
			err = json.Unmarshal(jsonBytes, &zettels)
			if err != nil {
				return nil, fmt.Errorf("❌ Failed to parse JSON: %w", err)
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("❌ Failed to check JSON file: %w", err)
	}
	return zettels, nil
}

// Insert a new Zettel into the JSON file
func InsertZettelToJson(zettel Zettel, config Config) error {
	zettels, err := LoadJson(config)
	if err != nil {
		log.Printf("❌ Error loading JSON: %v", err)
		return err
	}

	// Assign a new ID (incremental)
	newID := 1
	if len(zettels) > 0 {
		newID = len(zettels) + 1
	}
	zettel.ID = strconv.Itoa(newID)

	zettels = append(zettels, zettel)

	// Serialize JSON
	jsonBytes, err := json.MarshalIndent(zettels, "", "  ")
	if err != nil {
		return fmt.Errorf("❌ Failed to convert to JSON: %w", err)
	}

	err = os.WriteFile(config.ZettelJson, jsonBytes, 0644)
	if err != nil {
		return fmt.Errorf("❌ Failed to write JSON file: %w", err)
	}

	log.Println("✅ Successfully updated JSON file!")
	return nil
}
