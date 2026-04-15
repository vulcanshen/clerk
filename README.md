# claude-auto-digest

Claude Code session 結束時自動產生對話摘要，存成按日期組織的 markdown 檔案。

讓你以「某一天」為單位回顧工作內容，而不用逐 session 翻閱。

## 安裝

```bash
git clone https://github.com/anthropics/claude-auto-digest.git
cd claude-auto-digest
./install.sh
```

`install.sh` 會：
1. 檢查 `jq` 和 `claude` CLI 是否已安裝
2. 詢問摘要儲存目錄（`CLAUDE_AUTO_DIGEST_ROOT`）
3. 詢問安裝範圍（user-level 或 project-level）
4. 將 SessionEnd hook 設定合併到 `settings.json`

### 前置需求

- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code)
- [jq](https://jqlang.github.io/jq/) — `brew install jq`

## 使用方式

安裝後無需任何操作。每次 Claude Code session 結束時，會自動在背景產生摘要。

### 產出結構

```
$CLAUDE_AUTO_DIGEST_ROOT/
  20260415/
    sideproj-claude-auto-digest.md
    work-backend-api.md
  20260416/
    sideproj-claude-auto-digest.md
```

每個檔案依時段分隔，內容範例：

```markdown
## 14:30 session-abc123

_duration: 45m | branch: feature/foo | reason: user_exit_

- 建立了 `wrap-up.sh` 核心腳本，實作 SessionEnd hook 自動摘要功能
- 決定使用 shell script 而非 Go，因為核心邏輯只是 JSON 處理 + 呼叫 claude -p
- 產出檔案：`wrap-up.sh`、`install.sh`、`README.md`

## 16:45 session-def456

- 修復了 slug 產生邏輯的 edge case
- 加入 transcript 檔案不存在時的錯誤處理
```

## 設定

安裝時會要求指定摘要儲存目錄，寫入 `~/.claude/settings.json` 的 `env` 欄位：

```json
{
  "env": {
    "CLAUDE_AUTO_DIGEST_ROOT": "/Users/you/Documents/digests"
  }
}
```

如需修改，直接編輯 `~/.claude/settings.json` 中的 `CLAUDE_AUTO_DIGEST_ROOT` 即可。

## 運作原理

1. Claude Code session 結束時觸發 `SessionEnd` hook
2. `wrap-up.sh` 前景讀取 hook stdin 並寫入暫存檔，印出 `=== CLAUDE DIGESTING ===` 佔位符後立即 fork 背景程序退出（前景約 10-20ms）
3. 背景程序從暫存檔讀取 `transcript_path` 和 `cwd`
4. `cwd` 取最後兩層目錄轉為 slug（如 `sideproj-claude-auto-digest`）
5. 用 `jq` 精簡 transcript JSONL 為純文字對話（每則限 2000 字元）
6. 呼叫 `claude -p` 產生中文摘要
7. 移除佔位符，將摘要追加到 `$CLAUDE_AUTO_DIGEST_ROOT/YYYYMMDD/<slug>.md`

遞迴防護：背景 `claude -p` 啟動時設定 `WRAPUP_RUNNING=1`，防止其 SessionEnd 再次觸發摘要。

## 手動卸載

從 `~/.claude/settings.json` 移除 `hooks.SessionEnd` 中包含 `wrap-up.sh` 的項目，以及 `env.CLAUDE_AUTO_DIGEST_ROOT` 即可。

## License

GPL-3.0
