---
type: analysis
title: PR 2330 New Function Coverage Assessment
created: 2026-11-11
tags:
  - code-review
  - testing
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-related-tests]]'
  - '[[pr-2330-new-tests]]'
---

# New function coverage

Assessed the first unchecked **New code coverage** task in Document 4 against fixed PR base `de0b72766db0436b6c1887bb75768cb3960e67e5` and head `fb296d490cca4fdebbcb78d5107241f134152d19`. The new production file adds exactly two function/method declarations: `NewClient` and `(*Client).Search`, both in `internal/tools/keenable/keenable.go`. Other changed Go files extend existing functions or declarations; the completion edits add options to existing scripts. Those modified-code changes belong to the next task.

This is a static assessment of assertions and control flow, not measured runtime coverage or a passing-test claim. The checkout's Keenable source and tests match the fixed PR head. Graph discovery returned no Keenable nodes; the fixed Git diff and known test file supplied fallback evidence. A literal search of Go test files found no Keenable references outside `internal/tools/keenable/keenable_test.go`.

## Function assessment

| New function | At least one test? | Happy paths | Error cases | Edge cases |
| --- | --- | --- | --- | --- |
| `NewClient` (`keenable.go:34`) | Yes, indirectly: all three client tests construct it. | Construction and usable API-key question are exercised by successful searches; keyed test assigns `ApiKey.Value`. | No explicit error return or constructor error branch exists. | No direct assertions for plugin name, localized setup description/optional marker, `KEENABLE` environment prefix, or optional API-key question metadata. Configuration from environment and optional-key setup are not verified by these tests. |
| `(*Client).Search` (`keenable.go:65`) | Yes: `TestSearchWithoutKeyUsesPublicEndpoint`, `TestSearchWithKeyUsesKeyedEndpoint`, `TestSearchReturnsAPIError`. | Keyless POST path, absent key header, Fabric title header, query and result-limit payload; keyed path and key header; multi-result Markdown output and description preference. | One HTTP 429 response checks returned error contains the body detail. Request-construction, transport, and response-decoding failures are untested. Status-code formatting is not asserted. | URL-less result skipping, snippet fallback, and truncating 1,500 ASCII characters to 1,000 plus ellipsis are asserted. Empty results are returned by the keyed fixture, but output is discarded and only absence of an error is asserted. Other boundaries below remain untested. |

## Suggested cases for uncovered behavior

| Priority | Function / path | Missing assertion or scenario | Suggested test |
| --- | --- | --- | --- |
| High | `Search`: HTTP execution failure | No coverage of `httpClient.Do` failure or its localized error wrapper. | Use a deterministic failing transport and assert a non-nil execution error with the underlying failure detail; restore the shared client after the test. |
| High | `Search`: JSON decode failure | No malformed, empty-body, or incompatible-field-type response. | Serve HTTP 200 with invalid JSON, an empty body, and a non-array `results` value; assert decode errors and no result text. |
| Medium | `Search`: endpoint/key normalization | Keyed test uses only an unpadded key; no whitespace-only key. | Check empty and whitespace-only keys select public with no key header; check a padded key selects keyed and sends its trimmed value. |
| Medium | `Search`: rune truncation | ASCII-only test cannot distinguish rune slicing from byte slicing; exact threshold is untested. | Test snippets of 999, 1,000, and 1,001 runes, including multibyte characters; assert exact output, valid UTF-8, and ellipsis only above the limit. |
| Medium | `Search`: request-construction failure | Invalid endpoint never exercises `http.NewRequest` error handling. | Temporarily set the selected endpoint to a malformed URL and assert the create-request error; restore endpoint state. |
| Medium | `Search`: empty/fallback output | Empty results have no output assertion; missing/null results, all URL-less results, and empty description plus empty snippet are untested. | Assert empty string for no usable results and exact Markdown for a URL-bearing result with no text. Include a short snippet to establish fallback without truncation. |
| Medium | `NewClient`: plugin metadata | Constructor use alone does not verify setup metadata or optional API-key behavior. | With a controlled locale, assert environment prefix, API-key question identity/optional setting, and expected translated metadata using existing plugin test patterns. Test configuration loading in the later integration assessment. |
| Low | `Search`: request and API error details | Content-Type is set but never asserted; HTTP 429 assertion does not verify status inclusion. | Assert `application/json` and that a non-200 error includes both status code and returned detail; include an empty error body. |
| Low | `Search`: query preservation and description branch | Only ordinary queries and a short description are tested. | Verify quotes/newlines/Unicode survive JSON encoding. Assert a long description is preserved, since the current implementation truncates only fallback snippets; do not assume descriptions share that limit. |

The `json.Marshal` error branch cannot be induced through the current public inputs: its map contains only a string query and an integer constant, both JSON-supported values. Record this as structurally unreachable with the present payload rather than demanding artificial mocking or a production refactor solely to exercise it. The unchecked read error from a non-200 response body and response-size/redirect behavior may warrant separate robustness review, but there are no explicit error-handling branches for them in this implementation.

## Test-only helper

`withServer` (`internal/tools/keenable/keenable_test.go:13`) is the only new named test helper and is used by all three tests. It redirects both endpoint variables to a local HTTP server and registers cleanup that restores them and closes the server. It needs no standalone production-function coverage entry. Request-handler callbacks belong to these tests rather than new public functionality.

## Completion boundary and validation

Both new production functions have test invocations and happy-path exercise. Error coverage is partial for `Search`, and constructor metadata and several input/output boundaries lack assertions. The cases above are review recommendations, not implemented tests or confirmed production defects.

Only the **New code coverage** checkbox is completed. Modified-code coverage, consolidated coverage gaps, test-quality review, suite execution, and final `TEST_GAPS.md` remain for later runs. No source or test files were changed, and no test suite was run for this documentation-only assessment. Validation checked the fixed diff, exact source/test snapshot match, the two production function declarations and three test bodies, and documentation whitespace. No images were associated with this task or analyzed.
