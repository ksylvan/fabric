---
type: analysis
title: PR 2330 Coverage Gaps
created: 2026-11-11
tags:
  - code-review
  - testing
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-new-code-coverage]]'
  - '[[pr-2330-modified-code-coverage]]'
  - '[[pr-2330-related-tests]]'
---

# Coverage gaps

Consolidates only the **Coverage gaps** task in Document 4 for fixed PR base `de0b72766db0436b6c1887bb75768cb3960e67e5` and head `fb296d490cca4fdebbcb78d5107241f134152d19`. These are missing assertions identified through static inspection, not measured coverage percentages, runtime failures, or confirmed production defects. Test quality, suite execution, and the final `TEST_GAPS.md` are later tasks.

## New public functions and methods

Neither new public callable is entirely untested: all three client tests invoke `internal/tools/keenable/keenable.go:34`, `NewClient`, and `:65`, `(*Client).Search`. Their missing contracts are listed below. Other production changes extend existing functions or declarations.

| Priority | Source / function | Missing contract | Suggested cases |
| --- | --- | --- | --- |
| Medium | `internal/tools/keenable/keenable.go:34`, `NewClient` | No explicit plugin identity, localized setup metadata, environment prefix, or optional-question assertions. | Assert translated name/description, `KEENABLE` prefix, API-key question identity and optional status using a controlled locale. |
| Medium | `internal/tools/keenable/keenable.go:65`, `Search` | The keyed fixture returns empty results but discards output; exact empty-output behavior is unasserted. | Assert empty string for empty, missing, and null results and for all URL-less results. Check exact Markdown for a URL-bearing result with no description or snippet. |

## New branches and conditions

| Priority | Source / function | Untested condition or boundary | Suggested cases |
| --- | --- | --- | --- |
| Medium | `internal/tools/keenable/keenable.go:65`, `Search` | Key trimming is untested; existing tests cover only empty and unpadded keys. | Whitespace-only key selects public endpoint without `X-API-Key`; padded key selects keyed endpoint and transmits the trimmed value. |
| Medium | Same, snippet fallback/truncation | Only a 1,500-character ASCII snippet exercises truncation. Exact threshold and rune semantics have no assertions. | Use 999, 1,000, and 1,001 runes including multibyte characters; assert exact text, valid UTF-8, and ellipsis only above 1,000. Include short snippet fallback and a long description that remains untruncated under current behavior. |
| Low | Same, request construction | Content-Type and nontrivial query encoding are unasserted. | Assert `application/json`; round-trip quotes, newlines, and Unicode in the request query. |
| Medium | `internal/cli/tools.go:126`, `handleToolProcessing` | No explicit Keenable branch test; empty-query bypass, output/chat split, and ordering with other tools are unverified. | Assert no search for empty query, standalone output with no chat request, appended results with chat, and documented early-return/order behavior with another tool selected. |

Description preference, skipping URL-less results, normal keyless/keyed routing, POST query/result-limit payload, Fabric title header, and long ASCII snippet truncation already have assertions; they should not be reported as wholly missing coverage.

## Error handling paths

| Priority | Source / function | Missing path | Suggested cases |
| --- | --- | --- | --- |
| High | `internal/tools/keenable/keenable.go:65`, `Search` | HTTP execution failure has no test. | Install a deterministic failing transport; assert execution error includes underlying failure detail and output is empty. Restore shared client state. |
| High | Same, response decoding | No malformed JSON, empty HTTP 200 body, or incompatible results field type. | Serve each response with a local HTTP fixture; assert decode error and empty output. |
| Medium | Same, request creation | Malformed selected endpoint never exercises `http.NewRequest` failure. | Temporarily replace an endpoint with an invalid URL; assert create-request error and restore endpoint state. |
| Low | Same, non-200 response | HTTP 429 test asserts body detail only, leaving status formatting unverified. | Assert status code and body detail; add a non-200 response with an empty body. |
| High | `internal/cli/tools.go:126`, `handleToolProcessing` | No assertion that search or output errors reach the caller. | Return a deterministic search failure and force a controlled output failure; assert error propagation and no successful result handling. |

The `json.Marshal` failure branch is structurally unreachable for the current payload, whose values are a string and integer. It is not an actionable missing test unless the payload types change. The ignored read error on a non-200 body has no explicit handling branch; response bounds, redirects, and timeout policy are separate robustness concerns, not evidence of failing tests.

## Integration points

| Priority | Source / integration | Missing evidence | Suggested cases |
| --- | --- | --- | --- |
| High | `internal/cli/flags.go:76`, `Flags.KeenableSearch` / `Init`; `internal/cli/tools.go:126`, `handleToolProcessing` | Client-only tests bypass flag parsing, registry wiring, keyless dispatch, and downstream output/chat. | Parse `--keenable_search` with a query, assert missing-value error and default empty value, then exercise registry-backed dispatch using controlled HTTP responses. Assert query forwarding and keyless operation without a configured-key gate. |
| High | `internal/core/plugin_registry.go`, `NewPluginRegistry`, `Configure`, `SaveEnvFile` | Construction/save tests provide incidental execution but no Keenable-specific initialization, environment loading, or persisted-content assertions. | Isolate environment and database directory; assert non-nil client, load `KEENABLE_API_KEY`, permit absent key, persist configured key correctly, and preserve other provider entries. Check absent-key serialization against the existing optional-plugin contract. |
| Medium | Same, `runInteractiveSetup` | Optional-tools menu inclusion is unasserted. | Exercise menu construction with controlled input/output and assert Keenable appears in the optional group; reuse existing setup abstractions. |
| Medium | `internal/cli/help.go:63`, `flagDescriptionMap` / `getTranslatedDescription`; `internal/i18n/locales/*.json` | Locale key parity checks presence only; no new help mapping or formatted setup/error assertions. | Assert new flag renders its active-locale help text, setup optional marker, and error substitutions without formatting artifacts. |
| Medium | `completions/_fabric`, `completions/fabric.bash`, `completions/fabric.fish` | No completion-specific test located. | Verify each shell offers `--keenable_search` and consumes its following query argument; test bash's simple-argument branch. Syntax checks alone do not prove completion behavior. |

The broad `internal/cli/cli_test.go` test is skipped. Existing output-file tests verify output helpers independently and do not invoke Keenable dispatch. `TestSaveEnvFile` checks only errors; the Codex persistence test asserts another provider's content. `TestLocaleKeyParity` supplies useful incidental coverage of key presence across locales, but does not establish Keenable rendering or interpolation correctness.

## Evidence and validation

Reused [[pr-2330-new-code-coverage]] and [[pr-2330-modified-code-coverage]], checked the fixed source/test diff, and used codebase-memory graph discovery and snippets for client, registry, and locale tests. Graph results for the dispatcher lag current source, so the fixed Git diff supplied its new branch. Literal test searches found Keenable references only in its three-test client file and no direct dispatcher/help-map references; negative searches do not exclude all indirect execution. The fixed diff adds one test file and updates no existing tests.

Validated report paths, coverage of all four requested gap categories, exact checkbox-only task-line modification, and `git diff --check`. This review task changes documentation only: no production code or tests were added, and no Go suite was run. No images were associated with this task or analyzed.
