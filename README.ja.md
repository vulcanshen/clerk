# clerk

[![GitHub Release](https://img.shields.io/github/v/release/vulcanshen/clerk)](https://github.com/vulcanshen/clerk/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vulcanshen/clerk)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/vulcanshen/clerk/ci.yml?label=CI)](https://github.com/vulcanshen/clerk/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanshen/clerk)](https://goreportcard.com/report/github.com/vulcanshen/clerk)
[![License](https://img.shields.io/github/license/vulcanshen/clerk)](LICENSE)

[English](README.md) | [繁體中文](README.zh-TW.md) | [한국어](README.ko.md)

Claude Code のセッションはターミナルを閉じると消えてしまいます。clerk があれば、自分が何をしたか見失うことはありません。

## 課題

Claude Code を毎日使っているなら、こんな壁にぶつかったことがあるはずです：

- **コンテキストの喪失** — `-c` や `--resume` を付け忘れて、ゼロからやり直し。前のセッションには完全なコンテキストがあったのに、大量のセッション ID の中から見つけ出すのは至難の業。
- **セッションの混乱** — 複数のプロジェクト、複数のセッション、すべてが並行稼働。今朝 API サーバーで何をした？認証の修正はどのセッションだった？全く思い出せない。
- **週報パニック** — 金曜の午後、週報の時間。`git log` を掘り返して、今週実際に何をしたか必死に思い出そうとする。
- **手動の記録管理** — Claude に「要約を保存して」と頼んだけど、前回は忘れた。あるいはセッションがクラッシュした。あるいはターミナルを閉じた。コンテキストは消えた。

これらはすべて一つのことに帰結します：**Claude Code はセッション間の記憶を持たない。そしてあなたが覚えておく必要もないはずです。**

## 解決策

```bash
clerk install
```

以上です。clerk は Claude Code に連携し、バックグラウンドで静かに動作します：

| 課題 | clerk の解決方法 |
|------|----------------|
| コンテキストの喪失 | `/clerk-resume` — 以前のセッションからコンテキストを即座に復元 |
| セッションの混乱 | プロジェクトごとの日次要約を自動生成、日付別に整理 |
| 週報 | 要約 + キーワードタグ = `/clerk-search` で検索可能な作業履歴 |
| 手動の記録管理 | 完全自動 — 覚えるコマンドなし、身につける習慣なし |

clerk は**一度設定したら忘れていい**ツールです。一度インストールするだけで、すべてのセッションが自動的に要約、追跡、タグ付け、検索可能になります。コンテキストが必要になったら、スラッシュコマンド一つで呼び出せます。

> **注意：** clerk は AI メモリツールではありません。claude-mem のようなツールは AI が思い出すためのコンテキストを保存します。clerk は**あなた**が読むための要約を保存します — 日付別に整理され、キーワードで検索でき、週次レビューにすぐ使えます。

## 機能

- **自動要約** — Claude Code セッション終了時に増分要約を自動生成
- **コンテキスト復元** — `/clerk-resume` で前回のセッションからコンテキストを再構築
- **キーワード検索** — `/clerk-search` でタグによる過去の作業検索
- **セッション追跡** — 履歴検索のためにすべてのセッション開始を記録
- **タグシステム** — 要約からキーワードを自動抽出し、検索可能なインデックスを構築
- **カーソル追跡** — 前回以降の新しいメッセージのみを処理し、トークンと時間を節約
- **プロセス管理** — アクティブな feed の監視、強制終了、中断されたものの再試行
- **プロジェクトレベル設定** — プロジェクトごとに feed を無効化、グローバル設定を上書き
- **ワンコマンド設定** — `clerk install` でフック、MCP サーバー、スキルを一括設定
- クロスプラットフォーム：macOS、Linux、Windows
- シェル補完（bash、zsh、fish、powershell）

## 仕組み

### インストールされるもの

| コンポーネント | 機能 |
|----------------|------|
| **hook** | SessionStart でセッション ID を記録、SessionEnd で要約生成をトリガー |
| **mcp** | `clerk-resume` と `clerk-search` ツールを提供する MCP stdio サーバー |
| **skills** | Claude Code 用の `/clerk-resume` と `/clerk-search` スラッシュコマンド |

### 要約フロー

1. セッション終了 → フックが `clerk feed` をトリガー
2. Feed がバックグラウンドにフォーク（フックは即座に返る）
3. 前回以降の新しいメッセージのみを読み取り（カーソル追跡）
4. 既存の日次要約を読み込み、`claude -p` を呼び出してマージ
5. 更新された要約を保存 + 検索インデックス用のタグを抽出

### 復元フロー

1. Claude Code で `/clerk-resume` と入力
2. Claude がプロジェクトの作業ディレクトリを指定して `clerk-resume` MCP ツールを呼び出す
3. clerk がファイルパスを返す：日次要約 + 完全なトランスクリプトファイル
4. Claude がまず要約を読み取り、概要を素早く把握
5. より詳細が必要な場合、Claude がトランスクリプトファイルを読み取る
6. Claude が以前の作業内容を要約し、コンテキストが復元されたことを確認

### 検索フロー

1. Claude Code で `/clerk-search` と入力
2. Claude が検索したいキーワードを尋ねる（または引数として直接指定）
3. Claude がそのキーワードで `clerk-search` MCP ツールを呼び出す
4. clerk がタグインデックスと照合し、一致する要約とトランスクリプトのパスを返す
5. Claude がそれらのファイルを読み取り、関連するコンテキストを提示

```
~/.clerk/
├── 20260416/
│   ├── projects-my-app.md
│   └── work-frontend.md
├── .sessions/
│   ├── projects-my-app.md
│   └── work-frontend.md
├── .tags/
│   ├── mcp.md
│   ├── refactor.md
│   └── auth.md
├── .log/
│   └── 20260416-clerk.log
├── .running/
└── .cursor/
```

## インストール

### クイックインストール

macOS / Linux / Git Bash：

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/install.sh | sh
```

Windows（PowerShell）：

```powershell
irm https://raw.githubusercontent.com/vulcanshen/clerk/main/install.ps1 | iex
```

次にフック、MCP サーバー、スキルを設定します：

```bash
clerk install
```

### パッケージマネージャー

| プラットフォーム | コマンド |
|------------------|----------|
| Homebrew（macOS / Linux） | `brew install vulcanshen/tap/clerk` |
| Scoop（Windows） | `scoop bucket add vulcanshen https://github.com/vulcanshen/scoop-bucket && scoop install clerk` |
| Debian / Ubuntu | `sudo dpkg -i clerk_<version>_linux_amd64.deb` |
| RHEL / Fedora | `sudo rpm -i clerk_<version>_linux_amd64.rpm` |

### ソースからビルド

```bash
go install github.com/vulcanshen/clerk@latest
```

## コマンド一覧

| コマンド | 説明 |
|----------|------|
| `install` | すべてのコンポーネントをインストール（hook + mcp + skills） |
| `install hook` | SessionStart/SessionEnd フックのみをインストール |
| `install mcp` | MCP サーバーのみを登録 |
| `install skills` | スラッシュコマンドスキルのみをインストール |
| `uninstall` | すべてのコンポーネントを削除 |
| `config` | 現在の設定を表示（`config show` のエイリアス） |
| `config show` | マージされた設定とファイルパスを表示 |
| `config set <key> <value>` | プロジェクトレベルの設定値を変更 |
| `config set -g <key> <value>` | グローバル設定値を変更 |
| `status` | アクティブな feed プロセスと中断されたセッションを表示 |
| `status --watch` | ステータスをリアルタイム更新（毎秒） |
| `retry <slug>` | 指定した中断セッションを再試行 |
| `retry --all` | すべての中断セッションを再試行 |
| `kill <slug>` | 指定したアクティブ feed プロセスを強制終了 |
| `kill --all` | すべてのアクティブ feed プロセスを強制終了 |

内部コマンド（フックから呼び出されるもので、ユーザーが直接使用するものではありません）：

| コマンド | 説明 |
|----------|------|
| `feed` | セッションのトランスクリプトを処理し要約を生成 |
| `punch` | セッション開始時にセッション ID を記録 |
| `mcp` | MCP stdio サーバーを起動 |

## 設定

### 設定ファイル

- グローバル：`~/.config/clerk/.clerk.json`
- プロジェクト：`<cwd>/.clerk.json`（グローバル設定を上書き）

### 利用可能な設定

```json
{
  "output": {
    "dir": "~/.clerk/",
    "language": "en"
  },
  "summary": {
    "model": ""
  },
  "log": {
    "retention_days": 30
  },
  "feed": {
    "enabled": true
  }
}
```

| 設定項目 | デフォルト値 | 説明 |
|----------|-------------|------|
| `output.dir` | `~/.clerk/` | 要約の保存ルートディレクトリ |
| `output.language` | `zh-TW` | 要約の出力言語 |
| `summary.model` | `""`（claude デフォルト） | `claude -p` で使用するモデル |
| `log.retention_days` | `30` | ログとカーソルファイルの保持日数 |
| `feed.enabled` | `true` | このプロジェクトの feed を有効/無効にする |

### 使用例

```bash
# 特定のプロジェクトで feed を無効化
cd /path/to/unimportant-project
clerk config set feed.enabled false

# グローバルでより安価なモデルを使用
clerk config set -g summary.model haiku

# グローバルで出力言語を変更
clerk config set -g output.language en
```

## MCP ツール

MCP サーバーのインストール後に利用可能（`clerk install mcp`）：

| ツール | 説明 |
|--------|------|
| `clerk-resume` | コンテキスト復元のための要約 + トランスクリプトファイルパスを返す |
| `clerk-search` | キーワード/タグで過去のセッションを検索 |

## スキル

スキルのインストール後に利用可能（`clerk install skills`）：

| スキル | 説明 |
|--------|------|
| `/clerk-resume` | 前回のセッションからコンテキストを復元 — MCP ツールを呼び出し、ファイルを読み込み、コンテキストを再構築 |
| `/clerk-search` | キーワードで過去のセッションを検索 — MCP ツールを呼び出し、一致するファイルを読み込み |

## トラブルシューティング

ログは `~/.clerk/.log/YYYYMMDD-clerk.log` に保存されます：

```bash
cat ~/.clerk/.log/$(date +%Y%m%d)-clerk.log
```

よくある問題：

- **要約が生成されない** — `claude` が PATH にあるか確認
- **Hook cancelled** — clerk はバックグラウンドフォークで対応済み。最新版にアップデート
- **MCP ツールが見つからない** — `clerk install mcp` を実行してセッションを再起動

## シェル補完

```bash
# Zsh
mkdir -p ~/.zsh/completions
clerk completion zsh > ~/.zsh/completions/_clerk
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -Uz compinit && compinit' >> ~/.zshrc
source ~/.zshrc

# Bash
clerk completion bash > /etc/bash_completion.d/clerk

# Fish
clerk completion fish > ~/.config/fish/completions/clerk.fish

# PowerShell
clerk completion powershell > clerk.ps1
```

## ライセンス

[GPL-3.0](LICENSE)
