# Document 2: Go Code Quality Review

## Context

- **Playbook**: Fabric PR Review
- **Agent**: Fabric-PR-Review
- **Project**: /Users/kayvan/src/fabric
- **Date**: 2026-04-05
- **Working Folder**: /Users/kayvan/src/fabric/.maestro/playbooks

## Purpose

Perform a Go-specific code review focusing on Fabric's coding conventions, Go idioms, and best practices.

## Prerequisites

- `/Users/kayvan/src/fabric/.maestro/playbooks/REVIEW_SCOPE.md` exists from Document 1

## Tasks

### Task 1: Load Context

- [x] **Read scope**: Loaded `/Users/kayvan/src/fabric/.maestro/playbooks/REVIEW_SCOPE.md` and identified the Go files queued for later review: `internal/tools/youtube/youtube.go`, `internal/cli/cli.go`, `internal/cli/flags.go`, and `internal/cli/help.go`.

### Task 2: Check Go Idioms

- [x] **Error handling**: Verify Fabric's error patterns:
  - Errors are returned, not panicked (no `panic()` in library code)
  - Use `pkg/errors` for wrapping: `errors.Wrap(err, "context")`
  - Error messages are lowercase, no punctuation
  - Errors don't expose sensitive information
  Notes from targeted review/fix on 2026-04-05:
  - Reviewed the PR-touched Go files in scope: no PR-added `panic()` calls were introduced, and the new YouTube visual extraction path returns errors to callers instead of terminating the process.
  - `internal/tools/youtube/youtube.go` originally returned raw `ffmpeg` output and `tesseract` stderr from `GrabVisual`; updated the implementation and localized error strings so returned errors preserve context without exposing direct stream URLs or subprocess diagnostics.
  - The visual extraction error messages remain lowercase and do not add trailing punctuation after the cleanup.
  - Scoped files still lean on `fmt.Errorf` and `errors.New` more than `pkg/errors.Wrap`, so wrap-style consistency should be carried into the later consolidated Go issues document if the project wants to enforce that convention repo-wide.
  - Verified with `go test ./internal/tools/youtube ./internal/cli`.

- [x] **Context usage**: Check `context.Context` patterns:
  - Context is the first parameter where applicable
  - Context is propagated through call chains
  - Cancellation is handled for long operations
  - Timeouts are set for external calls
  Notes from targeted review on 2026-04-05:
  - Reviewed the PR-added visual extraction path in `internal/cli/cli.go` and `internal/tools/youtube/youtube.go`.
  - `internal/tools/youtube/youtube.go` adds a bounded execution window with `context.WithTimeout(context.Background(), 15*time.Minute)` and uses `exec.CommandContext` for `yt-dlp`, `ffmpeg`, and per-frame `tesseract` work, so the new subprocess-heavy flow does handle timeout-driven cancellation.
  - The main gap is context propagation: `processYoutubeVideo`, `YouTube.Grab`, and `YouTube.GrabVisual` do not accept a caller-provided `context.Context`, so CLI/request cancellation cannot flow into the OCR pipeline and callers cannot supply their own deadline.
  - `internal/cli/flags.go` and `internal/cli/help.go` only add surface area for the feature and do not affect context behavior.
  - Carry the missing caller-propagated context into the later consolidated Go issues document as a Major issue if Fabric wants request-scoped cancellation on long-running YouTube operations.

- [x] **Interface compliance**: Verify interfaces:
  - Functions accept interfaces, return concrete types
  - Interfaces are defined where they're used
  - No empty interfaces (`interface{}`) without good reason
  Notes from targeted review on 2026-04-05:
  - Reviewed the PR-touched Go files in scope: `internal/tools/youtube/youtube.go`, `internal/cli/cli.go`, `internal/cli/flags.go`, and `internal/cli/help.go`.
  - No `interface{}` or `any` usage was introduced in the reviewed files, so the PR does not add empty-interface surface area.
  - The only interface-typed parameters in scope are standard library I/O boundaries such as `detectError(io.Reader)` and the translated help writer functions that accept `io.Writer`, which is appropriate because those call sites benefit from stream abstraction.
  - The reviewed functions and methods continue to return concrete values (`string`, `[]string`, `*VideoInfo`, `*VideoMetadata`, `*TranslatedHelpWriter`) rather than interface types, so the new YouTube visual extraction path does not widen return contracts unnecessarily.
  - No new project-local interfaces were added in these files, so there is no new interface-placement issue to flag from this PR slice.
  - Verified with `go test ./internal/tools/youtube ./internal/cli`.

