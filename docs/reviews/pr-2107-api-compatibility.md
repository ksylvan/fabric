---
type: report
title: PR 2107 API Compatibility Review
created: 2026-10-03
tags:
  - code-review
  - api
  - compatibility
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-code-organization]]'
  - '[[pr-2107-correctness]]'
  - '[[pr-2107-data-flow]]'
  - '[[2_REVIEW_CODE]]'
---

# PR 2107 API compatibility review

Reviewed the **Breaking changes** checkbox for all three scoped files against base `ddf1aab968caa9adf90137a6536c768d02b237d6` and original PR head `f32656b529a1eb1c678c793703fe49a0755cd439`. Scoped source, REST handler, and options still match that head. No production changes were made; 0 images were supplied or analyzed. Interface-contract review remains a separate unchecked task.

## Assessment

The patch is source/schema compatible but intentionally changes behavior for `create_coding_feature` API calls and unsafe filesystem paths. It adds no function parameters, return types, REST fields, routes, or response-envelope changes. The changed tests do not alter an API. Security restrictions should ship without a grace period that preserves remote file writes; consumers depending on those writes must migrate.

| Surface | Before → after | Compatibility assessment |
| --- | --- | --- |
| REST file application (`internal/core/chatter.go:192`; `internal/server/chat.go:147`) | The pattern triggered parsing/application regardless of caller → requests with a nonnil `UpdateChan` skip both. Current REST handler always supplies a channel and requests streaming. | Intentional breaking change to a side effect: API clients can generate suggestions but no longer have this code apply them to the server's current directory. No opt-in field or compatibility mode is introduced. |
| API assistant session content (`internal/core/chatter.go:211`–`:215`) | Pattern response was replaced by its summary → API session appends the complete post-processing response, including the file-change marker/JSON. | Persisted response content changes. Session consumers and subsequent conversation turns can now see the payload. Existing sessions are not rewritten. This is a consequence of skipping the entire block, not a response-schema change. |
| REST streamed content | Content updates were forwarded before the pattern-specific application → same forwarding remains. | The live stream already contained the model payload; the new gate does not add the marker to the stream for the first time. Distinguish streamed content from saved session content. |
| CLI application (`internal/cli/chat.go:50`, `:84`; CLI `BuildChatOptions`) | CLI options omit `UpdateChan` → still omit it, so pattern processing remains enabled. | Normal streaming and non-streaming CLI flows keep their existing application/summary behavior. `Chatter.Stream` is independent of the channel guard. A custom internal caller supplying a channel now suppresses application even if it is a CLI consumer. |
| Parser paths (`internal/domain/file_manager.go:85`) | Rejected any path containing `..` → rejects paths for which `filepath.IsLocal` is false. | Absolute paths become rejected; safe filenames containing `..` and relative paths whose cleaned form stays local become accepted. This broadens valid local input while closing unsafe input. Host-platform semantics apply; Windows-specific acceptance was not tested. |
| Direct writer (`internal/domain/file_manager.go:157`) | No locality validation → nonlocal paths return an error before that entry mutates files. | Intentional behavioral restriction for callers bypassing parsing. Empty paths are also rejected. Validation remains per entry, so an error does not promise rollback of earlier writes; see [[pr-2107-correctness]]. |

## Deprecation and versioning

The original three-file diff contains no deprecation warning, migration note, or versioning change. No Go symbol is removed or renamed, so a symbol deprecation annotation is unnecessary. Warning users while continuing unsafe API writes would undermine the security fix. An API notice or release note can explain the immediate change without retaining that behavior.

The release configuration `.goreleaser.yaml:22` and `:37` injects `main.version` from the release version; this patch does not change that mechanism or prescribe a release number. No API version change is present. A mandatory major-version bump cannot be inferred from the scoped diff alone: release policy and supported guarantees are not established here. Release maintainers should explicitly classify the intentional behavior change and communicate it with the security release.

### Document the changed API behavior

**Severity:** Suggestion. **Locations:** `internal/core/chatter.go:192`, `data/patterns/create_coding_feature/README.md:21`.

The README describes the CLI workflow but does not explain that API calls now only generate suggestions or that saved API assistant messages retain the payload. Add release/migration guidance covering both effects, the relative-path requirement, and the new interpretation of `UpdateChan` for internal callers. Consumers needing application should review the generated changes and apply them through a separately authorized local workflow. Do not expose the internal channel as a REST permission switch.

The comment in `Send` describes the intended CLI/API split, but channel presence is still an implicit policy convention, as documented in [[pr-2107-code-organization]]. Source inspection supports the current REST boundary; it does not establish a safe default for every future caller. Earlier symlink containment and parser/data-loss findings remain unresolved by this compatibility assessment.

## Validation and limits

- Used the existing indexed graph for symbol discovery and source snippets: core `Send`, REST `HandleChat`, CLI `handleChatProcessing`/`BuildChatOptions`, `ChatOptions`, parser/writer, and existing tests. Rechecked the original diff, base post-processing block, scoped-file identity, README, and release configuration.
- Created and ran `TestReviewAPICompatibilityPaths`, with five cases covering ordinary relative paths, a filename containing `..`, a locally resolving parent component, escaping traversal, and an absolute path. Accepted cases preserve summary/records and write the expected content; rejected cases return errors and leave the project root empty. All passed on macOS.
- Probe retained at `Working/review_api_compatibility_probe_test.go` in the Auto Run folder. Its temporary copy in `internal/domain` was removed. To reproduce, copy into that package, run `go test ./internal/domain -run TestReviewAPICompatibilityPaths -v`, then remove the copy.
- Package regression validation: `go test ./internal/domain ./internal/core ./internal/server ./internal/cli`. Temporary storage and build cache use the authorized Auto Run Working folder. No permanent test or production file is changed by this documentation task.
- API session/gate conclusions are from before/after source inspection, not a new end-to-end test. No live provider, Windows, full-repository, or HTTP integration exercise was performed. This report does not claim shell-injection remediation; the scoped diff contains no template-extension execution changes.
