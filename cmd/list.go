/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list called")

		config, err := internal.LoadConfig()
		if err != nil {
			fmt.Println("1Error:", err)
			return
		}

		retention := time.Duration(config.Backup.Retention) * 24 * time.Hour

		err = internal.CleanupBackups(config.Backup.BackupDir, retention)
		if err != nil {
			fmt.Printf("2Backup cleanup failed: %v\n", err)
		}

		dir := config.NoteDir
		files, err := os.ReadDir(dir)
		if err != nil {
			fmt.Println("3Error:", err)
			return
		}

		pattern := regexp.MustCompile(`^\d{14}\.md$`)

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleRounded) // 罫線を柔らかいスタイルに
		t.Style().Options.SeparateRows = false

		t.AppendHeader(table.Row{"ID", "Title", "Type", "Tags", "Created", "Updated", "Project", "Links"})

		for _, file := range files {
			if !file.IsDir() && pattern.MatchString(file.Name()) {
				zettel, err := os.ReadFile(filepath.Join(dir, file.Name()))
				if err != nil {
					fmt.Println("4Error:", err)
				}
				frontMatter, err := internal.ParseFrontMatter(string(zettel))
				if err != nil {
					fmt.Println("5Error:", err)
					os.Exit(1)
				}

				t.AppendRows([]table.Row{
					{frontMatter.ID, frontMatter.Title, frontMatter.Type, frontMatter.Tags, frontMatter.CreatedAt, frontMatter.UpdatedAt, "-", "-"},
				})
			}

		}
		t.Render()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
