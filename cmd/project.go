package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func createNewProject(projectName string, tags []string, config internal.Config) (string, internal.Zettel, error) {
	tagName := strings.ReplaceAll(projectName, " ", "_")
	tags = append(tags, fmt.Sprintf("project:%v", tagName))

	t := time.Now()
	noteId := fmt.Sprintf("%d%02d%02d%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	createdAt := t.Format("2006-01-02 15:04:05")

	frontMatter := internal.FrontMatter{
		ID:        noteId,
		Title:     projectName,
		Type:      "project",
		Tags:      tags,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	frontMatterBytes, err := yaml.Marshal(frontMatter)
	if err != nil {
		return "", internal.Zettel{}, fmt.Errorf("❌ Failed to convert to YAML: %w", err)
	}

	content := fmt.Sprintf("---\n%s---\n\n## %s", string(frontMatterBytes), frontMatter.Title)

	filePath := fmt.Sprintf("%s/%s.md", config.NoteDir, noteId)
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", internal.Zettel{}, fmt.Errorf("❌ Failed to create file: %w", err)
	}

	zettel := internal.Zettel{
		NoteID:    noteId,
		NoteType:  "project",
		Title:     projectName,
		Tags:      tags,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
		NotePath:  filePath,
	}

	err = internal.InsertZettelToJson(zettel, config)
	if err != nil {
		return "", internal.Zettel{}, fmt.Errorf("❌ Failed to write to JSON: %w", err)
	}

	log.Printf("✅ Note created: %s", filePath)
	return filePath, zettel, nil
}

var projectCmd = &cobra.Command{
	Use:     "project",
	Short:   "Manage projects",
	Aliases: []string{"p"},
}

var projectNewCmd = &cobra.Command{
	Use:     "new [title]",
	Short:   "Create a new project",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"n"},
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]

		config, err := internal.LoadConfig()
		if err != nil {
			log.Printf("❌ Error loading config: %v", err)
			return
		}

		newProjectStr, _, err := createNewProject(projectName, tags, *config)
		if err != nil {
			log.Printf("❌ Failed to create project: %v", err)
			return
		}

		log.Printf("Opening %q (Title: %q)...", newProjectStr, projectName)
		time.Sleep(2 * time.Second)

		c := exec.Command(config.Editor, newProjectStr)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			log.Printf("❌ Failed to open editor: %v", err)
		}
	},
}

var projectAddCmd = &cobra.Command{
	Use:     "add [noteid] [project]",
	Short:   "Add a note to a project",
	Args:    cobra.ExactArgs(2),
	Aliases: []string{"a"},
	Run: func(cmd *cobra.Command, args []string) {
		noteID := args[0]
		projectName := args[1]

		config, err := internal.LoadConfig()
		if err != nil {
			log.Printf("❌ Error loading config: %v", err)
			return
		}

		zettels, err := internal.LoadJson(*config)
		if err != nil {
			log.Printf("❌ Error loading JSON: %v", err)
			return
		}

		for i := range zettels {
			if noteID == zettels[i].ID {
				noteByte, err := os.ReadFile(zettels[i].NotePath)
				if err != nil {
					log.Printf("❌ Error reading note file: %v", err)
					return
				}

				frontMatter, body, err := internal.ParseFrontMatter(string(noteByte))
				if err != nil {
					log.Printf("❌ Error parsing front matter: %v", err)
					return
				}

				projectTag := fmt.Sprintf("project:%s", strings.ReplaceAll(projectName, " ", "_"))
				if !contains(frontMatter.Tags, projectTag) {
					frontMatter.Tags = append(frontMatter.Tags, projectTag)
				}

				updatedMarkdown := internal.UpdateFrontMatter(&frontMatter, body)

				err = os.WriteFile(zettels[i].NotePath, []byte(updatedMarkdown), 0644)
				if err != nil {
					log.Printf("❌ Error writing updated note: %v", err)
					return
				}

				zettels[i].Tags = frontMatter.Tags
				if err := internal.SaveUpdatedJson(zettels, config); err != nil {
					log.Printf("❌ Error updating JSON file: %v", err)
					return
				}

				log.Printf("✅ Note and JSON updated: %s", zettels[i].NotePath)
				break
			}
		}
	},
}

func contains(slice []string, item string) bool {
	for _, val := range slice {
		if val == item {
			return true
		}
	}
	return false
}

func init() {
	projectCmd.AddCommand(projectNewCmd)
	projectCmd.AddCommand(projectAddCmd)
	rootCmd.AddCommand(projectCmd)
}
