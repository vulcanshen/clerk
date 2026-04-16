# clerk

[![GitHub Release](https://img.shields.io/github/v/release/vulcanshen/clerk)](https://github.com/vulcanshen/clerk/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vulcanshen/clerk)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/vulcanshen/clerk/ci.yml?label=CI)](https://github.com/vulcanshen/clerk/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanshen/clerk)](https://goreportcard.com/report/github.com/vulcanshen/clerk)
[![License](https://img.shields.io/github/license/vulcanshen/clerk)](LICENSE)

[English](README.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md)

Claude Code의 서기관 — 세션을 자동 요약.

clerk는 Claude Code의 `SessionEnd` 이벤트에 연결되는 CLI 도구로, 대화가 끝나면 자동으로 요약을 생성하여 정리된 markdown 파일로 저장합니다.

## 기능

- **자동 요약** — Claude Code 세션 종료 시 자동으로 요약 생성
- **증분 병합** — 각 세션은 같은 날 같은 프로젝트의 단일 요약 파일에 병합, 중복 없음
- **대화 필터링** — tool call을 제거하고 사용자와 AI의 대화 텍스트만 유지
- **날짜별 정리** — 요약은 `~/.clerk/YYYYMMDD/<프로젝트-slug>.md`에 저장
- **커서 추적** — 마지막 실행 이후의 새 메시지만 처리하여 토큰과 시간 절약
- **프로세스 관리** — 활성 feed 모니터링, 강제 종료, 중단된 것 재시도
- **설정 가능** — 출력 디렉토리, 언어, 모델, 로그 보존 일수 커스터마이즈 가능
- **원커맨드 설정** — `clerk hook install`로 모든 설정 완료
- **재귀 방지** — clerk가 `claude -p`를 호출할 때 무한 루프 방지
- 크로스 플랫폼: macOS, Linux, Windows
- 셸 자동 완성 (bash, zsh, fish, powershell)

## 작동 방식

Claude Code 세션이 종료되면 `SessionEnd` 훅이 `clerk feed`를 트리거합니다:

1. 백그라운드로 포크 (훅은 즉시 반환)
2. 마지막 실행 이후의 새 메시지만 트랜스크립트(JSONL)에서 읽기
3. 프로젝트의 기존 일일 요약 로드 (있는 경우)
4. `claude -p`를 호출하여 병합된 요약 생성
5. 요약 파일을 업데이트된 버전으로 덮어쓰기

```
~/.clerk/
└── 20260416/
    ├── projects-my-app.md
    ├── projects-api-server.md
    └── work-frontend.md
```

## 설치

### 원라인 설치

macOS / Linux / Git Bash:

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/vulcanshen/clerk/main/install.ps1 | iex
```

업데이트는 같은 명령어를 다시 실행하면 됩니다. 제거:

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/uninstall.sh | sh
```

```powershell
irm https://raw.githubusercontent.com/vulcanshen/clerk/main/uninstall.ps1 | iex
```

### 패키지 관리자

| 플랫폼 | 명령어 |
|--------|--------|
| Homebrew (macOS / Linux) | `brew install vulcanshen/tap/clerk` |
| Scoop (Windows) | `scoop bucket add vulcanshen https://github.com/vulcanshen/scoop-bucket && scoop install clerk` |
| Debian / Ubuntu | `sudo dpkg -i clerk_<version>_linux_amd64.deb` |
| RHEL / Fedora | `sudo rpm -i clerk_<version>_linux_amd64.rpm` |

`.deb` 및 `.rpm` 패키지는 [Releases 페이지](https://github.com/vulcanshen/clerk/releases)에서 다운로드할 수 있습니다.

### 소스에서 빌드

```bash
go install github.com/vulcanshen/clerk@latest
```

## 빠른 시작

```bash
# SessionEnd 훅 설치
clerk hook install
```

끝입니다. 훅을 설치하면 clerk는 완전히 백그라운드에서 실행됩니다 — 수동 조작도, 추가 명령어도 필요 없습니다. Claude Code 세션을 종료할 때마다 요약이 자동으로 생성되고 저장됩니다. 설치하고 나면 잊어도 됩니다.

## 명령어 목록

| 명령어 | 설명 |
|--------|------|
| `feed` | 세션 트랜스크립트를 처리하고 요약 생성 (훅에서 호출) |
| `config` | 현재 설정 표시 (`config show`의 별칭) |
| `config show` | 현재 설정과 설정 파일 경로 표시 |
| `config set <key> <value>` | 설정 값 변경 (키 탭 자동 완성 지원) |
| `hook install` | clerk를 Claude Code SessionEnd 훅으로 설치 |
| `hook uninstall` | Claude Code SessionEnd 훅에서 clerk 제거 |
| `status` | 활성 feed 프로세스와 중단된 세션 표시 |
| `status --watch` | 실시간 상태 업데이트 (매초) |
| `retry <slug>` | 지정한 중단 세션 재시도 |
| `retry --all` | 모든 중단 세션 재시도 |
| `kill <slug>` | 지정한 활성 feed 프로세스 강제 종료 |
| `kill --all` | 모든 활성 feed 프로세스 강제 종료 |

## 설정

설정 파일: `~/.config/clerk/config.json`

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

| 설정 항목 | 기본값 | 설명 |
|-----------|--------|------|
| `output.dir` | `~/.clerk/` | 요약 저장 루트 디렉토리 |
| `output.language` | `zh-TW` | 요약 출력 언어 |
| `summary.model` | `""` (claude 기본값) | `claude -p`에서 사용할 모델 |
| `log.retention_days` | `30` | 로그 및 커서 파일 보존 일수 |

`clerk config set`으로 설정:

```bash
clerk config set output.language en
clerk config set summary.model haiku
clerk config set log.retention_days 14
```

설정 파일은 선택 사항입니다 — 존재하지 않으면 clerk는 기본값을 사용합니다.

## 요약 형식

각 프로젝트는 하루에 하나의 파일, 증분적으로 병합됩니다:

```markdown
# projects-my-app

> Last updated: 14:30:25

### 핵심 작업
- JWT 토큰을 사용한 사용자 인증 구현
- WebSocket 핸들러의 경쟁 조건 수정

### 보조 작업
- GitHub Actions CI 파이프라인 추가
- README API 문서 업데이트

### 주요 결정 및 이유
- **결정**: 세션 대신 JWT 사용 → **이유**: 멀티리전 배포를 위한 무상태 스케일링

### 사용자 노트
- 최소한의 추상화 선호, 프레임워크보다 직접 코드 작성

### 버전 로그
- v1.0.0 — 인증 및 WebSocket 지원 포함 초기 릴리스
```

## 문제 해결

로그는 `~/.clerk/.log/YYYYMMDD-clerk.log`에 저장됩니다. 요약이 생성되지 않을 때 확인:

```bash
cat ~/.clerk/.log/$(date +%Y%m%d)-clerk.log
```

일반적인 문제:

- **요약이 생성되지 않음** — `claude`가 PATH에 있는지 확인
- **Hook cancelled** — clerk는 백그라운드 포크로 대응 완료. 최신 버전으로 업데이트
- **내용 중복** — 이전 버전 동작. 현재는 증분 병합 사용

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
