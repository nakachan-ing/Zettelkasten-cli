/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

var threshold float64

// `links:` フィールドを更新する
func AddLinkToFrontMatter(frontMatter *internal.FrontMatter, newLinks []string) *internal.FrontMatter {
	if frontMatter.Links == nil {
		frontMatter.Links = []string{}
		fmt.Println(frontMatter.Links)
	} else {
		for _, newLink := range newLinks {
			frontMatter.Links = append(frontMatter.Links, newLink)
		}

		fmt.Println(frontMatter.Links)
	}
	return frontMatter
}

// 📥 `zettel.json` からノートをロードし、TF-IDF を計算
func LoadNotesWithTFIDF(config *internal.Config) ([]internal.Note, error) {
	zettels, err := internal.LoadJson(*config)
	if err != nil {
		return nil, fmt.Errorf("❌ JSON 読み込みエラー: %v", err)
	}

	var notes []internal.Note

	// 各 `Zettel` を `Note` に変換
	for _, z := range zettels {
		content, err := os.ReadFile(z.NotePath)
		if err != nil {
			fmt.Printf("⚠️ ノートの読み込み失敗: %s (%v)\n", z.NotePath, err)
			continue
		}

		// MeCab でキーワード抽出
		keywords, err := internal.ExtractKeywordsMeCab(string(content))
		if err != nil {
			fmt.Printf("⚠️ キーワード抽出失敗: %s (%v)\n", z.NotePath, err)
			continue
		}

		// `Note` 構造体にマッピング
		notes = append(notes, internal.Note{
			ID:       z.ID,
			NoteId:   z.NoteID,
			Title:    z.Title,
			Content:  string(content),
			Keywords: keywords,
		})
	}

	// IDF を計算
	idf := internal.CalculateIDF(notes)

	// 各ノートの TF-IDF を計算
	for i := range notes {
		tf := internal.CalculateTF(notes[i].Keywords)
		tfidf := make(map[string]float64)
		for word, tfVal := range tf {
			tfidf[word] = tfVal * idf[word]
		}
		notes[i].TFIDF = tfidf
	}

	return notes, nil
}

// ✍️ **Markdown ファイルを更新**
func UpdateMarkdownFile(note internal.Note, relatedIDs []string, config *internal.Config) {
	content, err := os.ReadFile(note.Content)
	if err != nil {
		fmt.Println("❌ Error reading note:", err)
		return
	}

	frontMatter, body, err := internal.ParseFrontMatter(string(content))
	if err != nil {
		fmt.Println("❌ Error parsing front matter:", err)
		return
	}

	updatedFrontMatter := AddLinkToFrontMatter(&frontMatter, relatedIDs)
	updatedContent := internal.UpdateFrontMatter(updatedFrontMatter, body)

	err = os.WriteFile(note.Content, []byte(updatedContent), 0644)
	if err != nil {
		fmt.Println("❌ Error updating note:", err)
	}
}

// 💾 **zettel.json を保存**
func SaveUpdatedJson(zettels []internal.Zettel, config *internal.Config) {
	updatedJson, err := json.MarshalIndent(zettels, "", "  ")
	if err != nil {
		fmt.Println("⚠️ JSON の変換エラー:", err)
		return
	}

	err = os.WriteFile(config.ZettelJson, updatedJson, 0644)
	if err != nil {
		fmt.Println("⚠️ JSON 書き込みエラー:", err)
	} else {
		fmt.Println("✅ JSON 更新完了:", config.ZettelJson)
	}
}

