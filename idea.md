# claude-auto-digest

> Claude Code session 結束時自動產生對話摘要，存成按日期組織的 markdown 檔案

## 要解決的問題

Claude Code 使用者一天可能開很多個 session，但事後無法以「某一天」為單位回顧做了什麼。現有的 `--resume` 和 `--continue` 是 session-centric 的，必須一個一個翻。使用者需要一個 time-centric 的全局視角，讓每週回顧時能快速掌握每天的工作內容。

目標使用者：每天使用 Claude Code 的開發者。

## 核心功能（MVP）

- [ ] SessionEnd hook（async），session 結束時背景觸發
- [ ] Shell script 從 hook stdin JSON 用 `jq` 抽出 `transcript_path` 和 `cwd`
- [ ] `cwd` 轉換為 slug 作為檔名（取最後兩層目錄，用 `-` 連接）
- [ ] 呼叫 `claude -p` 讀取 transcript jsonl 產生中文摘要
- [ ] 摘要追加到 `<root>/YYYYMMDD/<slug>.md`，用 `## HH:MM` 作為時段分隔
- [ ] 遞迴防護：環境變數 `WRAPUP_RUNNING` 防止 `claude -p` 的 SessionEnd 再次觸發
- [ ] 設定項：環境變數 `WRAPUP_ROOT` 指定儲存根目錄

## 技術選擇

- **實作**：Shell script（bash）
- **JSON 處理**：jq（外部依賴）
- **AI 摘要**：`claude -p`（Claude Code CLI 的 print mode，使用者現有額度）
- **Hook 機制**：Claude Code SessionEnd hook（async）
- **部署**：GitHub repo，使用者手動設定 hook 到 `~/.claude/settings.json`

## 架構概述

```
Claude Code Session 結束
  │
  ▼
SessionEnd Hook (async)
  │ stdin: { session_id, transcript_path, cwd, reason }
  │
  ▼
wrap-up.sh
  ├── 檢查 WRAPUP_RUNNING 環境變數（遞迴防護）
  ├── jq 抽出 transcript_path 和 cwd
  ├── cwd → slug（取最後兩層目錄）
  ├── mkdir -p $WRAPUP_ROOT/YYYYMMDD/
  └── claude -p < transcript.jsonl >> YYYYMMDD/slug.md
        │
        ▼
      claude -p 結束 → 觸發 SessionEnd
        → wrap-up.sh → WRAPUP_RUNNING=1 → exit（遞迴終止）
```

產出的檔案結構：

```
<WRAPUP_ROOT>/
  20260415/
    sideproj-ideas.md    # 當天所有該專案的 session 摘要
    other-project.md
  20260416/
    sideproj-ideas.md
```

## 已知風險與待釐清事項

- **`claude -p` 能否讀取 transcript jsonl** — 需實測確認 `claude -p` 的 stdin 是否能接受大型 jsonl 檔案，以及它能理解多少 transcript 的格式
- **transcript jsonl 的大小** — 長 session 的 jsonl 可能很大，`claude -p` 的 context window 是否足夠處理
- **slug 命名衝突** — 不同路徑可能產生相同的 slug（例如 `a/foo-bar` 和 `foo/bar`），機率低但存在
- **`claude -p` 的費用** — 每次 session 結束都會消耗 token，長 session 的 transcript 可能不便宜
- **jq 依賴** — macOS 預設沒有 jq，需要使用者自行安裝（brew install jq）

## 後續可擴充

- Obsidian wikilink 格式支援，串聯跨日跨專案的關聯
- 每日索引頁（`YYYYMMDD/index.md`）自動列出當天所有專案摘要
- 設定摘要語言（中文/英文）
- 自訂 prompt template（讓使用者決定摘要的格式和重點）
- 支援其他 AI CLI 工具（Gemini CLI 等）
- 週報自動產生（讀取整週的 daily digest 產出週摘要）

## Brainstorm 筆記

### 核心定位決策

- 不做「給 AI 讀的記憶」— 市場上已有 claude-mem、claude-memory-compiler 在做跨 session 記憶，但它們的輸出是給下一個 AI session 用的，不是給人看的
- 定位是「給人看的工作日誌」— time-centric 視角，讓開發者回顧自己的工作

### 捨棄的方案

1. **手動 skill（`/wrap-up`）** — Claude 有完整 context 可以直接產摘要，最簡單。但使用者堅持要全自動，不想每次手動跑
2. **Go binary** — 最初考慮用 Go 寫以避免 jq 依賴，但核心邏輯只是 parse JSON + 呼叫 claude -p + 寫檔，用 Go 殺雞用牛刀
3. **純 parse 不用 AI** — 直接把 jsonl 轉成可讀 markdown，免費且快速，但 session 內容太長不好讀，失去「摘要」的意義
4. **SessionStart + tail -f 持續監聽** — 在 session 開始時啟動背景程式持續轉換 transcript，但只是把「需要摘要」的問題延後了
5. **PreSessionEnd hook** — 不存在，Claude Code 沒有這個 event

### 市場調研結果

- 沒有工具做到「自動產出人類可讀的 session 工作日誌」
- 現有 Memory 類工具（claude-mem 等）輸出對象是 AI，不是人
- Standup 類工具（git-standup 等）只看 git commits，看不到 AI 對話
- Obsidian 整合工具（Nexus Importer 等）需要手動 export，且針對 claude.ai 網頁版
- Gemini CLI 社群有類似需求（GitHub issue #2554、#5101），尚未實作

### 命名由來

- 最初考慮 ccautosave，但「autosave」暗示存檔而非摘要
- 考慮過 sesdigest、devdigest、aidigest
- 最終選擇 claude-digest：直接表明是 Claude Code 的工具，digest 精確描述功能
