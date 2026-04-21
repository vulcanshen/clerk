# Roadmap

> 這是目前的開發方向，不代表一定會實現。優先順序可能隨時調整。歡迎透過 [Issues](https://github.com/vulcanshen/clerk/issues) 提出建議或回饋。

## 待辦

### clerk upgrade — 自動升級
新增 `clerk upgrade` 指令，偵測新版本後詢問是否升級，確認後自動執行。`clerk version` 偵測到新版本時會提示使用 `clerk upgrade`。支援 Homebrew、Scoop、go install、install script，涵蓋 macOS、Linux、Windows。

### report.instruction 設定
Report 專用的自訂 prompt，機制同 `summary.instruction`。讓使用者控制報告風格、語言或重點。

```
clerk config set report.instruction "加上 action items 和 blockers"
```

### 修正 tab completion 預設行為
不接受檔案參數的指令不再顯示檔案/目錄建議。

### 舊指令相容
指令更名或 deprecated 時，舊指令不再只印訊息，而是顯示遷移提示後繼續執行對應的新指令。例如 `clerk report` 改名為 `clerk summary` 後，輸入 `clerk report` 會提示「請使用 clerk summary」，但仍然正常執行。

## 願景

### clerk import — 匯入任意檔案/目錄為摘要
將散落的筆記、文件、會議紀錄丟給 clerk，由 Claude 讀取並整理成結構化摘要，納入 clerk 管理。匯入後可透過 export、report、search 使用。

### Provider Adaptor — 解偶 AI provider
將 clerk 與 Claude Code 解偶，建立 adaptor 介面層。任何 AI coding tool 只要有對應的 adaptor，就能接入 clerk：hook 接收、transcript 解析、摘要產生，最終都整理到 clerk 統一的檔案結構中。目標支援 Gemini、Codex、Ollama 等。clerk 不只是 Claude Code 的 clerk，而是所有 AI session 的 clerk。
