# Document 1: Analyze PR Changes

## Context

- **Playbook**: Fabric PR Review
- **Agent**: Fabric-PR-Review
- **Project**: /Users/kayvan/src/fabric
- **Date**: 2026-04-05
- **Working Folder**: /Users/kayvan/src/fabric/.maestro/playbooks

## Purpose

Understand the scope and context of the Fabric pull request before diving into detailed review.

## Pull Request Information

**Pull Request**: https://github.com/danielmiessler/Fabric/pull/2073

> **NOTE**: Update the PR number above before running this playbook

## Tasks

### Task 1: Fetch PR Context

- [x] **Read the PR description**: Use `gh pr view XXXX` to fetch PR details. Note the stated goals, linked issues, and any breaking change warnings.
  Notes from `gh pr view 2073` on 2026-04-05:
  - Title: `feat(youtube): Implement visual text extraction via FFmpeg and OCR`
  - Stated goal: add a visual extraction pipeline to the `youtube` tool so on-screen text in videos is captured alongside transcript content.
  - Declared flags: `--visual`, `--visual-sensitivity`, and `--visual-fps`.
  - Declared implementation details: bounded OCR concurrency, 15-minute subprocess timeouts, separate stdout/stderr capture for Tesseract, race-safe temp directories, shell-safe `yt-dlp` invocation, and VTT-formatted OCR output.
  - Declared dependency requirement: `yt-dlp`, `ffmpeg`, and `tesseract` must be available on `PATH`; no new Go module dependencies are claimed.
  - Related issues: `closes #N/A (Community Feature Request)`.
  - Breaking change warnings: none explicitly called out in the PR description.

- [x] **Identify the base branch**: Determine what branch this PR is targeting (usually `main`).
  Confirmed from `gh pr view 2073 --json baseRefName` on 2026-04-05: base branch is `main`.

- [x] **Check PR size**: Fabric rejects PRs with 50+ files without justification. Count changed files early.
  Confirmed from `gh pr view 2073 --json changedFiles,files` on 2026-04-05: PR changes 19 files, so it is below Fabric's 50-file rejection threshold and does not require size-based justification.

### Task 2: Analyze Changed Files

- [x] **Get the diff summary**: Run `git diff --stat origin/main...HEAD` to see all changed files and their modification sizes.
  Notes from `git diff --stat origin/main...HEAD` on 2026-04-05:
  - Total diff size: 19 files changed, 321 insertions, 1 deletion.
  - Largest code change: `internal/tools/youtube/youtube.go` with 146 added lines.
  - CLI surface updates: `internal/cli/cli.go`, `internal/cli/flags.go`, and `internal/cli/help.go`.
  - Localization updates: 10 locale JSON files under `internal/i18n/locales/`.
  - Documentation/changelog updates: `.gitignore`, `README.md`, `README.zh.md`, and `cmd/generate_changelog/incoming/2073.txt`.

- [x] **Categorize changes**: Group files by Fabric's architecture:

  Notes from `git diff --name-only origin/main...HEAD` on 2026-04-05:
  - Core Components:
    - `cmd/`: `cmd/generate_changelog/incoming/2073.txt`
    - `internal/cli/`: `internal/cli/cli.go`, `internal/cli/flags.go`, `internal/cli/help.go`
    - `internal/core/`: none
    - `internal/chat/`: none
    - `internal/domain/`: none
  - Plugin System:
    - `internal/plugins/ai/`: none
    - `internal/plugins/db/`: none
    - `internal/plugins/strategy/`: none
    - `internal/plugins/template/`: none
  - Patterns & Strategies:
    - `data/patterns/`: none
    - `data/strategies/`: none
  - Infrastructure:
    - `internal/server/`: none
    - `internal/tools/`: `internal/tools/youtube/youtube.go`
    - `internal/i18n/`: `internal/i18n/locales/de.json`, `internal/i18n/locales/en.json`, `internal/i18n/locales/es.json`, `internal/i18n/locales/fa.json`, `internal/i18n/locales/fr.json`, `internal/i18n/locales/it.json`, `internal/i18n/locales/ja.json`, `internal/i18n/locales/pl.json`, `internal/i18n/locales/pt-BR.json`, `internal/i18n/locales/pt-PT.json`, `internal/i18n/locales/zh.json`
    - `internal/util/`: none
  - Other:
    - Test files (`*_test.go`): none in this PR
    - Configuration files: `.gitignore`
    - Documentation files: `README.md`, `README.zh.md`
    - Build/CI files: none

  **Core Components:**
  - `cmd/` - Entry points (fabric, code2context, to_pdf, generate_changelog)
  - `internal/cli/` - CLI flags, initialization, commands
  - `internal/core/` - Core chat functionality and plugin registry
  - `internal/chat/` - Chat coordination
  - `internal/domain/` - Domain models

  **Plugin System:**
  - `internal/plugins/ai/` - AI provider implementations
  - `internal/plugins/db/` - Database/storage plugins
  - `internal/plugins/strategy/` - Prompt strategies
  - `internal/plugins/template/` - Extension template system

  **Patterns & Strategies:**
  - `data/patterns/` - AI patterns (prompts)
  - `data/strategies/` - Prompt strategies (JSON)

  **Infrastructure:**
  - `internal/server/` - REST API server
  - `internal/tools/` - Utility tools
  - `internal/i18n/` - Internationalization
  - `internal/util/` - Shared utilities

  **Other:**
  - Test files (`*_test.go`)
  - Configuration files
  - Documentation files
  - Build/CI files

