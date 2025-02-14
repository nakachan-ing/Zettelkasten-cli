/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/text"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var taskStatus string    // --status の値（例: "todo", "doing", "done"）
var taskProject string   // --project の値（例: "Zettelkasten_CLI"）
var taskSortField string // --sort の値（例: "created", "updated", "priority"）
var taskTags []string    // --tag の値（複数指定可能、例: ["urgent", "review"])

func createNewTask(taskTitle, projectName string, config internal.Config) (string, internal.Zettel, error) {
	t := time.Now()
	noteId := fmt.Sprintf("%d%02d%02d%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	createdAt := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	var tags []string
	tags = append(tags, (fmt.Sprintf("project:%s", strings.Replace(projectName, " ", "_", -1))))

	// 統一フォーマットでフロントマターを作成
	frontMatter := internal.FrontMatter{
		ID:         fmt.Sprintf("%v", noteId),
		Title:      fmt.Sprintf("%v", taskTitle),
		Type:       "task",
		Tags:       tags,
		TaskStatus: "Not started",
		CreatedAt:  fmt.Sprintf("%v", createdAt),
		UpdatedAt:  fmt.Sprintf("%v", createdAt),
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
		ID:         "",
		NoteID:     fmt.Sprintf("%v", noteId),
		NoteType:   "task",
		Title:      fmt.Sprintf("%v", taskTitle),
		Tags:       tags,
		TaskStatus: "Not started",
		CreatedAt:  fmt.Sprintf("%v", createdAt),
		UpdatedAt:  fmt.Sprintf("%v", createdAt),
		NotePath:   filePath,
	}

	err = internal.InsertZettelToJson(zettel, config)
	if err != nil {
		return "", internal.Zettel{}, fmt.Errorf("⚠️ JSON書き込みエラー: %v", err)
	}

	fmt.Printf("✅ タスク %s を作成しました。\n", filePath)
	return filePath, zettel, nil

}

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
}

var taskAddCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a note to a project",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("task called")
		taskTitle := args[0]
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
		newTaskStr, _, err := createNewTask(taskTitle, projectName, *config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Opening %q (Title: %q)...\n", newTaskStr, taskTitle)

		time.Sleep(2 * time.Second)

		c := exec.Command(config.Editor, newTaskStr)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err = c.Run()
		if err != nil {
			log.Fatal(err)
		}

	},
}

