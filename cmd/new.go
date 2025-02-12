/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// var noteTitle string
var noteType string
var tags []string

var validTypes = map[string]bool{
	"fleeting":   true,
	"literature": true,
	"permanent":  true,
}

func validateNoteType(noteType string) error {
	if !validTypes[noteType] {
		return errors.New("invalid note type: must be 'fleeting', 'literature', or 'permanent'")
	}
	return nil
}

func CreateNewNote(title, noteType string, tags []string, config internal.Config) (string, error) {
	t := time.Now()
	noteId := fmt.Sprintf("%d%02d%02d%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	createdAt := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	// 統一フォーマットでフロントマターを作成
	frontMatter := internal.FrontMatter{
		ID:        fmt.Sprintf("%v", noteId),
		Title:     fmt.Sprintf("%v", title),
		Type:      noteType,
		Tags:      tags,
		CreatedAt: fmt.Sprintf("%v", createdAt),
		UpdatedAt: "",
	}

	// YAML 形式に変換
	frontMatterBytes, err := yaml.Marshal(frontMatter)
	if err != nil {
		return "", fmt.Errorf("⚠️ YAML 変換エラー: %v", err)
	}

	// Markdown ファイルの内容を作成
	content := fmt.Sprintf("---\n%s---\n\n", string(frontMatterBytes))

	// ファイルを作成
	filePath := fmt.Sprintf("%s/%s.md", config.NoteDir, noteId)
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("⚠️ ファイル作成エラー: %v", err)
	}

	// JSONファイルに書き込み
	zettel := internal.Zettel{
		ID:        "",
		NoteID:    fmt.Sprintf("%v", noteId),
		NoteType:  noteType,
		Title:     fmt.Sprintf("%v", title),
		Tags:      tags,
		CreatedAt: fmt.Sprintf("%v", createdAt),
		UpdatedAt: "",
		NotePath:  filePath,
	}

	err = internal.InsertZettelToJson(zettel, config)
	if err != nil {
		return "", fmt.Errorf("⚠️ JSON書き込みエラー: %v", err)
	}

	fmt.Printf("✅ ノート %s を作成しました。\n", filePath)
	return filePath, nil

}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var title string
		if len(args) > 0 {
			title = args[0]
		} else {
			fmt.Println("❌ タイトルを指定してください")
			os.Exit(1)
		}

		if err := validateNoteType(noteType); err != nil {
			fmt.Println(noteType)
			fmt.Println("Error:", err)
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

		newZettel, err := CreateNewNote(title, noteType, tags, *config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Opening %q (Title: %q)...\n", newZettel, title)

		time.Sleep(2 * time.Second)

		c := exec.Command(config.Editor, newZettel)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err = c.Run()
		if err != nil {
			log.Fatal(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVarP(&noteType, "type", "t", "fleeting", "Specify new note type")
	newCmd.Flags().StringSliceVar(&tags, "tag", []string{}, "Specify tags")

}
