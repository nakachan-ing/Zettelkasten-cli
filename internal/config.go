package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	NoteDir string `yaml:"note_dir"`
	Editor  string `yaml:"editor"`
	Backup  struct {
		Enable    bool   `yaml:"enable"`
		Frequency int    `yaml:"frequency"`
		Retention int    `yaml:"retention"`
		BackupDir string `yaml:"backup_dir"`
	}
}

func GetConfigPath() string {
	// 環境変数 `ZK_CONFIG` が設定されていれば、それを優先
	if customConfig := os.Getenv("ZK_CONFIG"); customConfig != "" {
		return customConfig
	}

	var configPath string

	switch runtime.GOOS {
	case "windows":
		// Windows の場合 `APPDATA\ztask\config.yaml`
		appData := os.Getenv("APPDATA")
		if appData != "" {
			configPath = filepath.Join(appData, "zettelkasten-cli", "config.yaml")
		} else {
			// `APPDATA` が取得できなかった場合は `USERPROFILE` を使う（レアケース）
			homeDir, _ := os.UserHomeDir()
			configPath = filepath.Join(homeDir, "AppData", "Roaming", "zettelkasten-cli", "config.yaml")
		}

	default: // macOS / Linux
		configDir, err := os.UserConfigDir()
		if err != nil {
			// `os.UserConfigDir()` が失敗した場合は `~/.zettelkasten-cli/config.yaml` にフォールバック
			homeDir, _ := os.UserHomeDir()
			configPath = filepath.Join(homeDir, ".zettelkasten-cli", "config.yaml")
		} else {
			configPath = filepath.Join(configDir, "zettelkasten-cli", "config.yaml")
		}
	}

	return configPath
}

// `~` をホームディレクトリに展開（Windows も考慮）
func expandHomeDir(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("設定ファイルを読み込めません: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("YAML のパースに失敗: %w", err)
	}

	// `~` を展開
	config.NoteDir = expandHomeDir(config.NoteDir)
	config.Backup.BackupDir = expandHomeDir(config.Backup.BackupDir)

	return &config, nil
}
