# Changelog

## v1.0.0 (2026-04-15)

Initial release.

### Features

- SessionEnd hook 自動觸發，session 結束時背景產生中文摘要
- 前景讀取 stdin 後立即 fork 背景退出（約 10-20ms），不阻塞 Claude Code
- 即時回饋：前景寫入 `=== CLAUDE DIGESTING ===` 佔位符，背景完成後替換為摘要
- 用 `jq` 精簡 transcript JSONL（每則限 2000 字元），避免超大 transcript 導致 timeout
- 摘要按日期目錄 + cwd slug 組織（`YYYYMMDD/<slug>.md`）
- 每段摘要含時段標題、session 名稱、duration、git branch、結束原因等 metadata
- `WRAPUP_RUNNING` 環境變數遞迴防護
- `install.sh` 互動式安裝：自訂儲存目錄、選擇 user-level 或 project-level 設定
- 錯誤日誌寫入 `wrap-up.err.log`
