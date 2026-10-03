---
type: analysis
title: Fabric PR 2107 Deduplicated Issue Counts
created: 2026-10-03
tags:
  - code-review
  - severity-counts
related:
  - '[[pr-2107-summary-input-review]]'
  - '[[pr-2107-merge-readiness]]'
  - '[[CODE_ISSUES]]'
  - '[[pr-2107-security-issues]]'
  - '[[TEST_GAPS]]'
  - '[[5_SUMMARIZE]]'
---

# Issue Counts

The deduplicated application/code inventory contains **26 issues: 0 Critical, 8 Major, 14 Minor, and 4 Suggestions**. Test work and an unconfirmed development-dependency advisory are reported separately below. These counts describe retained review findings at original PR head `f32656b529a1eb1c678c793703fe49a0755cd439`, not newly introduced regressions or remediation.

## Severity mapping and ledger

Security High and conditional Medium map to Major; Low maps to Minor. Code severity labels are retained. No High finding is promoted to Critical merely because it blocks this review. Severity and merge disposition are different: H01 and H02 remain the two required fixes in [[pr-2107-merge-readiness]], despite zero Critical-severity findings. H03 is a separate high-priority follow-up, and Medium findings retain their deployment conditions.

| Summary severity | Count | Source finding IDs |
| --- | ---: | --- |
| Critical | 0 | None established |
| Major | 8 | Security H01, H02, H03; M01, M02, M03, M04, M05 (conditional) |
| Minor | 14 | Security L01–L10; Code m1–m4 |
| Suggestions | 4 | Code S3, S4, S5, S6 |
| Total application/code issues | 26 | 18 security findings + 8 additional code findings |

Deduplication accounts for every code entry:

- Code M1 is Security H02 (symlink containment), counted once as Major.
- Code S1 is Security L01 (channel-based authority), counted once as Minor.
- Code S2 is a regression-coverage umbrella; its requested test work is represented by the separate [[TEST_GAPS]] inventory rather than another application issue.
- Code m1 already consolidates partial writes, warning-only outcomes and payload loss; retain it as one issue, as the source report does.
- Code S5 includes API/release guidance and reconciling the shell-remediation claim. Count it once as a communication suggestion; it is not another shell vulnerability.

Reconciliation: 11 code entries + 19 security entries − 2 duplicated application findings − 1 coverage umbrella − 1 separately tracked advisory = 26 application/code issues.

## Test-work inventory

Test priorities rank work and are not vulnerability severities. Preserve this separate inventory rather than converting missing tests into additional confirmed defects. Regression work can address an application finding already counted above (for example, test M3 and Security H02).

| Test priority | Count | TEST_GAPS finding IDs |
| --- | ---: | --- |
| High | 5 | M1, M2, M3, M4, I1 |
| Medium | 9 | M5, M6, M7, I2, I3, Q1, Q2, Q3, Q4 |
| Low | 4 | I4, I5, Q5, Q6 |
| Total test-work entries | 18 | 7 missing behaviors + 5 assertion gaps + 6 quality concerns |

Some entries share fixtures or contracts (M2/I2, I5/Q1), so 18 is a source-entry inventory, not a promise of 18 independent fixes or tests. Code S2 is covered here. Do not add these rows to the application severity table and call the result a vulnerability count.

## Dependency uncertainty

Security D01 is **one High-rated development-dependency advisory**, with unconfirmed Fabric runtime applicability. Track it separately, not as a confirmed application Major or merge blocker. The retained Go scan has 11 advisory IDs with differing match evidence; applicability is unresolved and they are not assigned additional confirmed-issue severities here. This task does not refresh advisory data.

## Evidence and validation

Reviewed the four required Auto Run reports and existing verdict. Verified source finding IDs, category counts, report snapshot hashes against [[pr-2107-summary-input-review]], and existing repository mirror equality. No production code or tests changed; no suites were rerun. Prior Go/frontend suite results remain those recorded in [[TEST_GAPS]]. No task images supplied or analyzed (0).

Only the count-issues checkbox in Document 5 is completed by this task. Ranking, the full summary, PR comment and final completeness review remain for subsequent runs.
