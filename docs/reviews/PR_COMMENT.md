---
type: report
title: PR 2330 Concise Review Comment
created: 2026-11-12
tags:
  - code-review
  - keenable
related:
  - '[[REVIEW_SUMMARY]]'
  - '[[CODE_ISSUES]]'
  - '[[SECURITY_ISSUES]]'
  - '[[TEST_GAPS]]'
---

## Code Review Summary

**Verdict: Request Changes** — PR https://github.com/danielmiessler/Fabric/pull/2330

Reviewed fixed head `fb296d490cca4fdebbcb78d5107241f134152d19` against base `de0b72766db0436b6c1887bb75768cb3960e67e5`. The optional-key integration follows existing provider patterns and includes useful local HTTP tests. Two issues need fixes and regression tests before merge.

### Key Findings

- **Unsafe redirects (Major; SEC-M1, source Medium)** — `internal/tools/keenable/keenable.go:27,83,87`: the default client forwards `X-API-Key` to unrelated hosts or HTTP destinations; 307/308 also replay the private query body. Exposure requires an unsafe redirect from the trusted endpoint or intermediary. Reject redirects, or enforce the original HTTPS scheme, host, effective port, and a finite hop limit. Stripping only the key still exposes replayed queries.
- **Unrestricted response consumption (Major; code M1 / SEC-M2, source Medium)** — `internal/tools/keenable/keenable.go:94,100,106,117`: error reads and success decoding lack byte budgets, while result counts and fields are unbounded. Enforce limits before decoding/formatting, detect overflow including valid JSON followed by excess bytes, and cap diagnostic output. Timeout and outbound `max_results` do not bound inbound allocation.
- **Follow-up quality and coverage** — return safe diagnostics without echoed credentials/private queries, preserve transport-read failures, and move fatal assertions out of HTTP-handler goroutines. Strengthen CLI/configuration, failure-path, formatting, and completion tests.

### Action Required

- [ ] Reject unsafe redirects before their destinations receive requests; test keyed/keyless 301/302/303/307/308, cross-host redirects, and HTTP downgrade.
- [ ] Bound success/error bodies, result counts, and useful field lengths; test exact boundaries, overflow, valid-JSON-prefix excess bytes, and transport-read failures.
- [ ] Rerun targeted/full Go suites and appropriate race checks after fixes, then reassess the verdict against the revised head.

<details>
<summary>Review scope, secondary findings, and validation limits</summary>

The deduplicated inventory is **0 Critical, 2 Major, 6 Minor, 4 Suggestions**. Security Medium maps to Major and Low to Minor in this summary. Code M1 / SEC-M2 and code m1 / SEC-L1 each count once; the 19 test recommendation rows are a separate inventory.

Nonblocking findings cover unsafe raw error diagnostics, pre-existing shared dotenv serialization and opt-in Wire logging, discarded error-body read failures, handler assertions, and optional rejection-audit metadata. Suggestions cover helper attribution, contract documentation, instance-scoped settings if needed, and future search-flow reuse. Shared hardening and optional audit work retain their original qualifications; historical dependency warnings require separate applicability checks.

Prior runs on November 11, 2026 passed the full Go suite and targeted race suite, including all three new client tests. Five full-suite skips included the broad CLI test (`-t` flag collision); no executed broad CLI integration evidence follows from that run. Synthetic redirect/resource probes characterize current flaws and are not committed regression tests or evidence of remediation. No live disclosure or out-of-memory attack was observed, and live-service compatibility was not established.

See [[REVIEW_SUMMARY]] for all locations, the 19-row test backlog, positive feedback, and recorded validation details. This comment draft introduces no fixes or new test results; findings apply to the fixed snapshot above. No GitHub comment or review has been posted.

</details>

---
*Automated review by Maestro*
