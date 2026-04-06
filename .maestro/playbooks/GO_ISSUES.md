---
type: report
title: Go Code Quality Issues
created: 2026-04-05
tags:
  - go
  - code-review
  - maestro
  - youtube
related:
  - '[[REVIEW_SCOPE]]'
  - '[[2_GO_QUALITY]]'
---
# Go Code Quality Issues

Findings captured from the Go-focused review documented in [[2_GO_QUALITY]] for the scoped files identified in [[REVIEW_SCOPE]].

## Critical Issues

No critical Go issues were identified in the reviewed scope. The branch does not introduce compiler errors, panics in library code, or static-analysis failures from the required `modernize`, `gofmt`, and `go vet` passes.

## Major Issues

### Issue 1: OCR workflow does not accept caller-provided context

- Severity: Major
- Files: `internal/tools/youtube/youtube.go:936`, `internal/tools/youtube/youtube.go:947`, `internal/cli/cli.go:117`, `internal/cli/cli.go:150`
- Issue: `GrabVisual` creates its own `context.Background()` timeout internally and `processYoutubeVideo` calls it without a caller-supplied `context.Context`. This prevents CLI or request-scoped cancellation from propagating into long-running `yt-dlp`, `ffmpeg`, and `tesseract` subprocesses.
- Suggested fix: Thread `context.Context` from the CLI/server entry points into `processYoutubeVideo`, `YouTube.Grab`, and `YouTube.GrabVisual`. Apply the 15-minute timeout as `context.WithTimeout(ctx, ...)` on top of the incoming context rather than starting from `context.Background()`.

### Issue 2: `GrabVisual` uses a wide positional API that is already straining

- Severity: Major
- File: `internal/tools/youtube/youtube.go:936`
- Issue: `GrabVisual(videoId string, language string, additionalArgs string, sensitivity float64, fps int)` carries five positional parameters, mixes option transport with execution logic, and makes the call site at `internal/cli/cli.go:150` harder to read and extend safely.
- Suggested fix: Refactor to `GrabVisual(ctx context.Context, videoID string, opts *VisualOptions)` or fold the visual extraction fields into an existing options struct so new controls do not require more positional arguments.

### Issue 3: Several AI provider implementations ignore the `context.Context` they accept

- Severity: Major
- Files: `internal/plugins/ai/ollama/ollama.go:93`, `internal/plugins/ai/ollama/ollama.go:107`, `internal/plugins/ai/gemini/gemini.go:63`, `internal/plugins/ai/gemini/gemini.go:127`, `internal/plugins/ai/vertexai/vertexai.go:64`, `internal/plugins/ai/vertexai/vertexai.go:182`, `internal/plugins/ai/lmstudio/lmstudio.go:55`, `internal/plugins/ai/lmstudio/lmstudio.go:92`
- Issue: Multiple provider methods name the context parameter `_ context.Context`, which makes cancellation and deadlines unavailable even though the interface surface implies they are supported.
- Suggested fix: Propagate the provided context into SDK calls, HTTP requests, and streaming loops. Where a provider truly cannot honor cancellation, document that limitation explicitly and consider tightening the interface contract.

## Minor Issues

### Issue 4: `youtube.go` has grown into a multi-responsibility file

- Severity: Minor
- File: `internal/tools/youtube/youtube.go:1`
- Issue: `internal/tools/youtube/youtube.go` is now 1067 lines long and bundles transcript fetching, playlist handling, metadata/comments, OCR frame extraction, subprocess/logging helpers, and flag wiring in one file. That makes the package harder to navigate and increases change coupling.
- Suggested fix: Split the file by responsibility, for example into `transcript.go`, `playlist.go`, `metadata.go`, `visual.go`, and shared subprocess/helper files.

### Issue 5: Older identifier names still use non-idiomatic initialisms

- Severity: Minor
- Files: `internal/tools/youtube/youtube.go:127`, `internal/tools/youtube/youtube.go:149`, `internal/tools/youtube/youtube.go:848`, `internal/tools/youtube/youtube.go:866`
- Issue: The package still contains pre-existing names such as `videoId`, `ChannelId`, and `YtDlpArgs`. Those spellings are inconsistent with standard Go initialism handling (`ID`).
- Suggested fix: Plan a focused rename pass to normalize identifier spellings (`videoID`, `channelID`, etc.) while preserving serialized field names and external behavior.

## Suggestions

### Suggestion 1: Standardize error wrapping at external-call boundaries

- Severity: Suggestion
- Files: `internal/tools/youtube/youtube.go:141`, `internal/tools/youtube/youtube.go:257`, `internal/tools/youtube/youtube.go:610`, `internal/cli/flags.go:309`
- Issue: The reviewed Go files still rely heavily on `fmt.Errorf(...)` and `errors.New(...)`, even where Fabric's convention prefers wrapped errors with context.
- Suggested fix: Standardize on a single wrapping approach such as `pkg/errors.Wrap` or modern `%w`-based wrapping so callers can inspect causes while still getting actionable context.

## Static Analysis Results

### Modernize

- Command: `go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest ./...`
- Result: Passed with no diagnostics.

### Gofmt

- Command: `gofmt -l .`
- Result: No unformatted Go files reported.

### Go Vet

- Command: `go vet ./...`
- Result: Passed with no diagnostics.

## Positive Observations

- The reviewed PR slice does not introduce new `panic()` usage in library code; the YouTube visual extraction path returns errors to callers instead of terminating the process.
- The new OCR pipeline uses `exec.CommandContext` for `yt-dlp`, `ffmpeg`, and `tesseract`, which is the right foundation for bounded subprocess execution once caller context is propagated.
- The scoped Go additions continue to accept interfaces only at genuine I/O boundaries such as `io.Reader` and `io.Writer`, while returning concrete types from package APIs.
- The new YouTube visual extraction configuration is now aligned with Fabric's existing CLI and YAML configuration flow, and the related static analysis checks stayed clean.
