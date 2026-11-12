---
type: report
title: PR 2330 Security Review Findings
created: 2026-11-11
tags:
  - code-review
  - security
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[CODE_ISSUES]]'
  - '[[pr-2330-authentication-review]]'
  - '[[pr-2330-deserialization-review]]'
  - '[[pr-2330-dependency-review]]'
  - '[[pr-2330-logging-review]]'
---

# Security Review Findings

Consolidated security review of PR 2330, **feat: add --keenable_search web search (no API key required)**. Fixed base: `de0b72766db0436b6c1887bb75768cb3960e67e5`; fixed head: `fb296d490cca4fdebbcb78d5107241f134152d19`. Line references below use this head.

Coverage is the 25 changed files (320 additions, 3 deletions) in [[REVIEW_SCOPE]]. All scoped working-tree files were verified byte-for-byte against the fixed head. This report records findings without implementing fixes. It deduplicates observations from the completed security checks rather than counting the same disclosure under several OWASP categories. Earlier reports contain historical “remaining tasks” statements; this consolidation completes the security checklist.

**Outcome:** no confirmed Critical or High application vulnerability introduced by this PR; two Medium findings and four Low findings/hardening observations. Existing dependency audit warnings are tracked separately, including npm's Critical package rating; that rating is not a demonstrated Critical exploit in the reviewed application.

## Critical Vulnerabilities

None confirmed in the scoped PR changes. This is not a clean dependency audit or a guarantee against undiscovered vulnerabilities. See the existing-component warning inventory below.

## High Risk Issues

None confirmed in the scoped PR changes. Remote Keenable service security, deployed infrastructure, and arbitrary downstream renderers were not audited.

## Medium Risk Issues

### SEC-M1: redirects can disclose API keys and queries

- **Severity:** Medium, conditional on a trusted endpoint/intermediary issuing an unsafe redirect.
- **Type:** credential disclosure (CWE-201); HTTPS downgrade also maps to OWASP A02 / CWE-319.
- **File and lines:** `internal/tools/keenable/keenable.go:27` (default client), `:83` (custom header), `:87` (request execution).
- **Description and impact:** the default redirect policy forwards custom `X-API-Key` headers to unrelated hosts, including HTTP destinations. For 307/308 the query POST body is replayed too. The initial URL is fixed HTTPS with normal TLS checks; query text and result URLs cannot directly select a redirect target. Exploitation requires control or misconfiguration of the trusted redirect response. The recipient can obtain the key and private query; downgrade additionally exposes them in transit. No live unsafe redirect or actual credential theft was observed.
- **Evidence:** synthetic transport probes exercise real `http.Client` redirect handling for 301/302/303/307/308 to unrelated HTTP and HTTPS hosts. They reproduce key forwarding and 307/308 body replay without external network calls.
- **Remediation:** reject redirects for this POST API unless required by its contract. Otherwise allow only the original HTTPS scheme, host, and effective port, with a finite hop limit. Reject before contacting the disallowed destination. Removing only the key does not protect replayed query bodies. Add keyed/keyless tests asserting rejected destinations receive no request.
- **Supporting reviews:** [[pr-2330-authentication-review]], [[pr-2330-sensitive-data-review]], [[pr-2330-misconfiguration-review]]. These describe one finding.

### SEC-M2: unrestricted response allocation permits local resource exhaustion

- **Severity:** Medium, conditional on a malicious or faulty upstream response.
- **Type:** uncontrolled resource consumption (CWE-400); OWASP A04 insecure design.
- **File and lines:** `internal/tools/keenable/keenable.go:94` (error read), `:100` (JSON decode), `:106` and `:117` (unrestricted result formatting).
- **Description and impact:** error bodies are read in full and copied into diagnostics; successful responses decode without a byte budget. Result count, titles, URLs, and descriptions have no local limits. Outbound `max_results: 10` is a hint, and the 1,000-rune fallback-snippet cap runs after allocation. A large response within the 30-second timeout can exhaust memory or flood stderr, terminating/disrupting the local CLI. No out-of-memory crash was attempted.
- **Evidence:** a bounded synthetic 2 MiB error body survives in the error; another bounded response containing 11 results with 2 MiB descriptions succeeds and preserves every full description. These demonstrate missing limits without a destructive stress test.
- **Remediation:** enforce a successful-response byte budget with overflow detection before decoding and a smaller diagnostic budget for failures. Reject oversized valid JSON prefixes rather than silently accepting truncation. Bound result count and useful field lengths before formatting. Preserve HTTP status and distinguish intentional diagnostic truncation from read failures; a timeout alone is insufficient.
- **Supporting reviews:** [[pr-2330-error-handling-review]], [[pr-2330-deserialization-review]]. This is the same issue as major robustness finding M1 in [[CODE_ISSUES]], with Medium security severity.

