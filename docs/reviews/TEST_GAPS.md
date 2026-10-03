---
type: report
title: Fabric PR 2107 Test Coverage Review
created: 2026-10-03
tags:
  - code-review
  - testing
  - security
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-coverage-gaps]]'
  - '[[pr-2107-test-assertions]]'
  - '[[pr-2107-test-structure]]'
  - '[[pr-2107-test-maintainability]]'
  - '[[pr-2107-test-mocking]]'
  - '[[pr-2107-test-execution]]'
  - '[[4_VERIFY_TESTS]]'
---

# Test Coverage Review

Consolidated review for PR https://github.com/danielmiessler/Fabric/pull/2107,
pinned to base `ddf1aab968caa9adf90137a6536c768d02b237d6` and head
`f32656b529a1eb1c678c793703fe49a0755cd439`. Changed source:
`internal/core/chatter.go` and `internal/domain/file_manager.go`; changed tests:
`internal/domain/file_manager_test.go`. Direct review covers 2 domain and 11
core test functions. No new public functions or methods were introduced.

Existing suites passed, but the security behavior lacks direct regression
coverage. The PR adds three rejection inputs to `TestApplyFileChanges`; it adds
no parser or file-write channel-gate regression. Priorities below rank test
work, not vulnerability severity. Missing tests are distinguished from weak
assertions and pre-existing maintenance concerns; overlapping findings are
counted once. No line/branch coverage percentage was measured.

## Missing Tests

| ID / Priority | Source file and function/method | Missing behavior | Suggested test cases |
| --- | --- | --- | --- |
| M1 / High | `internal/core/chatter.go`, `Chatter.Send` (gate at line 192) | Coding-pattern file application with nil/non-nil `UpdateChan`. Existing Send fixtures never reach this branch. | Cross streaming/non-streaming with nil/non-nil channel, using `create_coding_feature`, a valid file-change payload, and the existing `mockVendor`. Assert nil-channel application and summary persistence; non-nil channel leaves target files unchanged and preserves full assistant content. Add another-pattern no-write control. Consume updates concurrently. |
| M2 / High | `internal/domain/file_manager.go`, `ParseFileChanges` | New lexical validation distinctions are not tested: existing traversal input also failed the old substring check. | Reject a sandbox-derived absolute path and escaping traversal; accept `file..txt` and `subdir/../file.txt`. Assert exact decoded changes and summary. A mixed valid/invalid payload must return an error and no applicable changes. |
| M3 / High | `internal/domain/file_manager.go`, `ApplyFileChanges` | Resolved containment through intermediate or final-file symlinks. | Inside an isolated project root, create symlinks to controlled outside directories/files; exercise create/update and assert outside sentinels remain intact. If containment is the contract, expect rejection. Prior correctness review confirmed an intermediate-symlink escape; this desired security regression requires a production fix, not a test expecting the unsafe behavior. Skip only if symlink creation is unavailable. |
| M4 / High | `internal/server/chat.go`, `ChatHandler.HandleChat`; `internal/cli/chat.go`, `handleChatProcessing`; `internal/core/chatter.go`, `Chatter.Send` | Caller-level API/CLI write-boundary regression. | Through REST, supply a coding-pattern response from a fake vendor and assert content/completion plus no file mutation. Through CLI, assert intended writes and output/session behavior in streaming and non-streaming modes. REST currently supplies a channel; this is protection against future caller regressions, not a demonstrated present bypass. |
| M5 / Medium | `internal/domain/file_manager.go`, `ApplyFileChanges` | Update rejection, empty direct-call paths, accepted local edge paths, and native platform cases. | Reject empty paths and absolute/traversal update paths while preserving a sentinel; accept safe dotted/normalized local paths. On Windows, include drive-relative, absolute, UNC and reserved-name cases. Do not impose Windows separator semantics on Unix. |
| M6 / Medium | `internal/domain/file_manager.go`, `ApplyFileChanges` | Later-entry failure and filesystem errors. | Submit a valid entry followed by an invalid path; characterize current partial writes. Use a file as a required parent directory for `MkdirAll` failure and a directory as the write target for `WriteFile` failure. Assert contextual errors and unrelated-file preservation. Define an atomic contract before expecting rollback. |
| M7 / Medium | `internal/core/chatter.go`, `Chatter.Send` | Coding-pattern parse/application warnings and resulting assistant/session content. | Return malformed JSON or an invalid path; assert warning/content behavior and no writes. Force application failure with a directory conflict; assert returned error policy, persisted summary/payload and any partial writes. These branches predate the gate. |

