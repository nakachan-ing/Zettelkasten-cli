/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/nakachan-ing/Zettelkasten-cli/internal"
	"github.com/spf13/cobra"
)

var threshold float64
var manualFlag bool
var autoFlag bool

// Update `links:` field in front matter
func addLinkToFrontMatter(frontMatter *internal.FrontMatter, newLinks []string) *internal.FrontMatter {
	if frontMatter.Links == nil {
		frontMatter.Links = []string{}
	}

	for _, newLink := range newLinks {
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
	return frontMatter
}

// Allow user to select related notes
func selectRelatedNotes(relatedNotes []internal.Zettel) []string {
	var selected []string
	options := []string{}

	for _, note := range relatedNotes {
		options = append(options, fmt.Sprintf("%s: %s", note.ID, note.Title))
	}

	prompt := &survey.MultiSelect{
		Message: "Select notes to link:",
		Options: options,
	}

	survey.AskOne(prompt, &selected, nil)

	var selectedIDs []string
	for _, sel := range selected {
		selectedIDs = append(selectedIDs, strings.Split(sel, ": ")[0])
	}
	return selectedIDs
}

// Automatically link notes based on similarity
func autoLinkNotes(fromID string, threshold float64, config internal.Config, zettels []internal.Zettel, tfidfMap map[string]map[string]float64) error {
	cleanFromID := strings.TrimSpace(fromID)

	var fromZettel *internal.Zettel
	for i, zettel := range zettels {
		if strings.TrimSpace(zettel.ID) == cleanFromID {
			fromZettel = &zettels[i]
			break
		}
	}

	if fromZettel == nil {
		return fmt.Errorf("❌ Note with ID '%s' not found in zettels.json", cleanFromID)
	}

	fileID := fromZettel.NoteID
	filePath := fromZettel.NotePath

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("❌ Note file not found: %s", filePath)
	}

	relatedNotes := internal.FindRelatedNotes(*fromZettel, zettels, threshold, tfidfMap)
	if len(relatedNotes) == 0 {
		log.Printf("⚠️ No related notes found for: %s", fileID)
		return nil
	}

	selectedIDs := selectRelatedNotes(relatedNotes)
	if len(selectedIDs) == 0 {
		log.Println("⚠️ No notes selected. No links added.")
		return nil
	}

	for i := range zettels {
		if zettels[i].NoteID == fileID {
			zettels[i].Links = mergeUniqueLinks(zettels[i].Links, selectedIDs)
			break
		}
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("❌ Failed to read note: %w", err)
	}

	frontMatter, body, err := internal.ParseFrontMatter(string(content))
	if err != nil {
		return fmt.Errorf("❌ Failed to parse front matter: %w", err)
	}

	updatedFrontMatter := addLinkToFrontMatter(&frontMatter, selectedIDs)
	updatedContent := internal.UpdateFrontMatter(updatedFrontMatter, body)

	err = os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("❌ Failed to write updated note: %w", err)
	}

	internal.SaveUpdatedJson(zettels, &config)

	fmt.Printf("✅ Auto-linking completed: [%s] %s\n", fromZettel.NoteID, fromZettel.Title)
	return nil
}

// Merge existing and new links, removing duplicates
func mergeUniqueLinks(existingLinks, newLinks []string) []string {
	linkSet := make(map[string]bool)
	var merged []string

	for _, link := range existingLinks {
		linkSet[link] = true
		merged = append(merged, link)
	}

	for _, link := range newLinks {
		if !linkSet[link] {
			linkSet[link] = true
			merged = append(merged, link)
		}
	}

	return merged
}

func insertLinkInContext(filePath string, title string, fileName string, keywords []string) (string, error) {
	// ✅ Read note content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("❌ Failed to read note: %w", err)
	}
	text := string(content)

	// ✅ Separate front matter and body
	parts := strings.SplitN(text, "---", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("❌ Failed to parse front matter: missing `---` delimiters")
	}
	frontMatter := "---" + parts[1] + "---\n"
	body := parts[2]

	// ✅ Check if the link already exists
	markdownLink := fmt.Sprintf("[%s](%s)", title, fileName)
	if strings.Contains(body, markdownLink) {
		log.Printf("⚠️ Link already exists: %s", markdownLink)
		return text, nil
	}

	// ✅ Find keyword positions in the text
	var keywordPositions []string
	positionMap := make(map[string]int)
	for _, keyword := range keywords {
		re := regexp.MustCompile(fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(keyword)))
		loc := re.FindStringIndex(body)
		if loc != nil {
			position := fmt.Sprintf("After \"%s\"", keyword)
			keywordPositions = append(keywordPositions, position)
			positionMap[position] = loc[1] // Insertion position
		}
	}

	// ✅ Add an option to append to `### Links`
	keywordPositions = append(keywordPositions, "Append to ### Links")

	// ✅ Let the user select the insertion point
	var selectedPosition string
	prompt := &survey.Select{
		Message: "Select where to insert the link:",
		Options: keywordPositions,
	}
	err = survey.AskOne(prompt, &selectedPosition, nil)
	if err != nil {
		return "", fmt.Errorf("❌ Failed to get user input: %w", err)
	}

	// ✅ Insert the link at the chosen position
	inserted := false
	if selectedPosition != "Append to ### Links" {
		insertPos := positionMap[selectedPosition]
		body = body[:insertPos] + " " + markdownLink + body[insertPos:]
		inserted = true
		log.Printf("✅ Inserted link in context: %s", markdownLink)
	}

	// ✅ Append to `### Links` if no keyword was chosen
	if !inserted {
		if !strings.Contains(body, "### Links") {
			body += "\n\n### Links"
		}
		body += fmt.Sprintf("\n- %s", markdownLink)
		log.Printf("✅ Appended to `### Links`: %s", markdownLink)
	}

	// ✅ Return updated content
	updatedText := frontMatter + body
	return updatedText, nil
}

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Link notes together",
	Long: `Manually or automatically link notes.

Manual linking:
  zk link --manual <from> <to>

Automatic linking:
  zk link --auto <from>`,
	Args:    cobra.ArbitraryArgs,
	Aliases: []string{"ln"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// ✅ Prevent using both `--manual` and `--auto` at the same time
		if manualFlag && autoFlag {
			return fmt.Errorf("❌ Cannot use both `--manual` and `--auto` at the same time")
		}

		// ✅ Handle manual linking
		if manualFlag {
			if len(args) != 2 {
				return fmt.Errorf("❌ Usage: zk link --manual <from> <to>")
			}
			return runManualLink(args[0], args[1])
		}

		// ✅ Handle automatic linking
		if autoFlag {
			if len(args) != 1 {
				return fmt.Errorf("❌ Usage: zk link --auto <from>")
			}
			return runAutoLink(args[0])
		}

		return fmt.Errorf("❌ Please specify either `--manual` or `--auto`")
	},
}

