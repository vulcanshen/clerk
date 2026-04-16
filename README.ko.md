# clerk

[![GitHub Release](https://img.shields.io/github/v/release/vulcanshen/clerk)](https://github.com/vulcanshen/clerk/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vulcanshen/clerk)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/vulcanshen/clerk/ci.yml?label=CI)](https://github.com/vulcanshen/clerk/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanshen/clerk)](https://goreportcard.com/report/github.com/vulcanshen/clerk)
[![License](https://img.shields.io/github/license/vulcanshen/clerk)](LICENSE)

[English](README.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md)

Claude Code 세션은 터미널을 닫으면 사라집니다. clerk는 당신이 무엇을 했는지 절대 놓치지 않게 해줍니다.

## 문제

Claude Code를 매일 사용한다면, 이런 벽에 부딪힌 적이 있을 겁니다:

- **컨텍스트 유실** — `-c`나 `--resume`을 깜빡해서 처음부터 다시 시작. 이전 세션에 모든 컨텍스트가 있었는데, 수많은 세션 ID 더미에서 찾아내기란 거의 불가능.
- **세션 혼란** — 여러 프로젝트, 여러 세션이 동시에 진행 중. 오늘 아침 API 서버에서 뭘 했지? 인증 수정은 어느 세션이었지? 전혀 기억나지 않음.
- **주간 보고 공포** — 금요일 오후, 주간 보고 시간. `git log`를 뒤지며 이번 주에 실제로 뭘 했는지 재구성하려는 중.
- **수동 기록** — Claude에게 "요약 저장해"라고 했지만 지난번에는 깜빡함. 아니면 세션이 크래시됨. 아니면 터미널을 닫음. 컨텍스트는 사라짐.

이 모든 것은 결국 하나로 귀결됩니다: **Claude Code는 세션 간 기억을 유지하지 않으며, 당신이 직접 기억할 필요도 없어야 합니다.**

## 해결책

```bash
clerk install
```

끝입니다. clerk는 Claude Code에 연결되어 백그라운드에서 조용히 작동합니다:

| 문제점 | clerk의 해결 방법 |
|--------|-----------------|
| 컨텍스트 유실 | `/clerk-resume` — 이전 세션에서 컨텍스트를 즉시 복원 |
| 세션 혼란 | 프로젝트별 일일 요약 자동 생성, 날짜별 정리 |
| 주간 보고 | 요약 + 키워드 태그 = `/clerk-search`로 검색 가능한 작업 이력 |
| 수동 기록 | 완전 자동 — 기억할 명령어도, 들일 습관도 없음 |

clerk는 **설치하고 잊어버리는** 도구입니다. 한 번 설치하면 모든 세션이 자동으로 요약, 추적, 태그, 검색 가능해집니다. 컨텍스트가 필요할 때는 슬래시 명령어 하나면 됩니다.

> **참고:** clerk는 AI 메모리 도구가 아닙니다. claude-mem 같은 도구는 AI가 기억할 수 있도록 컨텍스트를 저장합니다. clerk는 **당신**이 읽을 수 있도록 요약을 저장합니다 — 날짜별로 정리되고, 키워드로 검색 가능하며, 주간 리뷰에 바로 사용할 수 있습니다.

## 기능

- **자동 요약** — Claude Code 세션 종료 시 증분 요약을 자동 생성
- **컨텍스트 복원** — `/clerk-resume`으로 이전 세션에서 컨텍스트 재구축
- **키워드 검색** — `/clerk-search`로 태그별 과거 작업 검색
- **세션 추적** — 이력 조회를 위해 모든 세션 시작을 기록
- **태그 시스템** — 요약에서 키워드를 자동 추출하여 검색 가능한 인덱스 구축
- **커서 추적** — 마지막 실행 이후의 새 메시지만 처리하여 토큰과 시간 절약
- **프로세스 관리** — 활성 feed 모니터링, 강제 종료, 중단된 것 재시도
- **프로젝트 레벨 설정** — 프로젝트별로 feed 비활성화, 전역 설정 재정의
- **원커맨드 설정** — `clerk install`로 훅, MCP 서버, 스킬을 일괄 설정
- 크로스 플랫폼: macOS, Linux, Windows
- 셸 자동 완성 (bash, zsh, fish, powershell)

## 작동 방식

### 설치되는 항목

| 컴포넌트 | 기능 |
|----------|------|
| **hook** | SessionStart에서 세션 ID 기록, SessionEnd에서 요약 생성 트리거 |
| **mcp** | `clerk-resume` 및 `clerk-search` 도구를 제공하는 MCP stdio 서버 |
| **skills** | Claude Code용 `/clerk-resume` 및 `/clerk-search` 슬래시 명령어 |

### 요약 흐름

1. 세션 종료 → 훅이 `clerk feed` 트리거
2. Feed가 백그라운드로 포크 (훅은 즉시 반환)
3. 마지막 실행 이후의 새 메시지만 읽기 (커서 추적)
4. 기존 일일 요약을 로드하고, `claude -p`를 호출하여 병합
5. 업데이트된 요약 저장 + 검색 인덱스용 태그 추출

### 복원 흐름

1. Claude Code에서 `/clerk-resume`을 입력
2. Claude가 프로젝트의 작업 디렉토리를 지정하여 `clerk-resume` MCP 도구를 호출
3. clerk가 파일 경로를 반환: 일일 요약 + 전체 transcript 파일
4. Claude가 먼저 요약을 읽어 빠른 개요 파악
5. 더 자세한 내용이 필요하면 Claude가 transcript 파일을 읽음
6. Claude가 이전에 수행한 작업을 요약하고, 컨텍스트가 복원되었음을 확인

### 검색 흐름

1. Claude Code에서 `/clerk-search`를 입력
2. Claude가 검색할 키워드를 물어봄 (또는 인수로 직접 제공)
3. Claude가 해당 키워드로 `clerk-search` MCP 도구를 호출
4. clerk가 태그 인덱스와 대조하여 일치하는 요약 및 transcript 경로를 반환
5. Claude가 해당 파일을 읽고 관련 컨텍스트를 제시

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

## 설치

### 빠른 설치

macOS / Linux / Git Bash:

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/vulcanshen/clerk/main/install.ps1 | iex
```

그런 다음 훅, MCP 서버, 스킬을 설정합니다:

```bash
clerk install
```

### 패키지 관리자

| 플랫폼 | 명령어 |
|--------|--------|
| Homebrew (macOS / Linux) | `brew install vulcanshen/tap/clerk` |
| Scoop (Windows) | `scoop bucket add vulcanshen https://github.com/vulcanshen/scoop-bucket && scoop install clerk` |
| Debian / Ubuntu | `sudo dpkg -i clerk_<version>_linux_amd64.deb` |
| RHEL / Fedora | `sudo rpm -i clerk_<version>_linux_amd64.rpm` |

### 소스에서 빌드

```bash
go install github.com/vulcanshen/clerk@latest
```

## 명령어 목록

| 명령어 | 설명 |
|--------|------|
| `install` | 모든 컴포넌트 설치 (hook + mcp + skills) |
| `install hook` | SessionStart/SessionEnd 훅만 설치 |
| `install mcp` | MCP 서버만 등록 |
| `install skills` | 슬래시 명령어 스킬만 설치 |
| `uninstall` | 모든 컴포넌트 제거 |
| `config` | 현재 설정 표시 (`config show`의 별칭) |
| `config show` | 병합된 설정과 파일 경로 표시 |
| `config set <key> <value>` | 프로젝트 레벨 설정 값 변경 |
| `config set -g <key> <value>` | 전역 설정 값 변경 |
| `status` | 활성 feed 프로세스와 중단된 세션 표시 |
| `status --watch` | 실시간 상태 업데이트 (매초) |
| `retry <slug>` | 지정한 중단 세션 재시도 |
| `retry --all` | 모든 중단 세션 재시도 |
| `kill <slug>` | 지정한 활성 feed 프로세스 강제 종료 |
| `kill --all` | 모든 활성 feed 프로세스 강제 종료 |

내부 명령어 (훅에서 호출되며, 사용자가 직접 사용하지 않음):

| 명령어 | 설명 |
|--------|------|
| `feed` | 세션 트랜스크립트를 처리하고 요약 생성 |
| `punch` | 세션 시작 시 세션 ID 기록 |
| `mcp` | MCP stdio 서버 시작 |

## 설정

### 설정 파일

- 전역: `~/.config/clerk/.clerk.json`
- 프로젝트: `<cwd>/.clerk.json` (전역 설정 재정의)

### 사용 가능한 설정

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

| 설정 항목 | 기본값 | 설명 |
|-----------|--------|------|
| `output.dir` | `~/.clerk/` | 요약 저장 루트 디렉토리 |
| `output.language` | `en` | 요약 출력 언어 |
| `summary.model` | `""` (claude 기본값) | `claude -p`에서 사용할 모델 |
| `log.retention_days` | `30` | 로그 및 커서 파일 보존 일수 |
| `feed.enabled` | `true` | 이 프로젝트의 feed 활성화/비활성화 |

### 예시

```bash
# 특정 프로젝트에서 feed 비활성화
cd /path/to/unimportant-project
clerk config set feed.enabled false

# 전역으로 더 저렴한 모델 사용
clerk config set -g summary.model haiku

# 전역으로 출력 언어 변경
clerk config set -g output.language en
```

## MCP 도구

MCP 서버 설치 후 사용 가능 (`clerk install mcp`):

| 도구 | 설명 |
|------|------|
| `clerk-resume` | 컨텍스트 복원을 위한 요약 + 트랜스크립트 파일 경로 반환 |
| `clerk-search` | 키워드/태그로 이전 세션 검색 |

## 스킬

스킬 설치 후 사용 가능 (`clerk install skills`):

| 스킬 | 설명 |
|------|------|
| `/clerk-resume` | 이전 세션에서 컨텍스트 복원 — MCP 도구 호출, 파일 읽기, 컨텍스트 재구축 |
| `/clerk-search` | 키워드로 과거 세션 검색 — MCP 도구 호출, 일치하는 파일 읽기 |

## 문제 해결

로그는 `~/.clerk/.log/YYYYMMDD-clerk.log`에 저장됩니다:

```bash
cat ~/.clerk/.log/$(date +%Y%m%d)-clerk.log
```

일반적인 문제:

- **요약이 생성되지 않음** — `claude`가 PATH에 있는지 확인
- **Hook cancelled** — clerk는 백그라운드 포크로 대응 완료. 최신 버전으로 업데이트
- **MCP 도구를 찾을 수 없음** — `clerk install mcp`를 실행하고 세션을 재시작

## 셸 자동 완성

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

## 라이선스

[GPL-3.0](LICENSE)
