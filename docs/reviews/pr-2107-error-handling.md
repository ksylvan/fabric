---
type: report
title: PR 2107 Error Handling Review
created: 2026-10-03
tags:
  - code-review
  - error-handling
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-correctness]]'
  - '[[pr-2107-data-flow]]'
  - '[[2_REVIEW_CODE]]'
---

# PR 2107 error handling review

Reviewed the **Error handling** task across `internal/core/chatter.go`, `internal/domain/file_manager.go`, and `internal/domain/file_manager_test.go`, including the REST caller and existing stream-error tests. Scope remains original PR head `f32656b529a1eb1c678c793703fe49a0755cd439` against base `ddf1aab968caa9adf90137a6536c768d02b237d6`; scoped source and REST handler match that head. No production code changed and no images were supplied or analyzed (0).

## Error handling assessment

| Boundary | Assessment |
| --- | --- |
| Parse and validate (`file_manager.go:30`, `:68`, `:77`) | Missing arrays, unmatched brackets, unrecoverable JSON, invalid operations, empty/nonlocal paths, and oversized content return explicit errors. Failed decoding retries existing escape repair once; final failure wraps the JSON error with `%w` and returns nil changes. Validation errors include a zero-based entry index and relevant field. Absence of the marker is intentionally a successful no-op. Existing repair/parser defects remain in [[pr-2107-data-flow]] and [[pr-2107-correctness]]. |
| Apply (`file_manager.go:155`) | The new nonlocal-path error occurs before mutation of that entry. Directory creation and file-write failures immediately return contextual errors with directory/file, index, and `%w`-wrapped cause. Later entries are not processed. Earlier mutations remain; this is not rollback. |
| Chat provider/session (`chatter.go:65`, `:173`, `:185`, `:215`) | Session construction, provider send, empty response, and final session persistence failures propagate through the named error return. Provider failure prevents file application. Empty response returns a nil session; other failure paths may return a nonnil session alongside an error. |
| CLI file application (`chatter.go:193`–`:211`) | Parse, current-directory, and application failures are caught and printed as warnings, then processing continues. The locally declared Getwd error does not set the named return error. The assistant response becomes the summary even after failure. This existing behavior is a diagnostic limitation, detailed below. |
| API file-write gate (`chatter.go:192`) | Current REST requests supply `UpdateChan`, so parsing/application and their local filesystem diagnostics are skipped. No newly added file error is routed to remote chat clients through this path. This conclusion depends on the present caller convention. |
| Changed tests (`file_manager_test.go:177`) | The added cases assert nonnil errors for three rejected paths. They do not assert cause types, diagnostics, no side effects, mkdir/write failures, or failure propagation to CLI callers. Existing successful write tests and parser error cases remain useful. |

## Findings and suggestions

### Minor: application failure is reported as successful chat completion

**Location:** `internal/core/chatter.go:193`–`:216`, especially `:205` and `:211`.

Parsing, Getwd, and application failures produce warning text but do not reach `Send`'s error return. If session saving succeeds (or is unnecessary), the caller sees nil error despite unsuccessful file application. The unconditional summary replacement can discard the attempted changes from the returned/saved assistant response; some parser boundary errors return the original response as their summary, so payload loss is not universal. Combined with partial application, callers cannot infer which writes occurred from the return value. This predates the PR; the new application path check adds another failure that direct application callers can encounter. Consolidate with the partial-write and payload-loss findings in [[pr-2107-correctness]] and [[pr-2107-data-flow]], rather than counting them again as independent regressions.

**Suggested fix:** make application status observable through the existing chat result/error contract, retain the original payload on failure, and report applied/failed entries when partial writes occur. Choose whether file application failure should fail the overall command explicitly; do not promise atomicity merely by propagating an error.

### Suggestion: warnings should use stderr and respect output policy

**Location:** `internal/core/chatter.go:195`, `:199`, `:202`, `:205`; related success output at `internal/domain/file_manager.go:173`.

Warnings use unconditional `fmt.Printf`, even when Quiet or buffered output is requested, whereas stream errors at `chatter.go:151` use stderr and honor Quiet. In CLI mode, warnings can contaminate stdout consumed by scripts. These are existing behaviors, unchanged by the gate.

**Suggested fix:** route warnings through the existing stderr/Quiet convention and keep machine-consumed response output predictable. Add coverage that captures stdout/stderr for a failed application and checks the chosen policy.

## Sensitive information and asynchronous errors

File-manager diagnostics contain operation/path metadata, filesystem paths, and wrapped OS/JSON errors; they do not explicitly dump the file-content string, entire model response, or credentials. The probes confirmed no content-sentinel disclosure for the tested mkdir, write, and JSON-syntax failures. Absolute paths can reveal local directory/user names, and JSON type errors may contain model-controlled field names; this is not proof that every error is safe to publish remotely. In the current REST path the gate avoids these file-manager diagnostics. Existing provider errors are still forwarded verbatim through stream updates or `server_chat_error`; arbitrary provider-secret redaction was not established by this review and is outside the changed file-write checks.

Go goroutines/channels implement the asynchronous path; promises and try/catch do not apply. `Send` uses a buffered error channel and `recordFirstStreamError` (`chatter.go:35`), so an error update plus a returned provider error cannot block on a second error enqueue. It waits for provider completion before reading the first error and returns before applying files. Existing tests cover propagation and the duplicate-error deadlock scenario. REST `HandleChat` uses a buffered `sendErrChan`, closes the stream after enqueueing a returned error, and `unreportedSendError` (`server/chat.go:226`) emits errors that did not arrive as updates while suppressing duplicates. Channel forwarding still assumes an active receiver and provider channel closure; client-disconnect cancellation/blocked-channel behavior was not exercised here. No new asynchronous mechanism is introduced by this PR.

## Validation and limits

- Used the existing codebase-memory graph for symbol discovery, snippets, and an inbound `ApplyFileChanges` trace; no new helper was introduced. No local CLAUDE.md or AGENTS.md exists; session instructions apply.
- Rechecked original three-file diff, unchanged scoped files, and English locale formats (`internal/i18n/locales/en.json:268`–`:275`) supporting `%w` wrapping and contextual messages.
- Added three temporary focused tests: deterministic mkdir failure using a file as parent, write failure using a directory as destination, and invalid JSON. All passed, proving wrapped causes, absence of later writes for tested filesystem failures, and absence of the content sentinel in tested diagnostics.
- Probe source retained in Auto Run `Working/review_error_handling_probe_test.go`; temporarily copied to `internal/domain` and removed after testing. To reproduce, copy it into that package, run `go test ./internal/domain -run TestReviewErrorHandlingProbes -v`, then remove the copy. Test temporary storage and build cache used the authorized Working folder.
- After removal, `go test ./internal/domain ./internal/core ./internal/server` passed. No permanent tests or source changes were added. Full repository, Windows, client-disconnect, and end-to-end CLI application-error tests were not run; warning-only behavior was established by source inspection.
- No newly introduced error-handling defect or confirmed content/credential leak was found in the changed logic. Existing limitations above remain actionable. Later organization, API-contract, consolidated-issues, and security tasks remain for future runs.
