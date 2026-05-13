# Roadmap

> 這是目前的開發方向，不代表一定會實現。優先順序可能隨時調整。歡迎透過 [Issues](https://github.com/vulcanshen/clerk/issues) 提出建議或回饋。

## 計畫中

### Open Summary Provider — 摘要引擎支援多 model provider
將摘要從 `claude -p` 解耦，支援 OpenAI-compatible API format，一個介面通吃多家 provider。

**動機：** 社群回饋最多的兩個需求 — token 成本太高、想用其他 provider。摘要本質上是 stateless 的 API call，不需要綁死 Claude。

**架構：**
- 保留 `claude -p` 作為預設（向後相容，零設定）
- 新增 `openai-compatible` provider — 支援任何兼容 `/v1/chat/completions` 的服務
- Config 設定 endpoint、model、API key

**支援範圍（OpenAI-compatible）：**
- 雲端：Gemini（免費）、Groq（免費）、OpenAI、Together AI、OpenRouter、Fireworks AI、Cloudflare Workers AI、DeepInfra
- 本地：Ollama、LM Studio、llama.cpp、vLLM

**分步實施：**
1. 基礎 — OpenAI-compatible HTTP client，config 擴充，provider 切換邏輯
2. Preset — `clerk config set summary.provider gemini` 自動填入 endpoint + 預設 model
3. Model 列表 — `clerk models` 呼叫 `/v1/models` 列出可用 model
4. 穩定性 — retry with backoff、rate limit 處理、額度用完通知、error message 優化

## 願景

### clerk import — 匯入任意檔案/目錄為摘要
將散落的筆記、文件、會議紀錄丟給 clerk，由 Claude 讀取並整理成結構化摘要，納入 clerk 管理。匯入後可透過 export、report、search 使用。

### Provider Adaptor — 解偶 AI provider
將 clerk 與 Claude Code 解偶，建立 adaptor 介面層。任何 AI coding tool 只要有對應的 adaptor，就能接入 clerk：hook 接收、transcript 解析、摘要產生，最終都整理到 clerk 統一的檔案結構中。目標支援 Gemini、Codex、Ollama 等。clerk 不只是 Claude Code 的 clerk，而是所有 AI session 的 clerk。
