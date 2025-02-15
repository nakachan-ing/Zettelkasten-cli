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

// Note represents a structured note
type Note struct {
	ID       string
	NoteId   string
	Title    string
	Content  string
	Keywords []string
	TFIDF    map[string]float64
}

// Extract keywords (nouns, verbs, adjectives) using MeCab
func ExtractKeywordsMeCab(text string) ([]string, error) {
	cmd := exec.Command("mecab")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stdin = strings.NewReader(text)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("❌ Failed to execute MeCab: %w", err)
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

// Extract key phrases using MeCab
func ExtractKeyPhrasesMeCab(text string) ([]string, error) {
	cmd := exec.Command("mecab")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stdin = strings.NewReader(text)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("❌ Failed to execute MeCab: %w", err)
	}

	lines := strings.Split(out.String(), "\n")
	var keyPhrases []string
	var currentPhrase string
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		features := strings.Split(parts[1], ",")
		if len(features) > 0 {
			wordType := features[0]
			if wordType == "名詞" || wordType == "動詞" {
				if currentPhrase == "" {
					currentPhrase = parts[0]
				} else {
					currentPhrase += parts[0]
				}
			} else {
				if currentPhrase != "" {
					keyPhrases = append(keyPhrases, currentPhrase)
					currentPhrase = ""
				}
			}
		}
	}
	if currentPhrase != "" {
		keyPhrases = append(keyPhrases, currentPhrase)
	}

	return removeDuplicates(keyPhrases), nil
}

// Remove duplicate phrases
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, str := range slice {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	return result
}

// Load notes from a directory
func LoadNotesFromDir(noteDir string) ([]Note, error) {
	var notes []Note
	files, err := os.ReadDir(noteDir)
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to read directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".md" {
			filePath := filepath.Join(noteDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("⚠️ Failed to read note: %v", err)
				continue
			}

			keywords, err := ExtractKeywordsMeCab(string(content))
			if err != nil {
				log.Printf("⚠️ Failed to extract keywords: %v", err)
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

// Calculate TF-IDF for notes
func CalculateTFIDF(notes []Note) {
	tfMap := make(map[string]map[string]float64)
	for _, note := range notes {
		tfMap[note.ID] = CalculateTF(note.Keywords)
	}

	idfMap := CalculateIDF(notes)

	for i := range notes {
		tfidf := make(map[string]float64)
		for word, tf := range tfMap[notes[i].ID] {
			idf := idfMap[word]
			tfidf[word] = tf * idf
		}
		notes[i].TFIDF = tfidf
	}
}

// Calculate Term Frequency (TF)
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

// Calculate Inverse Document Frequency (IDF)
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
		idf[word] = math.Log(totalDocs / (1 + idf[word])) // Prevent division by zero
	}
	return idf
}

// Compute Cosine Similarity
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

// Compute TF-IDF for notes
func ComputeTFIDFForZettels(zettels []Zettel) map[string]map[string]float64 {
	documents := make(map[string][]string)

	for _, zettel := range zettels {
		content, err := os.ReadFile(zettel.NotePath)
		if err != nil {
			log.Printf("⚠️ Failed to read note: %s (%v)", zettel.NotePath, err)
			continue
		}

		keywords, err := ExtractKeywordsMeCab(string(content))
		if err != nil {
			log.Printf("⚠️ Failed to extract keywords: %s (%v)", zettel.NotePath, err)
			continue
		}

		documents[zettel.NoteID] = keywords
	}

	return ComputeTFIDF(documents)
}

// Compute TF-IDF for a set of documents
func ComputeTFIDF(documents map[string][]string) map[string]map[string]float64 {
	tfMap := make(map[string]map[string]float64)
	idfCount := make(map[string]int)
	totalDocs := len(documents)

	for docID, words := range documents {
		tfMap[docID] = CalculateTF(words)

		seen := make(map[string]bool)
		for _, word := range words {
			if !seen[word] {
				idfCount[word]++
				seen[word] = true
			}
		}
	}

	idfMap := make(map[string]float64)
	for word, count := range idfCount {
		idfMap[word] = math.Log(float64(totalDocs) / (1.0 + float64(count)))
	}

	return tfMap
}

// 🔍 **Find related notes based on TF-IDF similarity**
func FindRelatedNotes(fromZettel Zettel, zettels []Zettel, threshold float64, tfidfMap map[string]map[string]float64) []Zettel {
	var relatedNotes []Zettel

	// ✅ Get TF-IDF vector for `fromZettel`
	fromTFIDF, exists := tfidfMap[fromZettel.NoteID]
	if !exists {
		log.Printf("⚠️ No TF-IDF data found for note: %s", fromZettel.NoteID)
		return relatedNotes
	}

	// ✅ Compute similarity with other notes
	for _, zettel := range zettels {
		if zettel.NoteID == fromZettel.NoteID {
			continue // 🔥 Skip itself
		}

		// Get TF-IDF vector for the current note
		noteTFIDF, exists := tfidfMap[zettel.NoteID]
		if !exists {
			continue
		}

		// Compute cosine similarity
		similarity := CosineSimilarity(fromTFIDF, noteTFIDF)
		if similarity >= threshold {
			relatedNotes = append(relatedNotes, zettel)
		}
	}

	// ✅ Warn if no related notes were found
	if len(relatedNotes) == 0 {
		log.Printf("⚠️ No related notes found for: %s", fromZettel.NoteID)
	}

	return relatedNotes
}
