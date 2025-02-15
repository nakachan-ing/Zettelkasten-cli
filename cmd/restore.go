package cmd

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

// Update `deleted:` and `archived:` fields in front matter
func updateRestoredToFrontMatter(frontMatter *internal.FrontMatter) *internal.FrontMatter {
	frontMatter.Deleted = false
	frontMatter.Archived = false
	return frontMatter
}

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:     "restore [noteID]",
	Short:   "Restore a note from archive or trash",
	Aliases: []string{"rs"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		restoreId := args[0]

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
			if restoreId == zettels[i].ID {
				found = true

				originalPath := zettels[i].NotePath
				restoredPath := filepath.Join(config.NoteDir, zettels[i].NoteID+".md")

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

				// Update `deleted:` and `archived:` fields
				updatedFrontMatter := updateRestoredToFrontMatter(&frontMatter)
				updatedContent := internal.UpdateFrontMatter(updatedFrontMatter, body)

				// Write back to file
				err = os.WriteFile(zettels[i].NotePath, []byte(updatedContent), 0644)
				if err != nil {
					log.Printf("❌ Error writing updated note file: %v", err)
					return
				}

				// Move note back to active notes directory
				err = os.Rename(originalPath, restoredPath)
				if err != nil {
					log.Printf("❌ Error restoring note: %v", err)
					return
				}

				zettels[i].NotePath = restoredPath
				zettels[i].Deleted = false
				zettels[i].Archived = false

				// Save updated JSON
				err = internal.SaveUpdatedJson(zettels, config)
				if err != nil {
					log.Printf("❌ Error updating JSON file: %v", err)
					return
				}

				log.Printf("✅ Note %s restored: %s", zettels[i].ID, restoredPath)
				break
			}
		}

		if !found {
			log.Printf("⚠️ Note with ID %s not found", restoreId)
		}
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}
