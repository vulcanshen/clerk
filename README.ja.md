# clerk

[![GitHub Release](https://img.shields.io/github/v/release/vulcanshen/clerk)](https://github.com/vulcanshen/clerk/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vulcanshen/clerk)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/vulcanshen/clerk/ci.yml?label=CI)](https://github.com/vulcanshen/clerk/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanshen/clerk)](https://goreportcard.com/report/github.com/vulcanshen/clerk)
[![License](https://img.shields.io/github/license/vulcanshen/clerk)](LICENSE)

[English](README.md) | [繁體中文](README.zh-TW.md) | [한국어](README.ko.md)

Claude Code の書記官 — セッションを自動要約。

clerk は Claude Code の `SessionEnd` イベントにフックする CLI ツールで、会話終了時に自動的に要約を生成し、整理された markdown ファイルとして保存します。

## 機能

- **自動要約** — Claude Code セッション終了時に自動的に要約を生成
- **増分マージ** — 各セッションは同日同プロジェクトの単一要約ファイルにマージ、重複なし
- **会話フィルタリング** — tool call を除去し、ユーザーと AI の会話テキストのみを保持
- **日付整理** — 要約は `~/.clerk/YYYYMMDD/<プロジェクト-slug>.md` に保存
- **カーソル追跡** — 前回以降の新しいメッセージのみを処理し、トークンと時間を節約
- **プロセス管理** — アクティブな feed の監視、強制終了、中断されたものの再試行
- **設定可能** — 出力ディレクトリ、言語、モデル、ログ保持日数をカスタマイズ可能
- **ワンコマンド設定** — `clerk hook install` で全自動セットアップ
- **再帰防止** — clerk が `claude -p` を呼び出す際の無限ループを防止
- クロスプラットフォーム：macOS、Linux、Windows
- シェル補完（bash、zsh、fish、powershell）

## 仕組み

Claude Code セッションが終了すると、`SessionEnd` フックが `clerk feed` をトリガーします：

1. バックグラウンドにフォーク（フックは即座に返る）
2. 前回以降の新しいメッセージのみをトランスクリプト（JSONL）から読み取り
3. プロジェクトの既存の日次要約を読み込み（存在する場合）
4. `claude -p` を呼び出してマージされた要約を生成
5. 要約ファイルを更新版で上書き

```
~/.clerk/
└── 20260416/
    ├── projects-my-app.md
    ├── projects-api-server.md
    └── work-frontend.md
```

## インストール

### ワンライナーインストール

macOS / Linux / Git Bash：

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/install.sh | sh
```

Windows（PowerShell）：

```powershell
irm https://raw.githubusercontent.com/vulcanshen/clerk/main/install.ps1 | iex
```

アップデートは同じコマンドを再実行するだけです。アンインストール：

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/uninstall.sh | sh
```

```powershell
irm https://raw.githubusercontent.com/vulcanshen/clerk/main/uninstall.ps1 | iex
```

### パッケージマネージャー

| プラットフォーム | コマンド |
|------------------|----------|
| Homebrew (macOS / Linux) | `brew install vulcanshen/tap/clerk` |
| Scoop (Windows) | `scoop bucket add vulcanshen https://github.com/vulcanshen/scoop-bucket && scoop install clerk` |
| Debian / Ubuntu | `sudo dpkg -i clerk_<version>_linux_amd64.deb` |
| RHEL / Fedora | `sudo rpm -i clerk_<version>_linux_amd64.rpm` |

`.deb` と `.rpm` パッケージは [Releases ページ](https://github.com/vulcanshen/clerk/releases) からダウンロードできます。

### ソースからビルド

```bash
go install github.com/vulcanshen/clerk@latest
```

## クイックスタート

```bash
# SessionEnd フックをインストール
clerk hook install
```

以上です。フックをインストールしたら、clerk は完全にバックグラウンドで動作します — 手動操作も追加コマンドも不要です。Claude Code セッションを終了するたびに、要約が自動的に生成・保存されます。インストールしたら忘れて大丈夫です。

## コマンド一覧

| コマンド | 説明 |
|----------|------|
| `feed` | セッションのトランスクリプトを処理し要約を生成（フックから呼び出し） |
| `config` | 現在の設定を表示（`config show` のエイリアス） |
| `config show` | 現在の設定と設定ファイルのパスを表示 |
| `config set <key> <value>` | 設定値を変更（キーはタブ補完対応） |
| `hook install` | clerk を Claude Code SessionEnd フックとしてインストール |
| `hook uninstall` | Claude Code SessionEnd フックから clerk を削除 |
| `status` | アクティブな feed プロセスと中断されたセッションを表示 |
| `status --watch` | ステータスをリアルタイム更新（毎秒） |
| `retry <slug>` | 指定した中断セッションを再試行 |
| `retry --all` | すべての中断セッションを再試行 |
| `kill <slug>` | 指定したアクティブ feed プロセスを強制終了 |
| `kill --all` | すべてのアクティブ feed プロセスを強制終了 |

## 設定

設定ファイル：`~/.config/clerk/config.json`

```json
{
  "output": {
    "dir": "~/.clerk/",
    "language": "zh-TW"
  },
  "summary": {
    "model": ""
  },
  "log": {
    "retention_days": 30
  }
}
```

| 設定項目 | デフォルト値 | 説明 |
|----------|-------------|------|
| `output.dir` | `~/.clerk/` | 要約の保存ルートディレクトリ |
| `output.language` | `zh-TW` | 要約の出力言語 |
| `summary.model` | `""`（claude デフォルト） | `claude -p` で使用するモデル |
| `log.retention_days` | `30` | ログとカーソルファイルの保持日数 |

`clerk config set` で設定：

```bash
clerk config set output.language en
clerk config set summary.model haiku
clerk config set log.retention_days 14
```

設定ファイルはオプションです — 存在しない場合、clerk はデフォルト値を使用します。

## 要約フォーマット

各プロジェクトは1日1ファイル、増分的にマージされます：

```markdown
# projects-my-app

> Last updated: 14:30:25

### コア作業
- JWT トークンによるユーザー認証を実装
- WebSocket ハンドラーの競合状態を修正

### サポート作業
- GitHub Actions CI パイプラインを追加
- README の API ドキュメントを更新

### 主要な決定と理由
- **決定**：セッションではなく JWT を使用 → **理由**：マルチリージョンデプロイのためのステートレススケーリング

### ユーザーノート
- 最小限の抽象化を好み、フレームワークよりも直接コードを書くスタイル

### バージョンログ
- v1.0.0 — 認証と WebSocket サポートを含む初回リリース
```

## トラブルシューティング

ログは `~/.clerk/.log/YYYYMMDD-clerk.log` に保存されます。要約が生成されない場合に確認：

```bash
cat ~/.clerk/.log/$(date +%Y%m%d)-clerk.log
```

よくある問題：

- **要約が生成されない** — `claude` が PATH にあるか確認
- **Hook cancelled** — clerk はバックグラウンドフォークで対応済み。最新版にアップデート
- **内容が重複** — 旧バージョンの動作。現在は増分マージを使用

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
