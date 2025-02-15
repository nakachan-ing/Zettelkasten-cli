package cmd

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

// Update `deleted:` field in front matter
func updateDeletedToFrontMatter(frontMatter *internal.FrontMatter) *internal.FrontMatter {
	frontMatter.Deleted = true
	return frontMatter
}

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete [id]",
	Short:   "Delete a note",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"rm"},
	Run: func(cmd *cobra.Command, args []string) {
		deleteId := args[0]

		config, err := internal.LoadConfig()
		if err != nil {
			log.Printf("❌ Error loading config: %v", err)
			return
		}

		// Cleanup backups and trash
		if err := internal.CleanupBackups(config.Backup.BackupDir, time.Duration(config.Backup.Retention)*24*time.Hour); err != nil {
			log.Printf("⚠️ Backup cleanup failed: %v", err)
		}
		if err := internal.CleanupTrash(config.Trash.TrashDir, time.Duration(config.Trash.Retention)*24*time.Hour); err != nil {
			log.Printf("⚠️ Trash cleanup failed: %v", err)
		}

		// Load JSON
		zettels, err := internal.LoadJson(*config)
		if err != nil {
			log.Printf("❌ Error loading JSON: %v", err)
			return
		}

		// Search for the note
		found := false
		for i := range zettels {
			if deleteId == zettels[i].ID {
				found = true

				originalPath := zettels[i].NotePath
				deletedPath := filepath.Join(config.Trash.TrashDir, zettels[i].NoteID+".md")

				note, err := os.ReadFile(zettels[i].NotePath)
				if err != nil {
					log.Printf("❌ Error reading note file: %v", err)
					return
				}

				// Parse front matter
				frontMatter, body, err := internal.ParseFrontMatter(string(note))
				if err != nil {
					log.Printf("❌ Error parsing front matter: %v", err)
					return
				}

				// Update `deleted:` field
				updatedFrontMatter := updateDeletedToFrontMatter(&frontMatter)
				updatedContent := internal.UpdateFrontMatter(updatedFrontMatter, body)

				// Write back to file
				err = os.WriteFile(zettels[i].NotePath, []byte(updatedContent), 0644)
				if err != nil {
					log.Printf("❌ Error writing updated note file: %v", err)
					return
				}

				// Move note to trash
				err = os.Rename(originalPath, deletedPath)
				if err != nil {
					log.Printf("❌ Error moving note to trash: %v", err)
					return
				}

				zettels[i].NotePath = deletedPath
				zettels[i].Deleted = true

				// Save updated JSON
				err = internal.SaveUpdatedJson(zettels, config)
				if err != nil {
					log.Printf("❌ Error updating JSON file: %v", err)
					return
				}

				log.Printf("✅ Note %s moved to trash: %s", zettels[i].ID, deletedPath)
				break
			}
		}

		if !found {
			log.Printf("⚠️ Note with ID %s not found", deleteId)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
