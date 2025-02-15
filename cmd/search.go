package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

var searchTitle bool
var searchTypes []string
var searchTags []string
var searchContext int
var interactive bool

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:     "search [query]",
	Short:   "Search notes",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"f"},
	Run: func(cmd *cobra.Command, args []string) {
		var keyword string
		if len(args) > 0 {
			keyword = args[0]
		} else {
			keyword = "" // Ensure it works even without a search keyword
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

		var rgArgs []string
		rgArgs = append(rgArgs, "--json", "--ignore-case", "--only-matching")
		rgArgs = append(rgArgs, "-C", fmt.Sprintf("%d", searchContext)) // Context lines

		// Validate that at least one search criteria is provided
		if keyword == "" && !searchTitle && len(searchTags) == 0 && len(searchTypes) == 0 && !interactive {
			log.Printf("❌ Please specify a search keyword, title, tag, type, or use --interactive mode.")
			os.Exit(1)
		}

		// If only --interactive is used, search all notes
		if interactive && keyword == "" && !searchTitle && len(searchTags) == 0 && len(searchTypes) == 0 {
			rgArgs = append(rgArgs, "-e", ".*") // Search everything
		}

		// Search by title (`--title`)
		if searchTitle {
			if keyword != "" {
				rgArgs = append(rgArgs, "-e", fmt.Sprintf("^title:\\s*.*%s", keyword))
			} else {
				rgArgs = append(rgArgs, "-e", "^title:\\s*") // Search all titles
			}
		}

		// Search by note type (`--type`)
		if len(searchTypes) > 0 {
			for _, t := range searchTypes {
				rgArgs = append(rgArgs, "-e", fmt.Sprintf(`^type:\s*%s`, t))
			}
		}

		// Search by tag (`--tag`)
		if len(searchTags) > 0 {
			rgArgs = append(rgArgs, "--multiline", "--multiline-dotall")
			for _, tag := range searchTags {
				rgArgs = append(rgArgs, "-e", fmt.Sprintf(`^tags:.*%s`, tag))
			}
		}

		// Full-text search
		if keyword != "" && !searchTitle && len(searchTags) == 0 && len(searchTypes) == 0 {
			rgArgs = append(rgArgs, "-e", keyword)
		}

		// Append search directory
		rgArgs = append(rgArgs, config.NoteDir)

		var out bytes.Buffer
		rgCmd := exec.Command("rg", rgArgs...)
		rgCmd.Stdout = &out
		rgCmd.Stderr = &out

		err = rgCmd.Run()
		output := out.String()

		if err != nil && output == "" {
			log.Printf("❌ Search failed: %v", err)
			os.Exit(1)
		}

		// Parse search results
		results, err := internal.ParseRipgrepOutput(output)

		if err != nil {
			log.Printf("❌ Failed to parse ripgrep output: %v", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			log.Println("❌ No matching notes found.")
			os.Exit(1)
		}

		// Interactive mode
		if interactive {
			internal.InteractiveSearch(results)
		} else {
			internal.DisplayResults(results)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().BoolVar(&searchTitle, "title", false, "Search by title")
	searchCmd.Flags().StringSliceVar(&searchTypes, "type", []string{}, "Filter by note type")
	searchCmd.Flags().StringSliceVar(&searchTags, "tag", []string{}, "Filter by tags")
	searchCmd.Flags().IntVar(&searchContext, "context", 0, "Show N lines before and after the search result")
	searchCmd.Flags().BoolVar(&interactive, "interactive", false, "Use interactive search with fzf")
}
