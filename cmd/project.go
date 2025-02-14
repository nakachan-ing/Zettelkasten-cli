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
	"strings"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func createNewProject(projectName string, tags []string, config internal.Config) (string, internal.Zettel, error) {
	tagName := strings.Replace(projectName, " ", "_", -1)
	fmt.Println(tagName)

	tags = append(tags, fmt.Sprintf("project:%v", tagName))

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
		Title:     fmt.Sprintf("%v", projectName),
		Type:      "project",
		Tags:      tags,
		CreatedAt: fmt.Sprintf("%v", createdAt),
		UpdatedAt: fmt.Sprintf("%v", createdAt),
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
		NoteType:  "project",
		Title:     fmt.Sprintf("%v", projectName),
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

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
}

// `zk project new [title]` コマンド
var projectNewCmd = &cobra.Command{
	Use:   "new [title]",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var projectName string
		if len(args) > 0 {
			projectName = args[0]
		} else {
			fmt.Println("❌ タイトルを指定してください")
			os.Exit(1)
		}

		config, err := internal.LoadConfig()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		err = internal.CleanupBackups(config.Backup.BackupDir, time.Duration(config.Backup.Retention)*24*time.Hour)
		if err != nil {
			fmt.Printf("Backup cleanup failed: %v\n", err)
		}
		err = internal.CleanupTrash(config.Trash.TrashDir, time.Duration(config.Trash.Retention)*24*time.Hour)
		if err != nil {
			fmt.Printf("Trash cleanup failed: %v\n", err)
		}

		newProjectStr, _, err := createNewProject(projectName, tags, *config)
		if err != nil {
			log.Fatal(err)
		}

		// fmt.Println(newZettel)
		fmt.Printf("Opening %q (Title: %q)...\n", newProjectStr, projectName)

		time.Sleep(2 * time.Second)

		c := exec.Command(config.Editor, newProjectStr)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err = c.Run()
		if err != nil {
			log.Fatal(err)
		}
	},
}

// `zk project add [noteid] [project]` コマンド
var projectAddCmd = &cobra.Command{
	Use:   "add [noteid] [project]",
	Short: "Add a note to a project",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		noteID := args[0]
		projectName := args[1]
		config, err := internal.LoadConfig()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		err = internal.CleanupBackups(config.Backup.BackupDir, time.Duration(config.Backup.Retention)*24*time.Hour)
		if err != nil {
			fmt.Printf("Backup cleanup failed: %v\n", err)
		}
		err = internal.CleanupTrash(config.Trash.TrashDir, time.Duration(config.Trash.Retention)*24*time.Hour)
		if err != nil {
			fmt.Printf("Trash cleanup failed: %v\n", err)
		}

		zettels, err := internal.LoadJson(*config)
		if err != nil {
			fmt.Println("Error:", err)
		}

		for i := range zettels {
			if noteID == zettels[i].ID {
				noteByte, err := os.ReadFile(zettels[i].NotePath)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}

				// マークダウンのフロントマター部分にproject:title をtagsに追加して更新したい
				frontMatter, body, err := internal.ParseFrontMatter(string(noteByte))
				if err != nil {
					fmt.Println("5Error:", err)
					os.Exit(1)
				}

				projectTag := fmt.Sprintf("project:%s", strings.Replace(projectName, " ", "_", -1))
				found := false

				for _, tag := range frontMatter.Tags {
					if tag == projectTag {
						found = true
						break
					}
				}

				if !found {
					frontMatter.Tags = append(frontMatter.Tags, projectTag)
				}

				updatedMarkdown := internal.UpdateFrontMatter(&frontMatter, body)

				// ノートを上書き保存
				err = os.WriteFile(zettels[i].NotePath, []byte(updatedMarkdown), 0644)
				if err != nil {
					fmt.Println("Error writing updated note:", err)
					return
				}
				// JSON データを更新
				zettels[i].Tags = frontMatter.Tags
				updatedJson, err := json.MarshalIndent(zettels, "", "  ")
				if err != nil {
					fmt.Println("Error converting JSON:", err)
					return
				}

				// `zettel.json` に書き込み
				err = os.WriteFile(config.ZettelJson, updatedJson, 0644)
				if err != nil {
					fmt.Println("Error writing JSON:", err)
					return
				}

				fmt.Println("✅ ノートと JSON の更新完了:", zettels[i].NotePath)
				break

			}

		}

	},
}

func init() {
	projectCmd.AddCommand(projectNewCmd)
	projectCmd.AddCommand(projectAddCmd)
	rootCmd.AddCommand(projectCmd)
}
