---
type: report
title: PR 2330 Test Coverage Review
created: 2026-11-11
tags:
  - code-review
  - testing
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-coverage-gaps]]'
  - '[[pr-2330-test-assertions]]'
  - '[[pr-2330-test-maintainability]]'
  - '[[pr-2330-test-mocking]]'
  - '[[pr-2330-test-structure]]'
  - '[[pr-2330-test-execution]]'
---

# Test Coverage Review

Review scope: PR 2330, base `de0b72766db0436b6c1887bb75768cb3960e67e5` to head `fb296d490cca4fdebbcb78d5107241f134152d19`. This consolidates the completed coverage, quality, and execution assessments for eight changed source paths, one added test file, and 14 locale files. No existing test was updated. Both new public callables, `keenable.NewClient` and `(*Client).Search`, are invoked by the three new tests.

Existing tests pass locally, but important error and integration behavior lacks direct assertions. The findings below are missing test evidence and quality recommendations, not confirmed production failures or measured coverage percentages. Detailed evidence is in the related review documents. No source or test changes were made for this consolidation.

## Missing Tests

| Priority | Source file and function/method | Missing evidence | Suggested test cases |
| --- | --- | --- | --- |
| High | `internal/tools/keenable/keenable.go`, `(*Client).Search` | Transport failure and JSON decoding failure are untested. | Use a deterministic failing transport, restoring shared state; serve malformed JSON, an empty HTTP 200 body, and a non-array `results` field. Require the relevant error detail and empty output. Reuse `withServer` for HTTP fixtures. |
| High | `internal/cli/flags.go`, `Flags.KeenableSearch` / `Init`; `internal/cli/tools.go`, `handleToolProcessing` | Client-only tests bypass flag parsing, registry-backed dispatch, standalone output, and chat routing. No direct Keenable dispatcher test was located. | Parse the query flag, default empty value, and missing-value error; exercise keyless dispatch without a configured-key gate. Assert query forwarding, standalone output, appended chat results, search/output error propagation, empty-query bypass, and ordering/early returns with another tool selected. |
| High | `internal/core/plugin_registry.go`, `NewPluginRegistry`, `Configure`, `SaveEnvFile` | Existing construction/save tests do not assert Keenable initialization, key loading, or persisted content. | With isolated environment and `t.TempDir()`, assert non-nil client, absent-key operation, `KEENABLE_API_KEY` loading, configured-key persistence, absent-key serialization under the optional-plugin contract, and preservation of other providers. Reuse the Codex persistence test pattern. |
| Medium | `internal/tools/keenable/keenable.go`, `NewClient` | Constructor metadata is used but never explicitly asserted. | Under a controlled locale, check translated identity/setup description, optional marker, `KEENABLE` prefix, and optional API-key question. |
| Medium | Same file, `Search` | No whitespace-key, malformed-endpoint, or Unicode/exact-boundary cases. | Whitespace-only key selects public with no key header; padded key selects keyed and sends the trimmed value. Invalid selected URL returns a request-creation error. Test 999/1,000/1,001-rune snippets with multibyte characters, exact output, valid UTF-8, and ellipsis only above the limit. |
| Medium | Same file, `Search` | Empty/missing/null results and fallback edges lack output assertions. | Assert empty string for no usable results, including all URL-less entries; exact Markdown for a URL-bearing result without text; short snippet fallback; long description preserved under current behavior. |
| Medium | `internal/core/plugin_registry.go`, `runInteractiveSetup` | Optional-tools menu membership is unverified. | Use controlled setup input/output to assert Keenable appears in the optional group. |
| Medium | `internal/cli/help.go`, `flagDescriptionMap` / `getTranslatedDescription`; `internal/i18n/locales/*.json` | Locale parity does not verify help mapping or message formatting. | Assert active-locale flag help, setup optional marker, and error substitutions without formatting artifacts. |
| Medium | `completions/_fabric`, `completions/fabric.bash`, `completions/fabric.fish`, existing option/argument handlers | No new-option completion assertions were located. | Verify each shell offers `--keenable_search` and consumes the query argument, including bash's simple-argument branch. Syntax checks alone cannot establish completion behavior. |
| Low | `internal/tools/keenable/keenable.go`, `Search` | Request content type and nontrivial query encoding are unchecked. | Assert `application/json` and round-trip quotes, newlines, and Unicode. Check actual key-header omission when absent. |

The current JSON payload contains only a string and integer, so its marshal-error branch is structurally unreachable through public inputs. Do not introduce artificial mocking solely to execute it. Response bounds, redirects, and timeout policy remain separate robustness concerns.

## Inadequate Tests

| Priority | Source / test function | Inadequacy | Suggested improvement |
| --- | --- | --- | --- |
| Medium | `keenable.go`, `Search` / `keenable_test.go`, `TestSearchWithoutKeyUsesPublicEndpoint` | Prefix/suffix assertions allow unexpected text between them. ASCII truncation cannot establish rune semantics. | Compare the entire expected Markdown with equality; add the boundary cases above. Keep the expected 1,000-rune contract independent of the production constant. |
| Low | Same source / `TestSearchWithKeyUsesKeyedEndpoint` | Empty-results fixture discards returned output; keyed method/body contract is unchecked. | Assert empty output and expected method, JSON payload, and content type. |
| Low | Same source / `TestSearchReturnsAPIError` | HTTP 429 test checks body detail only and discards output. | Require both status 429 and body detail, empty output, and a separate non-200 empty-body case. Prefer substring checks over exact localized error equality. |
| Medium | `internal/core/plugin_registry.go`, `SaveEnvFile` / `plugin_registry_test.go`, `TestSaveEnvFile` | Error-only assertions cannot detect omitted or incorrect Keenable entries. | Read isolated persisted content and compare keyless/keyed entries and preserved provider settings. This overlaps the integration gap above. |