## Inadequate Tests

| ID / Priority | Source and test function | Inadequacy | Suggested test cases/assertions |
| --- | --- | --- | --- |
| I1 / High | `internal/domain/file_manager.go`, `ApplyFileChanges`; `internal/domain/file_manager_test.go`, `TestApplyFileChanges` | Rejection cases accept any error and never prove absence of outside writes. `/etc/escape.txt` can mask failed validation with a permission error and depends on host layout. | Use a controlled temporary parent/root and absolute/traversal outside sentinels. Assert validation failure, unchanged/absent outside targets and no unexpected directories. Give create/update rejection cases independent names. |
| I2 / Medium | `internal/domain/file_manager.go`, `ParseFileChanges`; `internal/domain/file_manager_test.go`, `TestParseFileChanges` | Success checks only change count; summary and decoded values are discarded. | Compare ordered operation/path/content fields and summary. No-marker input should preserve the original summary; rejected input should return no applicable changes. Combine with M2 without duplicating fixtures. |
| I3 / Medium | `internal/core/chatter.go`, `Chatter.Send`; `internal/core/chatter_test.go`, `TestChatter_Send_StreamingMetadataPropagation` | Only total tokens is checked, leaving input/output usage and forwarding incomplete. | Assert fixture usage values 10/5/15 and, if one-for-one forwarding is intended, update type/content/count/order. Collect updates concurrently. |
| I4 / Low | `internal/core/chatter.go`, `recordFirstStreamError` and `Chatter.Send`; `internal/core/chatter_test.go`, `TestRecordFirstStreamError_ChannelFull`, `TestChatter_Send_StreamingErrorUpdateAndReturnDoesNotDeadlock` | First-error helper checks text rather than identity; deadlock test does not specify precedence of returned vs update errors. | Retain the original error and assert identity with `errors.Is` where appropriate. Add a separate precedence case only after defining the contract; preserve bounded completion checks. |
| I5 / Low | `internal/core/chatter.go`, `Chatter.Send` and `Chatter.BuildSession`; `internal/core/chatter_test.go`, `TestChatter_Send_StreamingBufferStreamDoesNotPrint`, `TestChatter_BuildSession_EndsWithUserMessage` | Stdout capture errors are ignored; delimiter-encoded message comparisons can become ambiguous as fixtures grow. | Check pipe read/close errors and returned-session presence; compare role/content sequences directly with a delimiter-containing fixture. |

## Test Quality Issues

| ID / Priority | Source file and function/method | Issue | Suggested improvement/validation |
| --- | --- | --- | --- |
| Q1 / Medium | `internal/core/chatter_test.go`, `TestChatter_Send_StreamingBufferStreamDoesNotPrint` (exercises `Chatter.Send`) | Global stdout restoration occurs only after normal return; reader is not closed; post-call draining can block on larger output. | Register cleanup immediately for restoration and both pipe ends; keep serial execution, check capture errors, and drain concurrently for larger fixtures. |
| Q2 / Medium | `internal/core/chatter_test.go`, `TestChatter_Send_StreamingErrorUpdateAndReturnDoesNotDeadlock` | Unexplained two-second timeout; stuck Send worker can outlive test fixtures. | Name/document a bounded timeout, use guaranteed context cancellation for cooperative work, and consider subprocess isolation for arbitrary deadlocks. This is a failure-path risk, not an observed flake. |
| Q3 / Medium | `internal/core/chatter_test.go`, `TestChatter_Send_StreamingMetadataPropagation` | Fixed channel capacity 10 is drained only after Send and can deadlock if the fixture grows. | Collect concurrently with explicit completion/channel ownership; verify bounded return when increasing fixture size beyond the old buffer. |
| Q4 / Medium | `internal/core/chatter_test.go`, `mockVendor.SendStream` | Fake ignores cancellation and uses unconditional sends. | Extend the existing fake with an optional context-aware stream callback for cancellation/backpressure cases; coordinate channels rather than sleeps. Assert bounded return after cancellation. Current mock remains adequate for ordinary delivery. |
| Q5 / Low | `internal/domain/file_manager_test.go`, `TestApplyFileChanges`, `TestParseFileChanges`; `internal/core/chatter_test.go`, Send tests and `TestChatter_BuildSession_SeparatesSystemSections` | Sequential application phases share state; JSON/Send setup repeats; strategy fixture hardcodes storage layout. | Use named independent application subtests and explicit update setup. Reuse `mockVendor` and storage APIs; introduce a small fixture helper only if repetition warrants it, with fresh state per call. Keep malformed JSON literal and expected values independent. |
| Q6 / Low | `internal/core/chatter_test.go`, `mockVendor.Send`, `mockVendor.SendStream` (exercises `Chatter.Send`) | Fixed responses do not establish request forwarding, raw-mode selection or provider configuration failures. | Where those contracts need coverage, use existing `sendFunc` to inspect meaningful inputs or a focused stream callback. Do not cite aggregation tests as provider integration coverage. |

