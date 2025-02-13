/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
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

func createNewNote(title, noteType string, tags []string, config internal.Config) (string, internal.Zettel, error) {
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
		return "", internal.Zettel{}, fmt.Errorf("⚠️ YAML 変換エラー: %v", err)
	}

	// Markdown ファイルの内容を作成
	content := fmt.Sprintf("---\n%s---\n\n## %s", string(frontMatterBytes), frontMatter.Title)

	// ファイルを作成
	filePath := fmt.Sprintf("%s/%s.md", config.NoteDir, noteId)
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", internal.Zettel{}, fmt.Errorf("⚠️ ファイル作成エラー: %v", err)
	}

	// JSONファイルに書き込み
	zettel := internal.Zettel{
		ID:        "",
		NoteID:    fmt.Sprintf("%v", noteId),
		NoteType:  noteType,
		Title:     fmt.Sprintf("%v", title),
		Tags:      tags,
		CreatedAt: fmt.Sprintf("%v", createdAt),
		UpdatedAt: fmt.Sprintf("%v", createdAt),
		NotePath:  filePath,
	}

	err = internal.InsertZettelToJson(zettel, config)
	if err != nil {
		return "", internal.Zettel{}, fmt.Errorf("⚠️ JSON書き込みエラー: %v", err)
	}

	fmt.Printf("✅ ノート %s を作成しました。\n", filePath)
	return filePath, zettel, nil

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
			// fmt.Println(noteType)
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

		retention = time.Duration(config.Trash.Retention) * 24 * time.Hour

		err = internal.CleanupTrash(config.Trash.TrashDir, retention)
		if err != nil {
			fmt.Printf("Trash cleanup failed: %v\n", err)
		}

		newZettelStr, newZettel, err := createNewNote(title, noteType, tags, *config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(newZettel)
		fmt.Printf("Opening %q (Title: %q)...\n", newZettelStr, title)

		time.Sleep(2 * time.Second)

		c := exec.Command(config.Editor, newZettelStr)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err = c.Run()
		if err != nil {
			log.Fatal(err)
		}

		updatedContent, err := os.ReadFile(newZettel.NotePath)
		if err != nil {
			fmt.Errorf("⚠️ マークダウンの読み込みエラー: %v", err)
		}

		frontMatter, _, err := internal.ParseFrontMatter(string(updatedContent))
		if err != nil {
			fmt.Println("5Error:", err)
			os.Exit(1)
		}

		newZettel.Title = frontMatter.Title
		newZettel.NoteType = frontMatter.Type
		newZettel.Tags = frontMatter.Tags
		newZettel.Links = frontMatter.Links
		newZettel.TaskStatus = frontMatter.TaskStatus
		newZettel.UpdatedAt = frontMatter.UpdatedAt

		// JSON を更新
		updatedJson, err := json.MarshalIndent(newZettel, "", "  ")
		if err != nil {
			fmt.Errorf("⚠️ JSON の変換エラー: %v", err)
		}

		// `zettel.json` に書き込み
		err = os.WriteFile(config.ZettelJson, updatedJson, 0644)
		if err != nil {
			fmt.Errorf("⚠️ JSON 書き込みエラー: %v", err)
		}

		fmt.Println("✅ JSON 更新完了:", config.ZettelJson)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVarP(&noteType, "type", "t", "fleeting", "Specify new note type")
	newCmd.Flags().StringSliceVar(&tags, "tag", []string{}, "Specify tags")

}
