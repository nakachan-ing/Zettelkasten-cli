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

var noteTitle string
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

		t := time.Now()
		noteId := fmt.Sprintf("%d%02d%02d%02d%02d%02d",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())
		createdAt := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())

		frontMatter := fmt.Sprintf(`---
id: %q
title: %q
type: %q
tags: %q
created_at: %q
updated_at:
---

## %q`, noteId, noteTitle, noteType, tags, createdAt, noteTitle)

		newZettel := filepath.Join(config.NoteDirectory, noteId+".md")
		err = os.WriteFile(newZettel, []byte(frontMatter), 0666)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Opening %q (Title: %q)...\n", newZettel, noteTitle)

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

	newCmd.Flags().StringVar(&noteTitle, "title", "No title", "Specify new note title")
	newCmd.Flags().StringVar(&noteType, "type", "fleeting", "Specify new note type")
	newCmd.Flags().StringSliceVar(&tags, "tag", []string{}, "Specify tags")

}