func AutoLinkNotes(fromID string, threshold float64, config internal.Config, zettels []internal.Zettel, tfidfMap map[string]map[string]float64) {
	// ✅ `fromID` の前後のスペースを削除
	cleanFromID := strings.TrimSpace(fromID)
	fmt.Println("🔍 クリーンな `fromID`:", cleanFromID)

	// ✅ `fromID` から `zettels.json` を使って `FileID` と `FilePath` を取得
	var fromZettel *internal.Zettel

	for i, zettel := range zettels {
		cleanZettelID := strings.TrimSpace(zettel.ID)
		if cleanZettelID == cleanFromID {
			fromZettel = &zettels[i]
			break
		}
	}

	// ✅ `fromZettel` が見つからなかった場合エラー
	if fromZettel == nil {
		fmt.Println("❌ 指定されたノートが `zettels.json` に見つかりません:", cleanFromID)
		return
	}

	// ✅ `FileID` と `FilePath` を取得
	fileID := fromZettel.NoteID
	filePath := fromZettel.NotePath
	fmt.Println("✅ `zettels.json` から取得した `FileID`:", fileID)
	fmt.Println("📄 `zettels.json` から取得した `FilePath`:", filePath)

	// ✅ ノートが存在するかチェック
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("❌ 指定されたノートのファイルが存在しません:", filePath)
		return
	}

	// ✅ 関連ノートを検索（自分自身を除外）
	relatedIDs := internal.FindRelatedNotes(*fromZettel, zettels, threshold, tfidfMap)
	if len(relatedIDs) == 0 {
		fmt.Println("⚠️ 関連ノートが見つかりませんでした:", fileID)
		return
	}

	// ✅ `zettels.json` の `Links` を更新
	for i := range zettels {
		if zettels[i].NoteID == fileID {
			zettels[i].Links = MergeUniqueLinks(zettels[i].Links, relatedIDs)
			break
		}
	}

	// ✅ ノートの内容を取得
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("❌ ノートの読み込みエラー:", err)
		return
	}

	frontMatter, body, err := internal.ParseFrontMatter(string(content))
	if err != nil {
		fmt.Println("❌ フロントマターの解析エラー:", err)
		return
	}

	// ✅ リンクを追加
	updatedFrontMatter := AddLinkToFrontMatter(&frontMatter, relatedIDs)
	updatedContent := internal.UpdateFrontMatter(updatedFrontMatter, body)

	// ✅ ファイルに書き戻し
	err = os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		fmt.Println("❌ 書き込みエラー:", err)
		return
	}

	// ✅ `zettels.json` を保存
	SaveUpdatedJson(zettels, &config)

	fmt.Printf("✅ 自動リンク完了: [%s]%s に関連ノートを追加しました\n", fromZettel.NoteID, fromZettel.Title)
}

// `existingLinks`（既存のリンク）と `newLinks`（追加するリンク）を統合し、重複を排除
func MergeUniqueLinks(existingLinks, newLinks []string) []string {
	linkSet := make(map[string]bool)
	var merged []string

	// 既存のリンクをセットに追加
	for _, link := range existingLinks {
		linkSet[link] = true
		merged = append(merged, link)
	}

	// 新しいリンクを追加（重複しないようにチェック）
	for _, link := range newLinks {
		if !linkSet[link] { // すでに存在しない場合のみ追加
			linkSet[link] = true
			merged = append(merged, link)
		}
	}

	return merged
}

// linkCmd represents the link command
var manualFlag bool
var autoFlag bool

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "ノート同士をリンクする",
	Long: `ノート同士を手動または自動でリンクします。

手動リンク:
  zk link --manual <from> <to>

自動リンク:
  zk link --auto <from>`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// ✅ `--manual` と `--auto` の両方が指定された場合はエラー
		if manualFlag && autoFlag {
			return fmt.Errorf("❌ `--manual` と `--auto` は同時に指定できません")
		}

		// ✅ `--manual` の場合 (手動リンク)
		if manualFlag {
			if len(args) != 2 {
				return fmt.Errorf("❌ `--manual` の場合は `zk link --manual <from> <to>` の形式で実行してください")
			}
			return runManualLink(args[0], args[1])
		}

		// ✅ `--auto` の場合 (自動リンク)
		if autoFlag {
			if len(args) != 1 {
				return fmt.Errorf("❌ `--auto` の場合は `zk link --auto <from>` の形式で実行してください")
			}
			return runAutoLink(args[0])
		}

		return fmt.Errorf("❌ `--manual` または `--auto` のどちらかを指定してください")
	},
}

