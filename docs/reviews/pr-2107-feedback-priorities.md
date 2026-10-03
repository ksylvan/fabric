---
type: analysis
title: Fabric PR 2107 Ranked Review Feedback
created: 2026-10-03
tags:
  - code-review
  - prioritization
related:
  - '[[pr-2107-issue-counts]]'
  - '[[pr-2107-merge-readiness]]'
  - '[[pr-2107-summary-input-review]]'
  - '[[CODE_ISSUES]]'
  - '[[pr-2107-security-issues]]'
  - '[[TEST_GAPS]]'
  - '[[5_SUMMARIZE]]'
---

# Ranked Review Feedback

**Request Changes remains the verdict.** Address H01 and H02 with secure-behavior regressions before approval; track H03 as a separate high-priority remediation. This ordered handoff covers all retained findings at original PR head `f32656b529a1eb1c678c793703fe49a0755cd439`. It changes neither severity nor merge disposition established in [[pr-2107-issue-counts]] and [[pr-2107-merge-readiness]].

## Ordering rules

The requested sequence is security Critical, correctness Critical, security High, code Major, test gaps, then Minor/Suggestions. Neither Critical category has an established finding. Code M1 is Security H02 and appears once in the High tier; there is no additional independent code Major. Conditional Medium security issues map to summary Major and follow High security before test work. D01 follows confirmed High application findings as an applicability investigation, not a confirmed application defect or blocker.

Within each tier, the order favors demonstrated impact, relevance to the changed boundary, and useful implementation dependencies. Test priorities are work priorities, not vulnerability severities. The ordinal sequence guides feedback presentation; implement related tests with their fixes, even though the test inventory appears later.

## 1–2: Critical security and correctness tiers

No confirmed Critical findings in either tier. No ordinal ranks are assigned to empty categories.

## 3: High security and advisory investigation

| Rank | Source ID / severity | Feedback and next action | Disposition and evidence limit |
| ---: | --- | --- | --- |
| 1 | Security H01 / High | Replace shell-source interpolation in the existing extension executor with executable/literal arguments; migrate templates deliberately and test double-quoted placeholders, substitutions, numbered arguments, separators, spaces and newlines. | Required fix for stated remediation. Safe probes demonstrated substitution; a vulnerable installed template and controllable value are prerequisites. Remote end-to-end reachability was not demonstrated. |
| 2 | Security H02 = Code M1 / High (summary Major) | Enforce project containment during filesystem resolution through rooted directory handles, including intermediate/final symlinks and concurrent replacement. Preserve outside sentinels in regressions and allow valid nested paths. | Required fix. Both symlink types redirected synthetic writes; current REST skips application. Resolve-then-write alone leaves a race. |
| 3 | Security H03 / High | Sanitize every assistant HTML-producing branch with explicit URL/attribute rules; render plain/Mermaid/fallback content as text. Validate streamed/saved messages and unsafe links in a browser; add compatible UI CSP. | Separate high-priority follow-up, inherited outside this patch. Unsafe markup retention was demonstrated; browser execution and cross-user exploitation were not. |
| 4 | Security D01 / High-rated advisory | Investigate development-tooling reachability; assess removing/replacing patch-package or constraining untrusted patterns. Update both supported lockfiles deliberately and rerun audits after selecting a fix. | One separate advisory, unconfirmed runtime applicability. Retained production-only npm scan had zero warnings. Avoid the suggested blind breaking downgrade. |

## 4: Remaining summary Major findings

All five entries retain their source Medium severity and deployment conditions. They are not additional unconditional merge blockers.

| Rank | Source ID | Feedback and next action | Condition |
| ---: | --- | --- | --- |
| 5 | Security M04 | Require trusted TLS termination and restricted backend access; validate remote provider/forwarding URLs with explicit cleartext opt-in while retaining local HTTP. | Untrusted-network HTTP exposure; deployed interception was not tested. |
| 6 | Security M01 | Generate high-entropy API keys, reject clearly weak configuration, and document rotation using existing auth facilities. | Guessable configured key; a one-character synthetic key was accepted. |
| 7 | Security M02 | Bound authentication/request attempts in existing middleware or a trusted proxy, including concurrency and trusted client identity. Coordinate with M01. | Exposed service without external controls; 100 wrong-key attempts returned 401 without throttling. |
| 8 | Security M03 | Specialize session storage to 0700 directories/0600 files, address existing files and define retention/backup controls; encrypt where required with separate key protection. | Sensitive plaintext history and traversable ancestors; private ancestors mitigate access. |
| 9 | Security M05 | Document full shared-key authority and use one instance/key per trust domain. Add principals, ownership and administrative scopes if multi-user isolation is required. | Mutually untrusted/read-only clients; shared authority is expected for a single owner. |

