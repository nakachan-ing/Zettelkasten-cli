/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

var noteType string

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
		t := time.Now()
		noteId := fmt.Sprintf("%d%02d%02d%02d%02d%02d",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())
		title := "Implement zk new command"

		frontMatter := fmt.Sprintf(`---
id: %q
title: %q
type: %q
tags: 
---

## %q`, noteId, title, noteType, title)

		newZettel := noteId + ".md"
		err := os.WriteFile(newZettel, []byte(frontMatter), 0666)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Opening %q (Title: %q)...", newZettel, title)
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

	newCmd.Flags().StringVar(&noteType, "type", "No title", "Specify new note name")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
