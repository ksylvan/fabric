---
type: analysis
title: PR 2330 Test Structure Review
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

# Test structure review

Completed only **Test structure** in Document 4: naming, setup/teardown, assertion focus, and independence. The fixed review snapshot is base `de0b72766db0436b6c1887bb75768cb3960e67e5` to head `fb296d490cca4fdebbcb78d5107241f134152d19`. The sole added test file is `internal/tools/keenable/keenable_test.go`; adjacent tests below are existing infrastructure, not new PR defects or proof of Keenable integration coverage.

## New client tests

| Criterion | Evidence and assessment |
| --- | --- |
| Behavior names | `TestSearchWithoutKeyUsesPublicEndpoint`, `TestSearchWithKeyUsesKeyedEndpoint`, and `TestSearchReturnsAPIError` describe observable routing or failure behavior. The keyless test also checks formatting/filtering/truncation, which its name does not reveal. |
| Setup and teardown | Each test calls `withServer`, creates a fresh client, and uses a local HTTP server. The helper captures both endpoint globals and uses `t.Cleanup` to restore them and close the server, including after a fatal assertion in the test goroutine. Adding `t.Helper()` would improve failure locations. |
| Assertion focus | Multiple request assertions in the keyed test support one routing contract; the API-error test has one clear outcome. The keyless test combines request shape, output formatting, URL-less filtering, description preference, snippet fallback, and truncation. Separate named subtests would localize regressions while reusing the server helper. One assertion per test is unnecessary. |
| Independence | No test consumes another test's output. Endpoint globals are restored, and no `t.Parallel` is present in this file. This supports sequential isolation; global endpoint mutation would become unsafe if these tests were parallelized. This is a future constraint, not an observed race. |

## Findings and suggested checks

| Priority | File and function | Structure issue | Suggested improvement and verification |
| --- | --- | --- | --- |
| Medium | `internal/tools/keenable/keenable_test.go`: keyless and keyed test handlers | `require.Equal`, `require.Empty`, `require.NoError`, and `require.EqualValues` run inside `httptest` handler goroutines. On failure, fatal assertions call `FailNow` outside the goroutine running the test, terminating the handler rather than stopping the test body as intended. | Capture request observations through a buffered channel and assert from the test goroutine after `Search`, or use nonfatal handler checks with explicit error responses/returns. Verify a deliberately incorrect path/body produces a direct, understandable test failure. |
| Low | `internal/tools/keenable/keenable_test.go`: `TestSearchWithoutKeyUsesPublicEndpoint` | A single broad test obscures which output behavior failed and stops at the first fatal check. | Use named cases for request shape, URL-less filtering, description preference, snippet fallback, and truncation; keep each case independent with its own fixture. |
| Low | `internal/tools/keenable/keenable_test.go`: `withServer` | Process-wide endpoint overrides constrain concurrency; helper lacks `t.Helper()`. | Keep these tests serial while overrides are global. If concurrent execution is needed, prefer client-local endpoint injection. Add `t.Helper()` when revising the helper and verify shuffled/repeated runs restore endpoints. |

## Adjacent test structure

- `internal/cli/flags_test.go`: flag-parser tests save and restore `os.Args` with deferred functions. YAML cases use descriptive subtest names and independently invoke `Init`; shared YAML content is read-only. They remain serial because arguments are global. `TestBuildChatOptions` checks several related fields of one returned contract, which is appropriate. No Keenable case is present.
- `internal/cli/output_test.go`: both active file-output tests register cleanup and check file contents, but use fixed filenames in the package working directory. Cleanup is registered only after a successful write. Prefer `filepath.Join(t.TempDir(), ...)` so concurrent processes and failed writes cannot leave or overwrite shared fixtures. **Low priority, pre-existing**; verify repeated/concurrent runs use separate directories.
- `internal/core/plugin_registry_test.go`: `TestSaveEnvFile` passes shared `os.TempDir()` to `fsdb.NewDb`, with no test-specific directory or explicit cleanup. **Medium priority, pre-existing**: use `t.TempDir()` and verify persistence against an isolated environment file. `TestNewPluginRegistry_CodexTokenPersist` already demonstrates a test-owned directory and builds its own seed state; it does not rely on the earlier save test.
- `internal/i18n/i18n_test.go`: translation cases are named by language and call `Init` for each case. `TestLocaleKeyParity` reads embedded locale files without writing shared fixtures. Translation initialization changes package state; these serial tests do not establish safety for parallel localization tests.
- `internal/cli/cli_test.go`: `TestCli` is unconditionally skipped. Its argument restore is correctly scoped but never executes; it supplies no runtime integration evidence. `TestCopyToClipboard` in `output_test.go` is also skipped.

The adjacent cases show global-state and shared-path constraints, not a demonstrated requirement to run tests in a specific order. No runtime order/race claim is made.

## Validation and boundary

Graph-first test-symbol discovery returned no matching nodes in the indexed checkout, so known paths from [[pr-2330-related-tests]] and direct source inspection supplied evidence. Checked the fixed PR test diff and reviewed all three new client tests plus adjacent parser, output, registry, and locale tests. Documentation validation checks referenced paths, front matter, the exact checkbox change, and `git diff --check`.

No production code or tests changed, so no tests were added or executed for this documentation-only assessment. Suite execution is a separate pending checkbox. Mocking, detailed assertion quality, maintainability, and consolidated `TEST_GAPS.md` are also pending. No images were associated with this task or analyzed.
