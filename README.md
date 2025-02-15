# Zettelkasten-cli (zk)

## 概要
Zettelkasten-cli (zk) は、Zettelkastenメソッドに基づいたノート管理をCLIで行うためのツールです。メモの作成、検索、リンク、タスク管理などを直感的なコマンドで操作できます。

## インストール方法
### 1. `go install` を使用する
以下のコマンドでインストールできます。
```sh
go install github.com/yourusername/zettelkasten-cli@latest
```

### 2. GitHub Releases からバイナリをダウンロード
#### Mac/Linux
```sh
wget https://github.com/yourusername/zettelkasten-cli/releases/latest/download/zk-linux-amd64 -O /usr/local/bin/zk
chmod +x /usr/local/bin/zk
```

#### Windows
1. [GitHub Releases](https://github.com/yourusername/zettelkasten-cli/releases) から最新の `zk-windows-amd64.exe` をダウンロード
2. `C:\Program Files\Zettelkasten-cli\zk.exe` などのディレクトリに配置し、環境変数 `PATH` に追加

## 必要な依存ツールのインストール
### MeCab
#### Mac/Linux
```sh
brew install mecab mecab-ipadic
```
#### Windows
[MeCab公式サイト](https://taku910.github.io/mecab/) からWindows用インストーラをダウンロードし、セットアップ

### fzf
#### Mac/Linux
```sh
brew install fzf
```
#### Windows
[公式リポジトリ](https://github.com/junegunn/fzf) からダウンロードし、環境変数 `PATH` に追加

## zk の設定 (`config.yaml`)
デフォルトの設定ファイルは `~/.config/zettelkasten-cli/config.yaml` に配置されます。
```yaml
# ノートを保存するディレクトリ
note_dir: "~/Library/Mobile Documents/com~apple~CloudDocs/Zettelkasten"

# 使用するエディタ
editor: "nvim"

# zettel.json:
zettel_json : "~/.config/zettelkasten-cli/zettel.json"

# アーカイブディレクトリ
archive_dir: "~/Library/Mobile Documents/com~apple~CloudDocs/Zettelkasten/archive"

# バックアップ設定
backup:
    enable: true
    frequency: 5  # 例: 5分ごとにバックアップを作成
    retention: 7   # 例: 過去7日分のバックアップを保持
    backup_dir: "~/Library/Mobile Documents/com~apple~CloudDocs/Zettelkasten/.backup"

# ゴミ箱設定
trash:
    frequency: 5
    retention: 30
    trash_dir: "~/.config/zettelkasten-cli/.Trash"
```

### 設定項目の説明
- `zettel.json` : ノートのメタデータを管理するJSONファイル
- `editor` : 使用するエディタ (vim, nvim, nano などを指定可能)
- `archive_dir` : アーカイブされたノートの保存ディレクトリ
- `backup_dir` : バックアップファイルの保存ディレクトリ
- `trash_dir` : 削除されたノートの保存ディレクトリ（一定期間後に完全削除）

## 実装済み機能とサンプルコマンド
### ノート作成・管理
- `zk new` (alias: `n`)
  - ノートを作成
  ```sh
  zk new "新しいメモのタイトル"
  ```
  - `--type (-t)`: ノートの種類を指定(`fleeting` / `literature` / `permanent` / `index` / `structure`)
  ```sh
  zk new -t fleeting "参考文献メモ"
  ```
  - `--tag`: タグを追加
  ```sh
  zk new --tag "devops","DevOpsについて"
  ```
- `zk show` (alias: `s`)
  - 指定したノートを表示
  ```sh
  zk show [id]
  ```
  - `--meta`: ノートのメタデータを表示
  ```sh
  zk show --meta [id]
  ```
- `zk list` (alias: `ls`)
  - すべてのノートを一覧表示
  ```sh
  zk list
  ```
  - `--tag` を使用してタグごとのノートを表示
  ```sh
  zk list --tag devops
  ```
- `zk edit` (alias: `e`): ノートをエディタで編集
  ```sh
  zk edit [id]
  ```
- `zk search` (alias: `f`)
  - キーワードで検索
  ```sh
  zk search "クラウド"
  ```
  - `--interactive`: 対話的検索
  ```sh
  zk search --interactive
  ```

### タスク管理
- `zk task add` (alias: `t a`): タスクを追加
  ```sh
  zk task add "新しいタスク"
  ```
- `zk task status` (alias: `t st`): タスクのステータスを変更(`Not started` / `In progress` / `Waiting` / `On hold` / `Done`)
  ```sh
  zk task status 1 "In progress"
  ```
- `zk task list` (alias: `t ls`)
  - `--limit`: 表示する件数を制限
  ```sh
  zk task list --limit 10
  ```

### ノート同期
- `zk sync` (alias: `sy`): ノートをクラウドと同期
  ```sh
  zk sync
  ```

## ライセンス
このプロジェクトは [MIT License](https://opensource.org/licenses/mit-license.php) のもとで提供されています。

## 貢献
バグ報告や機能追加の提案は [GitHub Issues](https://github.com/nakachan-ing/Zettelkasten-cli/issues) へお願いします。

