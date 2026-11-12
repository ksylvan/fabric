---
type: analysis
title: PR 2330 Hardcoded Secrets Review
created: 2026-11-11
tags:
  - code-review
  - security
  - secrets
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-sensitive-data-review]]'
  - '[[pr-2330-authentication-review]]'
---

# Hardcoded secrets review

Reviewed the fixed PR snapshot `de0b72766db0436b6c1887bb75768cb3960e67e5` to `fb296d490cca4fdebbcb78d5107241f134152d19`: 25 changed files, 320 added lines, three deleted lines. No apparent real hardcoded credential was introduced. This completes only the hardcoded-secrets checkbox; production code is unchanged. No images were supplied or analyzed.

## Assessment

| Requested category | Evidence and disposition |
| --- | --- |
| API keys | `internal/tools/keenable/keenable.go:47` creates an API-key setup question, `:72` reads the configured value, and `:86` assigns it to the request header. No production key literal is supplied. `internal/tools/keenable/keenable_test.go:51,56` uses the obvious synthetic value `secret`, checked against a loopback test server created at `:16`; this is a fixture, not evidence of a leaked provider key. |
| Passwords | No password literal introduced in the added code, localization, shell completions, README, or changelog. No login/password implementation added. |
| Private keys | No PEM/OpenSSH private-key header or private-key material identified in additions. |
| Connection strings | The fixed HTTPS API URLs at `internal/tools/keenable/keenable.go:22-23` contain no userinfo or embedded credential. No credential-bearing database URI or shared-access connection string identified in additions. |
| Tokens | No apparent provider-token literal, JWT, or other embedded bearer token identified in additions. Optional request authentication uses the configured API key. |

## Method and candidate triage

Used the existing code graph to locate `NewClient` and `TestSearchWithKeyUsesKeyedEndpoint`, then read the keyed test through `get_code_snippet`. Reviewed the fixed diff and credential handling, including non-code additions. Relevant working-tree implementation, locale, documentation, and completion files match the fixed head. No repository AGENTS.md or CLAUDE.md files were present; supplied session instructions apply.

Neither `gitleaks` nor `trufflehog` was available on PATH. Added an offline, review-only Python heuristic scan in the external Auto Run Working folder, reusing the fixed-snapshot/subprocess pattern from `validate_dependency_review.py`. It checks private-key headers, common provider token shapes, JWT shapes, credential-bearing URIs/shared-access keys, quoted credential assignments, literal authorization headers, and long high-entropy literals. Reports contain paths, line numbers, rule labels, and whether a line was added; matching values are not written to the report. No credentials were submitted to any service for validation.

The script scans complete head versions of all 25 changed files and marks candidates against the exact 320 added lines. Of 585 heuristic candidates, four occur in additions:

| Location | Rule | Manual disposition |
| --- | --- | --- |
| `internal/core/plugin_registry.go:47` | High-entropy literal | Go import path for the Keenable package. |
| `internal/tools/keenable/keenable.go:14` | High-entropy literal | Go import path for i18n. |
| `internal/tools/keenable/keenable.go:15` | High-entropy literal | Go import path for plugins. |
| `internal/tools/keenable/keenable_test.go:56` | Literal secret assignment | Synthetic API-key fixture used by the local HTTP test. |

The remaining 581 candidates are unchanged context, including localization identifiers and credential-related messages. They are outside this PR-additions assessment and were not individually adjudicated. Entropy and generic-name rules are deliberately noisy; candidate counts are not vulnerability counts. Manual review also covers the repeated test assertion at `:51`, which the literal-header rule does not match because it is an equality assertion.

## Validation and limits

Created `Working/scan_pr_secrets.py` and redacted `Working/hardcoded-secrets-scan.json` outside the repository. Seven positive controls cover each detector rule, and three negative controls cover an ordinary endpoint, a configured-value assignment, and a locale label. All ten passed. The script also asserts the 25-file/320-added-line scope. Reproduce from the repository root:

```sh
python3 /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/scan_pr_secrets.py
TMPDIR=/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/tmp go test -race -timeout 90s ./internal/tools/keenable
```

The complete Keenable package suite passed with the race detector (cached). No new executable application behavior warranted a production test; new checks test the review scanner itself. The full repository suite was not run. Reproduction of the scanner requires the external Working folder, as with prior review tasks.

This is a scoped heuristic and manual review, not an exhaustive repository/history secret audit or proof that all arbitrary token formats would be detected. Local `.env` files, user credentials, provider account validity, remote branches, and removed historical credentials were not investigated. No live provider requests or actual credentials were used. Existing redirect/error/logging disclosure findings remain separate from hardcoded secrets; see [[pr-2330-sensitive-data-review]]. Environment-variable verification, auth/authz checks, and the consolidated security report remain for later tasks.