### Task 3: Review Code Organization

- [x] **Package structure**: Check:
  - `internal/` packages are truly internal
  - No circular dependencies
  - Clear package boundaries
  - Appropriate file sizes
  Notes from targeted review on 2026-04-05:
  - Reviewed the PR-touched Go files in scope and the adjacent wiring points in `cmd/fabric/main.go`, `internal/core/plugin_registry.go`, and `internal/server/youtube.go` to trace package ownership and dependency direction.
  - The `internal/` boundaries remain properly internal in practice: the reviewed packages are only consumed from inside the module, with `cmd/fabric` entering through `internal/cli` and the rest of the YouTube flow staying under `internal/`.
  - No circular dependencies were introduced in the reviewed graph. `internal/cli` depends on `internal/core` and `internal/tools/youtube`, `internal/core` wires `youtube.NewYouTube()`, and `internal/tools/youtube` only depends downward on shared support packages such as `internal/plugins`, `internal/i18n`, and logging.
  - Package boundaries are otherwise clear for this PR slice: CLI orchestration lives in `internal/cli`, registry/setup ownership stays in `internal/core`, and REST exposure remains in `internal/server`.
  - The main organization concern is file sizing within `internal/tools/youtube`: `internal/tools/youtube/youtube.go` is now 987 lines and bundles transcript download, comments/metadata fetching, aggregate grab helpers, OCR frame extraction, and subprocess utilities in one file, so carry that forward as a Minor structure issue in the later consolidated Go issues document.
  - `internal/cli/cli.go` (203 lines) and `internal/cli/help.go` (289 lines) remain focused; `internal/cli/flags.go` is larger at 573 lines but still stays within the package's flag/config parsing responsibility.
  - Verified with `go test ./internal/tools/youtube ./internal/cli`.

- [x] **Naming conventions**: Verify:
  - CamelCase for exported identifiers
  - camelCase for unexported identifiers
  - Meaningful, descriptive names
  - Acronyms are all caps (HTTP, API, ID)
  Notes from targeted review/fix on 2026-04-05:
  - Reviewed the PR-touched Go files in scope: `internal/tools/youtube/youtube.go`, `internal/cli/cli.go`, `internal/cli/flags.go`, and `internal/cli/help.go`; no task images were attached for this checklist item.
  - The exported additions in scope remain descriptive and use CamelCase (`GrabVisual`, `VisualText`, `YouTubeVisualSensitivity`), while the new unexported locals added for the OCR path stay camelCase.
  - Normalized the new acronym-bearing identifiers introduced by this PR so `FPS` and `URL` now follow Go initialism style in the changed code (`YouTubeVisualFPS`, `VisualFPS`, `cmdURL`, `streamURL`) instead of mixed-case variants.
  - The package still contains older `Id` and `YtDlp` spellings such as `videoId`, `ChannelId`, and `YtDlpArgs`; those predate this task and should be treated as broader repo-level cleanup if strict Go initialism consistency is desired.
  - Verified with `go test ./internal/tools/youtube ./internal/cli`.

- [x] **Documentation**: Check:
  - Exported functions have doc comments
  - Package-level documentation exists
  - Complex logic is explained
  - No stale comments
  Notes from targeted review/fix on 2026-04-05:
  - Reviewed the PR-touched Go files in scope: `internal/tools/youtube/youtube.go`, `internal/cli/cli.go`, `internal/cli/flags.go`, and `internal/cli/help.go`; no task images were attached for this checklist item.
  - `internal/tools/youtube/youtube.go` already ships package documentation, and the PR-added exported method `GrabVisual` now has a more explicit doc comment describing its frame-sampling modes and VTT-like OCR output.
  - Added `internal/cli/doc.go` so the `internal/cli` package now has package-level documentation instead of relying only on file-local comments.
  - Filled in missing `description` tags for the new `--visual`, `--visual-sensitivity`, and `--visual-fps` flags so the flag definitions, custom translated help, and README text stay aligned.
  - Removed a stale inline comment in `Grab` that no longer matched the visual extraction call path after the PR started forwarding language, yt-dlp arguments, and sampling options.
  - Added a CLI help regression test to assert the three new visual flags remain documented in generated help output, and verified with `go test ./internal/cli ./internal/tools/youtube`.

### Task 4: Review Concurrency