## Low Risk Issues

### SEC-L1: raw service diagnostics disclose echoed data and permit forged log lines

- **Severity:** Low, conditional response-content disclosure.
- **Type:** CWE-209 / CWE-532; raw newlines also create a conditional CWE-117 concern (OWASP A09).
- **File and lines:** `internal/tools/keenable/keenable.go:94` and `:95`; stderr sink at `cmd/fabric/main.go:15`.
- **Description and impact:** non-200 bodies enter returned errors verbatim. If a service/gateway echoes a key, private query, or PII, captured stderr retains it. Remote newlines can resemble extra records in line-oriented logs. The client does not independently append its key/query to diagnostics, and a generic rejection does not reveal them. No actual sensitive response was observed.
- **Evidence:** synthetic 401/error bodies preserve fake credentials, a private-query marker, a reserved email, and embedded newlines. The final stderr sink was inspected statically.
- **Remediation:** return bounded, selected diagnostics or generic status errors; redact known credentials and encode diagnostic control characters. Avoid retaining private request text. A byte cap addresses SEC-M2 but does not resolve disclosure.
- **Supporting reviews:** [[pr-2330-authentication-review]], [[pr-2330-sensitive-data-review]], [[pr-2330-logging-review]]; same underlying issue as minor finding m1 in [[CODE_ISSUES]].

### SEC-L2: shared dotenv serialization accepts injected assignments

- **Severity:** Low hardening observation; pre-existing local configuration behavior.
- **Type:** configuration injection / CWE-74, related to OWASP A03.
- **File and lines:** `internal/plugins/plugin.go:198` (shared serializer), reached by new setup registration at `internal/core/plugin_registry.go:181`.
- **Description and impact:** a locally controlled API-key value containing a newline is serialized verbatim and can become another dotenv assignment. An attacker must already influence the configured value before saving it; search queries/responses do not populate the key setting. Configuration is loaded as dotenv rather than executed by a shell; no automatic command execution was demonstrated.
- **Evidence:** a synthetic multiline key produces a second variable when the serialized value is parsed by the existing dotenv loader.
- **Remediation:** reject CR/LF in single-line credentials or reuse the existing `godotenv.Marshal` quoting pattern, with round-trip tests. Harden the shared serializer rather than introducing a Keenable-specific duplicate.
- **Supporting review:** [[pr-2330-injection-review]].

### SEC-L3: opt-in Wire diagnostics retain private result content

- **Severity:** Low hardening observation; pre-existing shared logger behavior.
- **Type:** sensitive information in logs (CWE-532).
- **File and lines:** `internal/cli/tools.go:132`, `internal/cli/chat.go:20`, `internal/core/chatter.go:102`, `internal/log/log.go:28`.
- **Description and impact:** chat-mode search results become model input; explicitly enabling Wire logs the full message content. Private result text/URLs can enter retained or shared diagnostics. Logging defaults to Off. The API key is not deliberately inserted into messages. Requested standalone output/session persistence are intentional content destinations, not additional independent findings.
- **Evidence:** five verbosity probes using the actual dryrun chatter path show a reserved email at Wire and no such message log at Off/Basic/Detailed/Trace. This logger quotes newlines; it does not reproduce SEC-L1's raw newline behavior.
- **Remediation:** document the privacy cost of Wire diagnostics, prefer metadata-only records or deliberate redaction, and avoid sharing raw support logs containing result text.
- **Supporting reviews:** [[pr-2330-sensitive-data-review]], [[pr-2330-logging-review]].

### SEC-L4: search rejection events lack a dedicated audit trail

