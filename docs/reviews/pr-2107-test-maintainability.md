---
type: analysis
title: PR 2107 Test Maintainability Review
created: 2026-10-03
tags:
  - code-review
  - testing
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-test-structure]]'
  - '[[pr-2107-test-assertions]]'
  - '[[pr-2107-coverage-gaps]]'
---

# Test Maintainability Review

Assessed the 2 domain and 11 core direct test functions in
`internal/domain/file_manager_test.go` and `internal/core/chatter_test.go`.
Scope is pinned to base `ddf1aab968caa9adf90137a6536c768d02b237d6` and
head `f32656b529a1eb1c678c793703fe49a0755cd439`; both working test files
match that head. Findings describe static risks, not observed test failures.

## Findings by criterion

| Criterion | Assessment | Evidence |
| --- | --- | --- |
| Magic numbers without explanation | Two maintenance concerns | The deadlock guard uses 2 seconds without rationale; the metadata test uses channel capacity 10 unrelated to fixture length. Modes `0o755`/`0o644`, single-result channel capacity 1, expected message counts, and token fixture values 10/5/15 have clear conventional or local meaning. |
| Overly complex setup | Mostly straightforward, some repetition | Streaming tests repeatedly construct a temporary database, fake vendor, Chatter, request, and options. The system-section test manually arranges three filesystem entities and a temporary HOME; this supports its integration behavior but couples setup to the strategy storage path. |
| Brittle tests | Environment, timing, and fixture-growth risks | Absolute `/etc/escape.txt`, a fixed timeout, global stdout capture, and a channel drained only after Send returns create avoidable dependencies. Exact whitespace assertions are appropriate where formatting is the behavior under test. |
| Test data duplication | Moderate, localized | Parser cases repeat marker/JSON envelopes; application tests repeat content/path literals; six Send tests repeat model/request setup. Existing `mockVendor`, named parser cases, and named prompt cases already provide useful reuse and readability. |

## Suggested improvements

### Medium: make rejected paths independent of host filesystem layout

`internal/domain/file_manager_test.go:177`, `TestApplyFileChanges`, uses
`/etc/escape.txt`. Its meaning and OS permissions depend on the host; on Windows,
this path need not have the same absolute-path semantics. A validation regression
could reach a real host path or be masked by a permission error.

Use an absolute target derived from a controlled temporary sandbox, with a
separate application root inside it. Give each rejected path a named subtest
and assert outside targets remain unchanged or absent. Keep platform-specific
cases explicit. This is the same weak rejection oracle identified in
[[pr-2107-test-assertions]], assessed here for portability and fixture ownership.
No host writes were attempted during this review.

### Medium: explain and bound the deadlock timeout

`internal/core/chatter_test.go:359`,
`TestChatter_Send_StreamingErrorUpdateAndReturnDoesNotDeadlock`, uses
`time.After(2 * time.Second)` to distinguish completion from deadlock. The guard
is useful, but a slow or instrumented runner can exceed it without deadlocking.
The result channel's capacity 1 is appropriate for the single worker result.

Name the timeout and explain its budget; retain a bounded failure rather than
removing the guard. Guarantee cancellation for cooperative work and account for
the overall test deadline. A truly stuck worker may require subprocess
isolation. Worker cleanup is also documented in [[pr-2107-test-structure]].
This review does not establish that the existing test flakes.

### Medium: remove the metadata fixture's hidden buffering assumption

`internal/core/chatter_test.go:523`,
`TestChatter_Send_StreamingMetadataPropagation`, allocates an update channel
with capacity 10 and drains it only after synchronous Send returns. It currently
supplies two updates. Expanding the fixture or forwarding more updates can
exhaust the fixed buffer and block Send before the consumer starts.

Prefer collecting updates concurrently with explicit collector completion and
channel ownership. If a deliberately bounded fixture retains post-call draining,
derive and document capacity from the expected update count and enforce a
bounded Send duration. Token values 10/5/15 are readable data, not unexplained
magic constants; comparing all populated fields is an assertion improvement.

### Medium: preserve isolation when extending stdout capture

`internal/core/chatter_test.go:475`,
`TestChatter_Send_StreamingBufferStreamDoesNotPrint`, replaces process-wide
`os.Stdout`, restores it after Send, and reads the pipe only after closing the
writer. It must remain serial; cleanup should be registered immediately and
close both descriptors. If a future fixture or regression emits enough bytes,
a synchronous writer can fill the pipe before the test starts reading.

Use concurrent draining for larger output and check capture errors. Reuse an
existing output injection abstraction if one becomes available rather than
adding multiple capture helpers. These are failure-path and fixture-growth
risks; the current fixture is small. See [[pr-2107-test-structure]].

### Low: reduce repeated setup without hiding behavioral differences

The six Send tests (`SuppressThink`, `StreamingErrorPropagation`,
`StreamingErrorUpdateAndReturnDoesNotDeadlock`, `StreamingSuccessfulAggregation`,
`StreamingBufferStreamDoesNotPrint`, and `StreamingMetadataPropagation`) repeat
`fsdb.NewDb(t.TempDir())`, model name, user message, vendor, and options setup.
Continue using the existing `mockVendor`. If this repetition grows, consider a
small fixture constructor with `t.Helper` and fresh fixtures on every call;
keep Stream, UpdateChan, BufferStream, and error/chunk inputs visible in each
case. Graph discovery found no matching setupTest/newTest/makeTest/createTest
helper under core; no new helper is introduced by this review.

`TestChatter_BuildSession_SeparatesSystemSections` (`chatter_test.go:192`)
constructs `.config/fabric/strategies` directly. Prefer an existing storage API
or path resolver when revisiting the fixture so a storage-layout change does
not require updating unrelated prompt expectations. Its real filesystem setup
and temporary HOME are justified for the exercised integration.

### Low: simplify duplicated valid JSON and independent application cases

`TestParseFileChanges` (`file_manager_test.go:9`) repeats JSON envelopes and
content across cases. Additional valid cases could marshal explicit FileChange
fixtures or share a small envelope builder; retain literal malformed JSON so
invalid syntax remains intentional and visible. The marker is already reused
through `FileChangesMarker`.

`TestApplyFileChanges` (`file_manager_test.go:119`) repeats path/content strings
and couples update to earlier creation. Named independent cases with t.TempDir
and explicit update setup would make additions easier to select and diagnose.
Keep expected content readable and independently meaningful; avoid deriving
all expected results through the code being tested.

`TestChatter_BuildSession_EndsWithUserMessage` (`chatter_test.go:273`) serializes
message sequences with `:` and `|`. Direct role/content slice comparisons avoid
ambiguity if future fixture text includes those delimiters, as discussed in
[[pr-2107-test-assertions]]. Exact prompt and buffered-content comparisons remain
valuable behavioral contracts and should retain their whitespace checks.

## Validation and limits

Used the existing knowledge graph for symbol discovery and snippets of all 13
direct test functions and the existing mockVendor struct. Reviewed previous
structure and assertion findings to connect overlapping issues. Confirmed no
working-file differences from the pinned head. No repository AGENTS.md or
CLAUDE.md was present; supplied session instructions apply.

Documentation-only assessment: no source/test code changed, no new tests added,
and no tests executed. Suite execution and consolidated TEST_GAPS.md remain
separate unchecked tasks. No images were supplied or analyzed (0 images).
