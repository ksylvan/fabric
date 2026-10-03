---
type: report
title: PR 2107 Insufficient Logging Review
created: 2026-10-03
tags:
  - code-review
  - security
  - logging
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-authentication]]'
  - '[[pr-2107-sensitive-data]]'
  - '[[3_CHECK_SECURITY]]'
---

# PR 2107 insufficient logging review

Completed only the **Insufficient logging** checkbox: security-event logging and sensitive data in logs. Compared original base `ddf1aab968caa9adf90137a6536c768d02b237d6` with original head `f32656b529a1eb1c678c793703fe49a0755cd439`. No new logging regression was demonstrated. The PR adds application-layer path rejection but no dedicated audit event; existing generic access logs and CLI warnings remain. No production code changed. No images were supplied or analyzed (0).

## Assessment

| Check | Result |
| --- | --- |
| Authentication events | Both engines install Gin access logging before authentication (`internal/server/serve.go:37-41`, `internal/server/ollama.go:228-231`). Missing/wrong keys return 401; accepted keys reach the handler. Access logs record status, method, path, client IP, time and latency. The auth middleware emits no dedicated reason/event (`internal/server/auth.go:40-66`). Generic status records provide some detection evidence; logging is not wholly absent. |
| File security events | Parsing and applying reject nonlocal paths with returned errors (`internal/domain/file_manager.go:85-86`, `:157-158`). `Send` prints parse/apply warnings to stdout (`internal/core/chatter.go:195`, `:202`) and successful application messages (`:204-205`). Successful writes print operation/path (`file_manager.go:172`); direct rejection in `ApplyFileChanges` prints nothing. API file-application skipping (`chatter.go:192`) has no event. |
| Other inspected events | Unauthenticated startup warns through slog (`serve.go:43`, `ollama.go:233`); storage operational errors use slog (`internal/server/storage.go:37`); Ollama errors use standard log (`ollama.go:293-468`). Storage validation branches return 400 without a dedicated security event (`storage.go:31-35`). |
| Sensitive data | Header API keys were absent from ordinary access logs in the probe. Query-string values were retained. Wire diagnostics retain full request/response content. File success output omitted synthetic file contents, but includes model-controlled paths. |

## Low: generic access records and CLI output lack dedicated security-event context

- **Type:** insufficient security logging (CWE-778; OWASP security logging/monitoring category).
- **Locations:** `internal/server/auth.go:40-66`; `internal/domain/file_manager.go:155-176`; `internal/core/chatter.go:192-205`.
- **Evidence:** three requests produced access records for two 401 responses and one 200 response, without missing/wrong-key reasons or slog events. A direct rejected-path application returned an error with no stdout output. Existing `Send` callers print warnings, and success output includes operation/path.
- **Impact/preconditions:** operators can count 401 responses but cannot distinguish rejection reasons from those records alone; direct file-manager callers must record rejected paths themselves. CLI stdout is not a durable audit trail unless externally captured. Shared API keys also provide no distinct authenticated user identity. This is an observability hardening gap, not an authentication bypass. Absence of a log for every intentional API file-write skip is not independently a vulnerability.
- **Recommendation:** extend existing middleware and logging facilities with bounded, structured events for authentication failure and security validation rejection, using safe reason codes, request identifiers and trusted client metadata. Avoid keys, headers, bodies and raw model-controlled values. Establish collection, retention and alerting for deployed servers. Keep ordinary CLI feedback useful without equating it to a security audit trail.
- **Attribution:** generic logging and caller warnings predate the PR. The new apply-time rejection returns an error in the same existing style; no logging control was removed.

## Low: access logs preserve sensitive query-string values

- **Type:** sensitive information in logs (CWE-532).
- **Locations:** `internal/server/serve.go:37`; `internal/server/ollama.go:228` (default Gin logger).
- **Evidence:** the same Gin logger implementation, configured with a test writer, recorded `token=synthetic-query-secret` for requests to `/ping`, including rejected authentication requests. Header secrets were absent. The query token was synthetic and was not used for authentication.
- **Impact/preconditions:** clients must place a secret or PII in a URL; readers of captured access logs can then recover it. No Fabric endpoint requiring query-string credentials or real production disclosure was established. Logging a key supplied in an arbitrary query does not make query authentication valid.
- **Recommendation:** customize the existing access logger to omit queries or redact sensitive parameter values; retain route, status and safe request metadata. Clients should send credentials in supported headers and avoid sensitive URLs.
- **Attribution:** both logger installations are unchanged in the original PR range.

## Low: wire diagnostics expose message content

This is the same conditional CWE-532 finding already recorded in [[pr-2107-authentication]] and [[pr-2107-sensitive-data]], not an additional issue to count. `internal/core/chatter.go:77-80`, `:120-121`, and `:175-176` log request, streaming and response content without redaction at Wire. The reused real-`Send` probe confirmed synthetic request/response email markers at Wire and their absence at Off. Logging defaults to Off (`internal/log/log.go:28`). Streaming was inspected in source, not separately exercised here. Opt-in content diagnostics should have redaction and restricted access/retention; `%q` quoting is not redaction.

## Validation and limits

Used the existing indexed graph for discovery and snippets; cross-checked original diff, current line numbers and localized output formats. No root `CLAUDE.md` or `AGENTS.md` exists in the checkout; supplied graph instructions apply. Existing mock vendor and the retained sensitive-data logging probe were reused rather than adding production helpers.

Three temporary characterization tests passed: authentication status/event/query logging, file rejection/success output, and real non-streaming chat-content logging. They describe observed behavior, including gaps, rather than asserting a security fix. Formatted sources remain in the authorized Auto Run `Working` folder:

- `review_logging_server_probe_test.go` — copy temporarily to `internal/server`.
- `review_logging_domain_probe_test.go` — copy temporarily to `internal/domain`.
- `review_logging_core_probe_test.go` — copy temporarily to `internal/core`.

Reproduce with `go test ./internal/server ./internal/domain ./internal/core -run TestReviewLogging -v`, then remove the temporary copies. After removing probes, the server/core/domain/log package suites passed. Build cache and temporary files used the authorized Working directory. No permanent source or tests changed. No deployed log collector, alerts, retention, proxy logs, panic/recovery payloads, every provider log sink, browser telemetry or full repository test suite was assessed; this report does not certify all logs as secret-free. Remaining playbook tasks are pending.
