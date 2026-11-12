---
type: analysis
title: PR 2330 Error Handling Review
created: 2026-11-11
tags:
  - code-review
  - error-handling
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-logic-review]]'
  - '[[pr-2330-data-flow-review]]'
  - '[[pr-2330-code-style-review]]'
---

# PR 2330 Error Handling Review

Completed only the error-handling checkbox: error propagation, diagnostic quality, sensitive information in errors, and asynchronous failures. Scope remains the 25 files between `de0b72766db0436b6c1887bb75768cb3960e67e5` and `fb296d490cca4fdebbcb78d5107241f134152d19` from [[REVIEW_SCOPE]]. All scoped working-tree files match the recorded head byte for byte. Production source and original PR tests remain unchanged.

Found one major robustness issue and two minor diagnostic issues. The disclosure finding below is conditional on service behavior; no actual credential disclosure by Keenable was observed. These findings are for the subsequent consolidated issue report; this task does not implement fixes or complete later review checkboxes.

## Findings

### Major: response bodies have no byte limit

- **Location:** `internal/tools/keenable/keenable.go:94` and `:100`.
- Non-200 responses use unbounded `io.ReadAll`, then copy the body into a formatted error. Successful responses likewise decode unrestricted JSON before snippet truncation. The 30-second timeout bounds elapsed time, not the bytes delivered within that interval. The outgoing `max_results: 10` is a request hint, not a locally enforced body bound. A faulty service or proxy can exhaust memory; large failure bodies also flood stderr.
- **Evidence:** a mocked 503 with a 2 MiB body produces an error containing that entire body. This safely establishes lack of an error-body cap without trying to exhaust process memory. The success-path concern follows from source inspection, not a stress test.
- **Suggested fix:** bound successful JSON input with an explicit maximum plus overflow detection. Limit error diagnostics to a smaller budget, report truncation, and retain HTTP status. Enforce bounds before allocating/decoding the full body. Existing Serply and Firecrawl share this pattern, but this PR introduces another affected client.

### Minor: raw service error bodies can expose echoed sensitive information

- **Location:** `internal/tools/keenable/keenable.go:95`; propagation through `internal/cli/tools.go:129`, `internal/cli/cli.go:99`, and `cmd/fabric/main.go:15`.
- The entire untrusted error body is returned unchanged and eventually printed to stderr. A service or gateway response containing an echoed API key or private query is therefore copied into terminal output and any captured logs. The formatter itself does not append request headers or queries; disclosure requires the response to contain them. No live service was contacted and no real secrets were used.
- **Evidence:** a mocked 401 body containing `review-fake-key` and `review-private-query` retains both in the returned error.
- **Suggested fix:** prefer bounded, selected diagnostic fields or a generic status message. Redact known request secrets before displaying external diagnostics and sanitize control characters. Redacting only the API key would not remove arbitrary private query text. Review redirect/credential transport behavior separately during the security task.

### Minor: non-200 body-read errors are discarded

- **Location:** `internal/tools/keenable/keenable.go:94`.
- If the error response body is truncated, times out, or fails during reading, `body, _ := io.ReadAll` loses that cause. The caller still receives an HTTP failure, so the request is not falsely reported as successful, but the diagnostic can contain incomplete text without saying so. `errors.Is` cannot identify the read failure.
- **Evidence:** a synthetic 502 body returns partial text followed by `io.ErrUnexpectedEOF`; the returned error includes the partial text but neither wraps nor reports the read error. The corresponding 200 decode path does preserve `io.ErrUnexpectedEOF`.
- **Suggested fix:** retain HTTP status and safe partial diagnostics while wrapping the read error, distinguishing a diagnostic-size cap from a transport read failure.

## Error-path assessment

| Area | Assessment |
| --- | --- |
| Request encoding/construction | Both errors are checked and localized with `%w`. The current string/integer payload cannot ordinarily make `json.Marshal` fail; malformed endpoint construction is covered by the reused review test. |
| HTTP execution | Transport failures and actual client timeouts retain their causes through `%w`. The client waits synchronously for `Do`; no application goroutine or promise is left without an error consumer. A 25 ms test client exercises timeout propagation; production timeout remains 30 seconds. |
| Response cleanup | `defer resp.Body.Close()` runs on status, decode, and success paths after a successful `Do`. New tests check closure even when reads fail. Ignoring a close error on this read-only response is not a separate confirmed defect. |
| JSON failures | Syntax errors, type errors, empty bodies, and interrupted reads return errors with no partial Markdown. `errors.As` recovers JSON error types and `errors.Is` recovers EOF/read errors. No blind recovery or panic suppression is introduced. |
| HTTP diagnostics | Non-200 errors identify status and body, useful for rate limiting and authentication errors. Bounds, disclosure, and lost read causes are the findings above. No retries are introduced; absence of retry is not treated as a defect without a retry contract. |
| CLI propagation | Search failure returns immediately before result append/write/chat processing. `Cli` returns tool errors; `main` prints them and exits nonzero. New CLI cases preserve a transport sentinel and verify output-write failures are returned without overwriting an existing file. Existing output behavior prints results to stdout before attempting a file write. |
| Optional-key configuration | Registry intentionally ignores optional plugin configuration errors, matching neighboring clients. Keenable's single setting is optional, loads a string, and has no custom configuration callback; there is no newly fallible configuration operation being silently lost. `SaveEnvFile` still returns `Db.SaveEnv` errors. |
| Localization | All 14 changed locale files retain `%w` on create/execute/decode errors and `%d`/`%s` on status/body errors. This check validates formatting contracts, not translation quality. |
| Remaining scoped files | Flags/help, completions, README, and changelog add declarative entries without new error or async execution paths. The original test handlers' fatal-assertion goroutine issue remains recorded in [[pr-2330-code-style-review]]. |

## Validation and limits

Used the existing indexed knowledge graph for discovery and source inspection of search, plugin settings, registry persistence, CLI dispatch, and `Cli`. Graph symbols collide for several `Configure`/`WriteOutput` methods, so targeted source reads resolved the concrete receiver implementations. No repository or nested `AGENTS.md`/`CLAUDE.md` guidance was found in the relevant folders.

Added review-only characterization tests under the external Auto Run `Working` folder, reusing `reviewTransport`, `dataFlowTransport`, existing HTTP fixtures, and the prior logic/data-flow overlays. No new production helpers were created:

- `Working/keenable_error_handling_review_test.go`: large and echoed-sensitive error bodies, failed body reads, JSON error types, empty bodies, cleanup, and real HTTP timeout propagation. Assertions documenting flaws characterize current behavior; they are not the desired contracts for a fix.
- `Working/keenable_cli_error_review_test.go`: transport-cause and file-write failure propagation, with existing-file preservation.
- `Working/error-handling-review-overlay.json`: combines these with the earlier logic and data-flow review tests.

All five full package suites passed with `go test -race -timeout 90s -overlay /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/error-handling-review-overlay.json ./internal/tools/keenable ./internal/cli ./internal/core ./internal/plugins ./internal/i18n`. After adding the final CLI cases, the complete CLI suite was run again with the same race/overlay settings. Review tests remain outside the repository, following prior review tasks; reproducing them requires that Auto Run folder.

The first timeout fixture waited only for server-side request cancellation and delayed server cleanup; it was replaced with a bounded handler before the successful runs. All final fixtures are local servers or in-process transports. No live API/model calls, actual secrets, visual checks, or images were involved; zero images analyzed. Security-specific evaluation and remaining playbook tasks are still pending.
