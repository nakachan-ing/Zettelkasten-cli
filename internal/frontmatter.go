package internal

import (
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

type FrontMatter struct {
	ID        string   `yaml:"id"`
	Title     string   `yaml:"title"`
	Type      string   `yaml:"type"`
	Tags      []string `yaml:"tags"`
	CreatedAt string   `yaml:"created_at"`
	UpdatedAt string   `yaml:"updated_at"`
}

func ExtractFrontMatter(content string) (string, error) {
	re := regexp.MustCompile(`(?s)^---\n(.*?)\n---\n`)
	matches := re.FindStringSubmatch(content)
	if len(matches) < 2 {
		return "", fmt.Errorf("Front matter not found")
	}
	return matches[1], nil
}

func ParseFrontMatter(yamlContent string) (*FrontMatter, error) {
	var fm FrontMatter
	err := yaml.Unmarshal([]byte(yamlContent), &fm)
	if err != nil {
		return nil, err
	}
	return &fm, nil
}
