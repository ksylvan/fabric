---
type: analysis
title: PR 2330 Related Test Inventory
created: 2026-11-11
tags:
  - code-review
  - testing
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-test-context]]'
---

# Related test inventory

This inventory completes only **Find related tests** in Document 4. Scope is the eight source paths in [[REVIEW_SCOPE]], using PR base `de0b72766db0436b6c1887bb75768cb3960e67e5` and head `fb296d490cca4fdebbcb78d5107241f134152d19`. Paths below are relative to the repository root. A related test is a candidate for later review, not evidence that the changed behavior is covered.

## Source-to-test mapping

| Changed source | Corresponding or related tests | Discovery notes |
| --- | --- | --- |
| `internal/tools/keenable/keenable.go` | `internal/tools/keenable/keenable_test.go` | Direct same-package client tests: `TestSearchWithoutKeyUsesPublicEndpoint`, `TestSearchWithKeyUsesKeyedEndpoint`, `TestSearchReturnsAPIError`; shared `withServer` helper uses `httptest`. |
| `internal/cli/flags.go` | `internal/cli/flags_test.go` | Same-module tests include `TestInit`, `TestInitWithYAMLConfig`, `TestBuildChatOptions`, and `TestIsChatRequestWithPrintPrompt`. No Keenable-specific reference found outside the client test file in Go tests. |
| `internal/cli/help.go` | `internal/i18n/i18n_test.go` (indirect); `internal/cli/flags_test.go` (parser context) | No `help_test.go` or test reference to `flagDescriptionMap` found. `TestTranslation` and `TestLocaleKeyParity` exercise translation infrastructure and locale key presence, respectively; neither establishes the new help mapping works. |
| `internal/cli/tools.go` | `internal/cli/cli_test.go`, `internal/cli/flags_test.go`, `internal/cli/output_test.go` (indirect) | No `tools_test.go` or test reference to `handleToolProcessing` found. `TestCli` is explicitly skipped. Flag/chat classification and output file tests are adjacent test candidates, not direct search-dispatch tests. |
| `internal/core/plugin_registry.go` | `internal/core/plugin_registry_test.go` | Direct module match. `TestSaveEnvFile` constructs the registry and saves environment content; `TestNewPluginRegistry_CodexTokenPersist` also constructs the registry for a different provider. No explicit Keenable reference found in registry tests. |
| `completions/_fabric` | No corresponding test located | No completion-specific test/spec or completion reference in code tests located. |
| `completions/fabric.bash` | No corresponding test located | Same completion search as above. |
| `completions/fabric.fish` | No corresponding test located | Same completion search as above. |

## Supporting tests and reusable patterns

- `internal/tools/serply/serply_test.go` and `internal/tools/firecrawl/firecrawl_test.go` are sibling search-client examples for subsequent review; they test their own clients, not Keenable.
- `internal/i18n/i18n_test.go` is also relevant to the 14 changed locale JSON files. `internal/i18n/locale_test.go` and `internal/i18n/i18n_variants_test.go` provide locale detection/loading context.
- `internal/cli/chat_test.go` and `internal/cli/prompt_export_test.go` provide downstream chat/output context. Their presence does not prove Keenable integration coverage.
- Other core tests discovered (`chatter_test.go`, `plugin_registry_model_spelling_test.go`, `vendor_configuration_warning_test.go`) focus on chat/provider behavior rather than the new tool registration.

## Discovery method and limits

Used the existing codebase-memory graph and graph-augmented code search first, including test-function discovery under `internal/cli`, `internal/core`, and `internal/tools`, plus searches for Keenable, related providers, tool dispatch, help mappings, and completion references. Some graph queries returned no symbols despite files existing, so repository file-list and text searches supplied the fallback inventory. Inspected relevant test declarations and selected bodies to distinguish direct tests from adjacent infrastructure.

Searched test/spec filenames and conventional `__tests__`, `test`, and `tests` locations; related tests are colocated Go `*_test.go` files. No separate matching test/spec suite or completion test suite was located. Negative search results describe this checkout, not a guarantee against dynamically constructed coverage.

No source code or tests were added or changed. No test suite was run: execution is a later checkbox in this playbook, and this task only discovers related files. New-test verification, coverage/quality assessment, and `TEST_GAPS.md` remain pending. Validation for this documentation change consists of checking all mapped paths exist, the eight source paths are represented, and `git diff --check`. No images were associated with this task or analyzed.
