---
type: analysis
title: PR 2330 Test Assertion Review
created: 2026-11-11
tags:
  - code-review
  - testing
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-test-structure]]'
  - '[[pr-2330-test-mocking]]'
  - '[[pr-2330-coverage-gaps]]'
---

# Assertion review

Completed only **Assertions** in Document 4: meaningful assertions, correct expected values, and appropriate assertion types. Reviewed the fixed PR snapshot from `de0b72766db0436b6c1887bb75768cb3960e67e5` to `fb296d490cca4fdebbcb78d5107241f134152d19`, including all three added Keenable tests and adjacent parser, output, registry, and locale tests. No incorrect expected value was identified in the added tests; several assertions leave observable behavior unconstrained.

## Added client tests

| Test and location | Evidence | Assessment |
| --- | --- | --- |
| `TestSearchWithoutKeyUsesPublicEndpoint`, `internal/tools/keenable/keenable_test.go:25–32` | Checks POST, `/public`, empty key-header value, `Fabric` title, successful JSON decoding, query `golang generics`, and result limit 10. | These are meaningful request-contract assertions. Expected values agree with `Search`. `EqualValues(10, ...)` accommodates JSON numbers decoded into `any` as `float64`; strict equality with an `int` would be inappropriate for this representation. A typed request fixture could make the numeric contract clearer. `Empty(Header.Get(...))` proves the header has no value, rather than proving the header is absent. |
| Same test, lines 42–45 | Requires success, a specific Markdown prefix, absence of the URL-less result title, and a suffix of 1,000 ASCII characters plus an ellipsis. | Goes beyond a no-error smoke test: checks description preference, section order, snippet fallback, filtering, and the truncated suffix. Expected formatting and count agree with production. However, independently matching prefix and suffix permits arbitrary inserted content between them, including an extra prefix of snippet characters. It therefore does not establish exact output length or exclude all unwanted content. ASCII input does not distinguish byte from rune truncation. |
| `TestSearchWithKeyUsesKeyedEndpoint`, lines 50–58 | Checks `/keyed` and the synthetic key value `secret`; requires success for an empty result array. | Routing and authentication assertions are useful and correctly valued. The returned string is discarded, so fabricated nonempty output for an empty result array would pass. Method, JSON body, title, and content type are not asserted in this keyed case. |
| `TestSearchReturnsAPIError`, lines 63–68 | Returns HTTP 429 and a detail body; `require.ErrorContains` requires an error containing `rate limited`. | Appropriate substring assertion: proves both an error and propagation of the diagnostic body without coupling the test to the entire localized message. It does not establish inclusion of status 429 or empty output on failure because output is discarded. |

## Findings and suggested assertions

| Priority | Source/test function | Issue | Suggested verification |
| --- | --- | --- | --- |
| Medium | `keenable.Search` / `TestSearchWithoutKeyUsesPublicEndpoint` | Prefix/suffix checks leave a gap in the output oracle and return only Boolean failure diagnostics through `require.True`. | Compare the entire expected Markdown string with `require.Equal`, built from the fixed sections and exactly 1,000 repeated characters plus ellipsis. This rejects extra text, wrong separators, duplicate results, and excess snippet length while showing the actual difference. Add a multibyte boundary case as tracked in [[pr-2330-coverage-gaps]]. |
| Medium | `keenable.Search` / keyed and keyless request handlers | Fatal `require` assertions execute in HTTP-handler goroutines. Failure uses `FailNow` outside the test goroutine and can produce secondary request errors. | Capture request observations and assert in the test goroutine, or use nonfatal checks with explicit handler failure responses. Existing finding from [[pr-2330-test-structure]], not a separate production defect. |
| Low | `keenable.Search` / `TestSearchWithKeyUsesKeyedEndpoint` | Empty-result output is not checked. | Retain `got` and assert `require.Empty(t, got)` after success. Check the keyed request method/body and JSON content type if treating both endpoints as independent request contracts. |
| Low | `keenable.Search` / `TestSearchReturnsAPIError` | Body substring alone cannot detect a missing or wrong status in the error; failure output is unchecked. | Retain output and require it empty; check the error contains `429` as well as the fixture detail. Keep substring checks rather than exact localized-message equality. |
| Low | `keenable.Search` / keyless request assertions | No request content-type assertion, and key-header value emptiness does not prove absence. | Assert `application/json`. If header omission is part of the contract, check the canonical header-map entry is absent, in addition to an empty `Header.Get` value. |

These are review recommendations; tests and production code were not modified. Missing transport/decode and key-normalization cases remain coverage findings in [[pr-2330-coverage-gaps]].

## Adjacent assertion quality

- `internal/cli/flags_test.go`: `TestInit` checks the parsed copy value, YAML tests check field values and CLI precedence, and `TestBuildChatOptions` compares the complete expected options struct. These assert behavior beyond success. Exact float equality is reasonable for direct assignment/parsing of the same literals, with no arithmetic involved. Some cases use nonfatal `assert.NoError` before dereferencing returned pointers; a fatal prerequisite or guarded assertion would avoid a secondary panic if parsing/building starts returning nil on error. No Keenable-specific flag assertion is present.
- `internal/cli/output_test.go`: both active output tests read the written file and compare exact content, including newline behavior. Their error checks stop before dependent operations. These are suitable value assertions; the skipped clipboard test provides no executed assertion evidence.
- `internal/core/plugin_registry_test.go`: `TestSaveEnvFile` checks construction/save errors only, so it cannot detect omitted or wrong Keenable environment entries. Add content assertions in an isolated directory for keyless and keyed persistence. This is a pre-existing test weakness applied to the new integration, not an assertion introduced by this PR. `TestNewPluginRegistry_CodexTokenPersist` is a stronger existing pattern: checks concrete client type, callback availability, persisted token values, and preservation of unrelated content. It covers Codex, not Keenable persistence.
- `internal/i18n/i18n_test.go`: `TestTranslation` compares exact translations for an existing unrelated key. `TestLocaleKeyParity` reports each missing English key across locales; presence assertions indirectly include new Keenable keys, but do not establish nonempty values, correct formatting placeholders, or semantic translation correctness. Do not interpret key parity as validation of the full new message contract.

## Validation and boundary

Used the existing project index and graph-first discovery. The targeted Keenable test query returned no nodes, so inspected the known test file directly; graph snippets supplied `Search` and `TestBuildChatOptions`. Cross-checked all added assertions against the fixed PR test diff. No repository-root `AGENTS.md` or `CLAUDE.md` is present; followed the session-provided guidance.

This is a documentation-only assessment. No code or test changes were needed and no Go suite was run; execution is a separate pending checkbox. Validated report front matter, related-document references, source paths, unchanged task text apart from its checkbox, and whitespace. Maintainability, suite execution, and consolidated `TEST_GAPS.md` remain pending. No images were associated with this task or analyzed.
