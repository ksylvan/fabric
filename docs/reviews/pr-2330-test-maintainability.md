---
type: analysis
title: PR 2330 Test Maintainability Review
created: 2026-11-11
tags:
  - code-review
  - testing
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-related-tests]]'
  - '[[pr-2330-test-structure]]'
  - '[[pr-2330-test-assertions]]'
---

# Test maintainability review

Completed only **Test maintainability** in Document 4. Reviewed all three added client tests against the fixed PR base `de0b72766db0436b6c1887bb75768cb3960e67e5` and head `fb296d490cca4fdebbcb78d5107241f134152d19`, plus adjacent parser, output, registry, and locale tests. Recommendations below are review findings, not implemented changes or demonstrated runtime failures.

## Added tests

| Area | Evidence and assessment | Priority and recommendation |
| --- | --- | --- |
| Numeric literals | `internal/tools/keenable/keenable_test.go:32,36,45`: 10 is the expected request result limit; 1,500 creates an over-limit snippet; 1,000 is the expected truncation boundary (`snippetMaxChars` in production). Their purposes can be inferred, but the fixture does not explain the relationship. HTTP 429 already uses `http.StatusTooManyRequests`. | Low: name the fixture lengths and comment that the input exceeds the 1,000-rune contract. Keep the expected contract independent of the production constant so accidentally changing that constant does not silently change the oracle. Add named below/at/above-boundary cases when extending coverage. |
| Setup complexity and reuse | `withServer`, lines 13–21, centralizes endpoint replacement and cleanup for all three cases. Fresh clients and small inline JSON fixtures keep setup readable; there is no large fixture builder or unnecessary abstraction. The helper lacks `t.Helper()`. | Low: retain this helper and add `t.Helper()` when revising it. Reuse it for future serial HTTP cases rather than creating another server wrapper. |
| Global-state fragility | `withServer` temporarily replaces both package endpoint globals and restores them with `t.Cleanup`. No added test calls `t.Parallel`; current tests do not depend on another test's output. | Low: document the serial constraint. Introduce client-local endpoint injection only if concurrent tests become necessary; adding parallel execution to these tests without isolation would be unsafe. No race or order failure was established. |
| Failure diagnostics | Request checks in the keyless/keyed handlers use fatal `require` assertions outside the test goroutine. The keyless test combines request shape, filtering, fallback, and truncation, with Boolean prefix/suffix checks. | Medium: capture request observations and assert in the test goroutine; split distinct output behaviors into named cases and compare expected Markdown with an equality assertion. This repeats the underlying issues in [[pr-2330-test-structure]] and [[pr-2330-test-assertions]], not additional independent defects. |
| Brittle expectations | The keyless test repeats literal titles/URLs in its response and expected Markdown. These are observable output expectations, not private call-order assertions. Error checking uses a diagnostic substring rather than the full localized message. | Low: keep explicit expected output independent from the formatter. Small duplicated literals are acceptable; use local named fixture values if the fixture grows, without deriving expected output from production formatting. Exact Markdown remains an intentional contract check. |
| Test data duplication | Empty-results JSON appears only in the keyed test, error JSON only in the error test, and the longer fixture only in the keyless case. The synthetic key is repeated once for input and once for expectation. | No actionable duplication finding: a shared fixture framework would add indirection to these short tests. Reuse the existing helper, not a new cross-provider abstraction. |

## Adjacent tests: pre-existing observations

- **Medium — `internal/core/plugin_registry_test.go:24`, `TestSaveEnvFile`:** uses shared `os.TempDir()` without a test-owned directory or cleanup. This can leave persistent state or collide with other processes. Use `fsdb.NewDb(t.TempDir())`; the existing `TestNewPluginRegistry_CodexTokenPersist` already demonstrates isolated storage and content assertions. Reuse that pattern for Keenable persistence coverage.
- **Low — `internal/cli/output_test.go:18,40`:** fixed package-local filenames can collide across concurrent processes; cleanup is registered after successful writes. Use paths inside `t.TempDir()`. The two small tests duplicate a little file-reading scaffolding but cover distinct newline contracts; this alone does not justify a new helper.
- **Low — `internal/cli/flags_test.go:146`, `TestInitWithYAMLConfig`:** repeated argument save/restore and temporary-file setup lengthen the test. Named subtests and a shared read-only YAML fixture keep it understandable. Prefer its existing `t.TempDir()` pattern for future fixtures; introduce a helper only when repeated setup warrants it. Literal models/temperatures are intentional parser inputs, not unexplained timing thresholds. Global `os.Args` mutation constrains parallelization.
- **No new maintainability issue — `internal/i18n/i18n_test.go`:** locale parity reads embedded fixtures dynamically, avoiding a duplicated key list. Exact translation strings deliberately pin values and require updates for copy changes. Existing initialization mutates package state, so parallelizing translation cases would require isolation.

## Validation and boundary

Used graph search and source snippets for the new tests, helper, YAML setup, and registry save test. A broader adjacent-file graph query returned no results, so inspected the known inventory paths directly. Cross-checked the added test file against the fixed PR diff. No repository-root `AGENTS.md` or `CLAUDE.md` exists; followed session guidance.

Validated report metadata, related-document references, source paths, whitespace, and the exact playbook checkbox change. No production code or tests changed; no tests were added or run for this documentation-only assessment. Suite execution and consolidated `TEST_GAPS.md` are separate pending tasks. No images were associated with this task or analyzed.