- [x] **Goroutine safety**: Look for:
  - Race conditions on shared state
  - Proper channel usage (closing, direction)
  - Context-aware goroutines
  - No goroutine leaks
  Notes from targeted review/fix on 2026-04-05:
  - Reviewed the PR-touched Go files in scope: `internal/tools/youtube/youtube.go`, `internal/cli/cli.go`, `internal/cli/flags.go`, and `internal/cli/help.go`; no task images were attached for this checklist item.
  - The only new goroutine fan-out in scope is the OCR worker loop inside `internal/tools/youtube/youtube.go`; the CLI files do not introduce goroutines, channels, or other shared-state concurrency.
  - The current worker coordination is race-safe in the reviewed path: each goroutine writes to its own `results[idx]` slot, error aggregation is guarded by `sync.Mutex`, and semaphore tokens are released with `defer`, so the bounded worker pool does not leak permits on early returns.
  - The subprocess-heavy OCR work is at least internally cancellation-aware because each worker uses the shared timeout-bound context with `exec.CommandContext`, which means the goroutines do not outlive the 15-minute extraction deadline even if `tesseract` stalls.
  - Added `TestGrabVisualParallelOCRPreservesFrameOrder` to exercise the parallel OCR path with out-of-order worker completion and confirm the final output remains deterministic while the race detector stays clean.
  - Verified with `go test ./internal/tools/youtube` and `go test -race ./internal/tools/youtube`.

- [x] **Streaming**: For streaming responses:
  - Channels are properly buffered
  - Errors are communicated correctly
  - Cleanup happens on cancellation
  Notes from targeted review/fix on 2026-04-05:
  - Reviewed the PR-touched Go files in scope (`internal/tools/youtube/youtube.go`, `internal/cli/cli.go`, `internal/cli/flags.go`, `internal/cli/help.go`) plus the adjacent streaming path in `internal/core/chatter.go` and `internal/server/chat.go`, because YouTube command output ultimately flows through Fabric's shared chat streaming pipeline; no task images were attached for this checklist item.
  - The PR itself does not add a new Go streaming transport, but the existing SSE chat handler had two concrete issues that matter for stream correctness: it advertised `text/readystream` instead of `text/event-stream`, and it forwarded updates over an unbuffered server stream channel that could strand `Chatter.Send` after client cancellation.
  - Fixed the shared stream forwarding path in `internal/core/chatter.go` so updates sent to `opts.UpdateChan` now respect `ctx.Done()` instead of blocking indefinitely when the downstream consumer disappears.
  - Switched `internal/server/chat.go` to request-context cancellation (`c.Request.Context().Done()`), corrected the SSE content type to `text/event-stream`, and added a small buffer (`16`) on the handler-facing `streamChan` so HTTP writes do not have to rendezvous on every token.
  - Added regression coverage with `TestChatter_Send_CanceledUpdateChanDoesNotDeadlock`, `TestWriteSSEResponse_FormatsEventStreamChunk`, and `TestHandleChat_SetsEventStreamHeaders` to lock in cancellation cleanup, SSE framing, and header behavior.
  - Verified with `go test ./internal/core ./internal/server` and `go test ./...`.

### Task 5: Review API Changes

- [x] **Breaking changes**: If public APIs change:
  - Are changes backward compatible?
  - Are deprecation notices added?
  - Is the CHANGELOG updated?
  Notes from targeted review on 2026-04-05:
  - Reviewed the PR-touched public surface in `internal/cli/cli.go`, `internal/cli/flags.go`, `internal/cli/help.go`, `README.md`, `README.zh.md`, and `cmd/generate_changelog/incoming/2073.txt`; no task images were attached for this checklist item.
  - The user-facing CLI changes are additive rather than breaking: the PR adds `--visual`, `--visual-sensitivity`, and `--visual-fps` without removing or renaming existing flags, and existing YouTube transcript/comments/metadata flows keep their prior behavior unless callers opt into the new visual extraction path.
  - The exported Go additions in `internal/tools/youtube/youtube.go` are also additive (`GrabVisual`, extra `Options`/`VideoInfo` fields) and live under `internal/`, so they do not widen Fabric's external module API surface; a repo-wide search found no positional composite literals or call sites that would break because of the new fields.
  - Because no existing public behavior was removed or redefined, this PR does not need deprecation notices.
  - Changelog coverage is present through `cmd/generate_changelog/incoming/2073.txt`, and the README help text was updated in both English and Chinese to document the new CLI flags.
  - Verified the reviewed surface still builds and passes targeted regression coverage with `go test ./internal/cli ./internal/tools/youtube`.

