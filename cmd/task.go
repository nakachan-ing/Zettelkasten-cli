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

var taskStatus string
var taskProject string
var taskSortField string
var taskTags []string
var taskPageSize int

func createNewTask(taskTitle, projectName string, config internal.Config) (string, internal.Zettel, error) {
	t := time.Now()
	noteId := t.Format("20060102150405")
	createdAt := t.Format("2006-01-02 15:04:05")

	tags := []string{fmt.Sprintf("project:%s", strings.ReplaceAll(projectName, " ", "_"))}

	frontMatter := internal.FrontMatter{
		ID:         noteId,
		Title:      taskTitle,
		Type:       "task",
		Tags:       tags,
		TaskStatus: "Not started",
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	}

	frontMatterBytes, err := yaml.Marshal(frontMatter)
	if err != nil {
		return "", internal.Zettel{}, fmt.Errorf("❌ Failed to convert to YAML: %w", err)
	}

	content := fmt.Sprintf("---\n%s---\n\n## %s", string(frontMatterBytes), frontMatter.Title)

	filePath := fmt.Sprintf("%s/%s.md", config.NoteDir, noteId)
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", internal.Zettel{}, fmt.Errorf("❌ Failed to create file: %w", err)
	}

	zettel := internal.Zettel{
		NoteID:     noteId,
		NoteType:   "task",
		Title:      taskTitle,
		Tags:       tags,
		TaskStatus: "Not started",
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
		NotePath:   filePath,
	}

	err = internal.InsertZettelToJson(zettel, config)
	if err != nil {
		return "", internal.Zettel{}, fmt.Errorf("❌ Failed to write to JSON: %w", err)
	}

	log.Printf("✅ Task created: %s", filePath)
	return filePath, zettel, nil
}

var taskCmd = &cobra.Command{
	Use:     "task",
	Short:   "Manage tasks",
	Aliases: []string{"t"},
}

var taskAddCmd = &cobra.Command{
	Use:     "add [title] [project]",
	Short:   "Add a new task to a project",
	Args:    cobra.ExactArgs(2),
	Aliases: []string{"a"},
	Run: func(cmd *cobra.Command, args []string) {
		taskTitle := args[0]
		projectName := args[1]

		config, err := internal.LoadConfig()
		if err != nil {
			log.Printf("❌ Error loading config: %v", err)
			return
		}

		newTaskStr, _, err := createNewTask(taskTitle, projectName, *config)
		if err != nil {
			log.Printf("❌ Failed to create task: %v", err)
			return
		}

		log.Printf("Opening %q (Title: %q)...", newTaskStr, taskTitle)
		time.Sleep(2 * time.Second)

		c := exec.Command(config.Editor, newTaskStr)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			log.Printf("❌ Failed to open editor: %v", err)
		}
	},
}

var taskStatusCmd = &cobra.Command{
	Use:     "status [id] [status]",
	Short:   "Change task status",
	Aliases: []string{"st"},
	Run: func(cmd *cobra.Command, args []string) {
		taskId := args[0]
		status := args[1]

		config, err := internal.LoadConfig()
		if err != nil {
			log.Printf("❌ Error loading config: %v", err)
			return
		}

		tasks, err := internal.LoadJson(*config)
		if err != nil {
			log.Printf("❌ Error loading JSON: %v", err)
			return
		}

		found := false
		for i := range tasks {
			if taskId == tasks[i].ID {
				found = true
				tasks[i].TaskStatus = status

				if err := internal.SaveUpdatedJson(tasks, config); err != nil {
					log.Printf("❌ Error updating JSON: %v", err)
					return
				}

				log.Printf("✅ Task %s status updated to: %s", taskId, status)
				break
			}
		}

		if !found {
			log.Printf("⚠️ Task with ID %s not found", taskId)
		}
	},
}

var taskListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List tasks",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("🔍 Fetching task list...")

		config, err := internal.LoadConfig()
		if err != nil {
			log.Printf("❌ Error loading config: %v", err)
			return
		}

		tasks, err := internal.LoadJson(*config)
		if err != nil {
			log.Printf("❌ Error loading JSON: %v", err)
			return
		}

		filteredTasks := []table.Row{}

		for _, task := range tasks {
			// Apply delete filter
			if trash && !task.Deleted {
				continue
			}
			// Apply archive filter
			if archive && !task.Archived {
				continue
			}
			if !trash && !archive {
				if task.Deleted || task.NoteType != "task" {
					continue
				}

				// Filter by status
				if taskStatus != "" && strings.ToLower(task.TaskStatus) != strings.ToLower(taskStatus) {
					continue
				}

				// Filter by project
				if taskProject != "" {
					matchProject := false
					normalizedProject := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(taskProject), " ", "_"))
					for _, tag := range task.Tags {
						normalizedTag := strings.ToLower(strings.TrimSpace(tag))
						if strings.HasPrefix(normalizedTag, "project:") {
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

				// Filter by tags
				if len(taskTags) > 0 {
					match := false
					for _, filterTag := range taskTags {
						for _, t := range task.Tags {
							if strings.Contains(strings.ToLower(strings.TrimSpace(t)), strings.ToLower(strings.TrimSpace(filterTag))) {
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

			// Add to filtered list
			filteredTasks = append(filteredTasks, table.Row{
				task.ID, task.Title, task.TaskStatus, task.Tags,
				task.CreatedAt, task.UpdatedAt,
			})
		}

		// No tasks found
		if len(filteredTasks) == 0 {
			log.Println("⚠️ No matching tasks found.")
			return
		}

		// Pagination
		reader := bufio.NewReader(os.Stdin)
		page := 0

		log.Printf("📋 Displaying %d tasks\n", len(filteredTasks))

		// `--limit` がない場合は全件表示
		if pageSize == -1 {
			pageSize = len(filteredTasks)
		}

		for {
			start := page * pageSize
			end := start + pageSize
			// if end > len(filteredTasks) {
			// 	end = len(filteredTasks)
			// }

			// 範囲チェック
			if start >= len(filteredTasks) {
				fmt.Println("No more notes to display.")
				break
			}
			if end > len(filteredTasks) {
				end = len(filteredTasks)
			}

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.SetStyle(table.StyleDouble)
			t.Style().Options.SeparateRows = false

			t.AppendHeader(table.Row{"ID", "Title", "Status", "Tags", "Created", "Updated"})
			for _, row := range filteredTasks[start:end] {
				status := row[2].(string)
				var statusColored string

				switch status {
				case "Not started":
					statusColored = status
				case "In progress":
					statusColored = text.FgHiBlue.Sprintf(status)
				case "Waiting":
					statusColored = text.FgHiYellow.Sprintf(status)
				case "Done":
					statusColored = text.FgHiMagenta.Sprintf(status)
				case "On hold":
					statusColored = text.FgHiGreen.Sprintf(status)
				default:
					statusColored = status
				}

				t.AppendRow(table.Row{row[0], row[1], statusColored, row[3], row[4], row[5]})
			}
			t.Render()

			if pageSize == len(filteredTasks) {
				break
			}

			if end >= len(filteredTasks) {
				break
			}

			// Prompt for next page
			fmt.Print("\nPress Enter for next page (q to quit): ")
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

	taskListCmd.Flags().IntVar(&taskPageSize, "limit", -1, "Set the number of notes to display per page (-1 for all)")
}
