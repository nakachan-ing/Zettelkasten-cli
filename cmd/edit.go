/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
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

var editId string

func backupNote(notePath string, backupDir string) error {
	// 1. バックアップディレクトリの存在チェックと作成
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	// 2. 現在のタイムスタンプを取得
	t := time.Now()

	now := fmt.Sprintf("%d%02d%02dT%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	// 3. ノートファイル名からIDを抽出（例としてファイル名がID.mdの場合）
	base := filepath.Base(notePath)
	id := strings.TrimSuffix(base, filepath.Ext(base))

	// 4. バックアップファイル名を作成
	backupFilename := fmt.Sprintf("%s_%s.md", id, now)
	backupPath := filepath.Join(backupDir, backupFilename)

	// 5. ノートファイルの内容をコピー
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

		config, err := internal.LoadConfig()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		dir := config.NoteDirectory
		target := editId + ".md"
		lockFile := filepath.Join(dir, editId+".lock")
		files, err := os.ReadDir(dir)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if _, err := os.Stat(lockFile); err == nil {
			base := filepath.Base(lockFile)
			id := strings.TrimSuffix(base, filepath.Ext(base))
			fmt.Printf("[%q.md] is already under editing.:", id)
			os.Exit(1)
		} else {
			for _, file := range files {
				if file.Name() == target {
					zettelPath := filepath.Join(dir, file.Name())

					lockFile := filepath.Join(dir, editId+".lock")
					internal.CreateLockFile(lockFile)
					backupNote(zettelPath, config.BackupDirectory)
					fmt.Printf("Found %v, and Opening...\n", file.Name())
					time.Sleep(2 * time.Second)
					c := exec.Command(config.Editor, zettelPath)
					defer os.Remove(lockFile)
					c.Stdin = os.Stdin
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					err = c.Run()
					if err != nil {
						log.Fatal(err)
					}
					break
				}
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVar(&editId, "id", "", "Specify note id")
	editCmd.MarkFlagRequired("id")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// editCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// editCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
