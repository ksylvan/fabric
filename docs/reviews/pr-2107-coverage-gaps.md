---
type: report
title: PR 2107 Coverage Gap Inventory
created: 2026-10-03
tags:
  - code-review
  - testing
  - security
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-modified-code-coverage]]'
  - '[[pr-2107-related-tests]]'
---

# Coverage Gap Inventory

Completed Document 4's **Coverage gaps** task against pinned base
`ddf1aab968caa9adf90137a6536c768d02b237d6` and head
`f32656b529a1eb1c678c793703fe49a0755cd439`. This is a static inventory of
missing behavioral assertions, not measured line coverage or test execution.
The PR adds no public functions or methods; that category is not applicable.

## New branches and conditions

| Priority | Source and symbol | Missing coverage | Suggested regression cases |
| --- | --- | --- | --- |
| High | `internal/core/chatter.go:192`, `Chatter.Send` | No direct test combines `create_coding_feature`, a file-change response, and the new channel condition. | With a mocked vendor and isolated working directory, cross nil/non-nil `UpdateChan` with streaming/non-streaming chatter. Assert nil-channel cases apply valid changes and save the summary; non-nil cases create/overwrite no target files and retain the full assistant payload. Consume streamed updates to avoid a blocked send. Include another pattern as a no-write control. |
| High | `internal/domain/file_manager.go:82`, `ParseFileChanges` | Existing traversal rejection cannot distinguish the new check from the old substring check; absolute-path rejection is untested here. | Reject an absolute path derived from a temporary directory and escaping traversal. Accept `file..txt` and `subdir/../file.txt`; assert exact returned paths, operations, content, and summary rather than only count. Test mixed valid/invalid entries return an error and no changes. |
| Medium | `internal/domain/file_manager.go:157`, `ApplyFileChanges` | New rejection cases use only `create`; local-path edge behavior and empty direct-call paths are untested. | Repeat rejection inputs for `update`, preserving an existing sentinel. Reject an empty path. Verify successful writes for `file..txt` and normalized local paths. Add native Windows absolute/UNC/drive-relative/reserved-name cases on Windows; do not assume Unix treats backslashes as separators. |

## Error handling and filesystem effects

| Priority | Source and symbol | Missing coverage | Suggested regression cases |
| --- | --- | --- | --- |
| High | `internal/domain/file_manager.go:157-168`, `ApplyFileChanges` | Rejection tests assert only an error, so unrelated filesystem failures could mask a validation regression; no assertion proves absence of outside writes. | Use controlled temporary parent/root directories and outside sentinels for absolute and traversal targets. Assert path-validation failure, unchanged sentinels, and no unexpected files/directories for rejected single changes. Avoid system paths such as `/etc` as writable targets. |
| High | `internal/domain/file_manager.go:161-168`, `ApplyFileChanges` | No containment test covers parent-directory or final-file symlinks; lexical validation alone does not test resolved containment. | Place symlinks inside the project pointing at a controlled outside directory/file; attempt create/update through them and check outside sentinels. If containment is the required contract, expect rejection and no outside write; current code may fail that security regression. Skip only where symlink creation is unavailable. |
| Medium | `internal/domain/file_manager.go:156-170`, `ApplyFileChanges` | No mixed-batch or filesystem-error case establishes side effects before a later failure. | Apply a valid entry followed by an invalid path; assert the current partial-write behavior explicitly, or define an atomic contract before expecting rollback. Use a regular file where a parent directory is needed to exercise `MkdirAll` failure, and a directory as the target to exercise `WriteFile` failure. Assert error context and unchanged unrelated files. |
| Medium | `internal/core/chatter.go:194-210`, `Chatter.Send` | Coding-pattern parse/application warning paths and resulting assistant content lack tests. | Return invalid-path JSON and malformed JSON from a mocked vendor; assert no file writes and the warning/assistant-content behavior. Force an application failure with a conflicting directory; assert whether Send returns an error, what summary is persisted, and any partial writes. These are pre-existing branches affected by the new gate. |

## Integration points

| Priority | Source and symbol | Missing coverage | Suggested regression cases |
| --- | --- | --- | --- |
| High | `internal/server/chat.go:71`, `ChatHandler.HandleChat`; `internal/cli/chat.go:20`, `handleChatProcessing`; `internal/core/chatter.go:192`, `Chatter.Send` | No end-to-end coding-pattern test verifies caller setup preserves the API/CLI boundary. The server sets a channel, while CLI options come from `BuildChatOptions`; a unit test of Send alone cannot guard caller changes. | Through the REST handler, deliver a coding-pattern response from a test vendor and assert emitted content/completion plus no target-file mutation. Through CLI processing, verify intended application in both streaming and non-streaming modes and resulting output/session content. Reuse existing vendor and filesystem fixtures. Non-streaming core coverage is a gate permutation, not a claim that this REST handler currently supports non-streaming requests. |

## Existing coverage and limits

`TestParseFileChanges` already checks normal/no-marker input, malformed JSON,
invalid operation, empty path, and one traversal path. `TestApplyFileChanges`
checks real create/nested-create/update contents and three single-create path
rejections. Existing Send tests cover other chat behavior but do not use the
coding pattern. Adjacent server path tests do not substitute for file-application
containment tests.

The table distinguishes new branch gaps from pre-existing error/security paths
that matter to this change. It does not declare all proposed contracts implemented:
symlink containment and atomic application require explicit expectations, and
regressions may expose current implementation limitations. Other unchanged parser
branches are outside this focused inventory.

## Evidence and validation

Used the existing `fabric-pr-2107-codex` graph to discover symbols, read both
file-manager tests and modified source functions, trace Send callers, and inspect
the server/CLI integration snippets. Graph-backed searches across core, domain,
server, and CLI Go tests found file-manager calls only in its two direct tests
and no coding-pattern, symlink, or size-limit test literals in that scope.
Absence claims are scoped to this inspected suite and corroborate
[[pr-2107-modified-code-coverage]]. Verified the reviewed source/direct tests have
no local diff from the pinned head. Root `AGENTS.md` and `CLAUDE.md` are absent;
supplied session instructions apply.

Documentation only: no source or test changes, no tests added or executed, and
no pass/fail claims. Test quality, suite execution, and consolidated
`TEST_GAPS.md` remain separate playbook tasks. No images supplied or analyzed
(0 images).
