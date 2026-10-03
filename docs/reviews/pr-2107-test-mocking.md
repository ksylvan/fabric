---
type: analysis
title: PR 2107 Test Mocking Review
created: 2026-10-03
tags:
  - code-review
  - testing
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-related-tests]]'
  - '[[pr-2107-test-structure]]'
  - '[[pr-2107-coverage-gaps]]'
---

# Test Mocking Review

Completed Document 4's **Mocking** task for the direct tests in
`internal/core/chatter_test.go` and `internal/domain/file_manager_test.go`.
Comparison remains pinned to PR base `ddf1aab968caa9adf90137a6536c768d02b237d6`
and head `f32656b529a1eb1c678c793703fe49a0755cd439`. Both working test files
match the pinned head. Adjacent tests in [[pr-2107-related-tests]] are outside
this focused assessment.

## Findings by criterion

| Criterion | Assessment | Evidence |
| --- | --- | --- |
| External dependencies are mocked | Satisfied for AI services; local filesystem deliberately remains real | The six core Send tests inject `mockVendor` rather than a production provider. Its setup/configuration methods are inert; Send returns a fixture or invokes `sendFunc`, and SendStream emits fixture updates. No live provider credentials or network requests are needed in these cases. Core fixtures use real `fsdb.NewDb(t.TempDir())`; BuildSession reads real pattern/context/strategy files. Domain application tests create, update, and read actual files in a temporary directory. Parser tests need no external dependency. |
| Mocks are realistic | Adequate for the scenarios exercised, with limits | `mockVendor.SendStream` at lines 53–62 emits ordered typed updates, closes the response channel, then returns its configured error. Production `internal/plugins/ai/ollama/ollama.go`, `Client.SendStream`, likewise sends content/usage updates and closes the channel with a defer, including on error. Core fixtures cover content, usage, an error update, and a returned error, including both error mechanisms together. Non-streaming Send at lines 64–69 supplies text through the real Chatter processing path. |
| No over-mocking | No material over-mocking found | Chatter session construction, prompt assembly, aggregation, think filtering, output handling, domain parsing, and filesystem application are not replaced. Tests mostly inspect returned messages, output, errors, and forwarded updates. Mocking the vendor interface isolates remote model behavior without bypassing the code under review. Direct tests of `recordFirstStreamError` inspect an internal channel helper; this is a narrow helper test, not an excessive mock layer. |

## Limits and suggested improvements

### Medium: stream fake does not exercise cancellation or backpressure cancellation

`internal/core/chatter_test.go:53`, `mockVendor.SendStream`, discards the context
and uses unconditional channel sends. It models ordinary ordered delivery and
channel completion, but cannot establish prompt cancellation or unblock a send
when a downstream consumer stops. This is a limitation of the exercised
scenarios, not evidence that the production code fails cancellation.

For a cancellation regression, extend the existing fake with an optional
stream callback (following its existing `sendFunc` pattern) and use coordinated
channels plus context-aware sends. Verify bounded return when the context is
cancelled while an update cannot be delivered. Avoid sleeps or live services.
The worker-lifetime concern is separately recorded in [[pr-2107-test-structure]].

### Low: fixture responses do not validate vendor request inputs

`mockVendor.SendStream` ignores messages and options; the default non-streaming
Send returns fixed text. `NeedsRawMode` always returns false, and configuration
always succeeds. These choices are appropriate for aggregation and output tests,
but those tests cannot prove request forwarding, raw-mode selection, or provider
configuration failures. They should not be cited as provider integration coverage.

Where those behaviors need coverage, reuse `sendFunc` to inspect meaningful
message/options values and add an optional streaming callback only when needed.
Prefer observable results over exact internal call sequences. No new helper is
needed for the current review task.

### High: reuse this mock for the missing file-write gate regression

As already identified in [[pr-2107-coverage-gaps]], none of the direct Send
fixtures exercises `create_coding_feature` with model-produced file changes.
The absence of that case is a coverage gap, not over-mocking: the real file
application path remains available. Reuse `mockVendor` to return a valid file
change payload and real isolated filesystem fixtures to verify CLI application
and API suppression with nil/non-nil UpdateChan, for streaming and non-streaming
requests. Arrange the existing pattern fixture so execution reaches the gate,
and inspect file contents or absence rather than mocking ApplyFileChanges.

## Validation and limits

Used the existing knowledge graph for discovery and source snippets, then
confirmed the mock implementation directly because graph-augmented search
returned incomplete source coverage. Verified no diff between both direct test
files and the pinned PR head. No repository AGENTS.md or CLAUDE.md was available;
the supplied session instructions apply.

This is a static mocking review only. No source or test code changed, so no
new tests were added or executed. Assertions, maintainability, suite execution,
and consolidated TEST_GAPS.md remain separate unchecked tasks. No task images
were supplied or analyzed (0 images).
