/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zk",
	Short: "A CLI-based Zettelkasten note-taking system",
	Long: `zk is a command-line tool for managing notes, tasks, and projects 
based on the Zettelkasten method. It provides an efficient way to create, 
organize, and search notes, helping you build a structured knowledge system.

Features:
  - Create and manage notes (new, edit, list, delete, archive, search)
  - Organize tasks within notes (task new, task list, task status)
  - Associate notes with projects (project new, project add)

Examples:
  # Create a new note
  zk new "My First Note"

  # List all notes
  zk list

  # Search for a note
  zk search "Golang"

  # Add a task to a note
  zk task add 123 "Learn Kubernetes"

  # Change task status
  zk task status 123 Done
`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
