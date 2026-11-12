---
type: analysis
title: PR 2330 Prioritized Review Feedback
created: 2026-11-11
tags:
  - code-review
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[CODE_ISSUES]]'
  - '[[SECURITY_ISSUES]]'
  - '[[TEST_GAPS]]'
  - '[[pr-2330-issue-counts]]'
  - '[[pr-2330-merge-readiness]]'
---

# Prioritized Review Feedback

This ranking applies to PR 2330's fixed base `de0b72766db0436b6c1887bb75768cb3960e67e5` and head `fb296d490cca4fdebbcb78d5107241f134152d19`. It orders the existing findings for later summary drafting; it neither implements fixes nor creates the final summary. **REQUEST CHANGES** remains the verdict for two Major merge prerequisites. The inventory remains **0 Critical, 2 Major, 6 Minor, 4 Suggestions (12 total)** under the explicit mapping in [[pr-2330-issue-counts]]. Original security severities remain unchanged.

## 1. Security Critical Issues

None confirmed in the scoped PR. Historical dependency package ratings are separate maintenance evidence, not confirmed Critical PR vulnerabilities.

## 2. Correctness Critical Issues

None confirmed in the scoped PR.

## 3. Security High-Risk Issues

None confirmed in the scoped PR. The two security Medium findings below must not be relabeled High merely because they block approval.

## 4. Major Merge Prerequisites

Place the security-specific redirect finding first, then the overlapping correctness/resource-consumption finding. Both require fixes and regression evidence before merge; order within this tier does not make either optional.

| Rank | Source identifier(s) | Location at fixed head | Required action and acceptance evidence |
| --- | --- | --- | --- |
| 1 | SEC-M1 | `internal/tools/keenable/keenable.go:27`, `:83`, `:87` | Reject redirects unless required; otherwise restrict to the original HTTPS scheme, host, and effective port with a finite hop limit. Test keyed/keyless requests and 301/302/303/307/308, cross-host destinations, and HTTP downgrade. Rejected destinations must receive no request. Stripping only the key leaves replayed queries exposed. Source severity: Medium security; summary: Major. Exposure requires an unsafe redirect from the trusted endpoint/intermediary; no live disclosure was observed. |
| 2 | Code M1 + SEC-M2 | `internal/tools/keenable/keenable.go:94`, `:100`, `:106`, `:117` | Enforce a success-response byte budget with overflow detection before decoding and a smaller diagnostic budget preserving status and indicating truncation. Reject oversized bodies even when they start with valid JSON. Bound result count and useful field lengths before formatting. Test oversized success/error bodies and boundary behavior deterministically. Source severity: Major code / Medium security; one underlying issue. Synthetic bounded probes establish missing limits, not an observed out-of-memory crash. |

## 5. Test Coverage and Quality Recommendations

The labels below are local traceability labels: M1–M10 follow the ten Missing Tests rows, I1–I4 the four Inadequate Tests rows, and Q1–Q5 the five Test Quality Issues rows in [[TEST_GAPS]], in source order. They are not new defect IDs or additional issue counts. Priorities retain that report's labels. Regression tests for ranks 1–2 belong to those fixes and precede this backlog.

| Order | Source row(s) | Priority | Action |
| --- | --- | --- | --- |
| T1 | Missing M1 | High | Add deterministic transport errors, malformed JSON, empty HTTP 200 bodies, and non-array results; assert relevant error detail and empty output. Reuse the existing server helper and restore shared transport state. |
| T2 | Missing M2 | High | Exercise flag/default/missing-value parsing and registry-backed keyless dispatch, query forwarding, standalone output, appended chat results, search/output errors, empty-query bypass, and tool ordering/early returns. |
| T3 | Missing M3; Inadequate I4; Quality Q3 | High / Medium / Medium | Assert registry initialization, environment key loading, absent/keyed operation and persistence, optional-plugin serialization, and preservation of other settings. Read saved content using the existing isolated filesystem pattern and `t.TempDir()` instead of shared `os.TempDir()`. |
| T4 | Missing M5; Inadequate I1 | Medium / Medium | Assert whitespace-key/public routing, trimmed-key/keyed routing, invalid selected URLs, exact Markdown, and 999/1,000/1,001-rune multibyte snippet boundaries with valid UTF-8 and correct ellipsis. Keep expected boundaries independent of the production constant. |
| T5 | Missing M6 | Medium | Cover empty/missing/null results, all URL-less entries, URL-bearing entries without text, short fallback snippets, and long descriptions under the current contract. Update oversized-description expectations when implementing rank 2. |
| T6 | Missing M4, M7, M8 | Medium | Verify constructor identity/prefix/question/optional metadata, optional-tools setup membership, active-locale help mapping, optional marker, and error substitutions. Control locale/setup I/O and check for formatting artifacts. |
| T7 | Missing M9 | Medium | Test Bash/Zsh/Fish option suggestions and query-argument consumption, including Bash's simple-argument branch. Syntax checks alone do not establish completion behavior. |
| T8 | Quality Q1 | Medium | Move fatal handler assertions into the test goroutine via captured observations, or use nonfatal checks with explicit error responses. Same finding as Code m3 at rank 7 below; do not count twice. |
| T9 | Missing M10; Inadequate I2, I3 | Low | Assert content type, quotes/newlines/Unicode round-trip, actual absent-key header omission, keyed method/body and empty output, HTTP 429 status/detail with empty output, and a non-200 empty-body case. |
| T10 | Quality Q2, Q4, Q5 | Low | Keep endpoint/argument/locale mutations serial; add helper attribution and independent named output cases; isolate output files in `t.TempDir()`; guard prerequisites before pointer dereferences. Q2 overlaps Code S1/S3 below; fixed filenames and parser assertion weaknesses predate the PR. |

