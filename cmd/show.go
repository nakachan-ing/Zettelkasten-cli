/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

// var noteId string
var meta bool

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

		for _, zettel := range zettels {
			if noteId == zettel.ID {
				note, err := os.ReadFile(zettel.NotePath)
				if err != nil {
					fmt.Println("Error:", err)
				}

				titleStyle := color.New(color.FgCyan, color.Bold).SprintFunc()
				frontMatterStyle := color.New(color.FgHiGreen).SprintFunc()

				frontMatter, body, err := internal.ParseFrontMatter(string(note))
				if err != nil {
					fmt.Println("5Error:", err)
					os.Exit(1)
				}

				fmt.Printf("[%v] %v\n", titleStyle(frontMatter.ID), titleStyle(frontMatter.Title))
				fmt.Println(strings.Repeat("-", 50))
				fmt.Printf("type: %v\n", frontMatterStyle(frontMatter.Type))
				fmt.Printf("tags: %v\n", frontMatterStyle(frontMatter.Tags))
				fmt.Printf("links: %v\n", frontMatterStyle(frontMatter.Links))
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
	showCmd.Flags().BoolVar(&meta, "meta", false, "Optional")
}
