---
type: analysis
title: PR 2330 Logic Correctness Review
created: 2026-11-11
tags:
  - code-review
  - correctness
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-context]]'
---

# PR 2330 Logic Correctness Review

## Scope and outcome

Reviewed all 25 changed files between base `de0b72766db0436b6c1887bb75768cb3960e67e5` and head `fb296d490cca4fdebbcb78d5107241f134152d19`. Verified that every changed working-tree file matches that head before testing. No confirmed logic defect was found in this task. This is the first correctness checkbox only; the separate data-flow, best-practices, API, and categorized-issues tasks remain open. No production code was changed. No images were associated with the task or analyzed.

## File coverage

| Changed files | Logic checked and result |
| --- | --- |
| `internal/tools/keenable/keenable.go` | Constructor initializes the plugin and API-key question. Trimmed empty keys select the public endpoint and omit the key header; nonempty keys select the keyed endpoint. Marshal/request/transport/decode errors return without formatting results. Non-200 responses produce errors. Successful responses close their bodies. The finite range loop skips empty URLs, prefers descriptions, and truncates only fallback snippets after 1,000 runes, preserving Unicode. Empty/null result lists return empty output without error. |
| `internal/tools/keenable/keenable_test.go` | Existing cases exercise POST payload, public/keyed paths, title header, description preference, missing URLs, long snippets, and 429 errors. Tests do not run in parallel, so restoring shared endpoint variables through cleanup is consistent with current execution. Additional review-only cases are described below. |
| `internal/cli/tools.go` | Search errors return immediately. Successful output is appended; standalone requests write output and return, while chat requests retain results for the caller. This follows adjacent Serply/Firecrawl branches and the existing `IsChatRequest` predicate. Earlier standalone tools can return before later tools execute; this is the existing precedence behavior, not a new regression. |
| `internal/cli/flags.go`, `internal/cli/help.go` | The new string flag uses the same spelling as dispatch, and its help key exists in every changed locale. An empty flag value does not invoke search, consistent with adjacent tools. |
| `internal/core/plugin_registry.go` | The normal constructor initializes Keenable before configuration and dispatch. The optional key is loaded through existing settings configuration, serialized through existing environment-file support, and offered in optional setup. No key is required to use the public endpoint. No new loops or termination rules were introduced. |
| All 14 changed locale JSON files | Parsed every file and verified all seven new keys and matching formatting verbs: `%d`/`%s` for HTTP errors and `%w` for wrapped errors. |
| `completions/_fabric`, `completions/fabric.bash`, `completions/fabric.fish` | All register the implemented flag as accepting an argument; Bash also includes it in the simple-argument branch. Bash and Zsh syntax checks pass. Fish was reviewed statically. |
| `README.md`, `cmd/generate_changelog/incoming/2330.txt` | Flag spelling, public/keyed endpoint selection, optional-key behavior, and result routing agree with implementation. |

## Independent validation

All seven package suites passed:

```sh
go test ./internal/tools/keenable ./internal/tools/firecrawl ./internal/tools/serply ./internal/plugins ./internal/i18n ./internal/cli ./internal/core
```

Created review-only tests in the Auto Run `Working/keenable_logic_review_test.go` file and loaded them into the package with a Go overlay, leaving the reviewed source unchanged. They passed with the race detector:

```sh
go test -race -overlay /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/logic-review-overlay.json ./internal/tools/keenable -count=1
```

Additional cases cover empty/whitespace/padded keys, environment configuration and serialization, malformed JSON, incompatible result types, null/empty results, URL-less entries, exactly 1,000 and 1,001 Unicode runes, description precedence, request-construction failure, and transport-error wrapping with `errors.Is`. HTTP responses use local test servers; the transport failure uses a stub. Existing test helpers were reused.

Also passed `bash -n completions/fabric.bash`, `zsh -n completions/_fabric`, and programmatic JSON/key/format checks for all 14 changed locales. These tests do not call the live Keenable service or require credentials.

## Limits and later-review notes

The live service contract, availability, and translated prose accuracy were not verified. A zero-value manually assembled registry/client can contain nil fields; the normal constructor path initializes them, so this was not classified as a defect. The 1,000-rune bound intentionally applies only to fallback snippets, matching the recorded PR description.

`internal/tools/keenable/keenable.go:92-95` reads an entire non-200 body and ignores a read failure; successful JSON decoding also has no response-size bound. These are concrete robustness concerns to evaluate during the dedicated error-handling/security tasks, not evidence of an off-by-one or incorrect routing condition. Unknown/missing JSON fields produce an empty successful result under the current schema; whether that should instead be rejected depends on the service contract. No unsupported service-contract assumptions were promoted to confirmed findings.
