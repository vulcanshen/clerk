# clerk

[![GitHub Release](https://img.shields.io/github/v/release/vulcanshen/clerk)](https://github.com/vulcanshen/clerk/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vulcanshen/clerk)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/vulcanshen/clerk/ci.yml?label=CI)](https://github.com/vulcanshen/clerk/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanshen/clerk)](https://goreportcard.com/report/github.com/vulcanshen/clerk)
[![License](https://img.shields.io/github/license/vulcanshen/clerk)](LICENSE)

[English](README.md) | [日本語](README.ja.md) | [한국어](README.ko.md)

你的 Claude Code session 關掉終端機就消失了。clerk 確保你不會忘記自己做過什麼。

## 問題

如果你每天都在用 Claude Code，你一定遇過這些痛點：

- **上下文遺失** — 你忘了加 `-c` 或 `--resume`，然後一切從頭開始。上一個 session 明明有完整的上下文，但要從一堆 session ID 裡找回來？祝你好運。
- **Session 混亂** — 多個專案、多個 session，同時並行。今天早上在 API server 做了什麼？哪個 session 有那個 auth 修復？完全想不起來。
- **週報恐慌** — 週五下午，該寫週報了。你翻遍 `git log`，試圖拼湊出這週到底做了什麼。
- **手動記錄** — 你叫 Claude「存個摘要」，但上次忘了。或是 session 當掉了。或是你關了終端機。上下文就這樣消失了。

這些問題歸根結底就是一件事：**Claude Code 不會跨 session 記憶，而你不應該需要自己記。**

## 解決方案

```bash
clerk install
```

就這樣。clerk 掛載到 Claude Code 後，會在背景安靜地運作：

| 痛點 | clerk 如何解決 |
|------|--------------|
| 上下文遺失 | `/clerk-resume` — 立即從任何之前的 session 恢復上下文 |
| Session 混亂 | 按專案自動產生每日摘要，依日期整理 |
| 週報 | 摘要 + 關鍵字標籤 = 可搜尋的工作歷史，透過 `/clerk-search` |
| 手動記錄 | 全自動 — 不需要記任何指令，不需要養成任何習慣 |

clerk 是一個**裝完就忘**的工具。安裝一次，每個 session 都會自動摘要、追蹤、標記、並可搜尋。當你需要找回上下文時，只要一個斜線指令就行。

> **注意：** clerk 不是 AI 記憶工具。像 claude-mem 這類工具是儲存上下文讓 AI 回憶用的。clerk 儲存摘要是給**你**看的 — 按日期整理、可用關鍵字搜尋、隨時可用於你的週報。

## 功能特色

- **自動摘要** — Claude Code session 結束時自動產生增量摘要
- **上下文恢復** — `/clerk-resume` 從之前的 session 重建上下文
- **關鍵字搜尋** — `/clerk-search` 透過標籤搜尋過去的工作記錄
- **Session 追蹤** — 記錄每次 session 開始，供歷史查詢
- **標籤系統** — 自動從摘要中擷取關鍵字，建立可搜尋的索引
- **游標追蹤** — 只處理上次之後的新訊息，節省 token 和時間
- **Process 管理** — 監控進行中的 feed、強制終止、重試中斷的摘要
- **專案層級設定** — 按專案停用 feed，覆蓋全域設定
- **一鍵設定** — `clerk install` 自動配置 hook、MCP server 和 skills
- 跨平台：macOS、Linux、Windows
- Shell 自動補全（bash、zsh、fish、powershell）

## 運作原理

### 安裝了什麼

| 元件 | 功能 |
|------|------|
| **hook** | SessionStart 記錄 session ID，SessionEnd 觸發摘要產生 |
| **mcp** | MCP stdio server，提供 `clerk-resume` 和 `clerk-search` 工具 |
| **skills** | `/clerk-resume` 和 `/clerk-search` 斜線指令供 Claude Code 使用 |

### 摘要流程

1. Session 結束 → hook 觸發 `clerk feed`
2. Feed fork 到背景執行（hook 立刻返回）
3. 只讀取上次之後的新訊息（游標追蹤）
4. 載入現有的每日摘要，呼叫 `claude -p` 進行合併
5. 儲存更新後的摘要 + 擷取標籤供搜尋索引使用

### 恢復流程

1. 你在 Claude Code 中輸入 `/clerk-resume`
2. Claude 以你專案的工作目錄呼叫 `clerk-resume` MCP 工具
3. clerk 回傳檔案路徑：每日摘要 + 完整 transcript 檔案
4. Claude 先讀取摘要以快速了解概況
5. 如需更多細節，Claude 會讀取 transcript 檔案
6. Claude 總結之前完成的工作，並確認上下文已恢復

### 搜尋流程

1. 你在 Claude Code 中輸入 `/clerk-search`
2. Claude 詢問你要搜尋的關鍵字（或你以參數直接提供）
3. Claude 以該關鍵字呼叫 `clerk-search` MCP 工具
4. clerk 比對標籤索引，回傳符合的摘要和 transcript 路徑
5. Claude 讀取這些檔案並呈現相關的上下文

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

然後設定 hook、MCP server 和 skills：

```bash
clerk install
```

### 套件管理器

| 平台 | 指令 |
|------|------|
| Homebrew（macOS / Linux） | `brew install vulcanshen/tap/clerk` |
| Scoop（Windows） | `scoop bucket add vulcanshen https://github.com/vulcanshen/scoop-bucket && scoop install clerk` |
| Debian / Ubuntu | `sudo dpkg -i clerk_<version>_linux_amd64.deb` |
| RHEL / Fedora | `sudo rpm -i clerk_<version>_linux_amd64.rpm` |

### 從原始碼安裝

```bash
go install github.com/vulcanshen/clerk@latest
```

## 指令

| 指令 | 說明 |
|------|------|
| `install` | 安裝所有元件（hook + mcp + skills） |
| `install hook` | 僅安裝 SessionStart/SessionEnd hook |
| `install mcp` | 僅註冊 MCP server |
| `install skills` | 僅安裝斜線指令 skills |
| `uninstall` | 移除所有元件 |
| `config` | 顯示目前的設定（等同 `config show`） |
| `config show` | 顯示合併後的設定與檔案路徑 |
| `config set <key> <value>` | 設定專案層級的配置值 |
| `config set -g <key> <value>` | 設定全域配置值 |
| `status` | 顯示進行中的 feed process 和中斷的 session |
| `status --watch` | 即時重新整理狀態（每秒更新） |
| `retry <slug>` | 重試指定的中斷 session |
| `retry --all` | 重試所有中斷的 session |
| `kill <slug>` | 強制終止指定的 feed process |
| `kill --all` | 強制終止所有 feed process |

內部指令（由 hook 呼叫，非使用者直接使用）：

| 指令 | 說明 |
|------|------|
| `feed` | 處理 session 對話記錄並產生摘要 |
| `punch` | 在 session 開始時記錄 session ID |
| `mcp` | 啟動 MCP stdio server |

## 設定

### 設定檔

- 全域：`~/.config/clerk/.clerk.json`
- 專案：`<cwd>/.clerk.json`（覆蓋全域設定）

### 可用設定

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

| 設定項 | 預設值 | 說明 |
|--------|--------|------|
| `output.dir` | `~/.clerk/` | 摘要存放根目錄 |
| `output.language` | `en` | 摘要輸出語言 |
| `summary.model` | `""`（使用 claude 預設） | `claude -p` 使用的模型 |
| `log.retention_days` | `30` | Log 和 cursor 檔案保留天數 |
| `feed.enabled` | `true` | 啟用/停用此專案的 feed |

### 範例

```bash
# 停用特定專案的 feed
cd /path/to/unimportant-project
clerk config set feed.enabled false

# 全域使用較便宜的模型
clerk config set -g summary.model haiku

# 全域變更輸出語言
clerk config set -g output.language en
```

## MCP 工具

安裝 MCP server 後可用（`clerk install mcp`）：

| 工具 | 說明 |
|------|------|
| `clerk-resume` | 回傳摘要 + transcript 檔案路徑，用於恢復上下文 |
| `clerk-search` | 透過關鍵字/標籤搜尋之前的 session |

## Skills

安裝 skills 後可用（`clerk install skills`）：

| Skill | 說明 |
|-------|------|
| `/clerk-resume` | 從之前的 session 恢復上下文 — 呼叫 MCP 工具、讀取檔案、重建上下文 |
| `/clerk-search` | 透過關鍵字搜尋過去的 session — 呼叫 MCP 工具、讀取符合的檔案 |

## 疑難排解

Log 存放在 `~/.clerk/.log/YYYYMMDD-clerk.log`：

```bash
cat ~/.clerk/.log/$(date +%Y%m%d)-clerk.log
```

常見問題：

- **沒有產生摘要** — 確認 `claude` 是否在 PATH 中
- **Hook cancelled** — clerk 已改為 fork 背景執行來避免此問題，更新到最新版
- **MCP 工具找不到** — 執行 `clerk install mcp` 並重新啟動 session

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
