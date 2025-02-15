package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

var meta bool

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:     "show [id]",
	Short:   "Show a note",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"s"},
	Run: func(cmd *cobra.Command, args []string) {
		noteId := args[0]

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

		// Load notes from JSON
		zettels, err := internal.LoadJson(*config)
		if err != nil {
			log.Printf("❌ Error loading notes from JSON: %v", err)
			os.Exit(1)
		}

		// Find and display the requested note
		found := false
		for _, zettel := range zettels {
			if noteId == zettel.ID {
				found = true

				note, err := os.ReadFile(zettel.NotePath)
				if err != nil {
					log.Printf("❌ Error reading note file (%s): %v", zettel.NotePath, err)
					os.Exit(1)
				}

				titleStyle := color.New(color.FgCyan, color.Bold).SprintFunc()
				frontMatterStyle := color.New(color.FgHiGreen).SprintFunc()

				frontMatter, body, err := internal.ParseFrontMatter(string(note))
				if err != nil {
					log.Printf("❌ Error parsing front matter: %v", err)
					os.Exit(1)
				}

				fmt.Printf("[%v] %v\n", titleStyle(frontMatter.ID), titleStyle(frontMatter.Title))
				fmt.Println(strings.Repeat("-", 50))
				fmt.Printf("Type: %v\n", frontMatterStyle(frontMatter.Type))
				fmt.Printf("Tags: %v\n", frontMatterStyle(frontMatter.Tags))
				fmt.Printf("Links: %v\n", frontMatterStyle(frontMatter.Links))
				fmt.Printf("Task status: %v\n", frontMatterStyle(frontMatter.TaskStatus))
				fmt.Printf("Created at: %v\n", frontMatterStyle(frontMatter.CreatedAt))
				fmt.Printf("Updated at: %v\n", frontMatterStyle(frontMatter.UpdatedAt))

				// Render Markdown content unless --meta flag is used
				if !meta {
					renderedContent, err := glamour.Render(body, "dark")
					if err != nil {
						log.Printf("⚠️ Failed to render markdown content: %v", err)
					} else {
						fmt.Println(renderedContent)
					}
				}
				break
			}
		}

		if !found {
			log.Printf("❌ Note with ID %s not found", noteId)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().BoolVar(&meta, "meta", false, "Show only metadata without note content")
}
