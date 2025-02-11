/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

func backupNote(notePath string, backupDir string) error {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}
	t := time.Now()

	now := fmt.Sprintf("%d%02d%02dT%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	base := filepath.Base(notePath)
	id := strings.TrimSuffix(base, filepath.Ext(base))

	backupFilename := fmt.Sprintf("%s_%s.md", id, now)
	backupPath := filepath.Join(backupDir, backupFilename)

	input, err := os.ReadFile(notePath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(backupPath, input, 0644); err != nil {
		return err
	}
	return nil
}

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var editId string
		if len(args) > 0 {
			editId = args[0]
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

		dir := config.NoteDir
		lockFile := filepath.Join(dir, editId+".lock")

		if _, err := os.Stat(lockFile); err == nil {
			base := filepath.Base(lockFile)
			id := strings.TrimSuffix(base, filepath.Ext(base))
			fmt.Printf("[%q.md] is already under editing.:", id)
			os.Exit(1)
		} else {

			zettels, err := internal.LoadJson(*config)
			if err != nil {
				fmt.Println("Error:", err)
			}

			for i := range zettels {
				if editId == zettels[i].ID {
					// zettelPath := filepath.Join(dir, file.Name())

					lockFile := filepath.Join(dir, editId+".lock")
					internal.CreateLockFile(lockFile)
					backupNote(zettels[i].NotePath, config.Backup.BackupDir)
					fmt.Printf("Found %v, and Opening...\n", zettels[i].NotePath)
					time.Sleep(2 * time.Second)
					c := exec.Command(config.Editor, zettels[i].NotePath)
					defer os.Remove(lockFile)
					c.Stdin = os.Stdin
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					err = c.Run()
					if err != nil {
						log.Fatal(err)
					}
					updatedContent, err := os.ReadFile(zettels[i].NotePath)
					if err != nil {
						fmt.Errorf("⚠️ マークダウンの読み込みエラー: %v", err)
					}

					yamlContent, err := internal.ExtractFrontMatter(string(updatedContent))
					if err != nil {
						fmt.Println("Error extracting front matter:", err)
						return
					}

					frontMatter, err := internal.ParseFrontMatter(yamlContent)
					if err != nil {
						fmt.Println("5Error:", err)
						os.Exit(1)
					}

					zettels[i].Title = frontMatter.Title
					zettels[i].NoteType = frontMatter.Type
					zettels[i].Tags = frontMatter.Tags
					zettels[i].UpdatedAt = frontMatter.UpdatedAt

					fmt.Println(zettels[i].Title)
					fmt.Println(zettels[i].NoteType)
					fmt.Println(zettels[i].Tags)
					fmt.Println(zettels[i].UpdatedAt)

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

	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	// editCmd.Flags().StringVar(&editId, "id", "", "Specify note id")
	// editCmd.MarkFlagRequired("id")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// editCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// editCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
