---
type: report
title: PR 2107 Logic Correctness Review
created: 2026-10-03
tags:
  - code-review
  - correctness
  - security
related:
  - '[[REVIEW_SCOPE]]'
  - '[[2_REVIEW_CODE]]'
---

# PR 2107 logic correctness review

Reviewed the first unchecked task, **Check logic errors**, across all three changed files. Comparison is the original PR head `f32656b529a1eb1c678c793703fe49a0755cd439` against base `ddf1aab968caa9adf90137a6536c768d02b237d6`. Findings below describe that head; this review changes no production code. No images were supplied or analyzed (0).

## Changed logic

| Location | Assessment |
| --- | --- |
| `internal/core/chatter.go:192` | The pattern/channel conjunction is correct for the current REST and CLI callers. `internal/server/chat.go` supplies `UpdateChan: streamChan` before calling `Send`; CLI `BuildChatOptions` leaves the channel nil. Both CLI streaming and non-streaming retain automatic application. The channel is a caller convention, not an explicit permission: future callers leaving it nil also enable writes. |
| `internal/domain/file_manager.go:85` | `filepath.IsLocal` rejects empty, absolute and lexically escaping paths on the host platform. The explicit empty-path error still runs first. Names such as `version..txt` and paths such as `sub/../valid.txt` are intentionally newly accepted because their cleaned paths remain local. Symlinks are not checked. |
| `internal/domain/file_manager.go:156` | The new validation runs before directory creation or writing for the current entry; the finite range loop has no off-by-one or termination error. Nil/empty slices perform no writes and return nil. Invalid paths and mkdir/write failures return an error, but previously processed entries remain applied. |
| `internal/domain/file_manager_test.go:177` | The three new negative cases assert an error for parent traversal, an absolute path, and nested traversal. They exercise the application layer directly. Existing cases check creation, nested directories, and updates. They do not check filesystem side effects, symlinks, parser absolute-path rejection, or API gating. |

The changed condition adds no new nullable input contract: `Send` already dereferences its receiver, request and options earlier. Parser validation rejects an empty path and invalid operations before application. Existing JSON extraction loops are bounded by input length, but their bracket counting has the defect below. Host-platform path validation is appropriate for paths applied on that same host; Windows semantics were not executed on this macOS runner.

## Confirmed correctness findings

### Major: lexical containment does not prevent symlink escape

**Location:** `internal/domain/file_manager.go:157` and `:169`.

A project-local `link` pointing to an outside directory passes `IsLocal("link/escaped.txt")`; `MkdirAll` and `WriteFile` follow it and create the outside file. A file symlink can likewise redirect an overwrite. The directory-symlink case was reproduced with both directories inside disposable test storage. This behavior predates the PR and remains unresolved by its new containment check; it is a residual security gap, not a newly introduced exploit.

**Suggested fix:** use filesystem operations rooted in a directory handle that enforce containment throughout path resolution, and test intermediate and final symlinks. A lexical check or a separate resolve-then-write check alone does not protect against symlink races.

### Minor: validation of a later entry leaves a partially applied batch

**Location:** `internal/domain/file_manager.go:156`–`:169`; related warning handling at `internal/core/chatter.go:205`.

Applying `first.txt` followed by `../escape.txt` writes the first file, then returns an error for the second. Parser validation prevents this specific malformed batch in the current `Send` path, but direct application callers can encounter it. Runtime write failures can also leave earlier entries applied. `Send` logs application failure, still strips the file-change section from the saved response, and does not propagate that application error as its return value. There is no documented atomic-batch contract, so this is a correctness/diagnostic limitation rather than a claim of broken transactional behavior. Partial application and warning-only handling predate the PR; the new check adds another possible early return.

**Suggested fix:** validate every entry before mutation, document partial-write behavior, and preserve/report applied entries and failed entries. If atomicity is required, design staging and recovery separately; prevalidation cannot eliminate runtime filesystem failures.

### Minor: valid file content with unmatched brackets is rejected

**Location:** `internal/domain/file_manager.go:46`–`:59`.

The JSON boundary scanner counts `[` and `]` inside quoted content. A valid marshaled change whose content is `values[` is reported as unbalanced; content `values]` causes premature truncation and a JSON parse error. Both were reproduced. This parser defect predates the PR, but affects normal source-code generation and remains uncovered by the changed tests.

**Suggested fix:** decode a JSON array from the marker suffix with `json.Decoder`, preserving supported surrounding text and escape-repair behavior, or use a scanner that tracks string/escape state correctly. Add regressions for brackets in content and escaped quotes.

## Validation and limits

- `go test ./internal/domain ./internal/core` passed on Go 1.27.1, darwin/arm64, before the temporary probes were added.
- `go test ./internal/domain -run TestReviewCorrectnessProbes -v` passed all four observation probes: directory symlink escape, partial application, bracket-content rejection, and accepted local paths. A passing probe confirms observed behavior, including defects; it does not certify that behavior as correct.
- Temporary probe source is retained in the Auto Run `Working/review_correctness_probe_test.go` folder. It was removed from the source tree after execution so defect-characterization assertions do not become permanent regression expectations. To reproduce, temporarily copy it into `internal/domain`, run the command above, and remove it again. Test storage and build cache were placed under the authorized Auto Run Working folder.
- API gating was verified by graph-backed caller/source inspection, not a new end-to-end API test. Only the relevant domain/core suites were run; a full repository suite and Windows execution were outside this logic-review task.
- The original three-file diff contains no template-extension execution change. The PR title's shell-injection claim is not established by this task.

This report preserves correctness findings for later playbook tasks. Data-flow review, best-practice review, API-contract review, and the consolidated `CODE_ISSUES.md` task remain unchecked.
