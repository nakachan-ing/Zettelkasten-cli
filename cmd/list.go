package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/text"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

var listTypes []string
var noteTags []string
var trash bool
var archive bool
var pageSize int

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List notes",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
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

		filteredNotes := []table.Row{}

		for _, zettel := range zettels {
			// Apply filters
			if trash {
				if !zettel.Deleted {
					continue
				}
			} else if archive {
				if !zettel.Archived {
					continue
				}
			} else {
				if zettel.Deleted {
					continue
				}
				if zettel.NoteType == "task" {
					continue
				}

				// Type filter
				typeSet := make(map[string]bool)
				for _, listType := range listTypes {
					typeSet[strings.ToLower(listType)] = true
				}

				// Tag filter
				tagSet := make(map[string]bool)
				for _, tag := range noteTags {
					tagSet[strings.ToLower(tag)] = true
				}

				// If --type is specified but the note type does not match, skip
				if len(typeSet) > 0 && !typeSet[strings.ToLower(zettel.NoteType)] {
					continue
				}

				// If --tag is specified but no matching tags are found, skip
				if len(tagSet) > 0 {
					match := false
					for _, noteTag := range zettel.Tags {
						normalizedNoteTag := strings.ToLower(strings.TrimSpace(noteTag))
						for filterTag := range tagSet {
							if strings.Contains(normalizedNoteTag, filterTag) { // Partial match
								match = true
								break
							}
						}
					}
					if !match {
						continue
					}
				}
			}

			// Append filtered notes
			filteredNotes = append(filteredNotes, table.Row{
				zettel.ID, zettel.Title, zettel.NoteType, zettel.Tags,
				zettel.CreatedAt, zettel.UpdatedAt, len(zettel.Links),
			})
		}

		// Handle case where no notes match
		if len(filteredNotes) == 0 {
			fmt.Println("No matching notes found.")
			return
		}

		reader := bufio.NewReader(os.Stdin)
		page := 0

		fmt.Println(strings.Repeat("=", 30))
		fmt.Printf("Zettelkasten: %v notes shown\n", len(filteredNotes))
		fmt.Println(strings.Repeat("=", 30))

		// `--limit` がない場合は全件表示
		if pageSize == -1 {
			pageSize = len(filteredNotes)
		}

		// ページネーションのループ
		for {
			start := page * pageSize
			end := start + pageSize

			// 範囲チェック
			if start >= len(filteredNotes) {
				fmt.Println("No more notes to display.")
				break
			}
			if end > len(filteredNotes) {
				end = len(filteredNotes)
			}

			// テーブル作成
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.SetStyle(table.StyleDouble)
			t.Style().Options.SeparateRows = false

			t.AppendHeader(table.Row{
				text.FgGreen.Sprintf("ID"), text.FgGreen.Sprintf(text.Bold.Sprintf("Title")),
				text.FgGreen.Sprintf("Type"), text.FgGreen.Sprintf("Tags"),
				text.FgGreen.Sprintf("Created"), text.FgGreen.Sprintf("Updated"),
				text.FgGreen.Sprintf("Links"),
			})

			// フィルタされたノートをテーブルに追加
			for _, row := range filteredNotes[start:end] {
				noteType := row[2].(string)
				typeColored := noteType

				switch noteType {
				case "permanent":
					typeColored = text.FgHiBlue.Sprintf(noteType)
				case "literature":
					typeColored = text.FgHiYellow.Sprintf(noteType)
				case "fleeting":
					typeColored = noteType
				case "index":
					typeColored = text.FgHiMagenta.Sprintf(noteType)
				case "structure":
					typeColored = text.FgHiGreen.Sprintf(noteType)
				}

				t.AppendRow(table.Row{
					row[0], row[1], typeColored, row[3], row[4], row[5], row[6],
				})
			}

			t.Render()

			if pageSize == len(filteredNotes) {
				break
			}

			if end >= len(filteredNotes) {
				break
			}

			fmt.Print("\nPress Enter for the next page (q to quit): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input == "q" {
				break
			}

			page++
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringSliceVarP(&listTypes, "type", "t", []string{}, "Specify note type")
	listCmd.Flags().StringSliceVar(&noteTags, "tag", []string{}, "Specify tags")
	listCmd.Flags().BoolVar(&trash, "trash", false, "Show deleted notes")
	listCmd.Flags().BoolVar(&archive, "archive", false, "Show archived notes")
	listCmd.Flags().IntVar(&pageSize, "limit", -1, "Set the number of notes to display per page (-1 for all)")
}
