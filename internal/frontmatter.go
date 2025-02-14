package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type FrontMatter struct {
	ID         string   `yaml:"id"`
	Title      string   `yaml:"title"`
	Type       string   `yaml:"type"`
	Tags       []string `yaml:"tags"`
	Links      []string `yaml:"links"`
	TaskStatus string   `yaml:"task_status"`
	CreatedAt  string   `yaml:"created_at"`
	UpdatedAt  string   `yaml:"updated_at"`
	Archived   bool     `yaml:"archived"`
	Deleted    bool     `yaml:"deleted"`
}

func ExtractFrontMatter(content string) (string, error) {
	re := regexp.MustCompile(`(?s)^---\n(.*?)\n---\n`)
	matches := re.FindStringSubmatch(content)
	if len(matches) < 2 {
		return "", fmt.Errorf("Front matter not found")
	}
	return matches[1], nil
}

func ParseFrontMatter(content string) (FrontMatter, string, error) {
	if !strings.HasPrefix(content, "---") {
		return FrontMatter{}, content, fmt.Errorf("フロントマターが見つかりません")
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return FrontMatter{}, content, fmt.Errorf("フロントマターの形式が正しくありません")
	}

	frontMatterStr := strings.TrimSpace(parts[1])
	body := strings.TrimSpace(parts[2])

	// YAML をパース
	var frontMatter FrontMatter
	err := yaml.Unmarshal([]byte(frontMatterStr), &frontMatter)
	if err != nil {
		return FrontMatter{}, content, err
	}

	return frontMatter, body, nil
}

func UpdateFrontMatter(frontMatter *FrontMatter, body string) string {
	// YAML に再変換
	frontMatterBytes, _ := yaml.Marshal(frontMatter)

	// --- を維持して YAML と本文を結合
	return fmt.Sprintf("---\n%s---\n\n%s", string(frontMatterBytes), body)
}

// 💾 **zettel.json を保存**
func SaveUpdatedJson(zettels []Zettel, config *Config) {
	updatedJson, err := json.MarshalIndent(zettels, "", "  ")
	if err != nil {
		fmt.Println("⚠️ JSON の変換エラー:", err)
		return
	}

	err = os.WriteFile(config.ZettelJson, updatedJson, 0644)
	if err != nil {
		fmt.Println("⚠️ JSON 書き込みエラー:", err)
	} else {
		fmt.Println("✅ JSON 更新完了:", config.ZettelJson)
	}
}
