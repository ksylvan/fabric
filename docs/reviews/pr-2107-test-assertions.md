---
type: analysis
title: PR 2107 Test Assertion Review
created: 2026-10-03
tags:
  - code-review
  - testing
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-related-tests]]'
  - '[[pr-2107-coverage-gaps]]'
  - '[[pr-2107-test-structure]]'
---

# Test Assertion Review

Completed Document 4's **Assertions** task for the 2 domain and 11 core direct
test functions. Review remains pinned to base
`ddf1aab968caa9adf90137a6536c768d02b237d6` and head
`f32656b529a1eb1c678c793703fe49a0755cd439`. Both working test files match
that head. Adjacent tests in [[pr-2107-related-tests]] are outside this assessment.

## Findings by criterion

| Criterion | Assessment | Evidence |
| --- | --- | --- |
| Meaningful assertions | Mostly satisfied, with incomplete oracles | Application success tests read actual root/nested/updated file contents. Core tests compare prompt sections, message roles/content, aggregated text, buffered stdout, and error identity. Parser success checks only list length; rejection checks only error presence; metadata checks only total tokens. |
| Correct expected values | No inconsistent expected value found in the inspected fixtures | The valid parser fixture contains 2 changes; no-marker input yields 0. File contents match create/update fixtures. Stream fragments form `Hello world! This is a test.`; buffered fragments preserve the Go fence and newlines. Think suppression expects `visible`. Prompt fixtures expect strategy/context/pattern separated by newlines and a separate user message. Metadata total 15 matches input 10 plus output 5. |
| Appropriate assertion types | Mostly appropriate | Direct string/count/role equality is suitable for deterministic results. `errors.Is` in streaming error propagation permits wrapping while preserving the original error. Fatal setup/length failures protect later indexing; nonfatal field checks allow independent failures to be reported. The first-error helper's string comparison is weaker than error identity, and stdout read errors are ignored. |

## Assertion improvements

### High: path rejection does not establish absence of writes

`internal/domain/file_manager_test.go:177`, `TestApplyFileChanges`, accepts any
non-nil error for each of `../escape.txt`, `/etc/escape.txt`, and
`subdir/../../escape.txt`. An OS permission error after a validation regression
can satisfy this assertion, especially for the absolute `/etc` path. A write
followed by an error would also satisfy it. These checks establish error
presence, not rejection before filesystem side effects.

Use a controlled temporary sandbox containing an application root and outside
sentinel targets. Derive an absolute target within that sandbox, retain traversal
cases, and assert outside sentinels are unchanged or absent, inside contents are
unchanged, and no unwanted directories appear. Identify the validation failure
using a stable error category if available; avoid pinning a whole localized
message. This strengthens the existing rejection oracle and overlaps the
side-effect gap in [[pr-2107-coverage-gaps]].

### Medium: parser checks counts but discards decoded values and summary

`internal/domain/file_manager_test.go:104`, `TestParseFileChanges`, discards the
summary and checks only error presence and successful list length. Incorrect
operation/path/content values or ordering could pass if 2 changes are returned.
The error cases do not assert that no applicable changes are returned.

Compare the full ordered `[]FileChange` fixture (field equality or a structural
comparison), plus the expected summary before the marker. For no-marker input,
assert the original output is returned as the summary. For rejected inputs,
assert no applicable changes are returned and distinguish validation from
unrelated decoding failures where practical. Current malformed JSON and
validation fixtures have appropriate expected error booleans; their oracle is
incomplete rather than demonstrably incorrect.

### Medium: metadata assertion covers only one field

`internal/core/chatter_test.go:523`,
`TestChatter_Send_StreamingMetadataPropagation`, verifies a usage update exists,
its Usage pointer is non-nil, and TotalTokens equals 15. InputTokens and
OutputTokens could be lost or changed without failure; missing content forwarding
or duplicate usage updates also pass.

Compare all populated usage fields against the fixture (10, 5, 15). If the
intended contract is one-for-one forwarding, assert update count/order, type,
and content against the supplied two-update sequence. Keep this separate from
provider integration claims: the test uses a fake vendor.

### Low: error helpers and deadlock test have narrow error oracles

`internal/core/chatter_test.go:138`, `TestRecordFirstStreamError_ChannelFull`,
checks the retained error's text rather than its identity. Save the original
error and use `errors.Is` (or direct identity for this pass-through helper) to
ensure a replacement with identical text cannot pass.

`internal/core/chatter_test.go:359`,
`TestChatter_Send_StreamingErrorUpdateAndReturnDoesNotDeadlock`, appropriately
checks bounded completion, a non-nil error, and a session for its stated
purpose. It does not establish which of the update error and returned error
wins. Add a separate precedence assertion only after the intended contract is
specified; do not assume the vendor-returned error must win. The simpler
`StreamingErrorPropagation` test already checks vendor error identity.

### Low: buffered stdout assertion ignores capture failure

`internal/core/chatter_test.go:511`,
`TestChatter_Send_StreamingBufferStreamDoesNotPrint`, discards `io.ReadAll`'s
error (and the writer close error) before checking zero printed bytes. Check
capture errors before interpreting an empty buffer as evidence of silence.
Guard the returned session before accessing its last message to give a useful
failure diagnostic. Cleanup concerns remain in [[pr-2107-test-structure]].

### Low: message sequence comparison uses delimiter encoding

`internal/core/chatter_test.go:273`,
`TestChatter_BuildSession_EndsWithUserMessage`, joins role/content entries with
`|` before comparison. The current fixtures contain no ambiguous delimiter, so
their expected values are sound. Compare slices directly or inspect each role
and content if fixtures expand; delimiter collisions can hide a different
message sequence. Existing exact assertions in `SeparatesSystemSections`,
`JoinPromptSections`, `SuppressThink`, and `StreamingSuccessfulAggregation`
provide useful behavioral checks beyond successful return.

## Validation and limits

No repository AGENTS.md or CLAUDE.md was found during orientation; supplied
session instructions apply. Used the existing knowledge graph for discovery
and source snippets of all 13 direct test functions, and checked production
parser/application behavior against the working source. Confirmed both test
files have no diff against the pinned PR head. No images were supplied or
analyzed (0 images).

Static assertion assessment only: no source or test code changed, no new tests
were added, and no tests were executed. Maintainability, suite execution, and
consolidated TEST_GAPS.md remain separate unchecked tasks. This review makes
no executed pass/fail or measured coverage claim.
