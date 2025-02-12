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

## 今後の予定
- `zk delete`（メモの削除）
- `zk archive`（アーカイブ機能）
- `zk task`（タスク管理機能）
- `zk project`（プロジェクト管理機能）

## インストール方法
（インストール方法が決まり次第追加）

## 貢献
バグ報告や機能提案は Issue にて受け付けています。

