/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

var editId string

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		config, err := internal.LoadConfig()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		dir := config.NoteDirectory
		target := editId + ".md"
		files, err := os.ReadDir(dir)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		lockFile, err := os.Stat(dir + "/" + editId + ".lock")
		if err == nil {
			fmt.Printf("%q is already under editing.:", strings.Replace(lockFile.Name(), ".lock", ".md", 1))
			os.Exit(1)
		} else {
			for _, file := range files {
				if file.Name() == target {
					zettelPath := dir + "/" + file.Name()

					lockFile := (dir + "/" + editId + ".lock")
					internal.CreateLockFile(lockFile)

					fmt.Printf("Found %v, and Opening...\n", file.Name())
					time.Sleep(2 * time.Second)
					c := exec.Command(config.Editor, zettelPath)
					defer os.Remove(lockFile)
					c.Stdin = os.Stdin
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					err = c.Run()
					if err != nil {
						log.Fatal(err)
					}
					break
				}
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVar(&editId, "id", "", "Specify note id")
	editCmd.MarkFlagRequired("id")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// editCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// editCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
