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
		t := time.Now()
		noteId := fmt.Sprintf("%d%02d%02d%02d%02d%02d",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())
		// title := "Implement zk new command"
		createdAt := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())

		frontMatter := fmt.Sprintf(`---
id: %q
title: %q
type: %q
tags: %q
created_at: %q
---

## %q`, noteId, noteTitle, noteType, tags, createdAt, noteTitle)

		newZettel := noteId + ".md"
		err := os.WriteFile(newZettel, []byte(frontMatter), 0666)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Opening %q (Title: %q)...\n", newZettel, noteTitle)

		time.Sleep(2 * time.Second)

		c := exec.Command("vim", newZettel)
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
