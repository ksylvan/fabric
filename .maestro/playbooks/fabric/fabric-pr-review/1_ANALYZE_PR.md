# Document 1: Analyze PR Changes

## Context

- **Playbook**: Fabric PR Review
- **Agent**: Fabric-PR-Review
- **Project**: /Users/kayvan/src/fabric
- **Date**: 2026-04-06
- **Working Folder**: /Users/kayvan/src/fabric/.maestro/playbooks

## Purpose

Understand the scope and context of the Fabric pull request before diving into detailed review.

## Pull Request Information

**Pull Request**: https://github.com/danielmiessler/Fabric/pull/2073

> **NOTE**: Update the PR number above before running this playbook

## Tasks

### Task 1: Fetch PR Context

- [x] **Read the PR description**: Use `gh pr view XXXX` to fetch PR details. Note the stated goals, linked issues, and any breaking change warnings.
  Completed 2026-04-06 via `gh pr view 2073 --repo danielmiessler/Fabric`.
  PR title: `feat(youtube): Implement visual text extraction via FFmpeg and OCR`.
  Stated goals: add visual extraction to the `youtube` tool so Fabric can capture on-screen text from videos using `yt-dlp`, `ffmpeg`, and `tesseract`, with new `--visual`, `--visual-sensitivity`, and `--visual-fps` CLI flags.
  Related issues: `closes #N/A (Community Feature Request)`.
  Breaking changes: none stated in the PR description. Operational note: the feature requires `yt-dlp`, `ffmpeg`, and `tesseract` to be installed in `PATH`.

- [x] **Identify the base branch**: Determine what branch this PR is targeting (usually `main`).
  Completed 2026-04-06 via `gh pr view 2073 --repo danielmiessler/Fabric --json baseRefName` and GitHub PR metadata.
  Base branch: `main`.

- [ ] **Check PR size**: Fabric rejects PRs with 50+ files without justification. Count changed files early.

### Task 2: Analyze Changed Files

- [ ] **Get the diff summary**: Run `git diff --stat origin/main...HEAD` to see all changed files and their modification sizes.

- [ ] **Categorize changes**: Group files by Fabric's architecture:

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

- [ ] **Assess PR size**:
  - Small: < 100 lines
  - Medium: 100-500 lines
  - Large: > 500 lines
  - **Flag**: 50+ files = likely rejection without justification

- [ ] **Identify high-risk areas**: Flag files that:
  - Handle API keys/credentials (`*.env`, config loading)
  - Implement AI provider interfaces
  - Modify core chat flow
  - Change plugin registry behavior
  - Alter pattern loading/parsing
  - Touch authentication/OAuth flows

### Task 4: Identify Review Focus

- [ ] **Pattern changes**: Are any `data/patterns/` directories added or modified?

- [ ] **Plugin changes**: Are any `internal/plugins/ai/` providers added or modified?

- [ ] **API changes**: Are there changes to `internal/server/` endpoints?

- [ ] **CLI changes**: Are flags or commands modified in `internal/cli/`?

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
