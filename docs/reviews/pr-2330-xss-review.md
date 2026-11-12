---
type: analysis
title: PR 2330 Cross-Site Scripting Review
created: 2026-11-11
tags:
  - code-review
  - security
  - keenable
  - xss
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-security-context]]'
  - '[[pr-2330-sensitive-data-review]]'
  - '[[pr-2330-misconfiguration-review]]'
---

# Cross-site scripting review

Reviewed the fixed PR snapshot from `de0b72766db0436b6c1887bb75768cb3960e67e5` to `fb296d490cca4fdebbcb78d5107241f134152d19`. The inspected Keenable, CLI, sanitizer, chat component, and CSP files match that head. This completes only the XSS checkbox. Production code is unchanged; zero images were associated with this task or analyzed.

## Assessment

No exploitable XSS path introduced by this PR was demonstrated in the reviewed application. Keenable emits untrusted Markdown, including literal HTML and arbitrary URL schemes; it does not emit browser-safe HTML. The added flow writes text to stdout/a file or supplies text to a chat request. It introduces no browser HTML insertion sink or web route. The existing web chat sanitizes assistant Markdown before HTML insertion and escapes other message types. No direct Keenable-to-web integration was found in the scoped changes; testing the existing sanitizer with search output characterizes a protective boundary, not an end-to-end web-search feature.

| Check | Evidence and result |
| --- | --- |
| User input | `internal/tools/keenable/keenable.go:65` serializes the query as a JSON string. The client does not place the query directly into its returned Markdown. This is request encoding, not HTML sanitization; the external service can still return hostile content. |
| External content | `internal/tools/keenable/keenable.go:101-118` copies result title, URL, description, and fallback snippet into Markdown without escaping or URL-scheme validation. Snippet truncation limits length, not active content. JSON decoding restores literal angle brackets even when the response encoder escaped them. All four fields preserved the tested hostile payloads. |
| Output encoding | `internal/cli/tools.go:126-139`, `internal/cli/flags.go:524-539` append and print/save text. Hostile markup is preserved in the standalone file and chat-tool return value. No HTML renderer executes in that code. Non-200 response text is likewise diagnostic text (`internal/tools/keenable/keenable.go:93-96`); previous disclosure findings remain separate. |
| HTML insertion | Existing `web/src/lib/components/chat/ChatMessages.svelte:76-88,151-168` routes assistant Markdown through `renderMarkdown` before the sole `{@html}` insertion. Plain/Mermaid, user, and system content use Svelte text interpolation. The unsanitized `return content` branch is used as text, not raw HTML. No React `dangerouslySetInnerHTML` or equivalent was added by the PR. |
| Sanitizer | `web/src/lib/utils/sanitize-html.ts:9-18,28-35` runs Marked output through DOMPurify with the HTML profile, forbids forms/styles, and returns escaped text on SSR or parsing failure. Existing tests cover script/event removal, unsafe URL schemes, SVG/MathML, forms, streamed partial output, restored messages, and parser failure. All 48 synthetic Keenable outputs passed DOM-based checks for dangerous tags, event attributes, and active URL schemes. |
| CSP | Existing `web/svelte.config.js:87-109` configures automatic CSP with `script-src 'self'`, `object-src 'none'`, `base-uri 'self'`, `form-action 'none'`, `frame-ancestors 'self'`, `connect-src 'self'`, and an image-source allowlist. These settings are unchanged by the PR. No CSP is applicable to the CLI's plain-text stdout. Deployed headers and production browser enforcement were not tested. |

## Conditional downstream hardening observation

- **Type:** untrusted Markdown/HTML propagation; potential CWE-79 only if consumed by an unsafe browser renderer.
- **Location:** `internal/tools/keenable/keenable.go:118` and CLI output at `internal/cli/tools.go:132-135`.
- **Severity:** Low hardening recommendation; no demonstrated application XSS vulnerability.
- **Condition and impact:** an operator must open or import the emitted Markdown in a separate renderer that permits active HTML or unsafe links without sanitization. A malicious search result could then execute script in that renderer's origin. Literal preservation alone does not establish script execution; Fabric's inspected sanitizer rejected the tested active constructs.
- **Recommendation:** document that search results are untrusted. Browser consumers should reuse the existing `renderMarkdown` boundary and sanitize after Markdown parsing; preserve safe SSR/error fallbacks and CSP. Consider allowing only HTTP(S) result URLs and escaping title text if a stricter search-result format is desired. Avoid a new parallel sanitizer or treating JSON encoding as HTML protection.

This observation does not assert an exploit in any particular external Markdown viewer. Model prompt manipulation is a separate trust concern and is not evidence of XSS.

## Validation

Graph searches and symbol snippets identified `Search`, `AppendMessage`, `WriteOutput`, `renderMarkdown`, and `sanitizeHtml`. The graph's inbound trace returned no callers for Keenable `Search`, so the fixed diff and CLI source supplied the dispatch evidence. Literal searches covered raw HTML insertion and CSP settings. No repository-local `AGENTS.md` or `CLAUDE.md` exists; followed the supplied graph-discovery instructions and reused existing review fixtures.

Review-only artifacts are in the external Auto Run `Working` directory:

- `keenable_xss_review_test.go`: 48 combinations of four result fields, six hostile payloads (image event handler, script, JavaScript Markdown link, SVG event handler, HTML iframe, form), and public/keyed mode. Asserts exact payload preservation, absence of query reflection, and only one outbound request. Saves actual formatted results as `xss-results.json` for the renderer probe.
- `keenable_cli_xss_review_test.go`: verifies hostile markup/URL preservation through standalone file output and chat-tool routing, with no file written in chat mode.
- `xss-review-overlay.json`: extends the prior overlay, reusing transport helpers and existing regression probes.
- `xss-web-review.mjs`: feeds all 48 actual outputs through the repository's `renderMarkdown` with jsdom; checks parsed DOM for unsafe constructs and verifies ordinary Markdown formatting remains functional. This tests sanitizer output, not browser execution.
- `xss-vitest.config.mjs`: minimal configuration for the existing browser-DOM and no-DOM sanitizer suites, avoiding unrelated application plugins during these unit tests.

Both targeted Go package suites passed with the race detector (Go 1.27.3). All 48 sanitizer probes plus the ordinary Markdown control passed, and the two existing sanitizer suites passed all 10 tests (Node 25.9.0 / Vitest 4.1.11). The initial Vitest attempt lacked generated `.svelte-kit/tsconfig.json`; ran SvelteKit sync and the same suites passed. Installed locked web dependencies with `npm ci --ignore-scripts --no-audit --no-fund`; package manifests and lockfiles are unchanged. Installation deprecation notices were not assessed as vulnerabilities in this task.

```sh
# From the repository root; W is the existing Auto Run Working directory.
W=/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working
TMPDIR="$W/tmp" REVIEW_XSS_OUTPUT="$W/xss-results.json" go test -race -timeout 90s -overlay "$W/xss-review-overlay.json" ./internal/tools/keenable ./internal/cli
node "$W/xss-web-review.mjs"
cd web
./node_modules/.bin/svelte-kit sync
./node_modules/.bin/vitest run --config "$W/xss-vitest.config.mjs"
```

Reproduction requires the external Working fixtures, consistent with earlier review tasks. No live Keenable or model calls, actual credentials, screenshot inspection, deployed CSP checks, or full repository test run were needed. Later security categories remain pending.
