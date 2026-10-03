---
type: reference
title: PR 2107 Related Test Inventory
created: 2026-10-03
tags:
  - code-review
  - testing
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-test-context]]'
---

# Related Test Inventory

Completed Document 4's first unchecked task, **Find related tests**. The
comparison remains pinned to base `ddf1aab968caa9adf90137a6536c768d02b237d6`
and head `f32656b529a1eb1c678c793703fe49a0755cd439`.

## Direct test files

| Changed source | Corresponding test file | Tests to inspect in later tasks |
| --- | --- | --- |
| `internal/domain/file_manager.go` | `internal/domain/file_manager_test.go` | `TestParseFileChanges` and `TestApplyFileChanges`: parser cases and filesystem create/update behavior, including path rejection cases. |
| `internal/core/chatter.go` | `internal/core/chatter_test.go` | Eleven tests: two BuildSession tests; six Send tests for think suppression, streaming errors, deadlock prevention, aggregation, buffered output, and metadata; two stream-error helper tests; and `TestJoinPromptSections`. |

The domain test file is also one of the three modified files in the pinned PR
diff. Identifying it here does not complete the separate **Check for new tests**
task or establish coverage of the changed branches.

## Supporting and adjacent tests

- `internal/core/plugin_registry_test.go`: chatter construction and vendor/model
  selection; supporting setup, rather than direct Send behavior.
- `internal/server/chat_test.go`: `TestBuildPromptChatRequest_PreservesStrategyAndUserInput`,
  `TestUnreportedSendError`, and `TestUnreportedSendErrorWithEmptyChannel`;
  request construction and reporting Send errors at the server boundary.
- `internal/server/path_traversal_test.go`: adjacent server path security tests;
  discovery alone does not establish coverage of domain file application.
- `internal/cli/chat_test.go`: image generation compatibility/integration and
  notification tests; no matching file-manager or Chatter symbols were found.
- `internal/cli/workflow_test.go`: workflow loading, validation, variables and
  step input/output tests; no matching file-manager or Chatter symbols were found.
- `web/src/lib/components/chat/Chat.test.ts` and
  `web/src/lib/components/chat/read-file-content.test.ts`: frontend chat tests,
  separate from the changed Go implementation.

## Discovery and validation

The existing `fabric-pr-2107-codex` knowledge graph was available, so no indexing
was required. Used graph symbol search, graph-backed code search across Go tests,
and source snippets for both file-manager test functions. Confirmed the core
and server test names against the pinned head with `git show`.

A filename inventory supplemented graph discovery for `*.test.*`, `*.spec.*`,
Go's `*_test.go`, and `__tests__/`, `test/`, and `tests/` directories. Relevant
Go tests are colocated with their packages; no matching dedicated test directory
was found. No repository `AGENTS.md` or `CLAUDE.md` was found, including a nested
instruction-file search; the supplied session instructions apply.

This task changes documentation only. No test cases were added or executed;
coverage assessment, test quality, and suite execution remain separate tasks.
No task images were supplied or analyzed (0 images).
