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
	"path/filepath"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
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

		t := time.Now()
		noteId := fmt.Sprintf("%d%02d%02d%02d%02d%02d",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())
		createdAt := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())

		frontMatter := fmt.Sprintf(`---
id: %v
title: %v
type: %v
tags: %v
created_at: %v
updated_at:
---

## %v`, noteId, title, noteType, tags, createdAt, title)

		newZettel := filepath.Join(config.NoteDir, noteId+".md")
		err = os.WriteFile(newZettel, []byte(frontMatter), 0666)
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

	// newCmd.Flags().StringVarP(&noteTitle, "title", "t", "No title", "Specify new note title")
	newCmd.Flags().StringVar(&noteType, "type", "fleeting", "Specify new note type")
	newCmd.Flags().StringSliceVar(&tags, "tag", []string{}, "Specify tags")

}
