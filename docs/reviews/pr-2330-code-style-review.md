---
type: analysis
title: PR 2330 Code Style Review
created: 2026-11-11
tags:
  - code-review
  - code-style
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-data-flow-review]]'
  - '[[pr-2330-logic-review]]'
---

# PR 2330 Code Style Review

Reviewed the code-style checkbox only: naming, function sizes, single responsibility, duplication, dead code, and unused imports. Scope is all 25 files between base `de0b72766db0436b6c1887bb75768cb3960e67e5` and head `fb296d490cca4fdebbcb78d5107241f134152d19`, as recorded in [[REVIEW_SCOPE]]. All scoped working-tree files match that head byte for byte. Production source and PR tests are unchanged.

No critical or major style issues were identified. One minor test-practice issue and two optional maintenance suggestions follow. Error handling, module organization, API contracts, and the consolidated issue report remain separate playbook tasks.

## Findings

### Minor: fatal assertions run inside HTTP server goroutines

- **Location:** `internal/tools/keenable/keenable_test.go:25` (also lines 26–32 and 50–51).
- The callbacks passed to `httptest.NewServer` use `require.Equal`, `require.Empty`, and `require.NoError` with the parent test's `*testing.T`. Request handlers run in server goroutines. A failed `require` assertion invokes `FailNow`; `go doc testing.T.FailNow` explicitly requires calling it from the test goroutine. A mismatch can terminate the handler before it writes the response, producing an additional EOF error and obscuring the failing request check. This does not imply passing tests are currently incorrect.
- **Suggested fix:** use nonfatal `assert` checks in handlers, with an explicit handler return and error status if decoding fails; alternatively send captured request data/errors through a channel and perform `require` assertions in the test goroutine. Firecrawl has the same existing pattern, but the new tests extend it.

### Suggestion: mark the server helper as a test helper

- **Location:** `internal/tools/keenable/keenable_test.go:13`.
- Add `t.Helper()` at the start of `withServer` so future helper diagnostics identify the calling test. The existing `t.Cleanup` already restores both endpoint variables and closes the server. Endpoint globals are acceptable for the current serial tests; introducing parallel tests would require isolation.

### Suggestion: consider shared search dispatch when this code expands again

- **Location:** `internal/cli/tools.go:126` (compare Serply at line 90 and Firecrawl at line 108).
- Keenable adds a third search/append/write-or-chat sequence to the existing dispatcher. It correctly reuses `AppendMessage`, `Flags.IsChatRequest`, and `Flags.WriteOutput`; there is no existing shared search-dispatch helper to extend. Its absence of a required-key gate is intentional.
- A future helper accepting the query and a `func(string) (string, error)` search operation could centralize result routing, leaving provider-specific configuration checks at the call sites. This is optional: the 14-line branch is readable and consistent with adjacent providers, and a broad dispatcher refactor is unnecessary for this PR.

## Coverage and positive observations

| Files | Style assessment |
| --- | --- |
| `internal/tools/keenable/keenable.go` | `Client`, `NewClient`, `ApiKey`, embedded `PluginBase`, named returns, and localized errors follow Serply/Firecrawl conventions. Local `apiKey` uses normal initialism casing while the exported field matches existing clients. Both functions are fewer than 65 lines; request/decode/format phases remain readable without new helpers. The response type and HTTP settings remain package-local. All imports and declarations are used. |
| `internal/tools/keenable/keenable_test.go` | Descriptive test names, one reused HTTP fixture, automatic cleanup, and request/format assertions. Findings above apply. |
| `internal/cli/flags.go`, `internal/cli/help.go`, `internal/cli/tools.go` | Field, flag, and translation-key naming match the existing tools. No unused branch or unnecessary new abstraction; repeated routing is a suggestion above. |
| `internal/core/plugin_registry.go` | Initialization, field declaration, optional setup, environment serialization, and configuration follow the neighboring clients. Existing plugin/setup utilities are reused. |
| All 14 locale JSON files listed in [[REVIEW_SCOPE]] | Seven consistently prefixed Keenable keys per locale, placed alongside the neighboring service keys; existing two-space formatting is preserved. Translation quality is outside this style check. |
| `completions/_fabric`, `completions/fabric.bash`, `completions/fabric.fish` | New entries follow each shell's existing naming, quoting, indentation, and argument conventions. The long Bash option list predates this PR; this change adds one token. |
| `README.md`, `cmd/generate_changelog/incoming/2330.txt` | Help alignment and changelog structure match the existing documentation. No dead examples or unused configuration were introduced by the additions. |

The small client has one cohesive service responsibility. No shared formatter was found in the compared clients: Serply emits descriptions, Firecrawl emits scraped content with separators, and Keenable selects descriptions or truncated snippets. Extracting a common formatter here would add complexity for little reuse.

## Validation and limits

- Used the existing code graph to discover and inspect `handleToolProcessing` and `AddSetupQuestion`. Tool-client symbols were absent from the scoped graph search, so direct source/diff inspection supplied Keenable, Serply, and Firecrawl context. No repository-level `AGENTS.md` or `CLAUDE.md`, or nested versions under `internal`, `completions`, or `docs`, were present; the session's supplied instructions apply.
- `git diff --check` for the fixed PR snapshot passed. `gofmt -l` returned no paths for all six changed Go files.
- Parsed all 14 locale files and verified exact two-space JSON serialization, seven new keys each, and byte equality with the fixed head for all 25 paths.
- `go test -race ./internal/tools/keenable ./internal/tools/serply ./internal/tools/firecrawl ./internal/cli ./internal/core` passed for all five complete package suites, including Go's default vet checks. These passing tests do not exercise the assertion-failure path described above.
- No new executable behavior was added, so no new behavior tests were necessary. No live API/model calls, visual checks, or images were involved; zero images analyzed.
