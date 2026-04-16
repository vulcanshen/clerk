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
- **会話フィルタリング** — tool call を除去し、ユーザーと AI の会話テキストのみを保持
- **日付整理** — 要約は `~/.clerk/YYYYMMDD/<プロジェクト-slug>.md` に保存
- **追記モード** — 同日同プロジェクトの複数セッションは同一ファイルに追記
- **設定可能** — 出力ディレクトリ、言語、モデルをカスタマイズ可能
- **ワンコマンド設定** — `clerk hook install` で全自動セットアップ
- **再帰防止** — clerk が `claude -p` を呼び出す際の無限ループを防止
- クロスプラットフォーム：macOS、Linux、Windows
- シェル補完（bash、zsh、fish、powershell）

## 仕組み

Claude Code セッションが終了すると、`SessionEnd` フックが `clerk feed` をトリガーします：

1. セッションのトランスクリプト（JSONL）を読み取り
2. tool call を除去し、会話テキストのみを保持
3. `claude -p` を呼び出して要約を生成
4. 日付別の markdown ファイルに追記

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
| `config show` | 現在の設定と設定ファイルのパスを表示 |
| `hook install` | clerk を Claude Code SessionEnd フックとしてインストール |
| `hook uninstall` | Claude Code SessionEnd フックから clerk を削除 |

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
  }
}
```

| 設定項目 | デフォルト値 | 説明 |
|----------|-------------|------|
| `output.dir` | `~/.clerk/` | 要約の保存ルートディレクトリ |
| `output.language` | `zh-TW` | 要約の出力言語 |
| `summary.model` | `""`（claude デフォルト） | `claude -p` で使用するモデル |

設定ファイルはオプションです — 存在しない場合、clerk はデフォルト値を使用します。

## 要約フォーマット

各要約はタイムスタンプ付きで追記されます：

```markdown
---
### 14:30:25

## ユーザー入力の要約
- auth.ts の認証バグ修正を依頼
- 新しいログインページコンポーネントの作成を依頼

## AI 応答の要約
- トークン検証の問題を発見・修正
- フォーム処理付きの LoginPage コンポーネントを作成
```

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
