---
type: analysis
title: PR 2330 Security Logging Review
created: 2026-11-11
tags:
  - code-review
  - security
  - logging
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-sensitive-data-review]]'
  - '[[pr-2330-error-handling-review]]'
  - '[[pr-2330-authentication-review]]'
---

# Security logging review

Reviewed the fixed PR snapshot `de0b72766db0436b6c1887bb75768cb3960e67e5` to `fb296d490cca4fdebbcb78d5107241f134152d19` for security-event visibility and sensitive log content. Relevant implementation directories still match the recorded head. Only the insufficient-logging checkbox is completed; this is a review, with no production changes. No images were supplied or analyzed.

## Assessment

| Check | Evidence | Result |
| --- | --- | --- |
| Security events are logged | `internal/tools/keenable/keenable.go:90-105` returns transport, HTTP status, and decode errors. `internal/cli/tools.go:129-130` and `internal/cli/cli.go:97-99` propagate them. `cmd/fabric/main.go:13-16` prints errors to stderr and exits unsuccessfully. | Failures are visible to a CLI caller, including remote rejection/rate limiting, but no structured audit record, timestamp, correlation ID, or persistent event sink is added. Debug verbosity does not add search events. |
| No sensitive data in logs | Search does not directly log its key, query, request headers, or body. However `internal/tools/keenable/keenable.go:96-98` includes the raw non-200 response body in returned errors. | Conditional disclosure persists if the service echoes private data; default debug Off does not suppress the main stderr error. This carries forward the existing Low diagnostic finding. |
| Chat result privacy | `internal/cli/tools.go:132` routes results through `internal/cli/chat.go:20-23` into model messages. `internal/core/chatter.go:102-105` logs full message content at Wire; `internal/log/log.go:28-30` defaults to Off/stderr. | Existing opt-in Wire behavior can expose personal/private result text. Verified absent below Wire on the tested chatter path; no guarantee is made for all other providers or logging paths. |
| Authentication/session audit applicability | The PR adds an outbound optional service key, not Fabric login, session issuance, role changes, or an inbound protected endpoint. | No new local authentication lifecycle requires a server audit trail. Remote provider logs, alerting, retention, and service-side security-event coverage cannot be assessed here. |

## Findings and recommendations

### Low hardening observation: no search security-event audit trail

**Type:** OWASP A09 security logging and monitoring failures / CWE-778, conditional operational concern.

**Locations:** `internal/tools/keenable/keenable.go:90-105`; `internal/cli/tools.go:129-130`.

Remote authentication/authorization rejections and rate limits are returned to the invoking CLI, but no dedicated event records distinguish keyed/public mode, outcome, or timing. An unattended operator who discards stderr cannot reconstruct those events from a Fabric audit store. This is an observability limitation, not demonstrated authentication bypass or a merge-blocking audit regression: the integration is a local CLI tool, no persistent audit requirement is established, and visible nonzero errors already exist.

Where operational auditability is needed, use the existing logging abstraction for optional metadata-only events (service, status class, outcome, duration, correlation identifier) and arrange collection/retention at the CLI runner. Keep credentials, queries, response text, and private result URLs out of those records. Do not add verbose body logging to address this gap. This is a hardening observation for the consolidated review, not a claim that all local command invocations must be permanently logged.

### Low: raw response diagnostics can disclose sensitive data and forge log lines

**Type:** CWE-209 / CWE-532; newline handling also presents a conditional CWE-117 concern.

**Locations:** `internal/tools/keenable/keenable.go:96-98`, `cmd/fabric/main.go:15`.

The existing Low error-disclosure issue in [[pr-2330-sensitive-data-review]] applies to logging. A service/gateway body containing a synthetic key, private query, reserved email, and newline survives intact in the returned error. Main prints that error with `%s` to stderr. If collected as line-oriented logs, remote newlines can appear as additional records. This requires hostile or sensitive response content; normal generic HTTP errors do not automatically expose the configured key. The probe verifies the error content, while the final stderr sink was inspected statically. No live malicious service or real disclosure was observed.

Prefer bounded, selected status diagnostics, redact known credentials, and encode control characters at diagnostic boundaries. Preserve useful failure/status information without retaining raw bodies. Treat line injection as part of this diagnostic-boundary finding rather than a separate independent severity escalation.

### Low: existing Wire logs retain private result content

**Type:** CWE-532; same observation as [[pr-2330-sensitive-data-review]], not an additional issue.

**Locations:** `internal/core/chatter.go:102-105`, `internal/log/log.go:28-30`.

The shared chatter logs complete model input and output when Wire is explicitly enabled. Keenable titles, descriptions, snippets, and URLs can contain private information. A dryrun probe confirms a reserved email in result-shaped Markdown reaches the actual chatter Wire logger; it does not appear in that logger at Off/Basic/Detailed/Trace. The `%q` formatter quotes embedded newlines in these message records, so this specific wire-message sink does not reproduce the raw stderr newline issue. Search configuration keys are not deliberately added to model messages, but externally returned content could itself contain secrets.

Document the privacy cost of sharing Wire diagnostics; prefer metadata or explicit redaction when collection is needed. Standalone requested output and user-requested session persistence are separate intended content destinations, not evidence of an additional audit logger.

## Validation and limits

Used the existing indexed graph to find Search, tool dispatch, and chatter Send. The inbound Search trace returned no callers despite the visible dispatch call; read the actual dispatch snippet instead of treating that incomplete graph edge as proof of isolation. Confirmed relevant directories have no diff from the fixed head. No repository AGENTS.md or CLAUDE.md files were present; session guidance applies.

Created review-only `Working/keenable_logging_review_test.go`, `Working/keenable_wire_logging_review_test.go`, and `Working/logging-review-overlay.json` in the external Auto Run folder. Reused the existing `dataFlowTransport`, authentication overlay, error probes, and core dryrun/vendor setup patterns; no new production helper or dependency was added.

The 25 tool cases cover five verbosity levels across HTTP 200/401/403/429/500: one request per case, no direct tool debug/audit output, private result forwarding on success, and raw synthetic sensitive/newline content retained on failures. Five chatter cases exercise the actual Wire logging boundary using a synthetic result-shaped message and the dryrun vendor; no model or service network calls occur. These are characterization probes of current behavior, not desired secure regression contracts. Earlier overlay probes also run with their package suites.

All four complete targeted package suites passed with the race detector:

```sh
TMPDIR=/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/tmp go test -race -timeout 90s -overlay /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/logging-review-overlay.json ./internal/tools/keenable ./internal/cli ./internal/core ./internal/log
```

Reproduction requires the external Working folder, as in preceding review tasks. The full repository suite, physical terminal rendering, remote provider logs/monitoring, retention, and deployments were not evaluated. No live service calls, actual credentials, or images were used. Later secret scanning, auth/authz tasks, and consolidated SECURITY_ISSUES.md remain for subsequent runs.
