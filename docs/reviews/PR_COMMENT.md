---
type: report
title: Fabric PR 2107 Concise PR Review Comment
created: 2026-10-03
tags:
  - code-review
  - security
  - pr-comment
related:
  - '[[REVIEW_SUMMARY]]'
  - '[[pr-2107-merge-readiness]]'
---

## Code Review Summary

**Verdict: Request Changes**  
Reviewed head: `f32656b529a1eb1c678c793703fe49a0755cd439`  
PR: https://github.com/danielmiessler/Fabric/pull/2107

The patch improves lexical path validation and suppresses generated-file application for current REST callers. Two inherited defects leave the stated security remediation incomplete and require fixes before approval.

### Key Findings

- **H01 — shell interpolation** (`internal/plugins/template/extension_executor.go:52,78,83,86`): Values inserted inside double-quoted template placeholders remain executable through `sh -c`; command substitution/backticks were demonstrated. Exploitation requires control of a value used by a vulnerable installed extension; remote end-to-end reachability was not demonstrated.
- **H02 / Code M1 — symlink containment escape** (`internal/domain/file_manager.go:157-169`): Lexical locality checks do not stop intermediate/final symlinks from directing writes outside the project. Both cases overwrote synthetic outside sentinels. This requires an existing symlink and the CLI-like application branch; current REST chat skips application.
- **Missing boundary regressions**: The patch adds three writer rejection inputs, but no committed parser locality or coding-pattern channel-gate regression. Passing suites do not establish the missing security guarantees.

### Action Required

- [ ] Replace shell-source interpolation with a separate executable and literal arguments; migrate affected templates and add substitution, quoted/numbered-placeholder and literal-value regressions. Do not parse arguments with `strings.Fields`; SQL extensions also need parameter binding.
- [ ] Enforce project-root containment throughout resolution, including concurrent replacement. Add create/update tests for intermediate/final links that preserve outside sentinels and allow valid nested paths; resolve-then-write checks alone leave a race.
- [ ] Commit parser, writer and REST/CLI channel-gate regressions using existing fixtures; verify both streaming modes, file effects and saved-response behavior. Rerun relevant suites/platform checks and document API no-write and relative-path behavior.

<details>
<summary>Additional review context</summary>

The full review reconciles **26 application/code issues: 0 Critical, 8 Major, 14 Minor and 4 Suggestions**, plus 18 separate test-work entries and one development-dependency advisory with unconfirmed runtime applicability. Only H01/H02 are designated merge blockers.

Unsafe assistant HTML rendering (H03) is a separate high-priority follow-up outside this patch; markup retention was demonstrated, not browser execution or cross-user exploitation. Deployment-dependent findings and dependency applicability need assessment against the actual environment.

Previously recorded runs passed: Go (35 packages; 20 without tests) and web (36 tests in 7 files). This comment preparation did not rerun suites, refresh advisories or perform deployed end-to-end testing.

The patch follows existing modules, validates parser and writer paths, and preserves wrapped filesystem errors. See [[REVIEW_SUMMARY]] for the complete findings, recommendations and evidence limits. Reassess approval after remediation and regression evidence.

</details>

---
*Automated review by Maestro — prepared comment; not posted to GitHub.*
