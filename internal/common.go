package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

func LoadJson(config Config) ([]Zettel, error) {
	var zettels []Zettel
	if _, err := os.Stat(config.ZettelJson); err == nil {
		jsonBytes, err := os.ReadFile(config.ZettelJson)
		if err != nil {
			return []Zettel{}, fmt.Errorf("⚠️ JSON 読み込みエラー1: %v", err)
		}
		if len(jsonBytes) > 0 {
			err = json.Unmarshal(jsonBytes, &zettels)
			if err != nil {
				return []Zettel{}, fmt.Errorf("⚠️ JSON パースエラー: %v", err)
			}
		}
	} else if !os.IsNotExist(err) {
		// `Stat` がエラーになり、ファイルが存在しない場合以外のエラーをチェック
		return []Zettel{}, fmt.Errorf("⚠️ JSON ファイル確認エラー: %v", err)
	}
	return zettels, nil
}

func InsertZettelToJson(zettel Zettel, config Config) error {

	zettels, err := LoadJson(config)
	if err != nil {
		fmt.Println("Error:", err)
	}
	// 新しい ID を決定（最大の ID に +1 する）
	newID := 1
	if len(zettels) > 0 {
		newID = (len(zettels) + 1)
	}
	zettel.ID = strconv.Itoa(newID)

	zettels = append(zettels, zettel)

	// JSON にシリアライズ（見やすく整形）
	jsonBytes, err := json.MarshalIndent(zettels, "", "  ")
	if err != nil {
		return fmt.Errorf("⚠️ JSON 変換エラー: %v", err)
	}

	err = os.WriteFile(config.ZettelJson, jsonBytes, 0644)
	if err != nil {
		return fmt.Errorf("⚠️ JSON 書き込みエラー: %v", err)
	}
	fmt.Println("🎉 JSON 書き込み成功!")
	return nil
}
