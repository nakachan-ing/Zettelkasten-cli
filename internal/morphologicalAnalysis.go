package internal

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ノートを表す構造体
type Note struct {
	ID       string
	NoteId   string
	Title    string
	Content  string
	Keywords []string
	TFIDF    map[string]float64
}

// MeCab を使って名詞・動詞・形容詞を抽出
func ExtractKeywordsMeCab(text string) ([]string, error) {
	cmd := exec.Command("mecab")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stdin = strings.NewReader(text)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("MeCab の実行に失敗しました: %v", err)
	}

	lines := strings.Split(out.String(), "\n")
	var keywords []string
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) > 1 {
			features := strings.Split(parts[1], ",")
			if len(features) > 0 {
				wordType := features[0]
				if wordType == "名詞" || wordType == "動詞" || wordType == "形容詞" {
					keywords = append(keywords, parts[0])
				}
			}
		}
	}
	return keywords, nil
}

// 指定ディレクトリ内のノートを読み込む
func LoadNotesFromDir(noteDir string) ([]Note, error) {
	var notes []Note
	files, err := os.ReadDir(noteDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".md" {
			filePath := filepath.Join(noteDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("⚠️ ノート読み込みエラー: %v\n", err)
				continue
			}

			// MeCab でキーワードを抽出
			keywords, err := ExtractKeywordsMeCab(string(content))
			if err != nil {
				log.Printf("⚠️ キーワード抽出エラー: %v\n", err)
				continue
			}

			notes = append(notes, Note{
				ID:       strings.TrimSuffix(file.Name(), ".md"),
				Content:  string(content),
				Keywords: keywords,
				TFIDF:    make(map[string]float64),
			})
		}
	}
	return notes, nil
}

// **TF-IDF を計算**
func CalculateTFIDF(notes []Note) {
	// **TF（単語の出現頻度）を計算**
	tfMap := make(map[string]map[string]float64) // map[ノートID]map[単語]TF値
	for _, note := range notes {
		tfMap[note.ID] = CalculateTF(note.Keywords)
	}

	// **IDF（逆文書頻度）を計算**
	idfMap := CalculateIDF(notes)

	// **TF-IDF を計算**
	for i := range notes {
		tfidf := make(map[string]float64)
		for word, tf := range tfMap[notes[i].ID] {
			idf := idfMap[word]
			tfidf[word] = tf * idf // TF-IDF = TF × IDF
		}
		notes[i].TFIDF = tfidf
	}
}

// 🏷 **TF（単語の出現頻度）を計算**
func CalculateTF(words []string) map[string]float64 {
	tf := make(map[string]float64)
	totalWords := len(words)

	for _, word := range words {
		tf[word]++
	}
	for word := range tf {
		tf[word] /= float64(totalWords)
	}
	return tf
}

// 📊 **IDF（逆文書頻度）を計算**
func CalculateIDF(notes []Note) map[string]float64 {
	idf := make(map[string]float64)
	totalDocs := float64(len(notes))

	for _, note := range notes {
		seen := make(map[string]bool)
		for _, word := range note.Keywords {
			if !seen[word] {
				idf[word]++
				seen[word] = true
			}
		}
	}

	for word := range idf {
		idf[word] = math.Log(totalDocs / (1 + idf[word])) // IDF = log(総文書数 / (1 + 出現文書数))
	}
	return idf
}

// 🛠 **コサイン類似度を計算**
func CosineSimilarity(vec1, vec2 map[string]float64) float64 {
	var dotProduct, normA, normB float64
	for key, valA := range vec1 {
		if valB, found := vec2[key]; found {
			dotProduct += valA * valB
		}
		normA += valA * valA
	}
	for _, valB := range vec2 {
		normB += valB * valB
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// `from` のノートのキーワードを抽出
func GetNoteKeywords(note Note) map[string]float64 {
	tf := CalculateTF(note.Keywords) // TF を計算
	return tf
}

// 🔍 **関連ノートを検索**
func FindRelatedNotes(fromZettel Zettel, zettels []Zettel, threshold float64, tfidfMap map[string]map[string]float64) []string {
	var relatedIDs []string

	// ✅ `fromZettel` の TF-IDF を取得
	fromTFIDF, exists := tfidfMap[fromZettel.NoteID]
	if !exists {
		fmt.Println("⚠️ `TF-IDF` データが見つかりません:", fromZettel.NoteID)
		return relatedIDs
	}

	// ✅ 他のノートとの類似度を計算
	for _, zettel := range zettels {
		if zettel.NoteID == fromZettel.NoteID {
			continue // 🔥 **自分自身はスキップ**
		}

		// `zettel` の TF-IDF を取得
		noteTFIDF, exists := tfidfMap[zettel.NoteID]
		if !exists {
			continue
		}

		// コサイン類似度を計算
		similarity := CosineSimilarity(fromTFIDF, noteTFIDF)
		if similarity >= threshold {
			relatedIDs = append(relatedIDs, zettel.NoteID)
		}
	}

	return relatedIDs
}

// ✅ 各 `Zettel` から `TF-IDF` を計算する関数
func ComputeTFIDFForZettels(zettels []Zettel) map[string]map[string]float64 {
	// 🔹 全ノートの単語リストを格納
	documents := make(map[string][]string)

	// ✅ すべてのノートのテキストを処理
	for _, zettel := range zettels {
		// 🔹 ノートの内容を読み込む
		content, err := os.ReadFile(zettel.NotePath)
		if err != nil {
			fmt.Printf("⚠️ ノート読み込みエラー: %s (%v)\n", zettel.NotePath, err)
			continue
		}

		// 🔹 MeCab などを使ってキーワードを抽出
		keywords, err := ExtractKeywordsMeCab(string(content))
		if err != nil {
			fmt.Printf("⚠️ キーワード抽出失敗: %s (%v)\n", zettel.NotePath, err)
			continue
		}

		// 🔹 `NoteID` をキーに単語リストを保存
		documents[zettel.NoteID] = keywords
	}

	// ✅ `TF-IDF` を計算
	return ComputeTFIDF(documents)
}

// ✅ `TF-IDF` を計算する関数
func ComputeTFIDF(documents map[string][]string) map[string]map[string]float64 {
	// 🔹 各ノートの TF（単語の出現頻度）
	tfMap := make(map[string]map[string]float64)

	// 🔹 IDF（逆文書頻度）のための単語出現カウント
	idfCount := make(map[string]int)
	totalDocs := len(documents)

	// ✅ 各ノートの TF を計算
	for docID, words := range documents {
		tf := make(map[string]float64)
		for _, word := range words {
			tf[word]++
		}

		// 🔹 TF を正規化（単語数で割る）
		for word := range tf {
			tf[word] /= float64(len(words))
		}
		tfMap[docID] = tf

		// 🔹 各単語が何個のドキュメントに出現したかカウント
		seen := make(map[string]bool)
		for _, word := range words {
			if !seen[word] {
				idfCount[word]++
				seen[word] = true
			}
		}
	}

	// ✅ IDF（逆文書頻度）を計算
	idfMap := make(map[string]float64)
	for word, count := range idfCount {
		idfMap[word] = math.Log(float64(totalDocs) / (1.0 + float64(count)))
	}

	// ✅ TF-IDF を計算
	tfidfMap := make(map[string]map[string]float64)
	for docID, tf := range tfMap {
		tfidf := make(map[string]float64)
		for word, tfValue := range tf {
			tfidf[word] = tfValue * idfMap[word] // 🔥 TF × IDF
		}
		tfidfMap[docID] = tfidf
	}

	return tfidfMap
}