## Test Quality Issues

| Priority | File and function | Finding | Suggested improvement |
| --- | --- | --- | --- |
| Medium | `internal/tools/keenable/keenable_test.go`, keyless/keyed HTTP handlers | Fatal `require` assertions execute in server-handler goroutines. Failure may terminate the handler and create a secondary request error. | Capture request observations through a buffered channel and assert in the test goroutine, or use nonfatal checks with explicit handler failure responses. |
| Low | Same file, `withServer` and keyless test | Endpoint globals constrain parallelism; helper lacks `t.Helper()`; one broad case mixes several output contracts. | Keep endpoint-mutating cases serial, add `t.Helper()`, and use independent named cases. Existing cleanup restores globals; no race was observed. |
| Medium | `internal/core/plugin_registry_test.go`, `TestSaveEnvFile` (pre-existing) | Shared `os.TempDir()` fixture lacks test-owned isolation. | Use `fsdb.NewDb(t.TempDir())` and assert persisted content, following the existing Codex test. |
| Low | `internal/cli/output_test.go`, active file-output tests (pre-existing) | Fixed filenames and cleanup after successful writes can collide or leave fixtures. | Use paths inside `t.TempDir()`. |
| Low | `internal/cli/flags_test.go`, parser cases (pre-existing) | Some nonfatal prerequisite checks precede pointer dereferences; `os.Args` mutation constrains parallelism. | Use fatal prerequisite assertions or guard dependent operations; preserve serial argument handling. |

Small inline fixtures and repeated expected values do not justify a new shared fixture framework. No excessive duplication, incorrect expected value, or over-mocking was found in the added tests. Numeric fixture lengths would benefit from names explaining the truncation boundary.

## Test Improvements

1. First add deterministic transport/decode cases and registry-backed CLI/configuration tests: these cover the highest-priority unverified failure and integration contracts.
2. Strengthen existing client assertions and move handler assertions into the test goroutine while reusing `withServer`.
3. Add independent key-normalization, Unicode/boundary, empty-output, metadata, help/locale, and completion cases from the tables above.
4. Apply existing isolated-filesystem patterns when extending registry/output tests; avoid parallelizing tests that mutate global endpoints, arguments, or locale state.

These are recommendations for subsequent implementation. This task documents gaps; it does not fill them or add a test framework.

## Well-Tested Areas

- All three new tests exercise the real client against local HTTP stubs without live credentials. Serialization, HTTP handling, decoding, and formatting remain real; mocks do not replace the behavior under test.
- Keyless/keyed routing and synthetic key transmission have direct assertions. The keyless test also checks POST, query, result limit 10, and the Fabric title header.
- Description preference, URL-less filtering, snippet fallback, and long ASCII snippet truncation have useful partial output assertions; these behaviors are not wholly untested.
- The HTTP 429 case verifies propagation of the returned diagnostic body.
- Behavior-oriented names, fresh clients, and endpoint/server cleanup support sequential independence. Existing parser and output tests assert values; locale parity indirectly checks new key presence across all 14 locales, though not translation meaning or interpolation.

## Test Execution Results

Reused the completed execution task's results and re-counted its retained JSON logs during consolidation; no Go suite was rerun for this documentation-only task. Tested commit: `414906d31cac1f583e1ab6172e7d06f267d8df91` (fixed PR head plus review documentation). Environment: Go 1.27.3 on darwin/arm64, November 11, 2026. Module declares Go 1.26.10.

```bash
go test -json -count=1 -timeout=120s ./...
go test -json -race -count=1 -timeout=120s ./internal/plugins/... ./internal/server/... ./internal/core/... ./internal/tools/keenable
```

| Run | Exit | Elapsed | Packages passed | Packages without tests | Test/subtest pass events | Explicit test skips |
| --- | --- | --- | --- | --- | --- | --- |
| Full suite | 0 | 27.89 s | 46 | 14 | 1,416 | 5 |
| Race scope plus Keenable | 0 | 36.79 s | 28 | 1 | 1,052 | 0 |

All three added Keenable tests passed in both runs; no failures, timeouts, or race reports occurred. Counts include parent and subtest events, not just independent top-level functions. Stderr was empty.

Full-suite skips: `TestApplyFileChanges_RejectsGitShortName` (Windows-only feature), `TestCli` (existing `-t` flag collision), `TestCopyToClipboard` (explicit Docker/CI skip), `TestNotifySendProvider` (Linux-only), and `TestPowerShellProvider` (Windows-only). The broad CLI test supplies no executed integration evidence.

Warnings: the full run could not create `/some/path` on the read-only filesystem; both runs reported no content from an upstream Fabric server. Neither warning failed its run. These local results do not establish Linux/Windows matrix behavior, live service compatibility, or shell completion correctness.

Commands, environment/cache settings, skips, and log paths are detailed in [[pr-2330-test-execution]]. Raw evidence is under `/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/`: `test-execution-suite.jsonl`, `test-execution-race.jsonl`, corresponding `.stderr` files, and `test-execution-results.json`.

Validation: reconciled prior reports with the known client source/tests after graph discovery returned no matching Keenable nodes; checked report structure, source paths, matching report copies, recorded test counts, and whitespace. No images were associated with this task or analyzed.
