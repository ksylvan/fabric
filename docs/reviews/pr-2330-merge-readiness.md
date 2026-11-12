---
type: analysis
title: PR 2330 Merge Readiness Verdict
created: 2026-11-11
tags:
  - code-review
  - keenable
related:
  - '[[pr-2330-summary-inputs]]'
  - '[[REVIEW_SCOPE]]'
  - '[[CODE_ISSUES]]'
  - '[[SECURITY_ISSUES]]'
  - '[[TEST_GAPS]]'
---

# Merge Readiness Verdict

**Verdict: REQUEST CHANGES.** PR 2330 adds optional-key Keenable search with additive CLI, configuration, and localization integration. The reviewed snapshot is not ready to merge because the new HTTP client accepts unrestricted response bodies and follows redirects that can disclose credentials and private queries. These findings warrant fixes before approval even though the recorded tests pass and no Critical or High application vulnerability was confirmed.

This decision applies to base `de0b72766db0436b6c1887bb75768cb3960e67e5` and head `fb296d490cca4fdebbcb78d5107241f134152d19`. It synthesizes the existing review reports; it does not revalidate a later upstream revision.

## Required Before Merge

- **Bound response consumption** at `internal/tools/keenable/keenable.go:94` and `:100`. Code finding M1 and security finding SEC-M2 describe the same defect: error reads and successful JSON decoding have no byte budget. Recorded synthetic probes preserve a full 2 MiB error body and oversized successful descriptions; no destructive memory-exhaustion test was attempted. Enforce a successful-response budget with overflow detection before decoding and a smaller error-diagnostic budget, preserving status and indicating truncation. Add deterministic tests for both paths, including an oversized response with a valid JSON prefix. Security also recommends limiting result count and field lengths before formatting.
- **Constrain redirects** at `internal/tools/keenable/keenable.go:27`, `:83`, and `:87` (SEC-M1). Recorded synthetic probes show the default client forwards the custom API-key header to unrelated hosts and HTTP destinations; 307/308 also replay the query body. Exploitation depends on an unsafe redirect from the trusted endpoint or intermediary; no live disclosure was observed. Reject redirects unless required, or allow only the original HTTPS scheme, host, and effective port with a finite hop limit. Add keyed/keyless regression tests proving disallowed destinations receive no request; stripping the key alone leaves the query exposed.

These are merge-readiness decisions, not severity upgrades: M1 remains Major in the code report, and SEC-M1/SEC-M2 remain Medium in the security report. Avoid double-counting M1 and SEC-M2 in the later issue-count task.

## Supporting Findings

Raw diagnostics, lost body-read causes, and fatal assertions in handler goroutines remain actionable feedback in [[CODE_ISSUES]] and [[SECURITY_ISSUES]]. Existing shared dotenv/logging behavior, optional audit metadata, and maintenance suggestions do not independently block this feature. Historical dependency warnings need separate applicability/remediation triage; their package ratings do not establish new Critical PR vulnerabilities.

[[TEST_GAPS]] documents missing transport/decode assertions and registry-backed CLI/configuration coverage, along with boundary and output cases. Add targeted regression evidence for the required HTTP fixes and strengthen these high-priority integration checks. Missing coverage alone is not proof of a production defect.

## Validation and Scope Limits

The recorded full Go suite passed (46 packages) and scoped race run passed (28 packages). Those are prior results, not fresh execution for this decision. Passing review-only characterization probes demonstrate the current flaws rather than their remediation. Live service behavior, real credentials, remote deployment, and dependency advisories were not revalidated.

For this task, reread the consolidated input reports, checked the decision against their stated evidence and severity, verified the report copies match their Auto Run originals, and checked documentation whitespace. No source behavior changed; no tests were added or rerun. Zero images were associated with this task or analyzed.

Only the merge-readiness assessment is completed here. Combined severity counts, feedback ranking, the final summary, and an optional PR comment remain separate unchecked tasks. No GitHub review or comment was posted.