- [x] **Function signatures**: Verify:
  - Context is first parameter
  - Options pattern for many parameters
  - Error is last return value
  Notes from targeted review on 2026-04-05:
  - Reviewed the PR-touched function signatures in `internal/cli/cli.go` and `internal/tools/youtube/youtube.go`; no task images were attached for this checklist item.
  - Return ordering stays consistent in the reviewed code: the PR-added and PR-modified helpers continue to keep `error` in the final return position, including `Grab`, `GrabVisual`, and the CLI orchestration helpers.
  - The aggregate YouTube entry point already follows the options-pattern direction Fabric wants: `Grab(url string, options *Options)` absorbs the new visual extraction controls as added struct fields instead of forcing every caller through a wider positional signature.
  - The main signature gap is `GrabVisual(videoId, language, additionalArgs, sensitivity, fps)`, which now carries five positional parameters and still creates its own `context.Background()` timeout internally; that means the new OCR path does not meet the "context first parameter" guideline and is the clearest follow-up candidate for a `GrabVisual(ctx context.Context, videoID string, opts *VisualOptions)`-style refactor.
  - `processYoutubeVideo(flags, registry, videoId)` is only an unexported CLI helper with a narrow internal call graph, so its current signature is acceptable for now even though it also lacks caller-supplied context.
  - Carry the missing context-first and dedicated visual-options cleanup for the OCR path into the later consolidated Go issues document as a Major signature-design issue.
  - Verified with `go test ./internal/tools/youtube ./internal/cli`.

### Task 6: Check Fabric-Specific Patterns

- [x] **Plugin patterns**: For AI providers:
  - Implement `VendorPlugin` interface
  - Handle streaming via callbacks
  - Support model listing
  - Handle context cancellation
  Notes from targeted review on 2026-04-05:
  - Reviewed the PR diff (`git diff --name-only origin/main...HEAD`) plus the relevant provider wiring in `internal/plugins/ai/vendor.go`, `internal/plugins/ai/vendors.go`, and `internal/core/plugin_registry.go`; no task images were attached for this checklist item.
  - The branch does not modify any `internal/plugins/ai/*` provider implementation or the registry wiring that registers AI vendors, so this PR does not introduce a Fabric plugin-pattern regression.
  - Fabric's current provider contract is the `ai.Vendor` interface rather than a `VendorPlugin` type; `internal/core/plugin_registry.go` continues to register vendors through that interface, and the existing provider/registry tests still pass.
  - Streaming and model discovery remain on the established provider methods `SendStream(...)` and `ListModels(...)`; the PR's YouTube and CLI changes do not bypass, replace, or widen that contract.
  - Context cancellation is still a broader pre-existing inconsistency in some provider implementations because several `ListModels` and `SendStream` methods accept `context.Context` but ignore it via `_ context.Context`; that is worth carrying into the later consolidated Go issues document as existing tech debt, but it is not a regression introduced by this PR.
  - Verified with `go test ./internal/plugins/... ./internal/core`.

- [x] **Configuration patterns**: For config changes:
  - Environment variables via `godotenv`
  - Flags via `go-flags`
  - YAML config support
  Notes from targeted review/fix on 2026-04-05:
  - Reviewed the PR-touched configuration surface in `internal/cli/flags.go`, `internal/cli/cli.go`, `internal/cli/help.go`, `internal/plugins/db/fsdb/db.go`, and `internal/cli/flags_test.go`; no task images were attached for this checklist item.
  - The branch correctly exposes the new visual extraction settings through `go-flags`, and the existing environment-file path remains unchanged through `internal/plugins/db/fsdb/db.go`, which still loads persisted settings from `~/.config/fabric/.env` via `godotenv`.
  - The concrete regression in this PR slice was YAML support: the new `--visual`, `--visual-sensitivity`, and `--visual-fps` flags were added without `yaml` tags, so `config.yaml` defaults could not drive the new YouTube visual extraction feature even though the rest of Fabric's config loader expects tagged fields.
  - Added `yaml:"visual"`, `yaml:"visualSensitivity"`, and `yaml:"visualFPS"` to the new flags so they now participate in the established CLI-over-YAML merge path.
  - Extended `internal/cli/flags_test.go` to cover both YAML loading and CLI override behavior for the new visual settings, which locks in the expected configuration precedence.
  - Verified with `go test ./internal/cli ./internal/plugins/db/fsdb`.

