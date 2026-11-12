---
type: analysis
title: PR 2330 Deduplicated Issue Counts
created: 2026-11-11
tags:
  - code-review
  - keenable
related:
  - '[[CODE_ISSUES]]'
  - '[[SECURITY_ISSUES]]'
  - '[[TEST_GAPS]]'
  - '[[pr-2330-merge-readiness]]'
---

# Deduplicated Issue Counts

For PR 2330, fixed base `de0b72766db0436b6c1887bb75768cb3960e67e5` and head `fb296d490cca4fdebbcb78d5107241f134152d19`, the consolidated code/security inventory contains **12 distinct findings and recommendations**. This inventory includes explicitly identified shared hardening observations; it is not a count of newly introduced production vulnerabilities.

| Summary severity | Count |
| --- | ---: |
| Critical | 0 |
| Major | 2 |
| Minor | 6 |
| Suggestions | 4 |
| **Total** | **12** |

## Counting Policy

Retain the code report's Major/Minor/Suggestion classifications. Place redirect finding SEC-M1 in Major for this summary because the merge-readiness assessment requires its remediation; its original security severity remains Medium. Count SEC-M2 with code M1 once as Major, retaining its Medium security label. Place the four security Low findings in the summary's Minor bucket, merging SEC-L1 with code m1. This is an explicit summary grouping, not an upgrade of their security severity or a requirement to fix all Low observations before merge. SEC-L4 remains optional operational hardening despite its placement under Minor to preserve the source report's Low inventory.

## Count Reconciliation

Rows below are an inventory in source-report order, not the feedback ranking reserved for the next task.

| Source identifier(s) | Finding | Summary severity | Source classification / scope |
| --- | --- | --- | --- |
| Code M1 + SEC-M2 | Unbounded error reads and successful-response allocation | Major | Major code / Medium security; one issue |
| Code m1 + SEC-L1 | Raw diagnostics retain echoed private data and permit forged log lines | Minor | Minor code / Low security; one issue |
| Code m2 | Error-body read failures lose their cause | Minor | Minor code |
| Code m3 | Fatal assertions run in HTTP-handler goroutines | Minor | Minor code; also repeated in the test report |
| Code S1 | Add `t.Helper()` to the server helper | Suggestion | Optional code improvement |
| Code S2 | Reuse search dispatch flow if routing grows | Suggestion | Optional code improvement |
| Code S3 | Move mutable endpoint/client settings to instances if needed | Suggestion | Optional code improvement; no current race demonstrated |
| Code S4 | Clarify constructor, empty-result, and flag-value contracts | Suggestion | Optional documentation improvement |
| SEC-M1 | Redirects forward API keys and can replay private queries | Major | Medium security; merge prerequisite per verdict |
| SEC-L2 | Shared dotenv serialization permits injected assignments | Minor | Low; pre-existing local configuration hardening |
| SEC-L3 | Opt-in Wire logging retains private search content | Minor | Low; pre-existing shared logger hardening |
| SEC-L4 | No dedicated persistent search-rejection audit trail | Minor | Low; optional operational hardening, no audit requirement established |

Arithmetic: 8 code entries + 6 security entries - 2 duplicated entries = 12. Major: 1 code + 2 Medium security - 1 overlap = 2. Minor: 3 code + 4 Low security - 1 overlap = 6. Suggestions: 4 code entries. Neither report identifies a confirmed Critical issue. The existing verdict remains **REQUEST CHANGES** for the two Major entries; counting does not change the verdict.

## Test Recommendations and Other Exclusions

[[TEST_GAPS]] is a separate coverage/quality inventory, not an additional set of confirmed production defects: 10 missing-test rows, 4 inadequate-test rows, and 5 test-quality rows. These 19 rows overlap internally and with the code inventory (including fatal handler assertions, helper attribution, global state, and persistence coverage). Do not add 19 to the issue total or equate High test priority with Critical severity. Keep all those recommendations available to the later ranking and summary tasks; omission from the numeric defect inventory does not discard them. Regression tests needed for the two Major findings belong to those findings rather than additional issue counts.

Existing dependency scanner/package ratings, conditional unsafe-renderer observations, and general privacy/design limits in the security best-practices section stay outside these totals. Their applicability remains separately qualified in [[SECURITY_ISSUES]]. In particular, the historical npm Critical package rating does not create a confirmed Critical PR vulnerability. No advisory was revalidated for this task.

## Validation and Limits

Read the four consolidated prerequisite reports and prior merge verdict. Verified all input originals match their tracked copies, all 14 source finding identifiers occur exactly once in the reconciliation table after merging two duplicate pairs, and the severity totals reconcile to 12. Checked report-copy equality and documentation whitespace. No source behavior changed; no application tests were added or rerun. Prior passing test results remain historical evidence, and characterization probes do not remediate findings. Zero images were associated with this task or analyzed.

Only the severity-count checkbox is completed. Feedback ranking, final summary, optional PR comment, and final completeness review remain subsequent tasks. The Auto Run artifact `ISSUE_COUNTS.md` has an identical tracked copy at `docs/reviews/pr-2330-issue-counts.md`.