var taskStatusCmd = &cobra.Command{
	Use:   "status [id] [status]",
	Short: "Change task status",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("task called")
		taskId := args[0]
		status := args[1]
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
		tasks, err := internal.LoadJson(*config)
		if err != nil {
			fmt.Println("Error:", err)
		}

		for i := range tasks {
			if taskId == tasks[i].ID {
				task, err := os.ReadFile(tasks[i].NotePath)
				if err != nil {
					fmt.Println("Error:", err)
				}

				frontMatter, body, err := internal.ParseFrontMatter(string(task))
				if err != nil {
					fmt.Println("5Error:", err)
					os.Exit(1)
				}

				frontMatter.TaskStatus = status

				updatedContent := internal.UpdateFrontMatter(&frontMatter, body)

				// ✅ ファイルに書き戻し
				err = os.WriteFile(tasks[i].NotePath, []byte(updatedContent), 0644)
				if err != nil {
					fmt.Println("❌ 書き込みエラー:", err)
					return
				}

				if err != nil {
					fmt.Println("Error:", err)
					return
				}

				tasks[i].TaskStatus = frontMatter.TaskStatus

				// ✅ `tasks.json` を保存
				internal.SaveUpdatedJson(tasks, config)
				break
			}
		}

	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("task called")

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
		tasks, err := internal.LoadJson(*config)
		if err != nil {
			fmt.Println("Error:", err)
		}

		filteredTasks := []table.Row{}

		for _, task := range tasks {

			// delete フィルター
			if trash {
				if !task.Deleted {
					continue
				}
				// archive フィルター
			} else if archive {
				if !task.Archived {
					continue
				}
			} else {
				if task.Deleted {
					continue
				}

				// // --type フィルター
				// typeSet := make(map[string]bool)
				// for _, listType := range listTypes {
				// 	typeSet[strings.ToLower(listType)] = true
				// }

				// --tag フィルター
				// tagSet := make(map[string]bool)
				// for _, tag := range noteTags {
				// 	tagSet[strings.ToLower(tag)] = true
				// }

				// // --type に指定があり、かつノートのタイプがマッチしないならスキップ
				// if len(typeSet) > 0 && !typeSet[strings.ToLower(task.NoteType)] {
				// 	continue
				// }

				if task.NoteType != "task" {
					continue
				}

				// --status フィルター: 指定されている場合、task.Status と比較
				if taskStatus != "" && strings.ToLower(task.TaskStatus) != strings.ToLower(taskStatus) {
					continue
				}

				// --project フィルター: 指定されている場合、task.Tags 内の "project:" タグと比較する
				if taskProject != "" {
					matchProject := false
					normalizedProject := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(taskProject), " ", "_"))
					for _, tag := range task.Tags {
						normalizedTag := strings.ToLower(strings.TrimSpace(tag))
						if strings.HasPrefix(normalizedTag, "project:") {
							// "project:" の後ろの文字列を取得し、スペースをアンダースコアに置換する
							projectName := strings.TrimSpace(normalizedTag[len("project:"):])
							projectName = strings.ReplaceAll(projectName, " ", "_")
							if projectName == normalizedProject {
								matchProject = true
								break
							}
						}
					}
					if !matchProject {
						continue
					}
				}

				// --tag フィルター処理
				if len(taskTags) > 0 {
					match := false
					for _, filterTag := range taskTags {
						for _, t := range task.Tags {
							normalizedTaskTag := strings.ToLower(strings.TrimSpace(t))
							normalizedFilterTag := strings.ToLower(strings.TrimSpace(filterTag))
							if strings.Contains(normalizedTaskTag, normalizedFilterTag) {
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
			filteredTasks = append(filteredTasks, table.Row{
				task.ID, task.Title, task.TaskStatus, task.Tags,
				task.CreatedAt, task.UpdatedAt,
			})
		}
		// ページネーションの処理
		if len(filteredTasks) == 0 {
			fmt.Println("No matching notes found.")
			return
		}

		reader := bufio.NewReader(os.Stdin)
		page := 0

		fmt.Println(strings.Repeat("=", 30))
		fmt.Printf("Zettelkasten: %v tasks shown\n", len(filteredTasks))
		fmt.Println(strings.Repeat("=", 30))
		for {
			// ページ範囲を決定
			start := page * pageSize
			end := start + pageSize
			if end > len(filteredTasks) {
				end = len(filteredTasks)
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
				text.FgGreen.Sprintf("Status"), text.FgGreen.Sprintf("Tags"),
				text.FgGreen.Sprintf("Created"), text.FgGreen.Sprintf("Updated"),
			})
			// データを追加（Type によって色を変更）
			for _, row := range filteredTasks[start:end] {
				status := row[2].(string) // Type 列の値
				typeColored := noteType   // 初期値はそのまま

				// Type の値に応じて色を変更
				switch status {
				case "Not started":
					typeColored = status // デフォルト
				case "In progress":
					typeColored = text.FgHiBlue.Sprintf(status) // 明るい青
				case "Waiting":
					typeColored = text.FgHiYellow.Sprintf(status) // 明るい黄色
				case "Done":
					typeColored = text.FgHiMagenta.Sprintf(status) // 明るいマゼンタ
				case "On hold":
					typeColored = text.FgHiGreen.Sprintf(status) // 明るい緑
				}

				// 色付きの Type を適用して行を追加
				t.AppendRow(table.Row{
					row[0], row[1], typeColored, row[3], row[4], row[5],
				})
			}
			t.Render()

			// 最後のページなら終了
			if end >= len(filteredTasks) {
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
	taskCmd.AddCommand(taskAddCmd)
	taskCmd.AddCommand(taskStatusCmd)
	taskCmd.AddCommand(taskListCmd)
	rootCmd.AddCommand(taskCmd)

	taskListCmd.Flags().StringVar(&taskStatus, "status", "", "Specify note type")
	taskListCmd.Flags().StringVar(&taskProject, "project", "", "Specify tags")
	taskListCmd.Flags().StringVar(&taskSortField, "sort", "", "Optional")
	taskListCmd.Flags().StringSliceVar(&taskTags, "tag", []string{}, "Optional")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// taskCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// taskCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