## 5: Test coverage and quality work

Prefix IDs with Test to distinguish them from code/security IDs. Extend the existing vendor fake and real temporary filesystem/storage fixtures; avoid mocking away the writer boundary. Code S2 is represented by this tier, not a separate application issue.

| Rank | Source ID / priority | Action and acceptance evidence |
| ---: | --- | --- |
| 10 | Test M1 / High | Cross nil/non-nil channels with streaming/non-streaming in coding-pattern Send; assert intended writes, untouched targets, assistant/session content, and another-pattern no-write control. Consume updates concurrently. |
| 11 | Test M2 / High | Reject absolute/traversal parser inputs; accept safe dotted/normalized local paths. Assert exact decoded changes/summary and no applicable changes for mixed invalid payloads. |
| 12 | Test I1 / High | Replace host-dependent rejection targets with isolated outside sentinels; assert validation errors and absence of file/directory mutation for create/update. |
| 13 | Test M3 / High | With the H02 fix, reject intermediate/final symlink escapes and preserve outside sentinels. Expect secure behavior; skip only when symlink creation is unavailable. |
| 14 | Test M4 / High | Exercise REST/CLI caller configuration with a fake vendor; assert REST content/completion without writes and intended CLI writes/output/session behavior in both streaming modes. |
| 15 | Test M5 / Medium | Cover direct-writer empty/update rejection and valid local edge paths with sentinel preservation; run Windows drive/UNC/reserved-name cases on Windows. |
| 16 | Test M6 / Medium | Exercise later-entry and deterministic filesystem failures; assert contextual errors, unrelated-file preservation and documented partial results. Define atomicity before expecting rollback. |
| 17 | Test M7 / Medium | Assert coding-pattern parse/application warning, returned-error and persisted summary/payload behavior, including partial writes where applicable. |
| 18 | Test I2 / Medium | Strengthen parser checks to ordered operation/path/content and summary equality; preserve no-marker summary and require no applicable changes on rejection. Share M2 fixtures. |
| 19 | Test Q4 / Medium | Extend the existing fake with an optional context-aware streaming callback; coordinate cancellation/backpressure without sleeps and assert bounded return. |
| 20 | Test Q3 / Medium | Collect stream updates concurrently with explicit completion/channel ownership; verify larger fixtures cannot deadlock on the old buffer capacity. |
| 21 | Test Q1 / Medium | Register stdout/pipe cleanup immediately, check capture errors, retain serial execution and drain larger output concurrently. Coordinate with I5. |
| 22 | Test Q2 / Medium | Explain the bounded timeout and guarantee cooperative cancellation; consider subprocess isolation for arbitrary deadlocks so workers cannot outlive fixtures. |
| 23 | Test I3 / Medium | Assert input/output/total usage 10/5/15 and defined update type/content/count/order with concurrent collection. |
| 24 | Test I4 / Low | Assert first-error identity; specify returned/update error precedence before adding that case and preserve bounded completion checks. |
| 25 | Test I5 / Low | Check stdout read/close errors and returned sessions; compare role/content sequences directly using a delimiter-containing fixture. Coordinate with Q1. |
| 26 | Test Q5 / Low | Use named independent writer subtests and fresh update state; reuse vendor/storage APIs and add a small fixture helper only where repetition warrants it. |
| 27 | Test Q6 / Low | Use existing sendFunc or a focused stream callback for meaningful forwarding/raw-mode/configuration contracts; do not equate aggregation tests with provider integration coverage. |

## 6: Minor findings and optional suggestions

Minor correctness issues lead this tier because they can lose or corrupt generated content. Low security items remain conditional hardening concerns.

