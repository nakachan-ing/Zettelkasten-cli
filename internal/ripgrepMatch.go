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

	// fmt.Println("🔍 Parsing ripgrep JSON output...") // デバッグログ

	var currentFile string
	var currentMatch string

	for _, line := range lines {
		if line == "" {
			continue
		}

		var match RipgrepMatch
		err := json.Unmarshal([]byte(line), &match)
		if err != nil {
			// fmt.Println("❌ JSON parse error:", err, "Line:", line)
			continue
		}

		// match.Type のデバッグ出力
		// fmt.Println("📝 match.Type:", match.Type)

		// `match` の場合
		if match.Type == "match" {
			currentFile = match.Data.Path.Text
			currentMatch = fmt.Sprintf("📄 %s:%d\n   → %s",
				currentFile, match.Data.LineNumber, strings.TrimSpace(match.Data.Lines.Text))
			results[currentFile] = append(results[currentFile], currentMatch)
		}

		// `context` の場合
		if match.Type == "context" {
			if currentFile != match.Data.Path.Text {
				continue // 直前の `match` と同じファイルでなければスキップ
			}

			if strings.TrimSpace(match.Data.Lines.Text) == "" {
				continue // 空白行はスキップ
			}

			// `→` の後にインデントして `context` の行を追加
			contextLine := fmt.Sprintf("   → %s", strings.TrimSpace(match.Data.Lines.Text))
			results[currentFile] = append(results[currentFile], contextLine)
		}
	}

	// fmt.Println("🔍 Final parsed results:", results) // デバッグ
	return results
}

func DisplayResults(results map[string][]string) {
	if len(results) == 0 {
		fmt.Println("該当するノートが見つかりませんでした")
		return
	}

	fmt.Println("\n🔍 検索結果:")

	// ファイルごとに結果を表示
	for file, lines := range results {
		fmt.Printf("\n📄 %s\n", file) // ファイル名を明示
		for _, line := range lines {
			fmt.Printf("   %s\n", line) // インデントを追加
		}
	}
}

func InteractiveSearch(results map[string][]string) {
	if len(results) == 0 {
		fmt.Println("該当するノートが見つかりませんでした")
		return
	}

	// fzf に渡すフォーマット
	var fzfInput strings.Builder

	// ファイルごとに `match` と `context` を整理
	for file, lines := range results {
		fzfInput.WriteString(fmt.Sprintf("📄 %s\n", file)) // ファイル名を追加
		for _, line := range lines {
			fzfInput.WriteString("   " + line + "\n") // インデントを追加
		}
		fzfInput.WriteString("\n") // ファイル間の区切り
	}

	// fzf に渡すデータが空なら実行しない
	if fzfInput.Len() == 0 {
		fmt.Println("🔹 fzf に渡すデータがありません")
		return
	}

	// `fzf` の `--preview` コマンド修正
	previewCmd := `file=$(echo {} | cut -d':' -f1); line=$(echo {} | cut -d':' -f2); ` +
		`[ -f "$file" ] && bat --style=plain --color=always --line-range $line:$((line+5)) "$file"`

	// `fzf` の実行
	fzfCmd := exec.Command("fzf",
		"--delimiter", ":",
		"--preview", previewCmd,
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
