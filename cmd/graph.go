/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

var tag string

func filterByTag(zettels []internal.Zettel, tag string) []internal.Zettel {
	var filtered []internal.Zettel
	for _, z := range zettels {
		if z.Archived || z.Deleted {
			continue
		}

		// 指定タグを含むノートを取得
		for _, t := range z.Tags {
			if strings.EqualFold(t, tag) {
				filtered = append(filtered, z)
				break
			}
		}
	}
	return filtered
}

func buildGraph(zettels []internal.Zettel) (map[string][]string, map[string]string, map[string]string) {
	graph := make(map[string][]string)
	titles := make(map[string]string)
	types := make(map[string]string)

	for _, z := range zettels {
		titles[z.NoteID] = z.Title
		types[z.NoteID] = z.NoteType

		// リンク関係を登録
		for _, link := range z.Links {
			graph[z.NoteID] = append(graph[z.NoteID], link)
		}
	}

	return graph, titles, types
}

func findRootNodes(graph map[string][]string) []string {
	childNodes := make(map[string]bool)

	for _, children := range graph {
		for _, child := range children {
			childNodes[child] = true
		}
	}

	var rootNodes []string
	for node := range graph {
		if !childNodes[node] { // 他のノードから指されていなければルート
			rootNodes = append(rootNodes, node)
		}
	}
	return rootNodes
}

func printAsciiGraph(graph map[string][]string, titles map[string]string, node string, prefix string, branch string, visited map[string]bool, backlinkMap map[string][]string, filterRootOnly map[string]bool, isRoot bool) {
	if visited[node] {
		return
	}
	visited[node] = true

	// ルートノードがフィルター対象に含まれない場合、スキップ
	if isRoot && len(filterRootOnly) > 0 {
		if _, exists := filterRootOnly[node]; !exists {
			return
		}
	}

	// ルートノードならプレフィックスなし
	if isRoot {
		fmt.Println(color.CyanString("\n━━━━━━━━━━━━━━━━━━━━━━"))
		printNode("[" + node + "]" + titles[node])
	} else {
		fmt.Println(prefix + branch + "[" + node + "]" + titles[node])
	}

	children := graph[node]
	totalChildren := len(children)

	// **通常の子ノードを表示**
	for i, child := range children {
		childBranch := "├── "
		nextPrefix := prefix + "   " // 子ノードには `│` を適用
		if i == totalChildren-1 {
			childBranch = "└── "
		}
		printAsciiGraph(graph, titles, child, nextPrefix, childBranch, visited, backlinkMap, filterRootOnly, false)
	}

	// **ルートノードならバックリンクを表示しない**
	if isRoot {
		return
	}

	// **バックリンクを表示**
	backlinks, hasBacklinks := backlinkMap[node]
	if hasBacklinks {
		for i, src := range backlinks {
			backlinkBranch := "└── "      // バックリンクは常に最後のノード扱い
			nextPrefix := prefix + "    " // バックリンクには `│` を適用しない

			if i < len(backlinks)-1 {
				backlinkBranch = "├── "
			}

			printBacklink(nextPrefix + backlinkBranch + "↩ Backlink from " + "[" + node + "]" + titles[src])
		}
	}
}
func printHeader() {
	fmt.Println(color.YellowString("\n──────────────────────────────────────"))
	fmt.Println(color.YellowString("📂 Zettelkasten Graph (ASCII)"))
	fmt.Println(color.YellowString("──────────────────────────────────────"))
}

func printNode(title string) {
	c := color.New(color.FgCyan, color.Bold) // ノードタイトルを青 + 太字
	c.Println("🔹 " + title)
}

func printBacklink(title string) {
	c := color.New(color.FgHiBlack) // バックリンクをグレー
	c.Println(title)
}

// graphCmd represents the graph command
var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		config, err := internal.LoadConfig()
		if err != nil {
			log.Printf("❌ Error loading config: %v", err)
			os.Exit(1)
		}

		// Perform cleanup tasks
		if err := internal.CleanupBackups(config.Backup.BackupDir, time.Duration(config.Backup.Retention)*24*time.Hour); err != nil {
			log.Printf("⚠️ Backup cleanup failed: %v", err)
		}
		if err := internal.CleanupTrash(config.Trash.TrashDir, time.Duration(config.Trash.Retention)*24*time.Hour); err != nil {
			log.Printf("⚠️ Trash cleanup failed: %v", err)
		}

		// Load notes from JSON
		zettels, err := internal.LoadJson(*config)
		if err != nil {
			log.Printf("❌ Error loading notes from JSON: %v", err)
			os.Exit(1)
		}

		var rootFilter map[string]bool
		if tag != "" {
			filteredZettels := filterByTag(zettels, tag)
			rootFilter = make(map[string]bool)
			for _, z := range filteredZettels {
				rootFilter[z.NoteID] = true
			}
		}

		// グラフ構築
		graph, titles, _ := buildGraph(zettels)

		visited := make(map[string]bool)
		backlinkMap := make(map[string][]string)
		for parent, children := range graph {
			for _, child := range children {
				backlinkMap[child] = append(backlinkMap[child], parent)
			}
		}

		printHeader()

		// // ルートノード（親を持たないノート）を探して表示
		// rootNodes := findRootNodes(graph)

		// **ルートノードから出発**
		for _, z := range zettels {
			if _, exists := graph[z.NoteID]; exists {
				if !visited[z.NoteID] {
					printAsciiGraph(graph, titles, z.NoteID, "", "", visited, backlinkMap, rootFilter, true)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(graphCmd)
	graphCmd.Flags().StringVar(&tag, "tag", "", "Specify tags")
}
