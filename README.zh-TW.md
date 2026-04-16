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
- **增量合併** — 每次 session 合併到同一天同專案的單一摘要檔，不重複
- **對話過濾** — 過濾掉 tool call，只保留使用者與 AI 的對話文字
- **按日期整理** — 摘要存放在 `~/.clerk/YYYYMMDD/<專案-slug>.md`
- **游標追蹤** — 只處理上次之後的新訊息，節省 token 和時間
- **Process 管理** — 監控進行中的 feed、強制終止、重試中斷的摘要
- **可設定** — 輸出目錄、語言、模型、log 保留天數皆可自訂
- **一鍵設定** — `clerk hook install` 自動完成所有配置
- **遞迴防護** — 防止 clerk 呼叫 `claude -p` 時觸發無限循環
- 跨平台：macOS、Linux、Windows
- Shell 自動補全（bash、zsh、fish、powershell）

## 運作原理

當 Claude Code session 結束時，`SessionEnd` hook 觸發 `clerk feed`，流程如下：

1. Fork 到背景執行（hook 立刻返回）
2. 只讀取上次之後的新訊息（JSONL）
3. 載入該專案現有的當日摘要（如果有的話）
4. 呼叫 `claude -p` 產生合併後的摘要
5. 覆寫摘要檔為更新版本

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
| `config` | 顯示目前的設定（等同 `config show`） |
| `config show` | 顯示目前的設定與設定檔路徑 |
| `config set <key> <value>` | 設定配置值（key 可 tab 補全） |
| `hook install` | 將 clerk 安裝為 Claude Code SessionEnd hook |
| `hook uninstall` | 從 Claude Code SessionEnd hook 中移除 clerk |
| `status` | 顯示進行中的 feed process 和中斷的 session |
| `status --watch` | 即時重新整理狀態（每秒更新） |
| `retry <slug>` | 重試指定的中斷 session |
| `retry --all` | 重試所有中斷的 session |
| `kill <slug>` | 強制終止指定的 feed process |
| `kill --all` | 強制終止所有 feed process |

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
  },
  "log": {
    "retention_days": 30
  }
}
```

| 設定項 | 預設值 | 說明 |
|--------|--------|------|
| `output.dir` | `~/.clerk/` | 摘要存放根目錄 |
| `output.language` | `zh-TW` | 摘要輸出語言 |
| `summary.model` | `""`（使用 claude 預設） | `claude -p` 使用的模型 |
| `log.retention_days` | `30` | Log 和 cursor 檔案保留天數 |

用 `clerk config set` 設定：

```bash
clerk config set output.language en
clerk config set summary.model haiku
clerk config set log.retention_days 14
```

設定檔是選用的 — 不存在時 clerk 會使用預設值。

## 摘要格式

每個專案每天一份摘要檔，持續增量合併：

```markdown
# projects-my-app

> Last updated: 14:30:25

### 核心工作
- 實作 JWT token 的使用者認證
- 修復 WebSocket handler 的 race condition

### 輔助工作
- 新增 GitHub Actions CI pipeline
- 更新 README 的 API 文件

### 關鍵決策與理由
- **決策**：使用 JWT 而非 session → **理由**：多區域部署需要無狀態擴展

### 使用者備註
- 偏好最小抽象化，直接寫 code 而非依賴框架

### 版本紀錄
- v1.0.0 — 初始版本，包含認證和 WebSocket 支援
```

## 疑難排解

Log 存放在 `~/.clerk/.log/YYYYMMDD-clerk.log`，摘要沒出現時可以查看：

```bash
cat ~/.clerk/.log/$(date +%Y%m%d)-clerk.log
```

常見問題：

- **沒有產生摘要** — 確認 `claude` 是否在 PATH 中
- **Hook cancelled** — clerk 已改為 fork 背景執行來避免此問題，更新到最新版
- **內容重複** — 舊版行為；目前版本使用增量合併

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
