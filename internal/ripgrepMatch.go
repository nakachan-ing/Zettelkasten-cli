package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type RipgrepMatch struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text,omitempty"`
		} `json:"path"`
		LineNumber int `json:"line_number,omitempty"`
		Lines      struct {
			Text string `json:"text"`
		} `json:"lines,omitempty"`
		Submatches []struct {
			Match struct {
				Text string `json:"text"`
			} `json:"match"`
		} `json:"submatches"`
	} `json:"data"`
}

func ParseRipgrepOutput(output string) (map[string][]string, error) {
	results := make(map[string][]string) // Group results by file
	lines := strings.Split(output, "\n")

	var currentFile string

	for _, line := range lines {
		if line == "" {
			continue
		}

		var match RipgrepMatch
		err := json.Unmarshal([]byte(line), &match)
		if err != nil {
			log.Printf("⚠️ Failed to parse ripgrep output: %v", err)
			continue
		}

		currentFile = match.Data.Path.Text
		text := strings.TrimSpace(match.Data.Lines.Text)

		if text == "" {
			continue // Skip empty lines
		}

		// Format matched text for better readability
		if strings.HasPrefix(text, "tags:") {
			text = fmt.Sprintf("🏷️ %s", text)
		} else if strings.HasPrefix(text, "- ") {
			text = fmt.Sprintf("   🔹 %s", text) // List formatting
		} else if match.Type == "match" {
			text = fmt.Sprintf("🔍 %s", text) // Highlight matches
		} else {
			text = fmt.Sprintf("   → %s", text) // Context lines
		}

		results[currentFile] = append(results[currentFile], text)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no matching notes found")
	}

	return results, nil
}

func DisplayResults(results map[string][]string) {
	if len(results) == 0 {
		fmt.Println("❌ No matching notes found.")
		return
	}

	fmt.Println("\n🔍 Search Results:\n")

	for file, lines := range results {
		fmt.Printf("📄 %s\n", file) // Display file name once

		for _, line := range lines {
			formattedLine := strings.TrimSpace(line)

			// Skip unnecessary metadata
			if strings.HasPrefix(formattedLine, "links:") ||
				strings.HasPrefix(formattedLine, "created_at:") ||
				strings.HasPrefix(formattedLine, "updated_at:") {
				continue
			}

			// Improve readability
			if strings.HasPrefix(formattedLine, "tags:") {
				fmt.Println("   ", formattedLine)
			} else if strings.HasPrefix(formattedLine, "- ") {
				fmt.Println("     🔹", formattedLine) // List under tags
			} else if strings.HasPrefix(formattedLine, "##") {
				fmt.Println("   📌", formattedLine) // Highlight note titles
			} else {
				fmt.Println("   ", formattedLine)
			}
		}
		fmt.Println() // Separate files with a newline
	}
}

func InteractiveSearch(results map[string][]string) {
	if len(results) == 0 {
		fmt.Println("❌ No matching notes found.")
		return
	}

	var fzfInput strings.Builder

	// Prepare fzf input
	for file, lines := range results {
		for _, line := range lines {
			fzfInput.WriteString(fmt.Sprintf("%s:%s\n", file, line)) // `file:line` format
		}
	}

	// Execute `fzf`
	fzfCmd := exec.Command("fzf",
		"--delimiter", ":",
		"--preview", `file=$(printf "%s" {} | awk -F ":" '{print $1}'); file=$(realpath "$file"); [ -f "$file" ] && bat --color=always --style=header,grid --line-range :100 "$file"`,
		"--preview-window", "up:70%",
	)
	fzfCmd.Stdin = strings.NewReader(fzfInput.String())
	fzfCmd.Stdout = os.Stdout
	fzfCmd.Stderr = os.Stderr

	err := fzfCmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 130 {
			log.Println("🔹 fzf was interrupted by the user (Ctrl+C)")
			return
		}
		log.Printf("❌ Failed to execute fzf: %v", err)
	}
}