- [x] **Logging patterns**: For logging:
  - Use standard `log` package
  - Debug levels via `--debug` flag
  - No sensitive data in logs
  Notes from targeted review/fix on 2026-04-05:
  - Reviewed the PR-touched logging paths in `internal/tools/youtube/youtube.go`, `internal/cli/flags.go`, `internal/server/chat.go`, `README.md`, and the related regression tests; no task images were attached for this checklist item.
  - Preserved Fabric's existing split between standard `log` for operational errors and `internal/log` for debug output: server chat errors still use the standard logger, while request/progress messages now respect the configured `--debug` level instead of printing unconditionally.
  - Removed two concrete sensitive-data leaks introduced by the branch's debug path: YouTube trace logging now redacts signed URLs, auth headers, cookie/browser args, and password-like values before logging `yt-dlp` arguments or stderr lines, and YAML config loading now logs only the configured YAML key names instead of dumping the full parsed struct.
  - Updated the README debug-level documentation so the user-facing help matches the current `--debug` implementation, including level `4` (`wire`).
  - Added regression coverage in `internal/tools/youtube/youtube_logging_test.go`, `internal/cli/flags_test.go`, and `internal/server/chat_test.go` to lock in trace redaction and debug-gated request logging.
  - Verified with `go test ./internal/tools/youtube`, `go test ./internal/cli`, and `go test ./internal/server`.

### Task 7: Run Static Analysis

- [x] **Check for modernization**: Run:
  ```bash
  go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest ./...
  ```
  Notes from targeted review on 2026-04-05:
  - Ran `go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest ./...` from the repo root.
  - The analyzer exited successfully with no diagnostics, so this pass did not identify modernization changes to carry into `GO_ISSUES.md`.
  - No task images were attached for this checklist item.

- [x] **Check formatting**: Run:
  ```bash
  gofmt -l .
  ```
  Flag any unformatted files.
  Notes from targeted review on 2026-04-05:
  - Ran `gofmt -l .` from the repo root.
  - The command produced no output, so there are no currently unformatted Go files to carry into `GO_ISSUES.md`.
  - No task images were attached for this checklist item.

- [x] **Check vet**: Run:
  ```bash
  go vet ./...
  ```
  Note any issues.
  Notes from targeted review on 2026-04-05:
  - Ran `go vet ./...` from the repo root.
  - The analyzer exited successfully with no diagnostics, so there are no new vet findings to carry into `GO_ISSUES.md` from this pass.
  - No task images were attached for this checklist item.

### Task 8: Document Go Issues

- [x] **Create GO_ISSUES.md**: Write findings to `/Users/kayvan/src/fabric/.maestro/playbooks/GO_ISSUES.md`:

```markdown
# Go Code Quality Issues

## Critical Issues
[Must fix - compiler errors, data races, panics]

## Major Issues
[Should fix - error handling, context misuse, interface violations]

## Minor Issues
[Nice to fix - naming, documentation, style]

## Suggestions
[Optional improvements, modernization opportunities]

## Static Analysis Results

### Modernize
[Results from modernize tool]

### Gofmt
[Unformatted files if any]

### Go Vet
[Vet issues if any]

## Positive Observations
[Good Go practices observed]
```

For each issue include:
- File and line number
- Issue description
- Suggested fix
- Severity: Critical / Major / Minor / Suggestion

Notes from targeted review on 2026-04-05:
- Created `/Users/kayvan/src/fabric/.maestro/playbooks/GO_ISSUES.md` with YAML front matter, cross-links to `[[REVIEW_SCOPE]]` and `[[2_GO_QUALITY]]`, and a PR-scoped summary of the Go issues identified during this review.
- Captured three Major issues (missing caller-propagated context in the YouTube OCR path, the wide `GrabVisual` API surface, and provider implementations that ignore `context.Context`), two Minor issues (YouTube package/file organization and remaining initialism inconsistencies), and one Suggestion (error wrapping consistency).
- Recorded the clean static-analysis results from `modernize`, `gofmt -l`, and `go vet`, plus positive observations about timeout usage, config/help alignment, and the targeted regression coverage added during the review.
- No task images were attached for this checklist item.

## Success Criteria

- All Go files reviewed for idioms
- Error handling verified
- Context usage checked
- Static analysis completed
- GO_ISSUES.md created

## Status

Mark complete when Go review document is created.

---

**Next**: Document 3 will validate Fabric pattern system changes.
