/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

func AddLinkToFrontMatter(frontMatter *internal.FrontMatter, newLink string) *internal.FrontMatter {
	if frontMatter.Links == nil {
		// links フィールドがない場合、新しく作成
		frontMatter.Links = []string{newLink}
	} else {
		// すでに links フィールドがある場合、重複を防ぎつつ追加
		for _, link := range frontMatter.Links {
			if link == newLink {
				// すでに存在するリンクなら何もしない
				return frontMatter
			}
		}
		frontMatter.Links = append(frontMatter.Links, newLink)
	}

	return frontMatter
}

// linkCmd represents the link command
var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("link called")
		var sourceId string
		var destinationId string
		if len(args) > 0 {
			sourceId = args[0]
			destinationId = args[1]
		} else {
			fmt.Println("❌ タイトルを指定してください")
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
			if sourceId == zettels[i].ID {
				backupNote(zettels[i].NotePath, config.Backup.BackupDir)
				for ii := range zettels {
					if destinationId == zettels[ii].ID {
						sourceNoteContent, err := os.ReadFile(zettels[i].NotePath)
						if err != nil {
							fmt.Println("Error:", err)
						}
						frontMatter, body, err := internal.ParseFrontMatter(string(sourceNoteContent))
						if err != nil {
							fmt.Println("5Error:", err)
							os.Exit(1)
						}

						updatedFrontMatter := AddLinkToFrontMatter(&frontMatter, zettels[ii].NoteID)
						updatedContent := internal.UpdateFrontMatter(updatedFrontMatter, body)
						err = os.WriteFile(zettels[i].NotePath, []byte(updatedContent), 0644)
						if err != nil {
							fmt.Println("5Error:", err)
							os.Exit(1)
						}

						zettels[i].Links = frontMatter.Links
						// JSON を更新
						updatedJson, err := json.MarshalIndent(zettels, "", "  ")
						if err != nil {
							fmt.Errorf("⚠️ JSON の変換エラー: %v", err)
						}

						// `zettel.json` に書き込み
						err = os.WriteFile(config.ZettelJson, updatedJson, 0644)
						if err != nil {
							fmt.Errorf("⚠️ JSON 書き込みエラー: %v", err)
						}

						fmt.Println("✅ JSON 更新完了:", config.ZettelJson)
						break
					}
				}

			}
		}

	},
}

func init() {
	rootCmd.AddCommand(linkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// linkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// linkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
