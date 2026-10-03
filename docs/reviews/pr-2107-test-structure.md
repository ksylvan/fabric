---
type: analysis
title: PR 2107 Test Structure Review
created: 2026-10-03
tags:
  - code-review
  - testing
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-related-tests]]'
  - '[[pr-2107-coverage-gaps]]'
---

# Test Structure Review

Completed Document 4's **Test structure** task for the two direct test files:
`internal/domain/file_manager_test.go` (2 test functions) and
`internal/core/chatter_test.go` (11 test functions). Review is pinned to PR base
`ddf1aab968caa9adf90137a6536c768d02b237d6` and head
`f32656b529a1eb1c678c793703fe49a0755cd439`. The working test files match the
pinned head. Supporting tests in [[pr-2107-related-tests]] are outside this
focused structure assessment.

## Findings by criterion

| Criterion | Assessment | Evidence |
| --- | --- | --- |
| Clear names describing behavior | Mostly satisfied | Parser cases name valid input, malformed JSON, empty paths, invalid operations, and traversal. Core test names describe prompt composition, think suppression, stream aggregation, buffering, metadata, and error behavior. `TestApplyFileChanges` names only the function; its create, nested-create, update, and rejection checks have no named subtests. |
| Proper setup and teardown | Mostly satisfied, with cleanup exceptions | Application fixtures use `os.MkdirTemp` and deferred `os.RemoveAll`; core database fixtures use `t.TempDir`. Strategy setup uses `t.Setenv`, which restores the environment. Stdout capture and timeout handling need stronger cleanup as detailed below. |
| Single assertion focus where appropriate | Mostly satisfied | Core tests group assertions around one behavior, such as aggregated assistant output or separation of prompt sections. Parser subtests independently validate each input. Application tests combine multiple behaviors in one sequential scenario, making independent selection and failure diagnosis harder. |
| No test interdependencies | No explicit cross-test dependency found; within-test coupling remains | Tests create their own database/filesystem fixtures and do not call each other or use `t.Parallel`. BuildSession table cases share read-only pattern fixtures. Application update depends on the earlier create, and stdout capture temporarily changes process-wide state. Static review cannot establish order independence or race freedom. |

## Improvements

### Medium: guarantee stdout cleanup

`internal/core/chatter_test.go:501`,
`TestChatter_Send_StreamingBufferStreamDoesNotPrint`, replaces `os.Stdout` and
restores it only after `Send` returns normally. It closes the writer but never
closes the pipe reader. A recovered panic or a later early exit introduced in
this interval could leave the global output redirected; the reader also leaks
a descriptor until process exit.

Register restoration and closure of both pipe ends immediately after successful
pipe creation, using `t.Cleanup` or deferred cleanup. Keep the explicit writer
close before reading so the reader sees EOF. The test currently runs serially;
process-wide stdout capture needs to remain serial or use an injectable output
writer if such an abstraction is introduced later.

### Medium: bound the worker lifetime in the deadlock regression

`internal/core/chatter_test.go:397`,
`TestChatter_Send_StreamingErrorUpdateAndReturnDoesNotDeadlock`, starts `Send` in
a goroutine and fails after two seconds. Its buffered result channel prevents a
late result send from blocking, but the timeout does not stop a stuck `Send`.
On that failure path, the worker can outlive the test and its temporary fixture,
and can continue touching shared output or removed directories.

Use a cancellable context and guarantee cancellation in cleanup for cooperative
blocking paths. Cancellation alone cannot terminate an arbitrary deadlock;
consider subprocess isolation if the regression must exercise a truly stuck
worker without leaving it in the test process. This is a structural risk on the
failure path, not an observed leak from test execution.

### Low: give application scenarios independent names and fixtures

`internal/domain/file_manager_test.go:119`, `TestApplyFileChanges`, groups
root/nested creation, update, and three rejected paths in one fixture. A fatal
setup/application/read error in an earlier phase prevents later checks from
running. The rejection loop includes the path in its failure message, but its
cases cannot be selected individually with `go test -run`.

Use named subtests for create, nested-create, update, and each rejection path.
Give each its own `t.TempDir`; arrange an update fixture explicitly within the
update case. Preserve multiple assertions when they describe the same behavior
(e.g. successful application and expected file content). Prefer `t.TempDir`
over manual temporary-directory removal for integration with test cleanup.

## Validation and limits

Used knowledge-graph symbol discovery and source snippets for all 13 direct test
functions. Verified that both working test files have no diff against the pinned
PR head. No repository `AGENTS.md` or `CLAUDE.md` was found during orientation;
the supplied session instructions apply.

This task assesses structure only. No source or test code changed, so no new
tests were warranted or executed. Mock realism, assertion adequacy,
maintainability, and suite execution remain separate unchecked tasks. No images
were supplied or analyzed (0 images).
