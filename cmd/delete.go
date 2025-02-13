/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

// `deleted:` フィールドを更新する
func updateDeletedToFrontMatter(frontMatter *internal.FrontMatter, flag bool) *internal.FrontMatter {
	if !frontMatter.Deleted {
		frontMatter.Deleted = true
		fmt.Println(frontMatter.Deleted)
	}
	return frontMatter
}

// ✍️ **Markdown ファイルを更新**
func updateDeletedMarkdownFile(note internal.Note, flag bool, config *internal.Config) {
	content, err := os.ReadFile(note.Content)
	if err != nil {
		fmt.Println("❌ Error reading note:", err)
		return
	}

	frontMatter, body, err := internal.ParseFrontMatter(string(content))
	if err != nil {
		fmt.Println("❌ Error parsing front matter:", err)
		return
	}

	updatedFrontMatter := updateDeletedToFrontMatter(&frontMatter, flag)
	updatedContent := internal.UpdateFrontMatter(updatedFrontMatter, body)

	err = os.WriteFile(note.Content, []byte(updatedContent), 0644)
	if err != nil {
		fmt.Println("❌ Error updating note:", err)
	}
}

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete called")

		var deleteId string
		if len(args) > 0 {
			deleteId = args[0]
		} else {
			fmt.Println("❌ IDを指定してください")
			os.Exit(1)
		}
		config, err := internal.LoadConfig()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		retention := time.Duration(config.Backup.Retention) * 24 * time.Hour

		err = internal.CleanupBackups(config.Backup.BackupDir, retention)
		if err != nil {
			fmt.Printf("Backup cleanup failed: %v\n", err)
		}

		zettels, err := internal.LoadJson(*config)
		if err != nil {
			fmt.Println("Error:", err)
		}

		for i := range zettels {
			if deleteId == zettels[i].ID {
				flag := true
				note, err := os.ReadFile(zettels[i].NotePath)
				if err != nil {
					fmt.Println("Error:", err)
				}

				frontMatter, body, err := internal.ParseFrontMatter(string(note))
				if err != nil {
					fmt.Println("5Error:", err)
					os.Exit(1)
				}

				updatedFrontMatter := updateDeletedToFrontMatter(&frontMatter, flag)
				updatedContent := internal.UpdateFrontMatter(updatedFrontMatter, body)

				// ✅ ファイルに書き戻し
				err = os.WriteFile(zettels[i].NotePath, []byte(updatedContent), 0644)
				if err != nil {
					fmt.Println("❌ 書き込みエラー:", err)
					return
				}

				if err != nil {
					fmt.Println("Error:", err)
					return
				}

				originalPath := zettels[i].NotePath
				deletedPath := filepath.Join(config.Trash.TrashDir, zettels[i].NoteID+".md")

				err = os.Rename(originalPath, deletedPath)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}

				zettels[i].NotePath = deletedPath
				zettels[i].Deleted = flag
				// ✅ `zettels.json` を保存
				internal.SaveUpdatedJson(zettels, config)
				break
			}

		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