## Test Improvements

Implement M1/M2/I1 first: these directly guard the changed security conditions.
Add M4 to protect caller configuration and M3 alongside the containment fix.
Then cover the Medium error/edge cases and strengthen existing assertions.
Reuse the existing vendor fake and real temporary filesystem/database fixtures;
mocking the writer would conceal the behavior under review. Coordinate stream
consumers and cleanup explicitly. Keep platform-specific cases native to their
platform and document partial-write expectations.

These are proposed tests and improvements, not changes made by this review.
The PR's shell-injection claim is not supported by its three-file diff; this
report makes no claim of new template-extension coverage.

## Well-Tested Areas

- `TestParseFileChanges` exercises ordinary/no-marker output, malformed JSON,
  invalid operation, empty path and traversal; named cases isolate inputs.
- `TestApplyFileChanges` reads real root/nested-create and update contents and
  adds three direct-writer rejection inputs in this PR.
- Core tests check prompt section formatting, message roles/content, think
  suppression, stream aggregation, buffered output and propagated errors.
- Real temporary filesystem/database fixtures exercise local behavior while
  `mockVendor` isolates live AI services. No material over-mocking was found.
- Test names mostly express behavior; no explicit cross-test dependency was
  found. Within-test fixture coupling and process-global stdout remain caveats.

## Test Execution Results

Prior execution on 2026-10-03, recorded in [[pr-2107-test-execution]] and checked
against retained logs during consolidation:

| Command | Result |
| --- | --- |
| `go test -count=1 -timeout=5m ./...` (repository root) | Exit 0; 35 packages passed, 20 had no test files. |
| `npm ci --no-audit --no-fund` (`web`) | Exit 0; 442 packages installed and SvelteKit sync succeeded. |
| `npm test -- --run` (`web`) | Exit 0; 36 tests passed across 7 files. |

Both suite logs reported no failures or warnings. Installer warnings concerned
`svelte-reveal`, `eslint` and `lucide-svelte` deprecations, a SvelteKit Vite
output-format override, and the blocked `fsevents` install script. They did not
prevent the frontend suite passing; no manifest/policy changes were made.

Host was darwin/arm64 with Go 1.27.1, Node v25.9.0, npm 11.12.1 and Vitest
4.1.11. Logs are in the Auto Run folder's `Working/pr-2107-go-suite.log`,
`Working/pr-2107-npm-ci.log` and `Working/pr-2107-web-suite.log`.
No coverage collection, race detector, Windows/Linux run, live-provider test
or deployed end-to-end validation was performed. Passing existing tests does
not establish coverage of the gaps above.

## Evidence and Limits

Consolidated [[pr-2107-new-tests]], [[pr-2107-modified-code-coverage]],
[[pr-2107-coverage-gaps]] and the linked structure/mocking/assertion/
maintainability/execution reports. Rechecked writer source through the existing
knowledge graph and confirmed no working source/test/manifest differences from
the pinned PR head. Root `AGENTS.md` and `CLAUDE.md` are absent; supplied session
instructions apply. Prior security/correctness findings remain in
[[CODE_ISSUES]] and [[pr-2107-security-issues]].

This task changes documentation only. No tests were added or rerun because
behavior is unchanged and prior suite evidence is available. Validated required
sections, report-copy equality and checkbox preservation. The required Auto Run
`TEST_GAPS.md` and repository `docs/reviews/TEST_GAPS.md` are identical; the
repository copy preserves the review in the task commit. No images supplied or
analyzed (0 images).