- **Severity:** Low optional operational hardening, not a demonstrated access-control failure.
- **Type:** conditional OWASP A09 / CWE-778 concern.
- **File and lines:** `internal/tools/keenable/keenable.go:87` through `:102`; propagation at `internal/cli/tools.go:129`.
- **Description and impact:** failures are visible through returned errors, stderr, and nonzero exit, but no structured persistent event store records rejection/rate-limit outcomes. An unattended runner discarding stderr loses this evidence. No persistent CLI audit requirement or regression was established.
- **Evidence:** 25 cases across five verbosity levels and success/401/403/429/500 outcomes show one request per search, returned failures, and no direct search audit output.
- **Remediation:** where operational requirements justify it, reuse the existing logger for optional service/status/outcome/duration/correlation metadata and arrange runner collection/retention. Exclude keys, queries, response bodies, and private result URLs.
- **Supporting review:** [[pr-2330-logging-review]].

## Security Best Practices

### Existing dependency warnings: separate maintenance triage

No dependency, manifest, lockfile, or SDK change is introduced by this PR. The following are **captured prior-review audit results**, not a new live audit in this consolidation. Full package inventory, advisory links, rating discrepancies, scanner output paths, and applicability limits are in [[pr-2330-dependency-review]].

| Component/location | Recorded severity/type | Applicability, impact, and remediation |
| --- | --- | --- |
| DOMPurify, `web/package.json:75`; resolved `3.4.16` in `web/package-lock.json` | npm Critical package rating, OWASP A06; application exploit severity unconfirmed | Three advisory warnings. Critical SAFE_FOR_TEMPLATES/DOM/shadow-root mode, IN_PLACE/hook mode, and XHTML/SVG-animation mode are absent from the inspected `web/src/lib/utils/sanitize-html.ts:9` HTML/string wrapper. No exploit through that wrapper was demonstrated. Prior audit identifies `3.4.17` as patched; update to a verified patched supported version and rerun sanitizer tests/audit. |
| Local Go toolchain `net/http@go1.27.3`; existing server/client paths, including `internal/tools/keenable/keenable.go:87` | GO-2026-6859 scanner warning; no assigned application severity | Captured scanner reports an HTTP/1 server crash issue fixed in `go1.27.4`; advisory page was unavailable during the prior review. Symbol reachability is not proof that outbound search is exploitable. Rebuild with a verified patched supported toolchain and triage existing server handling. This is local-toolchain state, not a changed manifest version. |
| `golang.org/x/crypto` in unchanged `go.mod`/`go.sum` | GO-2026-5932 module-only warning; no assigned application severity | Scanner found no affected imported package/called symbol. Avoid introducing legacy OpenPGP usage; verify supported remediation using the captured advisory. A version bump alone does not repair unsafe legacy OpenPGP design. |
| Existing web development packages in `web/package-lock.json` | npm Moderate/Low package labels; OWASP A06 warnings | Complete captured audit: 15 affected package entries (1 Critical, 11 Moderate, 3 Low), including transitive propagation. Remaining warnings involve SvelteKit, Svelte plugin/deepmerge, Vitest/mocker, minimatch/TypeScript ESLint, and selector-parser/typography. Assess dev-server/build-input exposure and supported patches separately; do not blindly apply suggested major upgrades/downgrades. Production-only audit flagged DOMPurify alone. |

Audit warning exits were nonzero and were never counted as clean audits. Dependency findings predate this feature and are not included in its two-Medium/four-Low count.

### Preserve external-content and privacy boundaries

- **Low conditional downstream-renderer observation**, potential CWE-79 only with an unsafe consumer: `internal/tools/keenable/keenable.go:117` preserves untrusted HTML, Markdown, and URL schemes. No new browser sink or direct Keenable web route was added, and no application XSS was demonstrated. A separate unsafe renderer could execute malicious markup; browser consumers should reuse `renderMarkdown` at `web/src/lib/utils/sanitize-html.ts:28` after Markdown parsing, retain SSR/error escaping and CSP, and consider HTTP(S)-only URLs. Do not add a parallel sanitizer. See [[pr-2330-xss-review]]. This observation is not counted as a confirmed XSS vulnerability.
- Search results are untrusted model input. Document that retrieved text can contain instructions; preserving external content does not establish prompt isolation. No new model-tool execution privilege escalation was demonstrated by this review.
- Queries leave the process for Keenable in public and keyed modes; public means keyless, not anonymous or local. Chat results also reach the selected model. Provider retention/security and deployment headers were not verified.
- Shared plaintext key storage at `internal/plugins/db/fsdb/db.go:106` uses atomic writes with `0600`; directories are created with `0700`. This limits local access but is not encryption, does not tighten existing directory modes, and does not protect against the same user or backups. Consider an OS credential store for a stronger threat model. Stored-default masking does not establish hidden terminal input. These are existing design limits, not new PR regressions.

