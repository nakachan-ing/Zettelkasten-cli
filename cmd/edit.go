package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

func backupNote(notePath string, backupDir string) error {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	t := time.Now()
	now := fmt.Sprintf("%d%02d%02dT%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	base := filepath.Base(notePath)
	id := strings.TrimSuffix(base, filepath.Ext(base))

	backupFilename := fmt.Sprintf("%s_%s.md", id, now)
	backupPath := filepath.Join(backupDir, backupFilename)

	input, err := os.ReadFile(notePath)
	if err != nil {
		return fmt.Errorf("failed to read note file for backup: %w", err)
	}

	if err := os.WriteFile(backupPath, input, 0644); err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	return nil
}

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:     "edit [id]",
	Short:   "Edit a note",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"e"},
	Run: func(cmd *cobra.Command, args []string) {
		editId := args[0]

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

		dir := config.NoteDir
		lockFile := filepath.Join(dir, editId+".lock")

		if _, err := os.Stat(lockFile); err == nil {
			log.Printf("❌ Note [%q.md] is already being edited.", editId)
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
			if editId == zettels[i].ID {
				found = true

				// Create lock file
				lockFile := filepath.Join(dir, editId+".lock")
				if err := internal.CreateLockFile(lockFile); err != nil {
					log.Printf("❌ Failed to create lock file: %v", err)
					os.Exit(1)
				}

				// Backup note before editing
				if err := backupNote(zettels[i].NotePath, config.Backup.BackupDir); err != nil {
					log.Printf("⚠️ Backup failed: %v", err)
				}

				fmt.Printf("Found %v, opening...\n", zettels[i].NotePath)
				time.Sleep(2 * time.Second)

				// Open the note in the editor
				c := exec.Command(config.Editor, zettels[i].NotePath)
				defer os.Remove(lockFile) // Ensure lock file is deleted after editing
				c.Stdin = os.Stdin
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				if err := c.Run(); err != nil {
					log.Printf("❌ Failed to open editor: %v", err)
					os.Exit(1)
				}

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
			log.Printf("❌ Note with ID %s not found", editId)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
