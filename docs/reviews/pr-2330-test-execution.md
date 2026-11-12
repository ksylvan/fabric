---
type: report
title: PR 2330 Test Execution
created: 2026-11-11
tags:
  - code-review
  - testing
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-related-tests]]'
  - '[[pr-2330-coverage-gaps]]'
---

# Test execution

Completed only **Execute test suite** in Document 4. Tested review commit `414906d31cac1f583e1ab6172e7d06f267d8df91`, which contains the fixed PR head `fb296d490cca4fdebbcb78d5107241f134152d19` plus review artifacts. No production source or tests were changed, and no review-test overlays were applied. No images were associated with this task or analyzed.

## Environment and commands

Execution date: November 11, 2026 (America/Los_Angeles). Local runner: Go `1.27.3`, `darwin/arm64`; module declares Go `1.26.10`. These results apply to this local environment, not the Linux/Windows CI matrix. Followed the full-suite and race scopes in `.github/workflows/ci.yml`, adding Keenable to the race scope, JSON output for analysis, uncached test execution, and a per-package timeout.

```bash
export GOCACHE=/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/go-build
export GOMODCACHE=/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/go-mod
export TMPDIR=/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/tmp
export GOTOOLCHAIN=local
go test -json -count=1 -timeout=120s ./...
go test -json -race -count=1 -timeout=120s ./internal/plugins/... ./internal/server/... ./internal/core/... ./internal/tools/keenable
```

Reused the Working folder's caches and copied locally cached modules into its module cache. No dependency manifests changed. Each command also had a 600-second process limit; neither limit was reached.

## Results

| Run | Exit | Elapsed | Packages passed | Packages without tests | Test/subtest pass events | Test skips |
| --- | --- | --- | --- | --- | --- | --- |
| Full suite | 0 | 27.89 s | 46 | 14 | 1,416 | 5 |
| Race scope plus Keenable | 0 | 36.79 s | 28 | 1 | 1,052 | 0 |

Counts come from terminal Go JSON events. Test/subtest events include parent tests and their subtests; these are not counts of independent top-level test functions. Package-level skip events correspond to packages reporting `[no test files]`. Both stderr streams were empty. No test failures, build failures, timeouts, or race reports occurred.

All three added Keenable tests passed in both runs: `TestSearchWithoutKeyUsesPublicEndpoint`, `TestSearchWithKeyUsesKeyedEndpoint`, and `TestSearchReturnsAPIError`. Related CLI, core, and i18n package suites passed in the full run; the broad CLI test itself remains skipped, as detailed below.

## Skips and warnings

Full-suite skips:

| Test | Reported reason |
| --- | --- |
| `internal/domain.TestApplyFileChanges_RejectsGitShortName` | 8.3 short names are a Windows feature. |
| `internal/cli.TestCli` | Existing flag `-t` collision; explicitly skipped for now. |
| `internal/cli.TestCopyToClipboard` | Explicit skip citing Docker environment in CI. |
| `internal/tools/notifications.TestNotifySendProvider` | Linux notify-send test on a non-Linux platform. |
| `internal/tools/notifications.TestPowerShellProvider` | Windows PowerShell test on a non-Windows platform. |

The full run emitted `Warning: Could not create custom patterns directory /some/path: mkdir /some: read-only file system`. Both runs emitted `Warning: no content received from upstream Fabric server`. These were test-output warnings; both commands still exited successfully. Test names containing “warning” were not counted as runtime warnings.

## Evidence and limits

Raw logs are retained under `/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/`:

- `test-execution-suite.jsonl` and `test-execution-suite.stderr`
- `test-execution-race.jsonl` and `test-execution-race.stderr`
- `test-execution-results.json` (commands, exit codes, elapsed times)

Passing tests do not fill the missing cases recorded in [[pr-2330-coverage-gaps]], establish live Keenable service behavior, or validate shell completions. This task executes existing tests rather than adding cases or fixing previously identified coverage gaps. The next checkbox, creating `TEST_GAPS.md`, remains pending.
