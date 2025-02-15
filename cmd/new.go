package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var noteType string
var tags []string

var validTypes = map[string]bool{
	"fleeting":   true,
	"literature": true,
	"permanent":  true,
	"index":      true,
	"structure":  true,
}

func validateNoteType(noteType string) error {
	if !validTypes[noteType] {
		return fmt.Errorf("invalid note type: must be 'fleeting', 'literature', 'permanent', 'index' or 'structure'")
	}
	return nil
}

func createNewNote(title, noteType string, tags []string, config internal.Config) (string, internal.Zettel, error) {
	t := time.Now()
	noteId := fmt.Sprintf("%d%02d%02d%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	createdAt := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	// Create front matter
	frontMatter := internal.FrontMatter{
		ID:        noteId,
		Title:     title,
		Type:      noteType,
		Tags:      tags,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	// Convert to YAML format
	frontMatterBytes, err := yaml.Marshal(frontMatter)
	if err != nil {
		return "", internal.Zettel{}, fmt.Errorf("failed to convert to YAML: %w", err)
	}

	// Create Markdown content
	content := fmt.Sprintf("---\n%s---\n\n## %s", string(frontMatterBytes), frontMatter.Title)

	// Write to file
	filePath := fmt.Sprintf("%s/%s.md", config.NoteDir, noteId)
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", internal.Zettel{}, fmt.Errorf("failed to create note file (%s): %w", filePath, err)
	}

	// Write to JSON file
	zettel := internal.Zettel{
		ID:        "",
		NoteID:    noteId,
		NoteType:  noteType,
		Title:     title,
		Tags:      tags,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
		NotePath:  filePath,
	}

	err = internal.InsertZettelToJson(zettel, config)
	if err != nil {
		return "", internal.Zettel{}, fmt.Errorf("failed to write to JSON file: %w", err)
	}

	fmt.Printf("✅ Note %s has been created successfully.\n", filePath)
	return filePath, zettel, nil
}

// `newCmd` represents the new command
var newCmd = &cobra.Command{
	Use:     "new [title]",
	Short:   "Create a new note",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"n"},
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]

		if err := validateNoteType(noteType); err != nil {
			log.Printf("❌ Error: %v", err)
			os.Exit(1)
		}

		config, err := internal.LoadConfig()
		if err != nil {
			log.Printf("❌ Error loading config: %v", err)
			os.Exit(1)
		}

		// Perform cleanup tasks
		if err := internal.CleanupBackups(config.Backup.BackupDir, time.Duration(config.Backup.Retention)*24*time.Hour); err != nil {
			log.Printf("⚠️ Backup cleanup failed: %v", err)
		}
		if err := internal.CleanupTrash(config.Trash.TrashDir, time.Duration(config.Trash.Retention)*24*time.Hour); err != nil {
			log.Printf("⚠️ Trash cleanup failed: %v", err)
		}

		// Create a new note
		newZettelStr, newZettel, err := createNewNote(title, noteType, tags, *config)
		if err != nil {
			log.Printf("❌ Failed to create note: %v", err)
			os.Exit(1)
		}

		fmt.Printf("Opening %q (Title: %q)...\n", newZettelStr, title)

		time.Sleep(2 * time.Second)

		// Open the editor
		c := exec.Command(config.Editor, newZettelStr)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err = c.Run()
		if err != nil {
			log.Printf("❌ Failed to open editor: %v", err)
			os.Exit(1)
		}

		// Load JSON data
		zettels, err := internal.LoadJson(*config)
		if err != nil {
			log.Printf("❌ Error loading notes from JSON: %v", err)
			os.Exit(1)
		}

		found := false
		for i := range zettels {
			if newZettel.NoteID == zettels[i].NoteID {
				found = true

				// Read updated note content
				updatedContent, err := os.ReadFile(zettels[i].NotePath)
				if err != nil {
					log.Printf("❌ Failed to read updated note file: %v", err)
					os.Exit(1)
				}

				// Parse front matter
				frontMatter, _, err := internal.ParseFrontMatter(string(updatedContent))
				if err != nil {
					log.Printf("❌ Error parsing front matter: %v", err)
					os.Exit(1)
				}

				// Update note metadata
				zettels[i].Title = frontMatter.Title
				zettels[i].NoteType = frontMatter.Type
				zettels[i].Tags = frontMatter.Tags
				zettels[i].Links = frontMatter.Links
				zettels[i].TaskStatus = frontMatter.TaskStatus
				zettels[i].UpdatedAt = frontMatter.UpdatedAt

				// Convert to JSON
				updatedJson, err := json.MarshalIndent(zettels, "", "  ")
				if err != nil {
					log.Printf("❌ Failed to convert updated notes to JSON: %v", err)
					os.Exit(1)
				}

				// Write back to `zettel.json`
				if err := os.WriteFile(config.ZettelJson, updatedJson, 0644); err != nil {
					log.Printf("❌ Failed to write updated notes to JSON file: %v", err)
					os.Exit(1)
				}

				fmt.Println("✅ Note metadata updated successfully:", config.ZettelJson)
				break
			}
		}

		if !found {
			log.Printf("❌ Note with ID %s not found", newZettel.NoteID)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVarP(&noteType, "type", "t", "fleeting", "Specify new note type")
	newCmd.Flags().StringSliceVar(&tags, "tag", []string{}, "Specify tags")
}
