---
type: report
title: PR 2107 Data Flow Review
created: 2026-10-03
tags:
  - code-review
  - data-flow
  - correctness
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-correctness]]'
  - '[[2_REVIEW_CODE]]'
---

# PR 2107 data flow review

Reviewed **Verify data flow** across `internal/core/chatter.go`, `internal/domain/file_manager.go`, and `internal/domain/file_manager_test.go`. Findings refer to PR head `f32656b529a1eb1c678c793703fe49a0755cd439` against base `ddf1aab968caa9adf90137a6536c768d02b237d6`. This task changes documentation only; it does not fix production behavior. No images were supplied or analyzed (0).

## Initialization, types, transformations, and mutations

| Flow | Assessment |
| --- | --- |
| Model response → `message` (`chatter.go:61`–`:192`) | Initialized to an empty string, assigned by non-streaming `Send`, or accumulated from content updates. Provider errors return before applying changes. Think-block removal precedes parsing when enabled. The new gate reads initialized request/options fields; it adds no new pointer contract. |
| Execution context → write gate (`chatter.go:192`) | REST `HandleChat` allocates `streamChan`, sets `ChatOptions.UpdateChan` before `Send`, and forwards updates without file application. CLI `BuildChatOptions` omits that field, leaving it nil. CLI streaming is controlled independently by `Chatter.Stream`, so a streaming CLI response can still apply files. This convention is supported by current caller inspection, not an explicit permission type or an end-to-end API test. |
| Response → summary/JSON (`file_manager.go:30`–`:74`) | Marker and array indices are initialized and checked before slicing. `changeSummary` is the prefix before the marker; without a marker the entire response is returned. The bracket scanner and escape repair can alter/reject content, as documented below and in [[pr-2107-correctness]]. Parser errors return nil changes, preventing application of partially decoded/validated slices in `Send`. |
| JSON → `[]FileChange` (`file_manager.go:23`–`:27`, `:67`–`:95`) | Operation, path and content are strings with matching JSON tags; JSON decoding rejects incompatible non-null field types. Missing/null string fields become zero values. Operation and path checks reject their empty defaults; content permits empty strings. Content length measures decoded bytes, matching the later `[]byte` write. |
| Parsed paths → filesystem (`file_manager.go:155`–`:176`) | `IsLocal` reads the same path value later joined to the root. `filepath.Join` cleans lexical components without changing the input struct/slice. Root comes from checked `os.Getwd` in `Send`. `MkdirAll` precedes `WriteFile`; both errors stop further writes. Relative roots are also accepted by direct callers; `absPath` is only absolute when the root is. Symlink resolution and partial mutations remain limitations from [[pr-2107-correctness]]. |
| Operation/content → write (`file_manager.go:169`–`:173`) | Both create/update use the same overwrite-or-create operation; the operation string is only used in logging here. Parser operation/size validation is not repeated by direct application calls. These are existing semantics, with no create-only/update-only or transactional contract established by this review. A successful ordinary-content probe preserved exact UTF-8, newline, tab and quote bytes and left the input slice unchanged. |
| Write outcome → session (`chatter.go:193`–`:216`) | Getwd's local `err` shadows the named return error; parse/apply errors also remain local warning values. `message = summary` runs even after these failures, then the assistant summary is appended and a named session is saved. API mode skips this block and retains the full response, including its file-change payload. This is an intentional consequence of the new gate. Failure-time payload loss is an existing limitation. |
| New rejection tests (`file_manager_test.go:177`–`:181`) | Loop values are initialized strings and each call receives a fresh one-element slice. Tests assert errors but do not prove no filesystem mutation. Existing parser tests assert counts/errors, not exact content preservation, which leaves transformation defects undetected. |

## Confirmed findings

### Minor: valid JSON content is silently changed by escape repair

**Location:** `internal/domain/file_manager.go:65` and `:99`–`:152`.

The unconditional replacement of `\C` runs before attempting JSON decoding, including inside already valid escaped backslashes. A `json.Marshal` payload for content `C:\Code` returned content `C:\\Code` (two literal backslashes), with no error. That altered string then reaches `WriteFile`. The initial observation probe expected a parse error and failed; the output instead demonstrated silent mutation. The corrected characterization probe verifies the observed mutation and passes. This behavior exists in the base commit and is not introduced by the PR.

**Suggested fix:** attempt decoding the untouched JSON first. Apply fallback repair only after decoding fails, with a scanner that consumes escape pairs and preserves valid escapes. Add exact round-trip regressions for backslashes, quotes, control escapes, and Unicode; count-only parser tests are insufficient.

### Minor: null/missing content can erase an existing file

**Location:** `internal/domain/file_manager.go:26`, `:68`, `:89`, and `:169`.

JSON `"content": null` decodes into the string's empty zero value, passes size validation, and an update truncates the destination to zero bytes. Missing content likewise has no presence validation. A probe confirmed truncation of a disposable existing file using null content. Explicit empty content is valid for an empty file; the data model cannot distinguish that intent from omitted/null content. This is a pre-existing schema/validation limitation, not a PR regression.

**Suggested fix:** validate content presence and require a JSON string while continuing to accept an explicit empty string. Reuse the existing parser and `FileChange` model rather than adding an unrelated application helper.

### Existing limitations carried forward

[[pr-2107-correctness]] documents symlink escape (Major), partial batch application (Minor), and quoted-bracket parsing failures (Minor). The data-flow review additionally establishes that failed application/getwd/validation can discard the JSON payload from session history when `message` is replaced with the summary. Preserve the original response or record failed/applied entries for recovery. These findings should be included when the later consolidated issues task runs; they were not repaired in this review.

## Validation and limits

- Graph-backed symbol snippets and inbound traces were used for discovery, including `Send`, `ParseFileChanges`, `ApplyFileChanges`, `FileChange`, `fixInvalidEscapes`, `HandleChat`, `BuildChatOptions`, and the changed tests. The project was already indexed. No local `CLAUDE.md` or `AGENTS.md` was found; session-provided guidance applies.
- `go test ./internal/domain -run TestReviewDataFlowProbes -v` passed three characterization probes: ordinary byte preservation/input immutability, observed backslash corruption, and observed null-content truncation. Passing defect probes confirm observations, not correctness.
- Probe source is retained at `/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2107-codex/Working/review_data_flow_probe_test.go`. It was temporarily copied into `internal/domain` for execution and removed afterwards. To reproduce, copy it back, run the focused command, then remove it. Test temporary storage and build cache used the authorized Working folder.
- After probe removal, `go test ./internal/domain ./internal/core` passed. No permanent defect-characterization tests or production changes were added. Full repository, Windows, and end-to-end API tests were not run.
- Later best-practice, API-contract, consolidated-issues, and security tasks remain for subsequent runs. This task does not substantiate the PR title's shell-injection claim; the original diff contains no template-extension execution change.
