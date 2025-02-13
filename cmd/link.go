/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

var threshold float64

// `links:` フィールドを更新する
func addLinkToFrontMatter(frontMatter *internal.FrontMatter, newLinks []string) *internal.FrontMatter {
	if frontMatter.Links == nil {
		frontMatter.Links = []string{}
		fmt.Println(frontMatter.Links)
	} else {
		for _, newLink := range newLinks {
			// 重複チェック
			exists := false
			for _, existing := range frontMatter.Links {
				if existing == newLink {
					exists = true
					break
				}
			}
			if !exists {
				frontMatter.Links = append(frontMatter.Links, newLink)
			}
		}
	}
	return frontMatter
}

// 🔍 **ユーザーに関連ノートを選択させる**
func selectRelatedNotes(relatedNotes []internal.Zettel) []string {
	var selected []string
	options := []string{}

	for _, note := range relatedNotes {
		options = append(options, fmt.Sprintf("%s: %s", note.ID, note.Title))
	}

	prompt := &survey.MultiSelect{
		Message: "リンクするノートを選択してください:",
		Options: options,
	}

	survey.AskOne(prompt, &selected, nil)

	// 選択されたノートの ID を返す
	var selectedIDs []string
	for _, sel := range selected {
		selectedIDs = append(selectedIDs, strings.Split(sel, ": ")[0])
	}
	return selectedIDs
}

func autoLinkNotes(fromID string, threshold float64, config internal.Config, zettels []internal.Zettel, tfidfMap map[string]map[string]float64) {
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
	// fmt.Println("✅ `zettels.json` から取得した `FileID`:", fileID)
	// fmt.Println("📄 `zettels.json` から取得した `FilePath`:", filePath)

	// ✅ ノートが存在するかチェック
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("❌ 指定されたノートのファイルが存在しません:", filePath)
		return
	}

	// ✅ 関連ノートを検索（自分自身を除外）
	relatedNotes := internal.FindRelatedNotes(*fromZettel, zettels, threshold, tfidfMap)
	if len(relatedNotes) == 0 {
		fmt.Println("⚠️ 関連ノートが見つかりませんでした:", fileID)
		return
	}

	// ✅ **ユーザーに関連ノートを選択させる**
	selectedIDs := selectRelatedNotes(relatedNotes)
	if len(selectedIDs) == 0 {
		fmt.Println("⚠️ 何も選択されませんでした。リンクは追加されません。")
		return
	}

	// ✅ `zettels.json` の `Links` を更新
	for i := range zettels {
		if zettels[i].NoteID == fileID {
			zettels[i].Links = mergeUniqueLinks(zettels[i].Links, selectedIDs)
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
	updatedFrontMatter := addLinkToFrontMatter(&frontMatter, selectedIDs)
	updatedContent := internal.UpdateFrontMatter(updatedFrontMatter, body)

	// ✅ ファイルに書き戻し
	err = os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		fmt.Println("❌ 書き込みエラー:", err)
		return
	}

	// ✅ `zettels.json` を保存
	internal.SaveUpdatedJson(zettels, &config)

	fmt.Printf("✅ 自動リンク完了: [%s]%s に関連ノートを追加しました\n", fromZettel.NoteID, fromZettel.Title)
}