| Rank | Source ID / summary severity | Feedback and next action |
| ---: | --- | --- |
| 28 | Code m4 / Minor | Distinguish absent/null content from explicit empty strings before producing FileChange; reject absent/null values without truncating destinations. |
| 29 | Code m3 / Minor | Decode untouched JSON first; repair only failed input with escape-aware logic. Require exact backslash/quote/control/Unicode round trips. |
| 30 | Code m2 / Minor | Use JSON decoding or correct string/escape tracking for array extraction; test brackets and escaped quotes inside valid content while preserving supported surrounding text. |
| 31 | Code m1 / Minor | Prevalidate batches, retain failed payloads and expose applied/failed entries through existing result/error flow. Document runtime partial writes and define command failure policy; atomicity requires separate staging/recovery design. |
| 32 | Security L01 = Code S1 / Minor | Separate trusted default-denied file-application authority from output transport and grant it at CLI boundaries. No present REST bypass was established. |
| 33 | Security L10 / Minor | Add secret metadata to settings, mask stored prompt defaults and read secret input without echo; retain environment injection. |
| 34 | Security L02 / Minor | Keep content capture explicitly opt-in; redact sensitive message data and restrict diagnostic permissions/retention while preserving metadata. |
| 35 | Security L08 / Minor | Omit/redact URL queries in the existing access logger and retain safe route/status metadata; use credential headers. |
| 36 | Security L03 / Minor | Mask every nonempty short vendor credential and test the existing masking contract. No anonymous disclosure was shown. |
| 37 | Security L04 / Minor | Offer confidential-output permissions in the existing writer with review before application; preserve restrictive update modes. Ordinary shared source files need not be confidential. |
| 38 | Security L06 / Minor | Centralize appropriate success/error browser headers; apply tested CSP at the actual UI origin and HSTS at HTTPS termination. This does not replace H03 sanitization. |
| 39 | Security L07 / Minor | Extend existing logging with bounded safe reason codes, request IDs and trusted client metadata; define collection/alerts without keys, bodies or raw model values. |
| 40 | Security L09 / Minor | Document environment key injection before launch or resolve the key after dotenv loading with flag precedence preserved. Non-loopback unkeyed startup already fails closed. |
| 41 | Security L05 / Minor | Set production GIN_MODE=release or centralize an explicit mode policy while preserving deliberate development diagnostics. |
| 42 | Code S5 / Suggestion | Document API no-write behavior, saved-response and relative-path changes. Reconcile the shell-remediation claim with delivered code; maintainers classify version impact. Address this communication alongside H01/H02. |
| 43 | Code S4 / Suggestion | Document parser validation as a writer precondition, or share existing record validation if direct callers are supported; locality is not complete record validation. |
| 44 | Code S3 / Suggestion | Let orchestration report applied results and route warnings through existing stderr/Quiet conventions; test machine-consumed output behavior. |
| 45 | Code S6 / Suggestion | Extract focused orchestration only if Send grows, composing the existing parser/writer and keeping trusted authority visible. Avoid duplicate path helpers/writers. |

## Dependency scan follow-up without assigned severity

The retained Go scan's 11 advisory IDs are applicability evidence, not 11 confirmed Fabric vulnerabilities. As part of rank 4's dependency investigation, reconcile Ollama client versus daemon relevance, gRPC package-only evidence, OpenPGP module-only evidence, and conflicting version metadata in [[pr-2107-dependencies]]. Preserve the raw advisory inventory there; no new severity or runtime blocker is assigned here. No live advisory refresh was performed.

## Reconciliation and validation

The 45 ranked entries comprise 26 deduplicated application/code issues, 18 test-work entries and one separate advisory. Application severity totals remain 0 Critical, 8 Major, 14 Minor and 4 Suggestions. Every security H/M/L/D ID, code M/m/S ID and test M/I/Q ID is represented, including aliases Code M1 → H02, Code S1 → L01 and Code S2 → test tier. Overlapping test entries remain source-entry counts, not a promise of independent fixes.

Validated source-report snapshot hashes, existing mirror equality, complete ID coverage, rank continuity, severity totals, Markdown tables and preservation of the task line and remaining checkboxes. This is documentation-only prioritization; no source/tests changed and suites were not rerun. Prior recorded Go/frontend results remain those in [[TEST_GAPS]]; missing guarantees are not established by passing suites. No task images supplied or analyzed (0).

Only Document 5's Rank issues checkbox is completed. Creating REVIEW_SUMMARY.md, the optional PR comment and final completeness review remain for later runs.
