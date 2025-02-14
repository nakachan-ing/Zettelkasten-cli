package internal

import (
	"encoding/json"
	"fmt"
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

func ParseRipgrepOutput(output string) map[string][]string {
	results := make(map[string][]string) // ファイルごとにまとめる
	lines := strings.Split(output, "\n")

	var currentFile string

	for _, line := range lines {
		if line == "" {
			continue
		}

		var match RipgrepMatch
		err := json.Unmarshal([]byte(line), &match)
		if err != nil {
			continue
		}

		currentFile = match.Data.Path.Text
		text := strings.TrimSpace(match.Data.Lines.Text)

		if text == "" {
			continue // 空白行はスキップ
		}

		// `tags:` の行は `🏷️` をつける
		if strings.HasPrefix(text, "tags:") {
			text = fmt.Sprintf("🏷️ %s", text)
		} else if strings.HasPrefix(text, "- ") {
			text = fmt.Sprintf("   🔹 %s", text) // `tags:` のリストを見やすく
		} else if match.Type == "match" {
			text = fmt.Sprintf("🔍 %s", text) // `match` は強調
		} else {
			text = fmt.Sprintf("   → %s", text) // `context` の行
		}

		results[currentFile] = append(results[currentFile], text)
	}

	return results
}

func DisplayResults(results map[string][]string) {
	if len(results) == 0 {
		fmt.Println("該当するノートが見つかりませんでした")
		return
	}

	fmt.Println("\n🔍 検索結果:\n")

	for file, lines := range results {
		fmt.Printf("📄 %s\n", file) // ファイル名を一度だけ表示

		for _, line := range lines {
			formattedLine := strings.TrimSpace(line)

			// メタデータ (不要な情報) をスキップ
			if strings.HasPrefix(formattedLine, "links:") ||
				strings.HasPrefix(formattedLine, "created_at:") ||
				strings.HasPrefix(formattedLine, "updated_at:") {
				continue
			}

			// `tags:` の行を `🏷️` で見やすく
			if strings.HasPrefix(formattedLine, "tags:") {
				fmt.Println("   ", formattedLine)
			} else if strings.HasPrefix(formattedLine, "- ") {
				fmt.Println("     🔹", formattedLine) // `tags:` のリストをアイコン付きで表示
			} else if strings.HasPrefix(formattedLine, "##") {
				fmt.Println("   📌", formattedLine) // ノートのタイトルを見やすく
			} else {
				fmt.Println("   ", formattedLine)
			}
		}
		fmt.Println() // 各ファイルごとの改行
	}
}

func InteractiveSearch(results map[string][]string) {
	if len(results) == 0 {
		fmt.Println("該当するノートが見つかりませんでした")
		return
	}
	var fzfInput strings.Builder

	// ファイルごとに `match` と `context` を整理
	for file, lines := range results {
		for _, line := range lines {
			fzfInput.WriteString(fmt.Sprintf("%s:%s\n", file, line)) // `file:line` の形式に統一
		}
	}

	// デバッグ用
	// fmt.Println("🔍 fzf に渡すデータ:")
	// fmt.Println(fzfInput.String())

	// `fzf` の実行
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
			fmt.Println("🔹 fzf がユーザーによって中断されました (Ctrl+C)")
			return
		}
		fmt.Println("❌ fzf の実行に失敗しました:", err)
	}
}
