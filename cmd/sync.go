/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

func syncZettel(config internal.Config, zettels []internal.Zettel) error {
	// config, err := internal.LoadConfig()
	// if err != nil {
	// 	return fmt.Errorf("設定の読み込みに失敗: %w", err)
	// }

	// zettels, err := internal.LoadJson(*config)
	// if err != nil {
	// 	return fmt.Errorf("zettel.json の読み込みに失敗: %w", err)
	// }

	// notesDir := config.NoteDir // 設定で notes/ のパスを取得
	// noteFiles, err := scanDirectory(notesDir)
	// if err != nil {
	// 	return fmt.Errorf("ノートファイルのスキャンに失敗: %w", err)
	// }

	notesDir := config.NoteDir
	archiveDir := config.ArchiveDir
	trashDir := config.Trash.TrashDir

	updatedZettels := syncZettelData(zettels, notesDir, archiveDir, trashDir)

	// 変更がある場合、zettel.json を更新
	if err := saveZettelJson(updatedZettels, config); err != nil {
		return fmt.Errorf("zettel.json の保存に失敗: %w", err)
	}

	fmt.Println("zettel.json を最新の状態に同期しました！")
	return nil
}

// scanNotesDirectory は notes/ ディレクトリ内の Markdown ファイルを取得
func scanDirectory(dir string, sortAscending bool) (map[string]string, error) {
	noteFiles := make(map[string]string)
	var fileInfos []os.FileInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// ファイル情報を収集
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue // サブディレクトリと非 .md ファイルを無視
		}
		info, err := entry.Info()
		if err != nil {
			continue // 取得失敗した場合はスキップ
		}
		fileInfos = append(fileInfos, info)
	}

	// 作成日時順にソート
	sort.Slice(fileInfos, func(i, j int) bool {
		if sortAscending {
			return fileInfos[i].ModTime().Before(fileInfos[j].ModTime()) // 古い順
		}
		return fileInfos[j].ModTime().Before(fileInfos[i].ModTime()) // 新しい順
	})

	// ソート後のリストで map を作成
	for _, info := range fileInfos {
		noteID := info.Name()
		noteFiles[noteID] = filepath.Join(dir, noteID)
	}

	return noteFiles, nil
}

// syncZettelData は現在の zettels と notes の差分を同期
func syncZettelData(zettels []internal.Zettel, notesDir, archiveDir, trashDir string) []internal.Zettel {
	updatedZettels := []internal.Zettel{}
	existingNotes := make(map[string]bool)

	noteFiles, _ := scanDirectory(notesDir, true)
	archiveFiles, _ := scanDirectory(archiveDir, true)
	trashFiles, _ := scanDirectory(trashDir, true)

	// 既存の Zettel をチェックしながら更新
	for _, z := range zettels {
		t := time.Now()

		if path, exists := noteFiles[z.NoteID]; exists {
			// 既存ノートは更新（UpdatedAt を最新にする）
			z.UpdatedAt = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
				t.Year(), t.Month(), t.Day(),
				t.Hour(), t.Minute(), t.Second())
			z.NotePath = path
			updatedZettels = append(updatedZettels, z)
			existingNotes[z.NoteID] = true
		} else if path, exists := archiveFiles[z.NoteID]; exists {
			z.NotePath = path
			z.Archived = true
			z.Deleted = false
			updatedZettels = append(updatedZettels, z)
			existingNotes[z.NoteID] = true
		} else if path, exists := trashFiles[z.NoteID]; exists {
			// ノートが trash/ にある → 削除済み扱いにする
			z.NotePath = path
			z.Archived = false
			z.Deleted = true
			updatedZettels = append(updatedZettels, z)
			existingNotes[z.NoteID] = true
		} else {
			// どこにもない場合 → trash/ に移動し、削除フラグを立てる
			newPath := filepath.Join(trashDir, z.NoteID)
			os.Rename(z.NotePath, newPath) // ファイルを trash/ に移動
			z.UpdatedAt = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
				t.Year(), t.Month(), t.Day(),
				t.Hour(), t.Minute(), t.Second())
			z.NotePath = newPath
			z.Archived = false
			z.Deleted = true
			updatedZettels = append(updatedZettels, z)
		}
	}

	// 追加が必要なノートをチェック
	// **新しいノートを追加（notes, archive, trash を対象）**
	newID := len(zettels) + 1 // 既存の ID の次から開始

	for _, fileSet := range []struct {
		files    map[string]string
		archived bool
		deleted  bool
	}{
		{noteFiles, false, false},
		{archiveFiles, true, false},
		{trashFiles, false, true},
	} {
		for noteID, path := range fileSet.files {
			if _, exists := existingNotes[noteID]; !exists {
				note, err := os.ReadFile(path)
				if err != nil {
					fmt.Println("Error:", err)
					continue
				}
				frontMatter, _, err := internal.ParseFrontMatter(string(note))
				if err != nil {
					fmt.Println("Error parsing front matter:", err)
					continue
				}

				newZettel := internal.Zettel{
					ID:         strconv.Itoa(newID),
					NoteID:     frontMatter.ID,
					Title:      frontMatter.Title,
					NoteType:   frontMatter.Type,
					Tags:       frontMatter.Tags,
					TaskStatus: frontMatter.TaskStatus,
					Links:      frontMatter.Links,
					CreatedAt:  frontMatter.CreatedAt,
					UpdatedAt:  frontMatter.UpdatedAt,
					NotePath:   path,
					Archived:   fileSet.archived, // archive フラグ
					Deleted:    fileSet.deleted,  // trash フラグ
				}
				updatedZettels = append(updatedZettels, newZettel)
				newID++ // ID をインクリメント
			}
		}
	}

	return updatedZettels
}

// saveZettelJson は zettel.json を保存
func saveZettelJson(zettels []internal.Zettel, config internal.Config) error {
	file, err := os.Create(config.ZettelJson)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(zettels)
}

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sync called")

		config, err := internal.LoadConfig()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		zettels, err := internal.LoadJson(*config)
		if err != nil {
			fmt.Println("Error:", err)
		}

		err = syncZettel(*config, zettels)
		if err != nil {
			fmt.Println("Error:", err)
		}

	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
