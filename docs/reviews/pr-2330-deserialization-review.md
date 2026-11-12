---
type: analysis
title: PR 2330 Insecure Deserialization Review
created: 2026-11-11
tags:
  - code-review
  - security
  - keenable
  - deserialization
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-security-context]]'
  - '[[pr-2330-error-handling-review]]'
  - '[[pr-2330-xxe-review]]'
---

# Insecure deserialization review

Reviewed the fixed PR snapshot from `de0b72766db0436b6c1887bb75768cb3960e67e5` to `fb296d490cca4fdebbcb78d5107241f134152d19`. The working Keenable client, CLI dispatch, and registry files match that head. This completes only the insecure-deserialization checkbox. Production code is unchanged; zero images were associated with this task or analyzed.

## Assessment

No unsafe object reconstruction or executable deserialization path was identified. The external HTTP response is untrusted, but `encoding/json` decodes it into the fixed `searchResponse` structure with a slice of four string fields (`internal/tools/keenable/keenable.go:51-58,99-103`). The type is not selected by response metadata, and this structure has no custom unmarshal hook. There is no newly introduced gob, pickle, YAML object constructor, reflection-based class selection, or evaluation of serialized commands. Queries are accepted as strings and serialized outward with `json.Marshal`, rather than parsed as objects (`:62-67`). Registry changes construct/configure the client and reuse string settings; CLI changes consume the resulting Markdown without a second deserialization step.

Type checking occurs **during** decoding into that concrete Go type, before any result is formatted or used. A separate parser or unchecked generic-object cast is unnecessary here. JSON type errors can leave the local destination partly populated, but `Search` checks the error and returns before formatting, so no partial results escape. This does not establish strict schema validation: null/missing values, unknown fields, duplicate keys, and trailing data have permissive behavior described below.

| Check | Evidence and result |
| --- | --- |
| Invalid types | Wrong top-level array/string/boolean, object/scalar `results`, scalar result elements, and number/boolean/object/array values in each of title, URL, description, and snippet return wrapped `json.UnmarshalTypeError`, with empty output. |
| Partial or malformed JSON | A valid result followed by an invalid URL type produces no partial Markdown. Truncated input returns `io.ErrUnexpectedEOF`; 11,000 nested arrays return `json.SyntaxError` without panic. These samples do not constitute general fuzzing or resource-exhaustion testing. |
| Null and absent data | Top-level null, absent/null results, null result elements, and null URLs yield empty output without panic. Empty URLs are skipped. Acceptance may hide an API contract regression, but no authorization or executable-object boundary depends on these values. |
| Type metadata | An unknown `$type` and command-shaped string are ignored; the valid result is formatted normally. Static inspection establishes that no execution sink or custom decoder is present; the inert command string was never executed. |
| Parser permissiveness | Duplicate string fields use the later value; upper-case field names are accepted. A second JSON document or non-JSON text after a complete first document is ignored because only one `Decode` is called, with no EOF check. No second parser or signed-document validation boundary was found in the new flow, so these behaviors alone do not demonstrate an exploitable parser differential. |
| Cleanup and network boundary | Every new malformed/permissive probe closes its response body and makes exactly one intercepted request to the fixed API host. No result URL is fetched. |

## Medium: unbounded response allocation

- **Type:** uncontrolled resource consumption (CWE-400), during JSON decoding and formatting; this is the same response-size issue recorded as a major robustness finding in [[pr-2330-error-handling-review]], not an additional distinct issue.
- **Location:** `internal/tools/keenable/keenable.go:94,100,106-117`.
- **Condition and impact:** a malicious or faulty upstream service/proxy can send a large body within the request timeout, causing excessive allocation and potentially terminating the local CLI process through memory exhaustion. Successful decoding has no byte limit; result count, title, URL, and description lengths have no local cap. `max_results: 10` is only an outbound hint. The snippet cap runs after full decoding and applies only when description is empty. Non-200 `io.ReadAll` is likewise unbounded.
- **New evidence:** a bounded in-process response containing 11 results, each with a 2 MiB description, succeeds and preserves every full description in output. This demonstrates missing count/description caps without attempting an out-of-memory crash. The earlier error-body probe covers the failure path.
- **Recommendation:** reuse established bounded-read patterns where available, with an explicit byte budget and overflow detection before decoding; bound error diagnostics separately. Validate useful result-count/field-length limits before formatting. The 30-second timeout is not a memory budget. If the API requires exactly one JSON document, reject trailing content by checking for EOF after decoding. Unknown-field rejection should be an intentional compatibility decision, rather than an assumed security requirement.

No new code-execution or privilege-escalation finding was identified. The conditional upstream denial-of-service issue remains open for the consolidated security report. Untrusted text/URLs and raw error disclosure remain covered by earlier review documents.

## Validation and limits

Followed the supplied graph-discovery instructions, using graph search and a source snippet for `Search`, then checked the fixed diff for the typed structure, CLI integration, and registry changes. No repository-local `AGENTS.md` or `CLAUDE.md` was found. Reused `reviewTransport`, `errorReviewBody`, and the existing XXE overlay instead of creating production helpers or dependencies.

Added external review-only artifacts:

- `Working/keenable_deserialization_review_test.go`: 34 boundary cases in both keyed and keyless modes (68 subtests), plus one bounded response-size/count probe. Assertions for permissive parsing and missing caps characterize current behavior; they are not desired contracts for a future fix.
- `Working/deserialization-review-overlay.json`: extends the prior XXE overlay to include these probes and existing review regressions.

All three full targeted package suites passed with the race detector on Go 1.27.3:

```sh
W=/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working
TMPDIR="$W/tmp" go test -race -timeout 90s -overlay "$W/deserialization-review-overlay.json" ./internal/tools/keenable ./internal/cli ./internal/core
```

Reproduction requires the external Working fixtures, consistent with earlier review tasks. No live API/model calls, actual credentials, images, full repository suite, or remote-service implementation audit were involved. Dependency review and later checklist items remain pending.
