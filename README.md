# zettelkasten-cli

**zettelkasten-cli** は、Zettelkasten メソッドに基づいたメモ管理をコマンドラインで行うためのツールです。効率的に知識を整理し、関連するメモをリンクすることで、思考のネットワークを構築できます。

## 🚀 インストール方法

### **1. `go install` を使う（Go 環境がある場合）**
Go がインストールされている場合は、以下のコマンドで直接インストールできます。

```sh
go install github.com/yourname/zettelkasten-cli@latest
mv $(go env GOPATH)/bin/zettelkasten-cli $(go env GOPATH)/bin/zk
```

※ `$GOBIN` にバイナリが配置されるため、`PATH` に追加してください。

---

### **2. GitHub Releases からバイナリをダウンロード**
Go をインストールしていない場合は、[GitHub Releases](https://github.com/yourname/zettelkasten-cli/releases) から OS に対応したバイナリをダウンロードし、実行権限を付与して使うことができます。

**Mac/Linux**
```sh
wget https://github.com/yourname/zettelkasten-cli/releases/download/v1.0.0/zk-mac
mv zk-mac /usr/local/bin/zk
chmod +x /usr/local/bin/zk
```

**Windows（PowerShell）**
```sh
Invoke-WebRequest -Uri "https://github.com/yourname/zettelkasten-cli/releases/download/v1.0.0/zk.exe" -OutFile "C:\Program Files\zk.exe"
```

---

### **3. 必要な依存ツールのインストール**
`zettelkasten-cli` は、一部の機能で **MeCab** や **fzf** などのツールを利用するため、必要に応じてインストールしてください。

#### **`MeCab`（形態素解析：検索・リンク補完で使用）**
**Mac（Homebrew）**
```sh
brew install mecab mecab-ipadic
```
**Ubuntu/Debian**
```sh
sudo apt install mecab libmecab-dev mecab-ipadic
```

### **Windows**
1. [公式サイト](https://taku910.github.io/mecab/) から `mecab-0.996.exe` をダウンロード
2. インストール時に **「UTF-8 辞書 (`mecab-ipadic-utf8`)」** を選択
3. `mecab.exe` のパス `C:\Program Files\MeCab\bin` を環境変数 `PATH` に追加
4. インストール後、以下のコマンドで動作確認：
   ```powershell
   mecab --version
   echo "私はプログラマーです。" | mecab
   ```

#### **`fzf`（インタラクティブ検索）**
`zk list` や `zk search` で **インタラクティブにメモを選択** するために使用します。

**Mac（Homebrew）**
```sh
brew install fzf
```
**Ubuntu/Debian**
```sh
sudo apt install fzf
```

**Windows（Scoop）**
```sh
scoop install fzf
```

---

## 現在実装済みの機能

### 1. `zk new`
新しいメモを作成します。作成されたメモは一意のID（タイムスタンプ）を持ち、適切なメタデータが付与されます。

```sh
zk new "How to use Python in a Virtual Environment" --tags python,virtual-env
```

### 2. `zk show`
特定のメモの詳細を表示します。

```sh
zk show [id]
```

### 3. `zk list`
保存されているメモを一覧表示します。フィルタオプションを使用して、特定の種類（permanent, fleeting など）やタグで絞り込みが可能です。

```sh
zk list --type fleeting
zk list --tag Go
```
[![Image from Gyazo](https://i.gyazo.com/312c59c747b580254621c9f590d68fa8.png)](https://gyazo.com/312c59c747b580254621c9f590d68fa8)


### 4. `zk edit`
メモをエディタで開き、内容を編集できます。

```sh
zk edit [id]
```

### 5. `zk search`
メモの内容を検索します。`--context` オプションを使うと、指定した範囲のメタ情報を含めて検索結果を表示できます。
`--interactive` を追加することで、対話形式でファイル検索をすることができます。

```sh
zk search --type permanent --context 5
```
[![Image from Gyazo](https://i.gyazo.com/80f1085b56d9e6682af7e24c4cd1bbd9.png)](https://gyazo.com/80f1085b56d9e6682af7e24c4cd1bbd9)

```sh
zk search --interactive
```
[![Image from Gyazo](https://i.gyazo.com/59e6deccaf586bb5156bb6ab5599c6d3.png)](https://gyazo.com/59e6deccaf586bb5156bb6ab5599c6d3)


### 6. `zk link`
メモ同士をリンクし、ネットワークを構築できます。

```sh
zk link [link元id] [link先id]
```

## 🔥 **新機能: メモの整理とタスク管理**
### **7. `zk delete`（メモの削除）**
不要なメモを削除します。削除されたメモはゴミ箱 (`Trash/`) に移動され、`zk list --trash` で一覧表示できます。

```sh
zk delete [id]
```

**削除したメモを復元**
```sh
zk restore [id]
```

---

### **8. `zk archive`（アーカイブ機能）**
メモを削除せずにアーカイブし、通常の `zk list` で表示されないようにします。  
アーカイブされたメモは `archive/` フォルダに移動し、`zk list --archive` で一覧表示できます。

```sh
zk archive [id]
```

**アーカイブされたメモを復元**
```sh
zk restore [id]
```

---

### **9. `zk task`（タスク管理機能）**
メモをタスクとして管理し、`Not started` / `In progress` / `Wating` / `On hold` / `Done` のステータスを設定できます。

**タスクを一覧表示**
```sh
zk task list
```

**タスクのステータスを変更**
```sh
zk task status [id] "In progress"
zk task update [id] "Done"
```

---

### **10. `zk project`（プロジェクト管理機能）**
プロジェクトごとにメモを整理し、関連メモを簡単に管理できるようにします。

**新しいプロジェクトを作成**
```sh
zk project new "My DevOps Learning"
```

**プロジェクトにメモを追加**
```sh
zk project add [id] "My DevOps Learning"
```

**プロジェクトのメモを一覧表示**
```sh
zk list --tag project:My_DevOps_Learning
```

## **11. `zk sync`（メモと JSON の同期）**
`zk sync` は、Markdown ファイルと `zettel.json` を同期するためのコマンドです。

### **なぜ `zk sync` が必要か？**
- **ファイルを手動で編集・削除した場合**
  - `zk delete` や `zk archive` を使わずにメモを手動で移動・削除した場合、`zettel.json` のデータと実際のファイルが一致しなくなる。
- **外部ツールと連携した場合**
  - Obsidian やエディタで直接メモを変更した後に、`zk sync` を実行すると **JSON を最新の状態に更新** できる。

### **使い方**
#### **1. メモと `zettel.json` を同期**
```sh
zk sync
```
**すべてのメモをスキャンし、`zettel.json` の内容を最新化**

#### **2. メモのステータス（アーカイブ・ゴミ箱）を再確認**
```sh
zk sync --check-status
```
**削除済み or アーカイブ済みのメモが `zettel.json` に正しく反映されているかチェック**

#### **3. 不要なエントリを削除**
```sh
zk sync --clean
```
**既に存在しないメモのデータを `zettel.json` から削除**

## **貢献**
バグ報告や機能提案は Issue にて受け付けています。
