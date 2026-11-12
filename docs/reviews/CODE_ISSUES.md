---
type: report
title: PR 2330 Code Review Issues
created: 2026-11-11
tags:
  - code-review
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-logic-review]]'
  - '[[pr-2330-data-flow-review]]'
  - '[[pr-2330-code-style-review]]'
  - '[[pr-2330-error-handling-review]]'
  - '[[pr-2330-code-organization-review]]'
  - '[[pr-2330-api-compatibility-review]]'
  - '[[pr-2330-interface-contract-review]]'
---

# Code Review Issues

Consolidated findings for PR 2330, **feat: add --keenable_search web search (no API key required)**, using fixed base `de0b72766db0436b6c1887bb75768cb3960e67e5` and head `fb296d490cca4fdebbcb78d5107241f134152d19`. Coverage is all 25 changed files (320 additions, 3 deletions) in [[REVIEW_SCOPE]]. Line numbers refer to that head; all scoped working-tree files match it byte for byte. Production source and original PR tests remain unchanged. This report consolidates completed reviews; it does not implement fixes or replace the subsequent security review.

**Outcome:** no critical issues, one major issue, three minor issues, and four optional suggestions. No logic, data-flow, organization, backward-compatibility, or interface-contract defect was confirmed beyond the findings below.

## Critical Issues

None identified in the completed code-quality review. This is not a security sign-off.

## Major Issues

### M1: Bound response bytes before reading or decoding

- **Severity:** Major.
- **File and lines:** `internal/tools/keenable/keenable.go:94` and `internal/tools/keenable/keenable.go:100`.
- **Issue:** non-200 responses use unbounded `io.ReadAll`, then copy the entire body into an error. HTTP 200 responses decode unrestricted JSON before any snippet truncation. A faulty service or proxy can cause excessive allocation or memory exhaustion; large failure bodies also flood stderr. The 30-second timeout limits time rather than bytes, and outbound `max_results: 10` does not enforce an inbound size limit. Adjacent providers share this pattern, but the new client introduces another affected path.
- **Evidence:** the review-only mocked 503 with a 2 MiB body returns that entire body in its error. The unrestricted success path is established by source inspection; no memory-exhaustion stress test was attempted.
- **Suggested fix:** enforce a maximum successful-response byte budget with overflow detection before decoding. Apply a smaller error-diagnostic budget, preserve status, and explicitly mark truncation. A capped reader alone must not silently accept an oversized valid JSON prefix.
- **Source:** [[pr-2330-error-handling-review]].

## Minor Issues

### m1: Avoid returning raw service diagnostics containing sensitive values

- **Severity:** Minor (conditional disclosure).
- **File and line:** `internal/tools/keenable/keenable.go:95`; propagation through `internal/cli/tools.go:129`, `internal/cli/cli.go:99`, and `cmd/fabric/main.go:15`.
- **Issue:** an untrusted error body is returned verbatim and ultimately printed to stderr. If a service or gateway echoes an API key or private query, Fabric copies it into output and captured logs. The client does not itself append the request key/query to this error; disclosure depends on the returned body. No actual Keenable disclosure was observed.
- **Evidence:** a mocked 401 containing only fake values retains `review-fake-key` and `review-private-query` in the returned error.
- **Suggested fix:** select bounded, safe diagnostic fields or return a generic status message. Redact known request secrets and sanitize control characters; key redaction alone does not protect arbitrary private query text. The response-size cap in M1 does not by itself address disclosure.
- **Source:** [[pr-2330-error-handling-review]].

### m2: Preserve failures while reading non-200 response bodies

- **Severity:** Minor.
- **File and line:** `internal/tools/keenable/keenable.go:94`.
- **Issue:** `body, _ := io.ReadAll(resp.Body)` discards read failures. An interrupted response still returns an HTTP failure, but its diagnostic can be incomplete without explaining why, and callers cannot recover the read cause with `errors.Is`.
- **Evidence:** a synthetic 502 body returns partial text followed by `io.ErrUnexpectedEOF`; the returned error contains the partial text but loses the read cause. The success-path JSON decoder preserves the corresponding cause.
- **Suggested fix:** preserve HTTP status and safe partial diagnostics while wrapping the body-read error. Distinguish an intentional diagnostic cap from a transport failure.
- **Source:** [[pr-2330-error-handling-review]].

### m3: Keep fatal test assertions in the test goroutine

