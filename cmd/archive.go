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

// `archived:` フィールドを更新する
func updateArchivedToFrontMatter(frontMatter *internal.FrontMatter, flag bool) *internal.FrontMatter {
	if !frontMatter.Deleted {
		frontMatter.Deleted = true
		fmt.Println(frontMatter.Deleted)
	}
	return frontMatter
}

// archiveCmd represents the archive command
var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("archive called")
		var archiveId string
		if len(args) > 0 {
			archiveId = args[0]
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

		retention = time.Duration(config.Trash.Retention) * 24 * time.Hour

		err = internal.CleanupTrash(config.Trash.TrashDir, retention)
		if err != nil {
			fmt.Printf("Trash cleanup failed: %v\n", err)
		}

		zettels, err := internal.LoadJson(*config)
		if err != nil {
			fmt.Println("Error:", err)
		}

		for i := range zettels {
			if archiveId == zettels[i].ID {
				flag := true
				originalPath := zettels[i].NotePath
				archivedPath := filepath.Join(config.ArchiveDir, zettels[i].NoteID+".md")
				note, err := os.ReadFile(zettels[i].NotePath)
				if err != nil {
					fmt.Println("Error:", err)
				}

				frontMatter, body, err := internal.ParseFrontMatter(string(note))
				if err != nil {
					fmt.Println("5Error:", err)
					os.Exit(1)
				}

				updatedFrontMatter := updateArchivedToFrontMatter(&frontMatter, flag)
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

				err = os.Rename(originalPath, archivedPath)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}

				zettels[i].NotePath = archivedPath
				zettels[i].Archived = flag
				// ✅ `zettels.json` を保存
				internal.SaveUpdatedJson(zettels, config)
				break
			}

		}

	},
}

func init() {
	rootCmd.AddCommand(archiveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// archiveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// archiveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
