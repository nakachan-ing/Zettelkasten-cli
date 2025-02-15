# Zettelkasten-cli (zk)

## Overview
Zettelkasten-cli (zk) is a command-line tool for managing notes based on the Zettelkasten method. It allows intuitive commands for creating notes, searching, linking, and task management.

## Installation
### 1. Using `go install`
Install with the following command:
```sh
go install github.com/nakachan-ing/Zettelkasten-cli@latest
```

### 2. Downloading from GitHub Releases
#### Mac/Linux
```sh
wget https://github.com/nakachan-ing/Zettelkasten-cli/releases/download/v2.0.0/zk-mac -O /usr/local/bin/zk
chmod +x /usr/local/bin/zk

wget https://github.com/nakachan-ing/Zettelkasten-cli/releases/download/v2.0.0/zk-linux -O /usr/local/bin/zk
chmod +x /usr/local/bin/zk
```

#### Windows
1. Download the latest `zk-windows-amd64.exe` from [GitHub Releases](https://github.com/nakachan-ing/Zettelkasten-cli/releases)
2. Place it in a directory like `C:\Program Files\Zettelkasten-cli\zk.exe` and add it to the `PATH` environment variable

## Required Dependencies
### MeCab
#### Mac/Linux
```sh
brew install mecab mecab-ipadic
```
#### Windows
Download and install from the [MeCab official site](https://taku910.github.io/mecab/)

### fzf
#### Mac/Linux
```sh
brew install fzf
```
#### Windows
Download from the [official repository](https://github.com/junegunn/fzf) and add it to the `PATH`

## zk Configuration (`config.yaml`)
The default configuration file is located at `~/.config/zettelkasten-cli/config.yaml`.
```yaml
# Directory for storing notes
note_dir: "~/Library/Mobile Documents/com~apple~CloudDocs/Zettelkasten"

# Default text editor
editor: "nvim"

# Metadata file
zettel_json : "~/.config/zettelkasten-cli/zettel.json"

# Archive directory
archive_dir: "~/Library/Mobile Documents/com~apple~CloudDocs/Zettelkasten/archive"

# Backup settings
backup:
    enable: true
    frequency: 5  # Create a backup every 5 minutes
    retention: 7   # Keep backups for 7 days
    backup_dir: "~/Library/Mobile Documents/com~apple~CloudDocs/Zettelkasten/.backup"

# Trash settings
trash:
    frequency: 5
    retention: 30
    trash_dir: "~/.config/zettelkasten-cli/.Trash"
```

### Configuration Explanation
- `zettel.json`: Stores metadata of notes
- `editor`: Specifies the text editor (vim, nvim, nano, etc.)
- `archive_dir`: Directory for archived notes
- `backup_dir`: Directory for backup files
- `trash_dir`: Directory for deleted notes (permanently deleted after a retention period)

## Implemented Features and Sample Commands
### Note Creation and Management
- `zk new` (alias: `n`)
  - Create a new note
  ```sh
  zk new "New Note Title"
  ```
  - `--type (-t)`: Specify the type of note(`fleeting` / `literature` / `permanent` / `index` / `structure`)
  ```sh
  zk new -t reference "Reference Note"
  ```
  - `--tag`: Add tags
  ```sh
  zk new --tag "devops","About DevOps"
  ```
- `zk show` (alias: `s`)
  - Show a specific note
  ```sh
  zk show [id]
  ```
  - `--meta`: Show metadata
  ```sh
  zk show --meta [id]
  ```
- `zk list` (alias: `ls`)
  - List all notes
  ```sh
  zk list
  ```
  - List notes by tag
  ```sh
  zk list --tag devops
  ```
- `zk edit` (alias: `e`): Edit a note
  ```sh
  zk edit [id]
  ```
- `zk search` (alias: `f`)
  - Search by keyword
  ```sh
  zk search "cloud"
  ```
  - `--interactive`: Interactive search
  ```sh
  zk search --interactive
  ```

### Task Management
- `zk task add` (alias: `t a`): Add a task
  ```sh
  zk task add "New Task"
  ```
- `zk task status` (alias: `t st`): Change task status(`Not started` / `In progress` / `Waiting` / `On hold` / `Done`)
  ```sh
  ```sh
  zk task status 1 doing
  ```
- `zk task list` (alias: `t ls`)
  - `--limit`: Limit the number of displayed tasks
  ```sh
  zk task list --limit 10
  ```

### Note Synchronization
- `zk sync` (alias: `sy`): Sync notes with the cloud
  ```sh
  zk sync
  ```

## License
This project is provided under the [MIT License](https://opensource.org/licenses/mit-license.php).

## Contributions
Bug reports and feature requests are welcome at [GitHub Issues](https://github.com/nakachan-ing/Zettelkasten-cli/issues).

