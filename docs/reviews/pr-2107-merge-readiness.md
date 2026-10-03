---
type: analysis
title: Fabric PR 2107 Merge Readiness Verdict
created: 2026-10-03
tags:
  - code-review
  - merge-readiness
  - security
related:
  - '[[pr-2107-summary-input-review]]'
  - '[[REVIEW_SCOPE]]'
  - '[[CODE_ISSUES]]'
  - '[[pr-2107-security-issues]]'
  - '[[TEST_GAPS]]'
  - '[[5_SUMMARIZE]]'
---

# Merge Readiness Verdict

**Verdict: Request Changes.** The PR improves current REST write gating and lexical path validation, but unresolved shell interpolation and symlink containment defects prevent approval as the security remediation described in its title. This is a review recommendation based on the retained findings, not a submitted GitHub review or human sign-off.

## Changes required before approval

| Finding | Evidence and scope | Required change |
| --- | --- | --- |
| Security H01: residual shell injection | `internal/plugins/template/extension_executor.go:52,78,83,86`. Safe probes showed command substitution inside double-quoted placeholders. The original diff does not change the executor; execution precedes the new file-write gate. Exploitation requires control of a value used by a vulnerable installed extension; remote end-to-end reachability was not demonstrated. | Remove shell-source interpolation using separate executable/literal arguments in the existing executor, migrate affected templates deliberately, and add quoted-placeholder/substitution regressions. Reconcile the PR title/description with the actual delivered remediation. |
| Security H02 / Code M1: symlink containment escape | `internal/domain/file_manager.go:157-169`. Intermediate and final symlinks redirected writes to synthetic outside targets. This is one defect reported twice, predates the patch, and remains in the CLI-like application branch; current REST chat skips file application. | Enforce containment throughout path resolution using project-root directory handles, including symlinks and concurrent replacement. Add intermediate/final-link regressions that preserve outside sentinels and continue to permit valid nested paths. A separate resolve-then-write check is insufficient. |

Protect the changed boundary with committed regressions as described in [[TEST_GAPS]]: nil/non-nil channels crossed with streaming/non-streaming, REST/CLI caller behavior, parser absolute/traversal rejection and safe local edge paths, and rejection assertions proving no outside writes. Reuse the existing vendor fake and temporary filesystem fixtures. Retained probes characterize behavior but do not replace these expected-secure regressions.

## Other findings and disposition

Security H03 (unsafe assistant HTML rendering) warrants a separate high-priority remediation with sanitization and browser validation; unsafe markup retention was demonstrated, browser execution was not. It is inherited and outside this three-file patch. D01 is a retained development-only dependency advisory with unconfirmed runtime applicability; it is not established as a runtime merge blocker. The Go advisory matches also require applicability review before treating them as Fabric vulnerabilities.

Deployment-dependent Medium security findings and Low hardening concerns should retain their conditional scope. The four Minor code findings (partial/error outcomes, bracket extraction, escape corruption, missing/null content) and remaining maintenance/testing suggestions remain actionable follow-ups. No Critical vulnerability or new security regression was established. The request-changes recommendation rests on incomplete remediation of the stated security goals, not on relabeling inherited or conditional findings as new regressions.

## Evidence and validation

Assessment uses the four snapshot reports verified against the SHA-256 hashes in [[pr-2107-summary-input-review]], pinned to PR base `ddf1aab968caa9adf90137a6536c768d02b237d6` and head `f32656b529a1eb1c678c793703fe49a0755cd439`. Local changed source/tests and the extension executor still match that head. No live advisory refresh or new exploit investigation was performed.

Prior recorded suites passed: Go full suite (35 packages; 20 without tests) and frontend suite (36 tests in 7 files). Passing suites do not establish coverage of the missing security guarantees. This task changes documentation only; no production code/tests changed and no suites were rerun. Report structure, input hashes, checkbox preservation, and diff whitespace were checked. No task images supplied or analyzed (0).

Only Document 5's assess-merge-readiness checkbox is completed. Combined severity counts, ranked feedback, full `REVIEW_SUMMARY.md`, and optional `PR_COMMENT.md` remain for subsequent tasks.