## No Issues Found

These are scoped results, not assurances about remote services or the entire repository.

| Reviewed area | Result and supporting evidence |
| --- | --- |
| SQL, command, NoSQL, LDAP, XPath injection | No new execution/query sink. JSON serialization preserves hostile query strings as data; fixed endpoint selection and literal formatting do not execute commands. Shared local dotenv caveat is SEC-L2. [[pr-2330-injection-review]] |
| Password/session authentication | No new login, password comparison, session issuance, cookies, refresh token, or logout lifecycle. Optional outbound key replacement/clearing and cookie/token probes pass. Remote key expiry/revocation/rate limits remain external. [[pr-2330-auth-logic-review]], [[pr-2330-authentication-review]] |
| Secrets/environment configuration | No apparent real secret in all 320 added lines; synthetic fixtures/import-path candidates were classified offline. Optional `KEENABLE_API_KEY`, environment precedence, trimming, masking, and shared persistence reviewed. Not a full-history scan or credential-validity check. [[pr-2330-hardcoded-secrets-review]], [[pr-2330-environment-review]] |
| XML/XXE | JSON-only flow; no XML parser, DTD processing, or entity resolver introduced. Raw XML is rejected; XML-looking strings stay literal and trigger no entity requests. [[pr-2330-xxe-review]] |
| Access control/authorization | Intentional public service mode is not a Fabric login bypass. Keyed denials do not retry publicly. No new admin route, user-selected object lookup, or privilege escalation found; protected existing server route denial probes pass. Service entitlement enforcement cannot be verified locally. [[pr-2330-access-control-review]], [[pr-2330-authorization-review]] |
| Production configuration | No new production debug mode, default credential, inbound route, or applicable missing response-header requirement. Normal TLS verification and request timeout are retained; redirect policy remains SEC-M1. [[pr-2330-misconfiguration-review]] |
| XSS | No demonstrated application execution sink; existing sanitizer rejected 48 hostile output combinations and preserves safe fallback behavior. Existing DOMPurify warnings still require separate triage. [[pr-2330-xss-review]] |
| Executable deserialization | Fixed typed JSON strings with no metadata-driven object construction or custom executable decoder. Invalid types/malformed responses return errors without partial output. Permissive duplicate/null/trailing-content handling does not by itself establish an exploit; allocation remains SEC-M2. [[pr-2330-deserialization-review]] |

## Validation and Limits

For this consolidation, reread completed security reports, verified the 25-file fixed snapshot, and checked concrete source line numbers. Some earlier reports have shifted line references; this report uses the fixed-head source. The indexed graph search returned no Keenable Search node, so direct source inspection supplied the missing references. No repository-local CLAUDE.md/AGENTS.md exists; session instructions apply.

Reused 18 existing review-only fixtures by merging eight overlays with conflict/missing-file checks into external `Working/security-summary-overlay.json`. No duplicate production helper, dependency, or new application test was introduced. All eight complete targeted package suites passed with the race detector:

```sh
W=/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working
TMPDIR="$W/tmp" REVIEW_XSS_OUTPUT="$W/xss-results.json" go test -race -timeout 90s \
  -overlay "$W/security-summary-overlay.json" \
  ./internal/tools/keenable ./internal/cli ./internal/core ./internal/plugins \
  ./internal/plugins/db/fsdb ./internal/server ./internal/log ./internal/i18n
```

The merged suite covers hostile input, redirects, sensitive storage/TLS, malformed/large responses, timeout, logging, key lifecycle, and route denial, with original package tests. Characterization probes assert current flaws; passing tests do not fix them. Reproduction requires the external Working fixtures. Existing XSS renderer probes and sanitizer suites were also rerun for consolidation; dependency audit results above are historical evidence and were not refreshed.

No live Keenable/model calls, actual credentials, remote-service implementation audit, deployed CSP inspection, full-repository test run, or memory-exhaustion attack was performed. Zero images were associated with this task or analyzed.

The requested artifact is `/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/SECURITY_ISSUES.md`. An identical tracked copy at `docs/reviews/SECURITY_ISSUES.md` keeps the consolidated report with the review branch and its evidence documents.
