# GOAL.md

> 本檔案定義此分支（openai-compatible-api-summary-provider）的開發目標。

## 目標

將摘要引擎從 `claude -p` 解耦，支援 OpenAI-compatible API format，讓使用者可選擇 Groq、Gemini、OpenAI、Ollama 等任何兼容 `/chat/completions` 的服務。

## 核心功能

### Provider 抽象層
- [x] `Provider` interface（`Complete(ctx, prompt, systemPrompt)`）
- [x] `ClaudeCliProvider` — 現有 `claude -p`，預設
- [x] `OpenAIProvider` — HTTP POST to `/chat/completions`
- [x] 移除 `feed.CallClaude`，所有呼叫端改用 Provider

### Config
- [x] 三層結構：`summary.provider.name/model/endpoint/api_key`
- [x] `report.provider.*` 獨立設定，fallback 到 `summary.provider.*`
- [x] `summary.provider.model` 驗證放寬（非 claude provider 接受任意 model name）
- [x] API key 所有顯示處遮蔽（config show、config set、register、--json）

### Preset
- [x] 內建 preset（groq、gemini、openai、ollama）— 自動填入 endpoint + 預設 model
- [x] 使用者只需 `provider.name` + `api_key` 兩行設定

### 指令
- [x] `clerk provider` — 列出支援的 provider + 預設設定
- [x] `clerk provider models <name>` — 查詢 provider 的可用模型

### 移除
- [x] `logs --mask` 移除（使用者自行編輯 raw output）

## 穩定化

### 測試
- [x] `internal/provider/` 單元測試（resolve、preset、openai mock、retry、timeout、models）
- [x] 既有測試通過（`go test ./...`）

### Provider 穩定性
- [x] retry with exponential backoff（429 / 5xx）
- [x] API 額度用完的明確錯誤訊息
- [x] `claude -p` provider 處理 Claude CLI 不存在的情況

### 文件
- [x] CHANGELOG [Unreleased] 更新
- [x] 4 個 README config table、command table、token disclosure 更新
- [x] ROADMAP 更新

## 驗收標準

- [x] `go test ./...` 全通過
- [x] `go vet ./...` 零 warning
- [x] Groq end-to-end 測試通過（register + provider models）
- [ ] Claude 預設行為不變（向後相容）— 需要 brew reinstall 後實測
- [ ] 跨平台 CI 通過（macOS + Linux + Windows）— merge 後 CI 驗證
- [ ] merge 回 main，release