This maps all 19 source rows once while combining related work. Do not manufacture a JSON-marshal failure for the current string/integer payload, introduce a shared fixture framework for small inline fixtures, or equate High test priority with Critical defect severity. Review-only overlays provide additional historical evidence but do not add these cases to the committed PR tests.

## 6. Minor Issues and Suggestions

Ranks 3–8 are Minor in the summary; ranks 9–12 are Suggestions. Security Low observations retain their source scope, and SEC-L4 remains optional despite its Minor inventory bucket. These entries do not independently block merge.

| Rank | Source identifier(s) | Location at fixed head | Recommended feedback |
| --- | --- | --- | --- |
| 3 | Code m1 + SEC-L1 | `internal/tools/keenable/keenable.go:94–95` | Return selected safe bounded diagnostics or a generic status error, redact known credentials, avoid private query text, and encode control characters. Size caps alone do not prevent conditional echoed-data disclosure or forged log lines. One underlying finding. |
| 4 | SEC-L2 | `internal/plugins/plugin.go:198`; registration at `internal/core/plugin_registry.go:181` | Harden the shared dotenv serializer with CR/LF rejection or the existing `godotenv.Marshal` quoting pattern and round-trip tests. Pre-existing local configuration behavior; requires influence over the saved key, with no demonstrated shell execution. |
| 5 | SEC-L3 | `internal/cli/tools.go:132`, `internal/cli/chat.go:20`, `internal/core/chatter.go:102`, `internal/log/log.go:28` | Document Wire logging privacy costs and prefer metadata or deliberate redaction. Pre-existing opt-in logger behavior; logging defaults to Off. |
| 6 | Code m2 | `internal/tools/keenable/keenable.go:94` | Preserve HTTP status and safe partial diagnostics while wrapping body-read failures; distinguish intentional truncation from transport errors. Can be addressed with rank 2 without merging the distinct findings. |
| 7 | Code m3 | `internal/tools/keenable/keenable_test.go:25–32`, `:50–51` | Keep fatal assertions in the test goroutine; execute alongside T8. Passing handlers do not exercise the problematic failure path. |
| 8 | SEC-L4 | `internal/tools/keenable/keenable.go:87–102`; `internal/cli/tools.go:129` | Only if operational requirements justify it, reuse logging for service/status/outcome/duration/correlation metadata and arrange collection. Exclude keys, queries, bodies, and private URLs. Optional hardening; no persistent audit requirement established. |
| 9 | Code S1 | `internal/tools/keenable/keenable_test.go:13` | Add `t.Helper()` to `withServer`; coordinate with T10. |
| 10 | Code S4 | `internal/tools/keenable/keenable.go:60`; `internal/cli/flags.go:76` | Clarify constructor use, empty-result success, and omitted/missing/explicitly empty flag-value semantics. No contrary service contract establishes these as defects. |
| 11 | Code S3 | `internal/tools/keenable/keenable.go:20`, `:27` | Move mutable endpoints/client settings into instances if parallel tests or independent configurations are needed. No current race demonstrated; preserve serial cleanup meanwhile. |
| 12 | Code S2 | `internal/cli/tools.go:126` | If search routing grows, compose the existing search/append/output-or-chat flow while keeping provider configuration policy at call sites. The current branch does not require a broad refactor. |

## Separate Maintenance and Design Context

Retain the dependency warning inventory in [[SECURITY_ISSUES]] for separate applicability and supported-patch triage; captured scanner labels are historical and were not refreshed here. Conditional unsafe-renderer behavior, untrusted model input, external query/model destinations, and plaintext credential-storage limits remain qualified design context, not added confirmed defects. Reuse the existing renderer/sanitizer and shared configuration/logging infrastructure rather than adding parallel helpers.

## Validation and Limits

Read the consolidated code, security, test, scope, count, and verdict reports; verified all six Auto Run inputs match tracked copies. Reconciled the 12 ranked entries to all 14 source code/security IDs with two merged duplicate pairs, and mapped all 19 test rows exactly once. Checked report-copy equality and documentation whitespace. No application behavior changed; no application tests were added or rerun for this documentation-only ranking. Prior passing suites and characterization probes do not remediate the findings or establish live-service compatibility. Zero images were associated with this task or analyzed.

Only the ranking checkbox is completed. Final summary drafting, optional PR comment, and final completeness checking remain subsequent tasks. The Auto Run artifact `FEEDBACK_RANKING.md` has an identical tracked copy at `docs/reviews/pr-2330-feedback-ranking.md`.