// Manual linking of notes
func runManualLink(sourceId, destinationId string) error {
	config, err := internal.LoadConfig()
	if err != nil {
		return fmt.Errorf("❌ Failed to load config: %w", err)
	}

	zettels, err := internal.LoadJson(*config)
	if err != nil {
		return fmt.Errorf("❌ Failed to load JSON: %w", err)
	}

	var sourceZettel, destinationZettel *internal.Zettel
	for i := range zettels {
		if zettels[i].ID == sourceId {
			sourceZettel = &zettels[i]
		}
		if zettels[i].ID == destinationId {
			destinationZettel = &zettels[i]
		}
	}

	if sourceZettel == nil || destinationZettel == nil {
		return fmt.Errorf("❌ One or both notes not found: %s -> %s", sourceId, destinationId)
	}

	filePath := sourceZettel.NotePath
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("❌ Failed to read note: %w", err)
	}

	frontMatter, body, err := internal.ParseFrontMatter(string(content))
	if err != nil {
		return fmt.Errorf("❌ Failed to parse front matter: %w", err)
	}

	keyPhrases, err := internal.ExtractKeyPhrasesMeCab(string(body))
	if err != nil {
		return fmt.Errorf("❌ Failed to extract key phrases: %v", err)
	}

	updatedMarkdown, err := insertLinkInContext(filePath, destinationZettel.Title, destinationZettel.NoteID+".md", keyPhrases)
	if err != nil {
		return fmt.Errorf("❌ Failed to insert link into markdown: %v", err)
	}

	frontMatter, body, err = internal.ParseFrontMatter(updatedMarkdown)
	if err != nil {
		return fmt.Errorf("❌ Failed to parse front matter: %v", err)
	}

	updatedFrontMatter := addLinkToFrontMatter(&frontMatter, []string{destinationZettel.NoteID})
	finalMarkdown := internal.UpdateFrontMatter(updatedFrontMatter, body)

	err = os.WriteFile(filePath, []byte(finalMarkdown), 0644)
	if err != nil {
		return fmt.Errorf("❌ Failed to write updated note: %w", err)
	}

	sourceZettel.Links = mergeUniqueLinks(sourceZettel.Links, []string{destinationZettel.NoteID})
	internal.SaveUpdatedJson(zettels, config)

	fmt.Printf("✅ Linked [%s] %s to [%s] %s\n", sourceZettel.NoteID, sourceZettel.Title, destinationZettel.NoteID, destinationZettel.Title)
	return nil
}

func runAutoLink(fromID string) error {
	// ✅ Load configuration
	config, err := internal.LoadConfig()
	if err != nil {
		return fmt.Errorf("❌ Failed to load config: %w", err)
	}

	// ✅ Cleanup backups
	retention := time.Duration(config.Backup.Retention) * 24 * time.Hour
	if err := internal.CleanupBackups(config.Backup.BackupDir, retention); err != nil {
		log.Printf("⚠️ Backup cleanup failed: %v", err)
	}

	// ✅ Cleanup trash
	retention = time.Duration(config.Trash.Retention) * 24 * time.Hour
	if err := internal.CleanupTrash(config.Trash.TrashDir, retention); err != nil {
		log.Printf("⚠️ Trash cleanup failed: %v", err)
	}

	// ✅ Load `zettels.json`
	zettels, err := internal.LoadJson(*config)
	if err != nil {
		return fmt.Errorf("❌ Failed to load JSON file: %w", err)
	}

	// ✅ Compute TF-IDF for note similarity
	tfidfMap := internal.ComputeTFIDFForZettels(zettels)

	// ✅ Run auto-linking process
	if err := autoLinkNotes(fromID, threshold, *config, zettels, tfidfMap); err != nil {
		return fmt.Errorf("❌ Auto-linking failed: %w", err)
	}

	return nil
}

func init() {

	linkCmd.PersistentFlags().Float64VarP(&threshold, "threshold", "t", 0.5, "類似ノートをリンクするしきい値")
	rootCmd.AddCommand(linkCmd)
	linkCmd.Flags().BoolVarP(&manualFlag, "manual", "m", false, "手動でノートをリンク")
	linkCmd.Flags().BoolVar(&autoFlag, "auto", false, "関連ノートを自動でリンク")
}
