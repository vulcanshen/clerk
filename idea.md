# clerk (The Claude Code Clerk)

> Claude Code session 結束時自動幫你整理對話摘要的「書記官」，存成按日期組織的 markdown 檔案。

## 核心指令 (MVP)

- [ ] **clerk feed**：書記官核心指令，對接 `SessionEnd` hook。
  - 從 `stdin` 讀取 JSON 並過濾對話文字（僅保留 `user` 與 `assistant`）。
  - 將 `cwd` 轉為全小寫的 slug（取 `~` 之後的完整路徑）。
  - 呼叫 `claude -p` 產生摘要（注入 `CLERK_INTERNAL=1` 防遞迴）。
  - 存檔至 `<output_dir>/YYYYMMDD/<slug>.md`。
- [ ] **clerk config show**：印出目前生效的配置與配置路徑。
  - 配置路徑：`~/.config/clerk/config.json`。
  - 預設值：`output_dir: ~/.clerk/`, `output_language: zh-TW`。
- [ ] **clerk hook install/uninstall**：
  - `install`: 自動獲取 `clerk` 絕對路徑並寫入 `~/.claude/settings.json` 的 `hooks.SessionEnd`。
  - `uninstall`: 從 `settings.json` 移除該 hook 設定。

## 技術細節

### 1. 遞迴防護 (Recursion Guard)
- 在 `feed` 執行 `claude -p` 時注入環境變數 `CLERK_INTERNAL=1`。
- `feed` 啟動時若偵測到該變數則立即退出，避免摘要動作觸發無限循環。

### 2. 資料預處理
- 解析 `transcript.jsonl`，過濾掉 `tool_call` 與 `tool_result`，僅傳送純對話文字給 AI，節省 Token 並提升摘要品質。


## 發布規劃 (Release Strategy)

使用 **GoReleaser** 進行多平台發布與自動化打包：
- **支援平台**：macOS (Darwin), Linux, Windows。
- **套件管理**：
  - **macOS**: Homebrew (Tap)。
  - **Windows**: Scoop。
- **安裝腳本**：提供 Powershell, Git Bash, Bash, Sh 等一鍵安裝指令。
- **二進制分發**：透過 GitHub Releases 提供各平台預編譯檔。

## 開發規劃

1. `go mod init github.com/your-username/ccdigest`
2. 安裝 Cobra 框架。
3. 實作極簡 JSON 配置讀取與 `config show`。
4. 實作 `feed` 的對話過濾邏輯與 AI 摘要串接。
5. 實作 `hook install/uninstall` 的 JSON 注入邏輯。
6. 設定 `.goreleaser.yaml` 完成發布自動化。
