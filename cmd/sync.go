package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

func syncZettel(config internal.Config, zettels []internal.Zettel) error {
	notesDir := config.NoteDir
	archiveDir := config.ArchiveDir
	trashDir := config.Trash.TrashDir

	updatedZettels := syncZettelData(zettels, notesDir, archiveDir, trashDir)

	// Update `zettel.json` if there are changes
	if err := saveZettelJson(updatedZettels, config); err != nil {
		return fmt.Errorf("❌ Failed to save zettel.json: %w", err)
	}

	log.Println("✅ Synchronized zettel.json successfully!")
	return nil
}

// Scan directories for Markdown files
func scanDirectory(dir string, sortAscending bool) (map[string]string, error) {
	noteFiles := make(map[string]string)
	var fileInfos []os.FileInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			log.Printf("⚠️ Failed to get file info: %s (%v)", entry.Name(), err)
			continue
		}
		fileInfos = append(fileInfos, info)
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		if sortAscending {
			return fileInfos[i].ModTime().Before(fileInfos[j].ModTime())
		}
		return fileInfos[j].ModTime().Before(fileInfos[i].ModTime())
	})

	for _, info := range fileInfos {
		noteID := info.Name()
		noteFiles[noteID] = filepath.Join(dir, noteID)
	}

	return noteFiles, nil
}

// Synchronize zettel data
func syncZettelData(zettels []internal.Zettel, notesDir, archiveDir, trashDir string) []internal.Zettel {
	updatedZettels := []internal.Zettel{}
	existingNotes := make(map[string]bool)

	noteFiles, _ := scanDirectory(notesDir, true)
	archiveFiles, _ := scanDirectory(archiveDir, true)
	trashFiles, _ := scanDirectory(trashDir, true)

	for _, z := range zettels {
		t := time.Now()
		if path, exists := noteFiles[z.NoteID]; exists {
			z.UpdatedAt = t.Format("2006-01-02 15:04:05")
			z.NotePath = path
			updatedZettels = append(updatedZettels, z)
			existingNotes[z.NoteID] = true
		} else if path, exists := archiveFiles[z.NoteID]; exists {
			z.NotePath = path
			z.Archived = true
			z.Deleted = false
			updatedZettels = append(updatedZettels, z)
			existingNotes[z.NoteID] = true
		} else if path, exists := trashFiles[z.NoteID]; exists {
			z.NotePath = path
			z.Archived = false
			z.Deleted = true
			updatedZettels = append(updatedZettels, z)
			existingNotes[z.NoteID] = true
		} else {
			newPath := filepath.Join(trashDir, z.NoteID)
			if err := os.Rename(z.NotePath, newPath); err != nil {
				log.Printf("❌ Failed to move file to trash: %s (%v)", z.NotePath, err)
				continue
			}
			z.UpdatedAt = t.Format("2006-01-02 15:04:05")
			z.NotePath = newPath
			z.Archived = false
			z.Deleted = true
			updatedZettels = append(updatedZettels, z)
		}
	}

	newID := len(zettels) + 1
	for _, fileSet := range []struct {
		files    map[string]string
		archived bool
		deleted  bool
	}{
		{noteFiles, false, false},
		{archiveFiles, true, false},
		{trashFiles, false, true},
	} {
		for noteID, path := range fileSet.files {
			if _, exists := existingNotes[noteID]; !exists {
				note, err := os.ReadFile(path)
				if err != nil {
					log.Printf("❌ Error reading file: %s (%v)", path, err)
					continue
				}
				frontMatter, _, err := internal.ParseFrontMatter(string(note))
				if err != nil {
					log.Printf("❌ Error parsing front matter: %s (%v)", path, err)
					continue
				}

				newZettel := internal.Zettel{
					ID:         strconv.Itoa(newID),
					NoteID:     frontMatter.ID,
					Title:      frontMatter.Title,
					NoteType:   frontMatter.Type,
					Tags:       frontMatter.Tags,
					TaskStatus: frontMatter.TaskStatus,
					Links:      frontMatter.Links,
					CreatedAt:  frontMatter.CreatedAt,
					UpdatedAt:  frontMatter.UpdatedAt,
					NotePath:   path,
					Archived:   fileSet.archived,
					Deleted:    fileSet.deleted,
				}
				updatedZettels = append(updatedZettels, newZettel)
				newID++
			}
		}
	}

	return updatedZettels
}

// Save `zettel.json`
func saveZettelJson(zettels []internal.Zettel, config internal.Config) error {
	file, err := os.Create(config.ZettelJson)
	if err != nil {
		return fmt.Errorf("❌ Failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(zettels); err != nil {
		return fmt.Errorf("❌ Failed to write JSON file: %w", err)
	}

	return nil
}

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:     "sync",
	Short:   "Synchronize notes with local files",
	Aliases: []string{"sy"},
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("🔄 Syncing notes...")

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

		err = syncZettel(*config, zettels)
		if err != nil {
			log.Printf("❌ Error during sync: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
