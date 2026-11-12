---
type: analysis
title: PR 2330 Test Mocking Review
created: 2026-11-11
tags:
  - code-review
  - testing
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-related-tests]]'
  - '[[pr-2330-test-structure]]'
  - '[[pr-2330-coverage-gaps]]'
---

# Mocking review

Completed only **Mocking** in Document 4, covering external dependency isolation, fixture realism, and over-mocking. Review snapshot: base `de0b72766db0436b6c1887bb75768cb3960e67e5` to head `fb296d490cca4fdebbcb78d5107241f134152d19`. All three new tests are in `internal/tools/keenable/keenable_test.go`.

## Dependency boundaries

| Boundary | Evidence | Assessment |
| --- | --- | --- |
| Keenable HTTP service | `withServer` (lines 13–21) starts `httptest.NewServer`, redirects both endpoint globals to its `/public` and `/keyed` paths, and restores them and closes the server through `t.Cleanup`. Every new test uses this helper. | The service is replaced with a local HTTP stub. No live API credentials or Keenable service are required by these tests. The shared production HTTP client remains real. |
| Client behavior | Each test constructs `NewClient` and calls `Search`; no replacement of serialization, HTTP transport, decoding, filtering, or Markdown formatting occurs. | Good boundary choice: actual request/response behavior is exercised. No over-mocking or mock call-order expectations found. |
| Plugin setup/localization | Constructor setup questions and translation lookup remain real. The keyed test assigns the synthetic value `secret` directly to `ApiKey.Value`. | Appropriate for client unit tests, but does not establish environment loading, interactive setup, registry persistence, or locale selection. Those integration gaps are tracked in [[pr-2330-coverage-gaps]]. |

## Fixture realism and limitations

- `TestSearchWithoutKeyUsesPublicEndpoint` (lines 23–46) receives real POST requests and decodes the JSON body. It checks public routing, absence of the key header, the application-title header, query, and result limit. Its JSON response uses the fields consumed by production: description plus snippet, a result without a URL, and an oversized fallback snippet. These are useful contract-shaped fixtures, including deliberately incomplete data for filtering behavior.
- `TestSearchWithKeyUsesKeyedEndpoint` (lines 48–59) checks keyed routing and the synthetic key header, then returns an empty result array. This is sufficient to exercise authentication routing; it adds no evidence of keyed nonempty-result handling. The key is test data, not a credential requirement.
- `TestSearchReturnsAPIError` (lines 61–69) returns HTTP 429 with a JSON detail body. This realistically exercises a rate-limit response and body propagation, but does not emulate a full rate-limit policy, retry headers, or transport failure.
- Success fixtures omit an explicit response `Content-Type`; production decodes the body without checking that header, so this does not invalidate these tests. Setting `application/json` would make the stub closer to the expected API response. Request `Content-Type` is set by production but not checked by the current handlers.
- The local routes deliberately differ from production `/v1/search/public` and `/v1/search`. They distinguish endpoint selection but cannot detect a typo in either default production URL. Fixture shapes and production field tags agree; no captured response or schema validation in these tests independently proves live-service compatibility. No live API call was made for this review.

## Findings and suggested cases

| Priority | File/function | Gap or constraint | Suggested improvement |
| --- | --- | --- | --- |
| High | `keenable_test.go`: `Search` tests | Stubs cover success and one HTTP error, but not connection/transport errors or malformed/truncated JSON. | Reuse the server helper for malformed and truncated response bodies. For deterministic transport failure, use a narrowly scoped failing `RoundTripper` with restoration of the shared client, or inject a client if concurrency is needed; assert error propagation and no fabricated output. Avoid sleeps or a real remote outage. |
| Medium | `keenable_test.go`: keyed/keyless handlers | `require` fatal assertions run in server-handler goroutines. This can terminate a handler and obscure the original contract failure with a secondary client error. | Capture observations through a buffered channel and assert in the test goroutine, or use nonfatal handler checks and explicit failure responses. See [[pr-2330-test-structure]] for the existing finding. |
| Low | `keenable_test.go`: `withServer` | Endpoint replacement uses package globals, constraining parallel tests and bypassing default URL validation. Cleanup restores the globals; no current parallel execution is present. | Keep cases serial while globals are shared. Add a separate default-endpoint contract check if desired; use client-local configuration when parallelism is required. Do not replace `Search` itself with a mock. |
| Low | `keenable_test.go`: request/response fixtures | Synthetic happy-path fixtures do not cover whitespace-only/padded keys, Unicode snippets, or varied empty-result bodies. | Extend independent local-server cases for key normalization, multibyte truncation, and empty/missing results. Check JSON request content type and emit JSON response content type. Reuse the existing helper. |

The missing failure cases are coverage gaps, not evidence that current mocks are unrealistic or excessive. No new mocking framework or duplicate helper is needed.

## Adjacent tests

Inspected the related parser, output, registry persistence, and locale tests identified in [[pr-2330-related-tests]]. Parser tests exercise actual argument/YAML handling; output and registry tests use the real filesystem; locale tests use embedded locale resources and actual localization. Keeping these inexpensive local dependencies real is appropriate integration testing, not a failure to mock every dependency. `TestNewPluginRegistry_CodexTokenPersist` uses a test-owned directory and synthetic tokens without invoking a remote provider. `TestSaveEnvFile` uses shared `os.TempDir()` and output tests use fixed filenames: isolation issues already documented in [[pr-2330-test-structure]], not new over-mocking findings. Skipped CLI/clipboard tests provide no evidence of a mocked search-to-chat or desktop integration path.

## Validation and boundary

Used graph-first discovery and snippets for the three new tests, helper, and registry tests. The production-symbol graph query returned no matches, so direct source inspection supplied the client implementation; known adjacent paths supplied additional context. Cross-checked the added tests against the fixed PR diff. No repository-root `AGENTS.md` or `CLAUDE.md` exists; followed the session-provided graph-first guidance.

This is a documentation-only assessment: no source or tests changed, no new tests were added, and no Go tests were run. Suite execution remains a separate pending task. Validated referenced paths, front matter, the single checkbox transition, and whitespace with `git diff --check`. Assertions, maintainability, suite execution, and consolidated `TEST_GAPS.md` remain pending. No images were associated with this task or analyzed.