func runManualLink(sourceId, destinationId string) error {
	// ✅ 設定を読み込む
	config, err := internal.LoadConfig()
	if err != nil || config == nil {
		return fmt.Errorf("❌ 設定ファイルの読み込みエラー: %v", err)
	}

	zettels, err := internal.LoadJson(*config)
	if err != nil {
		return fmt.Errorf("❌ Jsonファイルの読み込みエラー: %v", err)
	}

	for i := range zettels {
		if zettels[i].ID == sourceId {
			sourceLinkId := zettels[i].NoteID
			filePath := fmt.Sprintf("%s/%s.md", config.NoteDir, sourceLinkId)
			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("❌ ノートの読み込みエラー: %v", err)
			}
			// ✅ フロントマターを解析
			frontMatter, body, err := internal.ParseFrontMatter(string(content))
			if err != nil {
				return fmt.Errorf("❌ フロントマターの解析エラー: %v", err)
			}

			// ✅ `frontMatter.Links` が `nil` の場合に初期化
			if frontMatter.Links == nil {
				frontMatter.Links = []string{}
			}
			for ii := range zettels {
				destLinkIds := []string{}
				if zettels[ii].ID == destinationId {
					destLinkId := zettels[ii].NoteID
					fmt.Println(destLinkId)
					destLinkIds = append(destLinkIds, destLinkId)
					// ✅ 既存の `Links` に `destinationId` を追加
					updatedFrontMatter := AddLinkToFrontMatter(&frontMatter, destLinkIds)

					// ✅ フロントマターを更新した新しい Markdown コンテンツを作成
					updatedContent := internal.UpdateFrontMatter(updatedFrontMatter, body)

					// ✅ Markdown を更新（書き込み処理）
					err = os.WriteFile(filePath, []byte(updatedContent), 0644)
					if err != nil {
						return fmt.Errorf("❌ Markdown 書き込みエラー: %v", err)
					}

					// ✅ `zettels.json` の `Links` も更新
					zettels[i].Links = MergeUniqueLinks(zettels[i].Links, []string{destLinkId})

					// ✅ `zettels.json` を保存
					SaveUpdatedJson(zettels, config)

					fmt.Printf("✅ ノート [%s]%s に [%s]%s をリンクしました\n", zettels[i].NoteID, zettels[i].Title, zettels[ii].NoteID, zettels[ii].Title)
					return nil
				}
			}

		}
	}
	return nil

}

func runAutoLink(fromID string) error {
	// ✅ 設定を読み込む
	config, err := internal.LoadConfig()
	if err != nil || config == nil {
		return fmt.Errorf("❌ 設定ファイルの読み込みエラー: %v", err)
	}

	// ✅ `zettels.json` をロード
	zettels, err := internal.LoadJson(*config)
	if err != nil {
		return fmt.Errorf("❌ JSONファイルの読み込みエラー: %v", err)
	}

	// ✅ TF-IDF を事前に計算
	tfidfMap := internal.ComputeTFIDFForZettels(zettels)

	// ✅ `fromID` から `zettels.json` の情報を取得するように変更
	AutoLinkNotes(fromID, threshold, *config, zettels, tfidfMap)

	return nil
}

func init() {

	linkCmd.PersistentFlags().Float64VarP(&threshold, "threshold", "t", 0.5, "類似ノートをリンクするしきい値")
	rootCmd.AddCommand(linkCmd)
	linkCmd.Flags().BoolVarP(&manualFlag, "manual", "m", false, "手動でノートをリンク")
	linkCmd.Flags().BoolVar(&autoFlag, "auto", false, "関連ノートを自動でリンク")
}
