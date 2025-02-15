package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	NoteDir    string `yaml:"note_dir"`
	Editor     string `yaml:"editor"`
	ZettelJson string `yaml:"zettel_json"`
	ArchiveDir string `yaml:"archive_dir"`
	Backup     struct {
		Enable    bool   `yaml:"enable"`
		Frequency int    `yaml:"frequency"`
		Retention int    `yaml:"retention"`
		BackupDir string `yaml:"backup_dir"`
	}
	Trash struct {
		Frequency int    `yaml:"frequency"`
		Retention int    `yaml:"retention"`
		TrashDir  string `yaml:"trash_dir"`
	}
}

func GetConfigPath() (string, error) {
	// Check if the environment variable `ZK_CONFIG` is set
	if customConfig := os.Getenv("ZK_CONFIG"); customConfig != "" {
		return customConfig, nil
	}

	var configPath string

	switch runtime.GOOS {
	case "windows":
		// Use `APPDATA\zettelkasten-cli\config.yaml` if available
		appData := os.Getenv("APPDATA")
		if appData != "" {
			configPath = filepath.Join(appData, "zettelkasten-cli", "config.yaml")
		} else {
			// Fallback to `USERPROFILE` if `APPDATA` is unavailable
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to determine home directory: %w", err)
			}
			configPath = filepath.Join(homeDir, "AppData", "Roaming", "zettelkasten-cli", "config.yaml")
		}

	default: // macOS / Linux
		configDir, err := os.UserConfigDir()
		if err != nil {
			// Fallback to `~/.zettelkasten-cli/config.yaml` if `os.UserConfigDir()` fails
			homeDir, homeErr := os.UserHomeDir()
			if homeErr != nil {
				return "", fmt.Errorf("failed to determine home directory: %w", homeErr)
			}
			configPath = filepath.Join(homeDir, ".zettelkasten-cli", "config.yaml")
			log.Printf("⚠️ Failed to get user config directory, using fallback: %s", configPath)
		} else {
			configPath = filepath.Join(configDir, "zettelkasten-cli", "config.yaml")
		}
	}

	return configPath, nil
}

// Expand `~` to the home directory (Windows included)
func expandHomeDir(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("⚠️ Failed to get home directory: %v", err)
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file (%s): %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Expand `~` in paths
	config.NoteDir = expandHomeDir(config.NoteDir)
	config.Backup.BackupDir = expandHomeDir(config.Backup.BackupDir)
	config.ZettelJson = expandHomeDir(config.ZettelJson)
	config.Trash.TrashDir = expandHomeDir(config.Trash.TrashDir)

	return &config, nil
}
