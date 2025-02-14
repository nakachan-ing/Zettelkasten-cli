/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
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

const pageSize = 20

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		config, err := internal.LoadConfig()
		if err != nil {
			fmt.Println("1Error:", err)
			return
		}

		retention := time.Duration(config.Backup.Retention) * 24 * time.Hour

		err = internal.CleanupBackups(config.Backup.BackupDir, retention)
		if err != nil {
			fmt.Printf("Backup cleanup failed: %v\n", err)
		}

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		filteredNotes := []table.Row{}

		zettels, err := internal.LoadJson(*config)
		if err != nil {
			fmt.Println("Error:", err)
		}

		for _, zettel := range zettels {

			// delete フィルター
			if trash {
				if !zettel.Deleted {
					continue
				}
				// archive フィルター
			} else if archive {
				if !zettel.Archived {
					continue
				}
			} else {
				if zettel.Deleted {
					continue
				}

				// --type フィルター
				typeSet := make(map[string]bool)
				for _, listType := range listTypes {
					typeSet[strings.ToLower(listType)] = true
				}

				// --tag フィルター
				tagSet := make(map[string]bool)
				for _, tag := range noteTags {
					tagSet[strings.ToLower(tag)] = true
				}

				// --type に指定があり、かつノートのタイプがマッチしないならスキップ
				if len(typeSet) > 0 && !typeSet[strings.ToLower(zettel.NoteType)] {
					continue
				}

				// --tag フィルター処理
				if len(tagSet) > 0 {
					match := false
					for _, noteTag := range zettel.Tags {
						normalizedNoteTag := strings.ToLower(strings.TrimSpace(noteTag))
						for filterTag := range tagSet {
							if strings.Contains(normalizedNoteTag, filterTag) { // 部分一致
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

			// 🔹 `--tag` がない場合でもここに到達するように修正
			filteredNotes = append(filteredNotes, table.Row{
				zettel.ID, zettel.Title, zettel.NoteType, zettel.Tags,
				zettel.CreatedAt, zettel.UpdatedAt, "-", len(zettel.Links),
			})
		}
		// ページネーションの処理
		if len(filteredNotes) == 0 {
			fmt.Println("No matching notes found.")
			return
		}

		reader := bufio.NewReader(os.Stdin)
		page := 0

		fmt.Println(strings.Repeat("=", 30))
		fmt.Printf("Zettelkasten: %v notes shown\n", len(filteredNotes))
		fmt.Println(strings.Repeat("=", 30))
		for {
			// ページ範囲を決定
			start := page * pageSize
			end := start + pageSize
			if end > len(filteredNotes) {
				end = len(filteredNotes)
			}

			// テーブル作成
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			// t.SetStyle(table.StyleRounded)
			t.SetStyle(table.StyleDouble)
			t.Style().Options.SeparateRows = false
			// t.SetStyle(table.StyleColoredBlackOnCyanWhite)

			t.AppendHeader(table.Row{
				text.FgGreen.Sprintf("ID"), text.FgGreen.Sprintf(text.Bold.Sprintf("Title")),
				text.FgGreen.Sprintf("Type"), text.FgGreen.Sprintf("Tags"),
				text.FgGreen.Sprintf("Created"), text.FgGreen.Sprintf("Updated"),
				text.FgGreen.Sprintf("Project"), text.FgGreen.Sprintf("Links"),
			})
			// データを追加（Type によって色を変更）
			for _, row := range filteredNotes[start:end] {
				noteType := row[2].(string) // Type 列の値
				typeColored := noteType     // 初期値はそのまま

				// Type の値に応じて色を変更
				switch noteType {
				case "permanent":
					typeColored = text.FgHiBlue.Sprintf(noteType) // 明るい青
				case "literature":
					typeColored = text.FgHiYellow.Sprintf(noteType) // 明るい黄色
				case "fleeting":
					typeColored = noteType // デフォルト
				case "index":
					typeColored = text.FgHiMagenta.Sprintf(noteType) // 明るいマゼンタ
				case "structure":
					typeColored = text.FgHiGreen.Sprintf(noteType) // 明るい緑
				}

				// 色付きの Type を適用して行を追加
				t.AppendRow(table.Row{
					row[0], row[1], typeColored, row[3], row[4], row[5], row[6], row[7],
				})
			}
			t.Render()

			// 最後のページなら終了
			if end >= len(filteredNotes) {
				break
			}

			// 次のページに進むか確認
			fmt.Print("\nEnterで次のページ (q で終了): ")
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
	listCmd.Flags().BoolVar(&trash, "trash", false, "Optional")
	listCmd.Flags().BoolVar(&archive, "archive", false, "Optional")
}
