# zettelkasten-cli

**zettelkasten-cli** は、Zettelkasten メソッドに基づいたメモ管理をコマンドラインで行うためのツールです。効率的に知識を整理し、関連するメモをリンクすることで、思考のネットワークを構築できます。

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

---

## **今後の予定**
- **メモのリマインダー機能**
- **日付ベースのタスク管理 (`zk task due`)**
- **自動タグ付け機能**
- **統計・分析機能 (`zk stats`)**

---

## **インストール方法**
（インストール方法が決まり次第追加）

---

## **貢献**
バグ報告や機能提案は Issue にて受け付けています。
