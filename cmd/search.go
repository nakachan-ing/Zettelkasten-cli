/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

var searchTitle bool
var searchTypes []string
var searchTags []string
var searchContext int
var interactive bool

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
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

		retention := time.Duration(config.Backup.Retention) * 24 * time.Hour

		err = internal.CleanupBackups(config.Backup.BackupDir, retention)
		if err != nil {
			fmt.Printf("Backup cleanup failed: %v\n", err)
		}

		var keyword string
		if len(args) > 0 {
			keyword = args[0]
		} else {
			keyword = "" // 検索ワードがない場合でも動作するようにする
		}
		var rgArgs []string
		// ripgrep の基本オプション
		rgArgs = append(rgArgs, "--json")
		rgArgs = append(rgArgs, "--ignore-case") // 大文字・小文字を無視
		rgArgs = append(rgArgs, "--only-matching")
		rgArgs = append(rgArgs, "-C", fmt.Sprintf("%d", searchContext)) // --context の指定

		// 検索条件が指定されていない場合、すべてのノートを対象にする
		if keyword == "" && !searchTitle && len(searchTags) == 0 && len(searchTypes) == 0 && !interactive {
			fmt.Println("❌ 検索ワード、タイトル、タグ、タイプ、または --interactive を指定してください")
			os.Exit(1)
		}

		// --interactive のみ指定された場合、すべてのノートを対象にする
		if interactive && keyword == "" && !searchTitle && len(searchTags) == 0 && len(searchTypes) == 0 {
			rgArgs = append(rgArgs, "-e", ".*") // すべてのノートを対象に全文検索
		}

		// タイトル検索 (`--title`)
		if searchTitle {
			if keyword != "" {
				rgArgs = append(rgArgs, "-e", fmt.Sprintf("^title:\\s*.*%s", keyword))
			} else {
				rgArgs = append(rgArgs, "-e", "^title:\\s*") // すべてのタイトルを検索
			}
		}

		// ノートタイプ検索 (`--type`)
		if len(searchTypes) > 0 {
			for _, t := range searchTypes {
				rgArgs = append(rgArgs, "-e", fmt.Sprintf(`^type:\s*%s`, t))
			}
		}

		// タグ検索 (`--tag`)
		if len(searchTags) > 0 {
			for _, tag := range searchTags {
				rgArgs = append(rgArgs, "-e", fmt.Sprintf(`^tags:\s*\[.*\b%s\b.*\]`, tag))
			}
		}

		// 通常の全文検索
		if keyword != "" && !searchTitle && len(searchTags) == 0 && len(searchTypes) == 0 {
			rgArgs = append(rgArgs, "-e", keyword)
		}

		// 検索ディレクトリを最後に追加
		rgArgs = append(rgArgs, config.NoteDir)

		var out bytes.Buffer
		rgCmd := exec.Command("rg", rgArgs...)
		rgCmd.Stdout = &out
		rgCmd.Stderr = &out
		output := out.String()
		// fmt.Printf("%T, %v", out, out)

		err = rgCmd.Run()
		if err != nil && output == "" {
			fmt.Println("検索に失敗しました:", err)
			// os.Exit(1)
		}

		// 結果の解析
		results := internal.ParseRipgrepOutput(out.String())

		// インタラクティブモード
		if interactive {
			internal.InteractiveSearch(results)
		} else {
			internal.DisplayResults(results)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().BoolVar(&searchTitle, "title", false, "Specify note id")
	searchCmd.Flags().StringSliceVar(&searchTypes, "type", []string{}, "Specify note type")
	searchCmd.Flags().StringSliceVar(&searchTags, "tag", []string{}, "Specify tags")
	searchCmd.Flags().IntVar(&searchContext, "context", 0, "Display N lines before and after the search result")
	searchCmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive search with fzf")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
