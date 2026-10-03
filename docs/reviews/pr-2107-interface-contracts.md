---
type: report
title: PR 2107 Interface Contract Review
created: 2026-10-03
tags:
  - code-review
  - api
  - contracts
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-api-compatibility]]'
  - '[[pr-2107-data-flow]]'
  - '[[pr-2107-error-handling]]'
  - '[[pr-2107-correctness]]'
  - '[[2_REVIEW_CODE]]'
---

# PR 2107 interface contract review

Reviewed the **Interface contracts** checkbox across `internal/core/chatter.go`, `internal/domain/file_manager.go`, and `internal/domain/file_manager_test.go`, comparing base `ddf1aab968caa9adf90137a6536c768d02b237d6` with original PR head `f32656b529a1eb1c678c793703fe49a0755cd439`. Function signatures, Go return types, and JSON field types are unchanged. No production code was changed. No images were supplied or analyzed (0).

## Signatures, returns, and parameter requirements

| Interface | Established contract | Assessment |
| --- | --- | --- |
| `ParseFileChanges(output string) (changeSummary string, changes []FileChange, err error)` (`file_manager.go:30`) | Without a marker, returns the original output, nil changes, and nil error. A valid marked array returns the prefix before the marker and decoded records; an empty array is successful. Errors return nil changes, with either the original output or the prefix depending on failure stage. | Names and types are clear. Callers must check `err` before consuming changes and must not assume the summary always has the same shape on failure. No marker is a valid no-op, not an error. |
| `FileChange` (`file_manager.go:23`–`:27`) | Parser requires operation `create` or `update`, a nonempty local path, and content no larger than `MaxFileSize` in decoded bytes. Missing/null operation or path becomes empty and fails validation. Non-null incompatible JSON field types fail decoding. Explicit empty content is accepted. | Required content presence is not enforced: missing/null content becomes an empty string. This is the pre-existing truncation finding in [[pr-2107-data-flow]], not a new finding introduced by the patch. Unknown JSON fields are ignored by ordinary unmarshalling. |
| `ApplyFileChanges(projectRoot string, changes []FileChange) error` (`file_manager.go:155`) | Nil/empty changes return nil without checking the root. Each record needs a local path. Root may be absolute or relative; an empty root resolves local paths against the process working directory. Entries create parent directories and create/overwrite the destination. Returns the first validation/filesystem error. | Signature is readable, but it presumes parser-validated records for operation and size. Nil error means writes completed for the supplied entries, not that operation labels were validated. An error does not promise rollback or no side effects; see [[pr-2107-correctness]]. Lexical locality does not guarantee containment through symlinks. |
| `Chatter.Send(ctx context.Context, request *domain.ChatRequest, opts *domain.ChatOptions) (session *fsdb.Session, err error)` (`chatter.go:61`) | Requires an initialized chatter/vendor and nonnil request/options; no nil-pointer guards are added. Uses context for provider calls. Mutates options (model, session ID, context-length defaults and sometimes raw mode), and session construction can initialize or transform the request message. | Pointer types and returns remain appropriate for an internal orchestration API. Treat request/options as mutable per-call values. A returned error can accompany a populated session; empty provider output explicitly returns nil session. Callers must check `err` before relying on the session. |
| `ChatRequest` and `ChatOptions` (`domain.go:15`–`:62`) | Request message can be nil and is initialized by `BuildSession`. Empty pattern/context/session names skip their respective lookups; a named session is persisted. `UpdateChan` is optional for streaming delivery and excluded from JSON with `json:"-"`. | The patch gives channel presence a second meaning: nonnil skips file parsing/application for this pattern, including for a non-streaming chatter. Streaming is controlled independently by `Chatter.Stream`. A nil channel permits application for this pattern; it is not proof of CLI authorization for arbitrary future callers. |
| Changed tests (`file_manager_test.go:177`–`:181`) | Three invalid direct-writer paths must return nonnil errors. Existing tests cover ordinary create/update content and parser counts/errors. | These assertions exercise the return contract but do not establish full containment, atomicity, required content presence, or API/CLI gating. Passing tests should not be read as guarantees for those behaviors. |