- **Severity:** Minor.
- **File and lines:** `internal/tools/keenable/keenable_test.go:25` (also lines 26–32 and 50–51).
- **Issue:** HTTP server handlers call `require` assertions with the parent test's `*testing.T`. On failure, these assertions invoke `FailNow` from a server goroutine, contrary to its test-goroutine requirement. This can terminate the handler before a response is written, adding an EOF error that obscures the request mismatch. Passing assertions do not demonstrate a current production failure.
- **Evidence:** the handler callbacks contain fatal assertions; passing package suites do not exercise this failure path.
- **Suggested fix:** capture request data/errors through a channel and assert in the test goroutine, or use nonfatal assertions in the handler with an explicit error response and return when decoding fails.
- **Source:** [[pr-2330-code-style-review]].

## Suggestions

| ID | Severity | File and line | Optional improvement |
| --- | --- | --- | --- |
| S1 | Suggestion | `internal/tools/keenable/keenable_test.go:13` | Add `t.Helper()` to `withServer` to attribute future helper diagnostics to callers. Existing cleanup already restores endpoints and closes the server. |
| S2 | Suggestion | `internal/cli/tools.go:126` | If search routing grows further, consolidate the repeated search/append/write-or-chat sequence shared with Serply and Firecrawl. Keep provider-specific configuration policy at call sites. The current short branch is readable; a broad refactor is unnecessary for this feature. This deduplicates the same suggestion from the style and organization reviews. |
| S3 | Suggestion | `internal/tools/keenable/keenable.go:20` and `internal/tools/keenable/keenable.go:27` | If parallel tests or independent client configurations are needed, move mutable endpoint/client settings into each client with constructor defaults. Current serial tests restore globals; no current race was demonstrated. |
| S4 | Suggestion | `internal/tools/keenable/keenable.go:60` and `internal/cli/flags.go:76` | Clarify constructor use, empty-result success, and omitted/missing/explicitly empty flag-value semantics in documentation. Empty results and whitespace-query forwarding are not classified as defects without a contrary service contract. |

These suggestions are optional maintenance improvements rather than merge prerequisites. See [[pr-2330-code-style-review]], [[pr-2330-code-organization-review]], and [[pr-2330-interface-contract-review]].

## Positive Observations

- Optional-key setup reuses existing plugin configuration and environment serialization. An absent or whitespace-only key selects the public endpoint without a key header; a trimmed nonempty key selects the keyed endpoint without mutating the stored setting.
- CLI dispatch reuses existing append/output/chat behavior, preserves input flags, and returns search/write failures. No new required-key gate blocks keyless searches.
- Request, transport, and JSON failures preserve their causes. Response bodies close on success and error paths; errors do not produce partial Markdown.
- Formatting preserves result order, skips URL-less entries, prefers descriptions, and truncates fallback snippets by Unicode runes.
- The flag, registry integration, configuration, and translations are additive. Existing signatures/settings remain intact; no deprecation or breaking-version increment is needed for this diff.
- All 14 locale catalogs add the same seven keys with compatible formatting placeholders. Help, README, changelog, and Bash/Zsh/Fish completions agree with the implemented flag and optional-key behavior.
- The provider stays small and follows existing module boundaries without adding dependencies, storage schemas, or a parallel configuration mechanism.

## Validation and Limits

For consolidation, reread the completed review reports, checked cited source locations, verified all 25 scoped paths against the fixed head, and passed fixed-diff whitespace checks. The existing graph index returned no Keenable symbols for the targeted search, so direct source inspection supplied those references.

Reused the existing review-only tests without adding duplicate fixtures. All seven complete package suites passed:

```sh
go test -race -timeout 90s -overlay /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/interface-contract-review-overlay.json ./internal/tools/keenable ./internal/tools/serply ./internal/tools/firecrawl ./internal/cli ./internal/core ./internal/plugins ./internal/i18n
```

Keenable, Serply, Firecrawl, plugins, and i18n returned cached passes; CLI and core returned fresh passes. Existing overlay cases cover key setup/persistence, requests, Unicode, malformed responses, response fields, cleanup, errors/timeouts, CLI parsing, output/chat routing, and unchanged state. Characterization tests for flaws assert current behavior, not the desired contract after a fix. Passing tests do not resolve M1 or the minor findings. Fixtures remain in the external Auto Run `Working` folder; reproducing the command requires that folder.

No live API/model calls, real credentials, or remote-schema verification were used. Translation prose quality and live-service availability are not established by these tests. Fish completion was reviewed statically in the earlier review; Bash and Zsh syntax checks passed there. Security-specific redirect/credential transport and external-content evaluation remain for Document 3. No images were associated with this task; zero images analyzed.

The requested artifact is `/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/CODE_ISSUES.md`. An identical version is tracked at `docs/reviews/CODE_ISSUES.md` so the consolidated findings are available with the review branch and its supporting reports.
