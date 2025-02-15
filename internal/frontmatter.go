package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

// Parse front matter from note content
func ParseFrontMatter(content string) (FrontMatter, string, error) {
	if !strings.HasPrefix(content, "---") {
		return FrontMatter{}, content, fmt.Errorf("❌ Front matter not found")
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return FrontMatter{}, content, fmt.Errorf("❌ Invalid front matter format")
	}

	frontMatterStr := strings.TrimSpace(parts[1])
	body := strings.TrimSpace(parts[2])

	// Parse YAML
	var frontMatter FrontMatter
	err := yaml.Unmarshal([]byte(frontMatterStr), &frontMatter)
	if err != nil {
		return FrontMatter{}, content, fmt.Errorf("❌ Failed to parse front matter: %w", err)
	}

	return frontMatter, body, nil
}

// Update front matter in note content
func UpdateFrontMatter(frontMatter *FrontMatter, body string) string {
	// Convert to YAML
	frontMatterBytes, err := yaml.Marshal(frontMatter)
	if err != nil {
		log.Printf("❌ Failed to convert front matter to YAML: %v", err)
		return body
	}

	// Preserve `---` and merge YAML with body
	return fmt.Sprintf("---\n%s---\n\n%s", string(frontMatterBytes), body)
}

// Save updated JSON to `zettel.json`
func SaveUpdatedJson(zettels []Zettel, config *Config) error {
	updatedJson, err := json.MarshalIndent(zettels, "", "  ")
	if err != nil {
		return fmt.Errorf("❌ Failed to convert to JSON: %w", err)
	}

	err = os.WriteFile(config.ZettelJson, updatedJson, 0644)
	if err != nil {
		return fmt.Errorf("❌ Failed to write JSON file: %w", err)
	}

	log.Printf("✅ Successfully updated JSON file: %s", config.ZettelJson)
	return nil
}