// `existingLinks`（既存のリンク）と `newLinks`（追加するリンク）を統合し、重複を排除
func mergeUniqueLinks(existingLinks, newLinks []string) []string {
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

// ✍️ **文脈の中に `[title](ファイル名)` を挿入**
// ✍️ **文脈の中に `[title](ファイル名)` を挿入**
func insertLinkInContext(filePath string, title string, fileName string, keywords []string) (string, error) {
	// ✅ ノートの内容を取得
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("❌ ノートの読み込みエラー: %v", err)
	}
	text := string(content)

	// ✅ **フロントマター部分と本文を分離**
	parts := strings.SplitN(text, "---", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("❌ フロントマターの解析エラー: `---` の区切りが不足しています")
	}
	frontMatter := "---" + parts[1] + "---\n" // フロントマター
	body := parts[2]                          // 本文

	// ✅ **すでに同じリンクがある場合は何もしない**
	markdownLink := fmt.Sprintf("[%s](%s)", title, fileName)
	if strings.Contains(body, markdownLink) {
		fmt.Println("⚠️ 既にリンクが存在するため、追加しません:", markdownLink)
		return text, nil
	}

	// ✅ **キーワードの出現位置を取得**
	var keywordPositions []string
	positionMap := make(map[string]int)
	for _, keyword := range keywords {
		re := regexp.MustCompile(fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(keyword)))
		loc := re.FindStringIndex(body)
		if loc != nil {
			position := fmt.Sprintf("%s の後", keyword)
			keywordPositions = append(keywordPositions, position)
			positionMap[position] = loc[1] // 挿入する位置
		}
	}

	// ✅ **選択肢を作成**
	keywordPositions = append(keywordPositions, "### Links に追加")

	// ✅ **ユーザーに挿入場所を選択させる**
	var selectedPosition string
	prompt := &survey.Select{
		Message: "リンクを挿入する場所を選択してください:",
		Options: keywordPositions,
	}
	survey.AskOne(prompt, &selectedPosition, nil)

	// ✅ **ユーザーが選択した場所にリンクを挿入**
	inserted := false
	if selectedPosition != "### Links に追加" {
		insertPos := positionMap[selectedPosition]
		body = body[:insertPos] + " " + markdownLink + body[insertPos:]
		inserted = true
		fmt.Println("✅ 文脈の中にリンクを挿入しました:", markdownLink)
	}

	// ✅ **適切な場所が見つからなかった場合、`### Links` に追加**
	if !inserted {
		if !strings.Contains(body, "### Links") {
			body += "\n\n### Links"
		}
		body += fmt.Sprintf("\n- %s", markdownLink)
		fmt.Println("✅ `### Links` に追加しました:", markdownLink)
	}

	// ✅ **更新された内容を返す**
	updatedText := frontMatter + body
	return updatedText, nil
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

	err = internal.CleanupBackups(config.Backup.BackupDir, time.Duration(config.Backup.Retention)*24*time.Hour)
	if err != nil {
		fmt.Printf("Backup cleanup failed: %v\n", err)
	}
	err = internal.CleanupTrash(config.Trash.TrashDir, time.Duration(config.Trash.Retention)*24*time.Hour)
	if err != nil {
		fmt.Printf("Trash cleanup failed: %v\n", err)
	}

	zettels, err := internal.LoadJson(*config)
	if err != nil {
		return fmt.Errorf("❌ Jsonファイルの読み込みエラー: %v", err)
	}

	var sourceZettel *internal.Zettel
	var destinationZettel *internal.Zettel

	for i := range zettels {
		if zettels[i].ID == sourceId {
			sourceZettel = &zettels[i]
		}
		if zettels[i].ID == destinationId {
			destinationZettel = &zettels[i]
		}
	}

	// ✅ `sourceId` または `destinationId` が見つからない場合
	if sourceZettel == nil || destinationZettel == nil {
		return fmt.Errorf("❌ 指定されたノートが見つかりません: %s -> %s", sourceId, destinationId)
	}

	// ✅ `sourceZettel` のノートを取得
	filePath := sourceZettel.NotePath
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("❌ ノートの読み込みエラー: %v", err)
	}

	// ✅ フロントマターを解析
	frontMatter, body, err := internal.ParseFrontMatter(string(content))
	if err != nil {
		return fmt.Errorf("❌ フロントマターの解析エラー: %v", err)
	}

	// ✅ **MeCab でキーワードを抽出**
	// keywords, err := internal.ExtractKeywordsMeCab(body)
	// if err != nil {
	// 	return fmt.Errorf("❌ キーワード抽出エラー: %v", err)
	// }

	// ✅ **MeCab でフレーズを抽出**
	keyPhrases, err := internal.ExtractKeyPhrasesMeCab(string(body))
	if err != nil {
		return fmt.Errorf("❌ フレーズ抽出エラー: %v", err)
	}

	// ✅ **ユーザーに挿入場所を選択させる**
	updatedMarkdown, err := insertLinkInContext(filePath, destinationZettel.Title, destinationZettel.NoteID+".md", keyPhrases)
	if err != nil {
		return fmt.Errorf("❌ Markdown 内へのリンク挿入エラー: %v", err)
	}

	// ✅ **フロントマターを更新**
	frontMatter, body, err = internal.ParseFrontMatter(updatedMarkdown)
	if err != nil {
		return fmt.Errorf("❌ フロントマターの解析エラー: %v", err)
	}
	updatedFrontMatter := addLinkToFrontMatter(&frontMatter, []string{destinationZettel.NoteID})

	// ✅ **フロントマターを適用**
	finalMarkdown := internal.UpdateFrontMatter(updatedFrontMatter, body)
	fmt.Println(finalMarkdown)

	// ✅ **Markdown を最終更新（書き込み処理）**
	err = os.WriteFile(filePath, []byte(finalMarkdown), 0644)
	if err != nil {
		return fmt.Errorf("❌ 最終 Markdown 書き込みエラー: %v", err)
	}

	// ✅ `zettels.json` の `Links` も更新
	sourceZettel.Links = mergeUniqueLinks(sourceZettel.Links, []string{destinationZettel.NoteID})

	// ✅ `zettels.json` を保存
	internal.SaveUpdatedJson(zettels, config)

	fmt.Printf("✅ ノート [%s] %s に [%s] %s をリンクしました\n", sourceZettel.NoteID, sourceZettel.Title, destinationZettel.NoteID, destinationZettel.Title)
	return nil

}

func runAutoLink(fromID string) error {
	// ✅ 設定を読み込む
	config, err := internal.LoadConfig()
	if err != nil || config == nil {
		return fmt.Errorf("❌ 設定ファイルの読み込みエラー: %v", err)
	}

	retention := time.Duration(config.Backup.Retention) * 24 * time.Hour

	err = internal.CleanupBackups(config.Backup.BackupDir, retention)
	if err != nil {
		fmt.Printf("Backup cleanup failed: %v\n", err)
	}

	retention = time.Duration(config.Trash.Retention) * 24 * time.Hour

	err = internal.CleanupTrash(config.Trash.TrashDir, retention)
	if err != nil {
		fmt.Printf("Trash cleanup failed: %v\n", err)
	}

	// ✅ `zettels.json` をロード
	zettels, err := internal.LoadJson(*config)
	if err != nil {
		return fmt.Errorf("❌ JSONファイルの読み込みエラー: %v", err)
	}

	// ✅ TF-IDF を事前に計算
	tfidfMap := internal.ComputeTFIDFForZettels(zettels)

	// ✅ `fromID` から `zettels.json` の情報を取得するように変更
	autoLinkNotes(fromID, threshold, *config, zettels, tfidfMap)

	return nil
}

func init() {

	linkCmd.PersistentFlags().Float64VarP(&threshold, "threshold", "t", 0.5, "類似ノートをリンクするしきい値")
	rootCmd.AddCommand(linkCmd)
	linkCmd.Flags().BoolVarP(&manualFlag, "manual", "m", false, "手動でノートをリンク")
	linkCmd.Flags().BoolVar(&autoFlag, "auto", false, "関連ノートを自動でリンク")
}
