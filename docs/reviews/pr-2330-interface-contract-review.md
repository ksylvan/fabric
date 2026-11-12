---
type: analysis
title: PR 2330 Interface Contract Review
created: 2026-11-11
tags:
  - code-review
  - interface-contracts
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-api-compatibility-review]]'
  - '[[pr-2330-data-flow-review]]'
  - '[[pr-2330-error-handling-review]]'
---

# PR 2330 Interface Contract Review

Completed only the interface-contract checkbox for the fixed 25-file snapshot between `de0b72766db0436b6c1887bb75768cb3960e67e5` and `fb296d490cca4fdebbcb78d5107241f134152d19`. All scoped working-tree files match the head byte for byte. No new confirmed interface-contract defect was found. Production source and original PR tests remain unchanged; consolidated issues are a later task. Existing robustness and diagnostic findings in [[pr-2330-error-handling-review]] remain open.

## Signatures and return values

| Surface | Contract and assessment |
| --- | --- |
| Construction | `internal/tools/keenable/keenable.go:34`: `NewClient() *Client` initializes the embedded plugin base and API-key question. No caller arguments or error result are needed for this in-memory construction. Use the constructor: a zero-value client has nil embedded state and is not ready for `Search`. Registry construction at `internal/core/plugin_registry.go:55` supplies the initialized client. No zero-value usability guarantee is introduced. |
| Search | `internal/tools/keenable/keenable.go:62`: `Search(query string) (string, error)` accepts a query and returns formatted Markdown, consistent with neighboring providers. Empty results, missing/null results, or all URL-less items yield an empty string with nil error. Request, transport, non-200, and decode failures return an empty string and an error from this method. Wrapped request/transport/decode errors retain their causes; the existing non-200 read-cause finding remains applicable. |
| Formatting | `internal/tools/keenable/keenable.go:106`: result order is preserved; nonempty URLs are required for inclusion. Title may be absent. Description takes precedence; absent/empty description falls back to snippet, truncated to 1,000 runes plus an ellipsis. Descriptions are not truncated. The final Markdown is trimmed. Neither URL validation nor Markdown escaping is promised by this signature. |
| CLI dispatch | `internal/cli/tools.go:14` and `:126`: the existing `(messageTools string, err error)` contract is preserved. A successful search appends Markdown, writes output for a standalone request, or defers it to chat when `IsChatRequest` is true. Search errors return before append/write. Output-write errors propagate. Earlier tool contributions can already exist in `messageTools`; an error must still be honored by callers. |
| Configuration | `internal/tools/keenable/keenable.go:46` and `internal/core/plugin_registry.go:531`: the key is explicitly optional. Inherited configuration loads it; absence does not block search. Registry configuration follows existing optional-plugin conventions. The new setting participates in existing serialization and setup interfaces without changing their return types. |

## Required and optional inputs

- **CLI query:** `internal/cli/flags.go:76` defines an optional string flag. When supplied without a value, the parser rejects it. Both `--keenable_search query` and `--keenable_search=query` preserve the string, including Unicode. An explicitly empty value parses but skips the branch because dispatch tests `!= ""`. Whitespace-only values are nonempty and are forwarded; `Search` itself does not validate query content. These are local semantics, not a claim that the remote service accepts empty queries.
- **API key:** absent or whitespace-only values select `/v1/search/public` without `X-API-Key`; a trimmed nonempty value selects `/v1/search` and supplies the header. Trimming does not mutate the stored setting. The stable environment name is `KEENABLE_API_KEY`. No configuration gate is required for the public endpoint.
- **Outbound request:** `internal/tools/keenable/keenable.go:64` sends POST JSON with string `query` and integer `max_results: 10`, plus `Content-Type: application/json` and `X-Keenable-Title: Fabric`. No caller-selectable result limit or context/cancellation parameter is exposed; the shared HTTP client supplies a 30-second timeout.
- **Inbound response:** `internal/tools/keenable/keenable.go:51` expects a top-level `results` array with string `title`, `url`, `description`, and `snippet` fields. Go decoding tolerates omitted/null optional fields and unknown properties. Wrong non-null field types cause a decode error with no partial Markdown. Missing/null results are accepted as empty output. HTTP 200 is the only success status; a 204 is an HTTP error. This describes the implementation, without asserting a separately verified external schema.
- **Other changed files:** help, README, and all three completion scripts agree on the string-valued flag; locales provide matching formatting placeholders. Registry setup/persistence additions use existing interfaces. No REST handler, dependency manifest, or storage schema changes in the fixed scope.

## Optional documentation suggestion

**Suggestion — `internal/tools/keenable/keenable.go:60` and `internal/cli/flags.go:76`:** document constructor use, empty-result success, and the distinction between an omitted flag, a missing argument, and an explicitly empty argument. The short method comment currently explains Markdown and keyless access. Clarifying these edge semantics would help future callers; no established contract requires treating them as errors. Do not classify missing-results acceptance as a service-schema violation without an authoritative schema requirement.

## Validation and limits

Used graph discovery and source snippets for the constructor, search, response type, shared settings, chat predicate, and dispatcher; inspected the fixed diff and relevant tests. No on-disk `AGENTS.md` or `CLAUDE.md` was found in the repository; the session-supplied graph-first instructions were followed. All 25 scope paths match the fixed head, and fixed-diff whitespace checks passed.

Added review-only cases under the external Auto Run `Working` folder, reusing the original `withServer` helper and earlier review overlays:

- `keenable_contract_review_test.go`: missing results, unknown properties, null optional fields, and incorrect URL/description types.
- `keenable_cli_contract_review_test.go`: absent flag, separate/equals values, Unicode, missing value, and explicitly empty value.
- `interface-contract-review-overlay.json`: combines these with existing logic, data-flow, and error-handling fixtures. Earlier tests cover optional key loading/serialization, request types and headers, Unicode truncation, output/chat routing, unchanged input state, repeated calls, body closure, typed errors, and failure propagation.

All seven full suites passed with `go test -race -timeout 90s -overlay /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/interface-contract-review-overlay.json ./internal/tools/keenable ./internal/tools/serply ./internal/tools/firecrawl ./internal/cli ./internal/core ./internal/plugins ./internal/i18n`. Keenable, CLI, and core returned fresh passes; the other four returned cached passes. The review fixtures remain outside Git, consistent with earlier tasks, so reproducing the overlay requires the Auto Run folder.

No live API or model calls, real credentials, external-service schema verification, or images were involved. Zero images analyzed. This local contract assessment does not establish the remote service's current behavior or resolve the previously reported unbounded-body and diagnostic issues.
