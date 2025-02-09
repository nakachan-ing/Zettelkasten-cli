/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

// var noteId string
var meta bool

func extractBody(content string) (string, error) {
	re := regexp.MustCompile(`(?s)^---\n(.*?)\n---\n`)
	body := re.ReplaceAllString(content, "") // フロントマター部分を削除
	body = strings.TrimSpace(body)           // 余分な空白や改行を除去
	return body, nil
}

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var noteId string
		if len(args) > 0 {
			noteId = args[0]
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
		target := noteId + ".md"
		files, err := os.ReadDir(dir)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		for _, file := range files {
			if file.Name() == target {
				zettelPath := dir + "/" + file.Name()

				zettel, err := os.ReadFile(zettelPath)
				if err != nil {
					fmt.Println("Error:", err)
				}

				titleStyle := color.New(color.FgCyan, color.Bold).SprintFunc()
				frontMatterStyle := color.New(color.FgHiGreen).SprintFunc()

				yamlContent, err := internal.ExtractFrontMatter(string(zettel))
				if err != nil {
					fmt.Println("Error extracting front matter:", err)
					return
				}

				frontMatter, err := internal.ParseFrontMatter(yamlContent)
				if err != nil {
					fmt.Println("5Error:", err)
					os.Exit(1)
				}

				body, err := extractBody(string(zettel))
				if err != nil {
					fmt.Println("Error:", err)
				}

				fmt.Printf("[%v] %v\n", titleStyle(frontMatter.ID), titleStyle(frontMatter.Title))
				fmt.Println(strings.Repeat("-", 50))
				fmt.Printf("type: %v\n", frontMatterStyle(frontMatter.Type))
				fmt.Printf("tags: %v\n", frontMatterStyle(frontMatter.Tags))
				fmt.Printf("created_at: %v\n", frontMatterStyle(frontMatter.CreatedAt))
				fmt.Printf("updated_at: %v\n", frontMatterStyle(frontMatter.UpdatedAt))
				if !meta {
					renderedContent, _ := glamour.Render(body, "dark")
					fmt.Println(renderedContent)
				} else {
					fmt.Printf("\n")
				}
				break
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(showCmd)

	// showCmd.Flags().StringVar(&noteId, "id", "", "Specify note id")
	showCmd.MarkFlagRequired("id")
	showCmd.Flags().BoolVar(&meta, "meta", false, "Optional")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// showCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// showCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