`Send` can return nil error after parse, working-directory, or application failure because those failures are logged as warnings. Its returned error primarily covers session/provider processing and persistence, not successful file application. This existing limitation and payload replacement are documented in [[pr-2107-error-handling]] and [[pr-2107-data-flow]]. Keep that distinction explicit in caller documentation or return a structured application outcome if consumers need it.

Current REST `HandleChat` (`internal/server/chat.go:136`–`:152`) supplies a channel, while CLI `BuildChatOptions` leaves it nil. This supports the patch's intended current caller split; it does not establish an explicit permission contract. See [[pr-2107-api-compatibility]] for changed saved-response behavior and [[pr-2107-code-organization]] for the policy coupling suggestion. A streaming caller must consume the supplied channel; forwarding uses a blocking send and `Send` does not close `opts.UpdateChan` itself. The REST caller owns that channel's closure.

## Findings and decisions

### Suggestion: state the writer's validation precondition

**Locations:** `internal/domain/file_manager.go:154`–`:173`.

The exported writer is described as applying parsed changes, but accepts any `[]FileChange`. Its new path check does not validate operation or content size. A direct call with operation `delete` writes the content and logs a delete operation, although the parser rejects that label. Graph tracing shows the sole production direct caller is `Chatter.Send`, which parses first, so this is an existing contract/documentation limitation rather than an established current production bypass.

Document that operation/size validation must precede application, or share those checks between parser and writer if direct callers are supported. Extend the existing functions rather than introducing a separate writer. Do not describe the newly added locality check as complete record validation. Base-source inspection confirms operation/size omission predates this PR.

### Both operation labels intentionally create or overwrite

A probe confirmed that `create` overwrites an existing file and `update` creates a missing file. This agrees with `data/patterns/create_coding_feature/system.md` under “File Creation and Modification,” which explicitly allows creation of missing files and overwriting of existing ones. No create-only/update-only defect is raised. Operation labels identify intended changes and do not select distinct filesystem primitives.

### Existing findings remain applicable

Missing/null content acceptance (Minor), warning-only file-application failures (Minor), partial application (Minor), and symlink escape (Major) remain unresolved. See the linked reports for evidence and suggested fixes. No new merge-blocking defect was identified specifically in signatures or return types. The optional channel's new policy meaning should be documented or replaced with explicit trusted execution policy, as already suggested by the organization review.

## Validation and limits

- Used the already indexed code graph for symbol discovery, snippets and inbound writer traces. Reviewed parser/writer, `FileChange`, core `Send`/`BuildSession`, request/options, REST handler and both existing domain tests. Rechecked the scoped diff and base writer; no local CLAUDE.md or AGENTS.md was present.
- Added and ran seven temporary characterization subtests in `TestReviewInterfaceContracts`: absent marker, empty array, required fields/type rejection, explicit empty content, nil batch, both operation labels creating/overwriting, and unsupported operation accepted by the direct writer. All passed on macOS. Passing characterization probes describe existing behavior, including limitations; they do not assert all behavior is desirable.
- Retained probe source at `Working/review_interface_contracts_probe_test.go` in the Auto Run folder and removed its temporary copy from `internal/domain`. To reproduce, copy it into that package, run `go test ./internal/domain -run TestReviewInterfaceContracts -v`, then remove the copy.
- After removal, ran `go test ./internal/domain ./internal/core ./internal/server ./internal/cli`; all passed. Test temporary storage and build cache used the authorized Auto Run Working folder. No permanent tests or production source changed.
- Channel gating, pointer requirements, mutation and session-return conclusions use source inspection, not a new live-provider or HTTP integration test. Windows and full-repository testing were not performed. The scoped diff still contains no template-extension execution changes and does not substantiate the title's shell-injection claim.
- Only the interface-contract checkbox is completed by this task. Consolidated CODE_ISSUES.md and later security tasks remain for subsequent runs.