### Task 3: Understand the Scope

- [x] **Assess PR size**:
  - Small: < 100 lines
  - Medium: 100-500 lines
  - Large: > 500 lines
  - **Flag**: 50+ files = likely rejection without justification
  Notes from `gh pr view 2073 --json additions,deletions,changedFiles` on 2026-04-05:
  - GitHub PR metadata reports 321 additions and 1 deletion across 19 changed PR files, for a 322-line delta.
  - Size assessment: `Medium` (falls within the 100-500 line range).
  - File-count flag: not triggered because the PR is well below the 50-file rejection threshold.
  - Note: local `git diff origin/main...HEAD` also includes this playbook file, so PR size classification is based on GitHub's PR file list rather than the local working branch diff.

- [x] **Identify high-risk areas**: Flag files that:
  - Handle API keys/credentials (`*.env`, config loading)
  - Implement AI provider interfaces
  - Modify core chat flow
  - Change plugin registry behavior
  - Alter pattern loading/parsing
  - Touch authentication/OAuth flows
  Notes from targeted diff review on 2026-04-05:
  - None of the 19 PR files touch Fabric's explicitly listed highest-risk cross-cutting areas: there are no API key or credential-loading changes, no AI provider or plugin registry edits, no authentication/OAuth changes, no pattern loading/parsing updates, and no `internal/core/` or `internal/chat/` modifications.
  - `internal/tools/youtube/youtube.go` is the primary code-risk file because it adds a new subprocess-heavy path (`yt-dlp` -> `ffmpeg` -> `tesseract`), bounded OCR concurrency, temp-directory lifecycle handling, and VTT-style visual text output generation.
  - `internal/cli/cli.go` is the secondary risk file because it changes the `youtube` command execution flow by wiring visual extraction into the message-building path and broadening the condition that triggers extraction work.
  - `internal/cli/flags.go`, `internal/cli/help.go`, and the locale JSON files are low-to-medium risk surface changes because they expose the new CLI flags and i18n strings, but they do not alter core plugin, auth, or provider architecture.

### Task 4: Identify Review Focus

- [x] **Pattern changes**: Are any `data/patterns/` directories added or modified?
  Confirmed from `git diff --name-only origin/main...HEAD` on 2026-04-05: no files under `data/patterns/` or `data/strategies/` are added or modified, so no pattern validation review is needed for this PR.

- [x] **Plugin changes**: Are any `internal/plugins/ai/` providers added or modified?
  Confirmed from `git diff --name-only origin/main...HEAD` on 2026-04-05: no files under `internal/plugins/ai/` or other plugin provider directories are changed, so no plugin architecture review is needed for this PR.

- [x] **API changes**: Are there changes to `internal/server/` endpoints?
  Confirmed from `git diff --name-only origin/main...HEAD` on 2026-04-05: there are no `internal/server/` changes, so no API endpoint review is needed for this PR.

- [x] **CLI changes**: Are flags or commands modified in `internal/cli/`?
  Confirmed from `git diff origin/main...HEAD -- internal/cli/cli.go internal/cli/flags.go internal/cli/help.go` on 2026-04-05:
  - `internal/cli/flags.go` adds the new `--visual`, `--visual-sensitivity`, and `--visual-fps` YouTube flags.
  - `internal/cli/help.go` maps those flags to new i18n-backed help text keys.
  - `internal/cli/cli.go` changes `processYoutubeVideo` so visual extraction can run independently of transcript/comments/metadata and appends OCR-derived VTT output when `--visual` is enabled.
  - Review implication: yes, CLI behavior and flag-surface review are required for this PR.

### Task 5: Create Scope Document

- [ ] **Write REVIEW_SCOPE.md**: Create `/Users/kayvan/src/fabric/.maestro/playbooks/REVIEW_SCOPE.md` with:

```markdown
# Fabric PR Review Scope

## PR Information
- **URL**: [PR URL]
- **Title**: [PR Title]
- **Base Branch**: [target branch]
- **Size**: [small/medium/large]
- **File Count**: [X files] [FLAG if 50+]

## Changed Files by Category

### Core Components
[List files in cmd/, internal/cli/, internal/core/, internal/chat/, internal/domain/]

### Plugin System
[List files in internal/plugins/]

### Patterns & Strategies
[List files in data/patterns/, data/strategies/]

### Infrastructure
[List files in internal/server/, internal/tools/, internal/i18n/, internal/util/]

### Tests
[List *_test.go files]

### Other
[Documentation, config, CI files]

## High-Risk Areas
[Files requiring extra scrutiny]

## Review Focus
- [ ] Pattern validation needed: [Yes/No]
- [ ] Plugin architecture review needed: [Yes/No]
- [ ] API endpoint review needed: [Yes/No]
- [ ] CLI changes review needed: [Yes/No]

## PR Requirements Checklist
- [ ] PR is focused (not 50+ files without justification)
- [ ] Tests included for new functionality
- [ ] No obvious formatting issues
```

## Success Criteria

- PR context fetched and understood
- All changed files identified and categorized
- High-risk areas flagged
- Review focus areas identified
- REVIEW_SCOPE.md created

## Status

Mark complete when scope document is created.

---

**Next**: Document 2 will perform Go-specific code quality review.
