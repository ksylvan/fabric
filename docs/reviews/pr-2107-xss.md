---
type: report
title: PR 2107 Cross-Site Scripting Review
created: 2026-10-03
tags:
  - code-review
  - security
  - xss
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-security-context]]'
  - '[[pr-2107-misconfiguration]]'
  - '[[3_CHECK_SECURITY]]'
---

# PR 2107 cross-site scripting review

Completed only the **XSS (Cross-Site Scripting)** checkbox, including input handling, output encoding, CSP, and raw HTML sinks. Compared original base `ddf1aab968caa9adf90137a6536c768d02b237d6` with original head `f32656b529a1eb1c678c793703fe49a0755cd439`. No new XSS regression was established in the three changed Go files. The adjacent web UI has a pre-existing High unsafe assistant-HTML rendering path. No production code changed. No images were supplied or analyzed (0).

## Assessment

| Check | Result |
| --- | --- |
| Input handling | User and system messages use Svelte text interpolation. Assistant output is untrusted: it can reflect prompts, retrieved material, or model output. `cleanPatternOutput` removes wrappers/whitespace, not HTML or unsafe URLs (`ChatService.ts:92-105`). Language handling adds a prefix or returns unchanged content (`ChatService.ts:22-25`). Neither is a sanitizer. |
| Output encoding | `ChatMessages.svelte:86` returns `marked.parse` output directly. Plain responses, selected-pattern Mermaid responses, and parser-error fallback return raw content (`:65-92`). All assistant branches enter the same `{@html}` sink at `:161`. Fenced code is encoded by the tested parser; this does not protect other Markdown or plain-response content. |
| Raw HTML equivalents | The graph-backed web search found the runtime chat `{@html}` sink and detached-textarea entity decoding (`transcriptService.ts:9-13`). The latter returns text; no separate execution sink was demonstrated there. The apparent HelpModal match is prose mentioning the pattern `sanitize_broken_html_to_markdown`, not sanitization code or a sink. Build-time mdsvex highlighting also emits HTML; successful highlighting escapes Svelte syntax, while fallback interpolates code into markup (`web/svelte.config.js:44-57`). Those inputs are repository Markdown in the inspected configuration, not a demonstrated remote-input route. |
| CSP | No explicit CSP configuration or response policy was found in `web/svelte.config.js`, `web/src`, or `internal/server`. Existing missing API headers are documented in [[pr-2107-misconfiguration]]. Deployment/proxy CSP was not available; a suitable policy could reduce exploitability but would not repair this renderer. |
| Original PR boundary | Filesystem path validation and the automatic-write gate introduce no browser HTML sink. API responses continue carrying generated text; preventing API file writes does not sanitize that text for web display. There are no web changes in the original PR range. |

## High: assistant content reaches HTML insertion without sanitization

- **Type:** XSS / injection (CWE-79); potentially persistent through chat history.
- **Locations:** `web/src/lib/components/chat/ChatMessages.svelte:75-92,159-161`; `web/src/lib/services/ChatService.ts:92-105`; message accumulation/persistence at `web/src/lib/store/chat-store.ts:16-28,127-147`; saved-session ingress at `web/src/lib/store/session-store.ts:93-101` and `web/src/lib/components/chat/SessionSelector.svelte:38-41`.
- **Evidence:** The actual extracted rendering functions, using lockfile versions marked 18.0.14 and Svelte 5.57.1, preserve `<img ... onerror="...">` in Markdown, selected-pattern, plain, and selected-pattern Mermaid cases. Markdown also retains a `javascript:` link. A synthetic parser exception returns raw HTML. Compilation of the actual component confirms `$.html` receives `renderContent` output, while user/system interpolation compiles into text updates. These tests preserve synthetic markup without executing event handlers or contacting a model.
- **Reachability:** Live SSE content is decoded by `ChatService.createMessageStream` (`:107-206`) and accumulated as assistant messages by the chat store. Browser localStorage persists those messages and restores them on load. Selecting a saved session loads its `Message` array directly into the same store. Stored or live assistant-role text therefore reaches the same rendering boundary; normal user-role text does not. JSON parsing/serialization does not make the eventual HTML insertion safe.
- **Impact/preconditions:** An attacker must influence assistant output or assistant-role history that a victim displays, for example through hostile external content reflected by a model or an attacker-writable shared session. Executable event attributes can run with the web origin's authority when rendered without an effective blocking CSP; unsafe links require user activation. Potential impact includes reading/modifying local chat history and making requests with whatever authority the web origin actually has. No arbitrary remote history write, credential theft, cross-user exploit, or browser execution was demonstrated. A self-supplied prompt alone is not evidence of a cross-user attack.
- **Recommendation:** Reuse a maintained sanitizer at the existing rendering boundary after Markdown conversion, with an explicit tag/attribute and URL-protocol policy. Route every HTML-producing branch through that boundary, including fallback responses. Render plain and Mermaid text through ordinary Svelte interpolation; provide a safe text fallback on parser/sanitizer errors. Preserve user/system text interpolation. Add behavioral browser tests for event handlers, unsafe links, streamed fragments, saved history, and plain/error cases. Set a compatible CSP for the actual UI origin as defense in depth.
- **Attribution:** All identified rendering, persistence, and session-loading paths are unchanged by the original PR. This is an existing application vulnerability, not a new changed-line regression.

## Transport and validation

The real server `writeSSEResponse` (`internal/server/chat.go:257-269`) JSON-escapes markup and preserves one SSE frame even when payload text contains newlines resembling another event. Both synthetic payloads round-trip back to their original content on JSON decode, confirming transport escaping is not HTML sanitization. `detectFormat` labels the tested payloads Markdown (`:271-281`). No direct reflected HTML response exploit was established in the inspected API paths.

Eight Node characterization tests and one Go transport test with two payload cases passed. Probe sources are retained in the authorized Auto Run Working folder:

- `Working/review_xss_probe.test.mjs`: run from the repository root with `node --test /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2107-codex/Working/review_xss_probe.test.mjs`. Isolated dependency versions are recorded under `Working/xss-runtime`; no application dependency or lockfile changed.
- `Working/review_xss_server_probe_test.go`: copy temporarily into `internal/server`, run `go test ./internal/server -run TestReviewXSS -v`, then remove the copy.

After removing the temporary Go test, the existing server/core/domain suites passed on Go 1.27.1, darwin/arm64. Node probes ran on Node 25.9.0. Used the existing indexed graph for discovery and snippets, with exact-line/source verification afterward; no root `AGENTS.md` or `CLAUDE.md` is present. No new helper or production test was added for this review-only task. Full frontend/repository suites, live browser execution, external deployment policy, and subsequent playbook categories were not assessed. The passing probes characterize the unsafe output rather than assert that the vulnerability has been fixed.
