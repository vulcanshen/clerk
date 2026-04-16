# clerk

[![GitHub Release](https://img.shields.io/github/v/release/vulcanshen/clerk)](https://github.com/vulcanshen/clerk/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vulcanshen/clerk)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/vulcanshen/clerk/ci.yml?label=CI)](https://github.com/vulcanshen/clerk/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanshen/clerk)](https://goreportcard.com/report/github.com/vulcanshen/clerk)
[![License](https://img.shields.io/github/license/vulcanshen/clerk)](LICENSE)

[English](README.md) | [日本語](README.ja.md) | [한국어](README.ko.md)

Claude Code 的書記官 — 自動摘要你的對話。

clerk 是一個 CLI 工具，掛載在 Claude Code 的 `SessionEnd` 事件上，在對話結束時自動產生摘要並存成有組織的 markdown 檔案。

## 功能特色

- **自動摘要** — Claude Code session 結束時自動產生摘要
- **對話過濾** — 過濾掉 tool call，只保留使用者與 AI 的對話文字
- **按日期整理** — 摘要存放在 `~/.clerk/YYYYMMDD/<專案-slug>.md`
- **附加模式** — 同一天同一個專案的多次 session 會附加到同一檔案
- **可設定** — 輸出目錄、語言、模型皆可自訂
- **一鍵設定** — `clerk hook install` 自動完成所有配置
- **遞迴防護** — 防止 clerk 呼叫 `claude -p` 時觸發無限循環
- 跨平台：macOS、Linux、Windows
- Shell 自動補全（bash、zsh、fish、powershell）

## 運作原理

當 Claude Code session 結束時，`SessionEnd` hook 觸發 `clerk feed`，流程如下：

1. 讀取 session 對話記錄（JSONL）
2. 過濾掉 tool call，只保留對話文字
3. 呼叫 `claude -p` 產生摘要
4. 將摘要附加到按日期組織的 markdown 檔案

```
~/.clerk/
└── 20260416/
    ├── projects-my-app.md
    ├── projects-api-server.md
    └── work-frontend.md
```

## 安裝

### 快速安裝

macOS / Linux / Git Bash：

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/install.sh | sh
```

Windows（PowerShell）：

```powershell
irm https://raw.githubusercontent.com/vulcanshen/clerk/main/install.ps1 | iex
```

更新只需重新執行相同指令。解除安裝：

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/uninstall.sh | sh
```

```powershell
irm https://raw.githubusercontent.com/vulcanshen/clerk/main/uninstall.ps1 | iex
```

### 套件管理器

| 平台 | 指令 |
|------|------|
| Homebrew（macOS / Linux） | `brew install vulcanshen/tap/clerk` |
| Scoop（Windows） | `scoop bucket add vulcanshen https://github.com/vulcanshen/scoop-bucket && scoop install clerk` |
| Debian / Ubuntu | `sudo dpkg -i clerk_<version>_linux_amd64.deb` |
| RHEL / Fedora | `sudo rpm -i clerk_<version>_linux_amd64.rpm` |

`.deb` 和 `.rpm` 套件可從 [Releases 頁面](https://github.com/vulcanshen/clerk/releases)下載。

### 從原始碼安裝

```bash
go install github.com/vulcanshen/clerk@latest
```

## 快速開始

```bash
# 安裝 SessionEnd hook
clerk hook install
```

就這樣。安裝完 hook 之後，clerk 會完全在背景運作 — 不需要手動操作，不需要額外指令。每次你結束 Claude Code session，摘要就會自動產生並存檔。裝完就可以忘了它。

## 指令

| 指令 | 說明 |
|------|------|
| `feed` | 處理 session 對話記錄並產生摘要（由 hook 呼叫） |
| `config show` | 顯示目前的設定與設定檔路徑 |
| `hook install` | 將 clerk 安裝為 Claude Code SessionEnd hook |
| `hook uninstall` | 從 Claude Code SessionEnd hook 中移除 clerk |

## 設定

設定檔路徑：`~/.config/clerk/config.json`

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

| 設定項 | 預設值 | 說明 |
|--------|--------|------|
| `output.dir` | `~/.clerk/` | 摘要存放根目錄 |
| `output.language` | `zh-TW` | 摘要輸出語言 |
| `summary.model` | `""`（使用 claude 預設） | `claude -p` 使用的模型 |

設定檔是選用的 — 不存在時 clerk 會使用預設值。

## 摘要格式

每份摘要以時間戳分隔附加：

```markdown
---
### 14:30:25

## 使用者輸入摘要
- 要求修復 auth.ts 中的認證 bug
- 要求建立新的登入頁面元件

## AI 回應摘要
- 找到並修復了 token 驗證問題
- 建立了 LoginPage 元件並處理表單邏輯
```

## Shell 自動補全

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

## 授權條款

[GPL-3.0](LICENSE)
