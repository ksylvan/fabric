---
type: report
title: PR 2107 Consolidated Security Review Findings
created: 2026-10-03
tags:
  - security
  - code-review
  - fabric
related:
  - '[[REVIEW_SCOPE]]'
  - '[[3_CHECK_SECURITY]]'
  - '[[pr-2107-injection]]'
  - '[[pr-2107-access-control]]'
  - '[[pr-2107-xss]]'
  - '[[pr-2107-dependencies]]'
  - '[[pr-2107-authentication]]'
  - '[[pr-2107-sensitive-data]]'
  - '[[pr-2107-misconfiguration]]'
  - '[[pr-2107-logging]]'
  - '[[pr-2107-environment]]'
  - '[[pr-2107-authorization]]'
---

# Security Review Findings

## Scope and disposition

Consolidates the completed security reviews for https://github.com/danielmiessler/Fabric/pull/2107, comparing base `ddf1aab968caa9adf90137a6536c768d02b237d6` with original head `f32656b529a1eb1c678c793703fe49a0755cd439`. Later checkout commits contain review documentation; application sources still match the original head. No root CLAUDE.md or AGENTS.md is present; supplied graph-discovery instructions apply.

**No Critical vulnerability or newly introduced security regression was established. Three unresolved High application findings remain**, plus one inherited High development-dependency advisory with unconfirmed runtime applicability. Five Medium findings are deployment-dependent; ten Low findings are hardening opportunities. Repeated observations across auth, logging, data exposure, and authorization reports are counted once here. All application findings describe inherited behavior, except L01, which hardens the PR's new channel-based authority convention. This review implements no fixes.

The PR improves the current REST file-write boundary and lexical path rejection, but does not fully establish project containment. Its shell-execution-removal claim is unsupported by the actual three-file diff: `sh -c` remains. Recommend addressing H01/H02 before describing this PR as a complete remediation of its stated security goals; track H03 separately. These are review recommendations, not human sign-off or evidence of a remote exploit.

## Critical Vulnerabilities

None established within the reviewed boundary. This is not a whole-application security certification.

## High Risk Issues

### H01: Command substitution survives double-quoted extension placeholders

- **Severity:** High. **Type:** OWASP Injection; CWE-78.
- **File and line:** `internal/plugins/template/extension_executor.go:52,78,83,86`; shipped example `internal/plugins/template/Examples/sqlite3_demo.yaml:17`.
- **Description and impact:** Values are single-quoted then interpolated into shell source executed through `sh -c`. Within a template's double quotes those single quotes are literal, so `$(printf REVIEW_INJECTED)` and backticks execute. An attacker controlling a value used by a vulnerable installed extension can execute commands with process privileges. Standalone unquoted placeholders resisted the probes; remote end-to-end reachability was not demonstrated. Template execution occurs before the new file-write gate.
- **Remediation:** Reuse the executor/registry with separate executable and literal argument elements, remove shell interpolation, and deliberately migrate templates. SQL extensions also need database parameter binding. Cover quoted contexts, numbered placeholders, substitutions, separators, spaces and newlines; `strings.Fields` is insufficient for parsing arguments.
- **Evidence/attribution:** Three safe substitution cases and two controls in [[pr-2107-injection]]; unchanged by this PR.

### H02: Project-local symlinks allow generated writes outside the project

- **Severity:** High. **Type:** OWASP Broken Access Control; CWE-59.
- **File and line:** `internal/domain/file_manager.go:157-169`; call at `internal/core/chatter.go:192-205`.
- **Description and impact:** `filepath.IsLocal` checks lexical locality. `MkdirAll`/`WriteFile` follow intermediate and final symlinks, allowing model-directed overwrite of outside files writable by the process. Both link types overwrote synthetic outside markers. Preconditions include an existing project symlink and entry into the CLI-like application branch; current REST chat skips that branch. No OS privilege elevation was shown.
- **Remediation:** Perform writes through a project-root directory handle with containment enforced during resolution, including symlinks and races. A separate `EvalSymlinks` check followed by ordinary writes leaves a race. Add regressions for file/directory links alongside valid nested paths.
- **Evidence/attribution:** [[pr-2107-access-control]] and [[pr-2107-correctness]] describe the same inherited defect; it remains after the PR.

### H03: Assistant messages reach HTML insertion without sanitization

- **Severity:** High. **Type:** OWASP Injection/XSS; CWE-79.
- **File and line:** `web/src/lib/components/chat/ChatMessages.svelte:75-92,159-161`; ingress `web/src/lib/services/ChatService.ts:92-105`, `web/src/lib/store/chat-store.ts:127-147`, `web/src/lib/store/session-store.ts:93-101`.
- **Description and impact:** Markdown, selected-pattern/plain, and parser-error branches preserve unsafe HTML/event attributes; saved assistant history uses the same rendering boundary. Attacker-influenced assistant content displayed by a victim can execute with the UI origin's authority absent an effective CSP. Unsafe links require activation. Probes confirmed unsafe markup retention and compiled HTML insertion, not browser execution, credential theft or a cross-user exploit.
- **Remediation:** Sanitize every HTML-producing branch after Markdown conversion using a maintained sanitizer and explicit URL/attribute policy. Render plain/Mermaid content and error fallbacks as text. Add browser behavior coverage for streamed/saved content and unsafe links; deploy a compatible UI CSP.
- **Evidence/attribution:** [[pr-2107-xss]]; rendering and persistence are unchanged by the PR.

### D01: High advisory in development-only braces dependency

- **Severity:** High advisory rating; Fabric runtime exploitability unconfirmed. **Type:** OWASP Vulnerable Components; CWE-674.
- **File and line:** `web/package.json:36`, `web/package-lock.json:2533`, `web/pnpm-lock.yaml:890`.
- **Description and impact:** The retained 2026-10-03 audits report braces 3.0.3 through patch-package/find-yarn-workspace-root/micromatch. Deeply nested attacker-controlled patterns could exhaust a Node stack if they reach this development tooling. npm's four affected-package entries and pnpm's one advisory represent one underlying warning; production-only npm audit reports zero warnings. No public runtime path was established.
- **Remediation:** Assess whether patch-package can be removed/replaced, constrain untrusted patterns and track an upstream fix. Update both supported lockfiles deliberately and rerun audits; do not blindly apply the suggested breaking downgrade.
- **Evidence/attribution:** [[pr-2107-dependencies]] and retained audit JSON; inherited, no dependency inputs changed. Advisory snapshot: https://github.com/advisories/GHSA-vfj7-8cjw-p6xm (CVE-2026-93687). This consolidation does not refresh advisory metadata.

## Medium Risk Issues

### M01: Weak shared API keys are accepted

- **Severity:** Medium, conditional. **Type:** OWASP Identification and Authentication Failures; CWE-521.
- **File and line:** `internal/server/auth.go:20-22,40-65`.
- **Description and impact:** A one-character synthetic key passed startup validation and authentication. Guessable configured keys can expose all shared-key authority; digest comparison does not add entropy.
- **Remediation:** Generate high-entropy keys during setup, reject clearly weak configuration, and document generation/rotation using existing authentication facilities.
- **Evidence/attribution:** [[pr-2107-authentication]]; inherited configuration-dependent gap.

### M02: Authentication attempts lack throttling

- **Severity:** Medium, conditional. **Type:** OWASP Identification and Authentication Failures; CWE-307.
- **File and line:** `internal/server/auth.go:40-65`, `internal/server/serve.go:35-43`, `internal/server/ollama.go:223-234`; `docs/rest-api.md:479-481`.
- **Description and impact:** Source has no limiter; 100 wrong-key attempts all returned 401 without throttling. Exposed servers without proxy controls permit unrestricted guessing/traffic, particularly harmful with weak keys. A finite probe does not rule out external controls.
- **Remediation:** Add bounded authentication/request limiting in existing middleware or a trusted deployment proxy, with deliberate client identity and concurrency handling.
- **Evidence/attribution:** [[pr-2107-authentication]]; inherited.

### M03: Sensitive histories use plaintext and permissive local permissions

- **Severity:** Medium, conditional. **Type:** OWASP sensitive-data exposure; CWE-732/CWE-200.
- **File and line:** `internal/plugins/db/fsdb/storage.go:25-26,103-108,269-272`, `internal/plugins/db/fsdb/sessions.go:42-43`, `internal/plugins/db/fsdb/db.go:49-50`.
- **Description and impact:** Under umask 022, session JSON was plaintext with mode 0644 and directory mode 0755. Other users can read it only if ancestors are traversable; private home/configuration ancestors mitigate this. The new API gate retains complete generated responses in existing named-session storage (`internal/core/chatter.go:192-215`), but no new unauthenticated disclosure was shown.
- **Remediation:** Specialize existing session storage to 0700 directories/0600 files, address existing files, and define retention/backup controls; encrypt where required with separately protected keys.
- **Evidence/attribution:** [[pr-2107-sensitive-data]]; storage behavior predates the PR, which changes response retention.

### M04: Unprotected remote HTTP exposes keys and message data

- **Severity:** Medium, conditional. **Type:** OWASP Cryptographic Failures; CWE-319.
- **File and line:** `internal/server/serve.go:79`, `internal/server/ollama.go:214-219,500-521`, `internal/plugins/ai/openai/openai.go:133-136`.
- **Description and impact:** Servers use plain Gin Run; forwarding/custom provider URLs permit remote HTTP. Direct access across untrusted networks permits interception of keys/content. A loopback backend behind properly configured HTTPS termination mitigates this; no deployment or interception was tested.
- **Remediation:** Require documented TLS termination and restricted backend access for network exposure; validate non-loopback provider/forwarding URLs and require explicit opt-in for cleartext while retaining local HTTP support.
- **Evidence/attribution:** [[pr-2107-sensitive-data]]; inherited.

### M05: Shared keys provide no user or administrative isolation

- **Severity:** Medium only when distinct-user/read-only permissions are required. **Type:** OWASP Broken Access Control; CWE-639/CWE-862 under that deployment contract.
- **File and line:** `internal/server/auth.go:40-65`, `internal/server/storage.go:57-64,69-97`, `internal/server/configuration.go:18-26`, `internal/server/sessions.go:15-18`.
- **Description and impact:** A valid key can read/delete any named history and mutate configuration. No principals, ownership or administrative scopes exist in this flow. Sharing a key among mutually untrusted/read-only clients gives excessive authority; this is expected authorization within a single-owner trust domain. No anonymous IDOR was found.
- **Remediation:** Use one instance/key per trust domain and document full key authority. Multi-user deployments need principals, object ownership and scoped administrative checks; random object names are insufficient.
- **Evidence/attribution:** [[pr-2107-access-control]] and [[pr-2107-authorization]] describe one inherited limitation.

## Low Risk Issues

Each item below is a conditional hardening concern, not a demonstrated new exploit.

### L01: Output-channel convention determines file-write authority

- **Severity:** Low. **Type:** OWASP Broken Access Control hardening.
- **File and line:** `internal/core/chatter.go:192`, `internal/server/chat.go:139-151`.
- **Description and impact:** Nil channels enable generated-file application; nonnil channels suppress it independent of streaming. Current REST sets the channel, but a future server caller omitting it could enable writes.
- **Remediation:** Add explicit, default-denied file-application authority granted by trusted CLI entry points, separate from output transport.
- **Evidence/attribution:** Four gate cases in [[pr-2107-access-control]]; concerns the PR's new convention, with no current bypass found.

### L02: Wire diagnostics preserve credentials and PII in message content

- **Severity:** Low. **Type:** OWASP sensitive-data exposure; CWE-532.
- **File and line:** `internal/core/chatter.go:77-80,120-121,175-176`; default Off at `internal/log/log.go:28`.
- **Description and impact:** Synthetic request/response email markers appear at Wire, absent at Off. Sensitive content in opt-in captured diagnostics is available to log readers; streaming was source-inspected. Formatting with percent-q does not redact data.
- **Remediation:** Preserve metadata diagnostics, explicitly opt into content capture, redact sensitive data, and restrict permissions/retention.
- **Evidence/attribution:** [[pr-2107-authentication]], [[pr-2107-sensitive-data]] and [[pr-2107-logging]] describe one inherited issue.

### L03: Short vendor keys are returned intact

- **Severity:** Low. **Type:** sensitive-data exposure; CWE-200.
- **File and line:** `internal/server/configuration.go:32-37,78-88`.
- **Description and impact:** Keys of at most four characters are returned unredacted to configuration callers. A custom provider must accept such a credential; typical long keys are partially masked and no anonymous disclosure was shown.
- **Remediation:** Extend existing maskAPIKey to redact every nonempty short credential and test that contract.
- **Evidence/attribution:** [[pr-2107-authentication]]; inherited.

### L04: Newly generated confidential content uses general source-file modes

- **Severity:** Low. **Type:** permissive permissions; CWE-732.
- **File and line:** `internal/domain/file_manager.go:164-168`.
- **Description and impact:** New files use 0644 under umask 022; existing 0600 files retain their mode. Secret-bearing outputs can become locally readable with traversable ancestors. Ordinary shared source files need not be confidential.
- **Remediation:** Offer an explicit confidential-output policy in the existing writer with restrictive modes and review before application; preserve restrictive update permissions.
- **Evidence/attribution:** [[pr-2107-sensitive-data]]; inherited, distinct from session storage.

### L05: Production servers inherit Gin debug mode

- **Severity:** Low. **Type:** OWASP Security Misconfiguration; CWE-489.
- **File and line:** `internal/server/serve.go:35-38`, `internal/server/ollama.go:225-229`; `scripts/docker/Dockerfile` runtime configuration.
- **Description and impact:** With GIN_MODE unset outside tests, the framework defaults to debug. Route/handler metadata reaches captured stdout; no remote debugger or secret leak was demonstrated.
- **Remediation:** Set GIN_MODE=release in production configuration or centralize an explicit mode policy while preserving deliberate development diagnostics.
- **Evidence/attribution:** [[pr-2107-misconfiguration]]; inherited.

### L06: Browser security headers are missing/inconsistent

- **Severity:** Low. **Type:** OWASP Security Misconfiguration.
- **File and line:** `internal/server/serve.go:35-46`, `internal/server/ollama.go:225-234`, `internal/server/storage.go:20-24,33-35,51-53`; UI configuration `web/svelte.config.js:1`.
- **Description and impact:** Tested success/401 responses lack CSP, nosniff, frame restrictions and HSTS; only a storage validation response has HSTS. No explicit UI CSP was found. Without proxy policy, browser deployments lose defense in depth; JSON responses have narrower exposure. HSTS over HTTP cannot supply TLS.
- **Remediation:** Centralize appropriate success/error headers in existing middleware/proxy; apply a tested CSP at the actual UI origin and HSTS at HTTPS termination, reviewing subdomain implications.
- **Evidence/attribution:** [[pr-2107-misconfiguration]] and [[pr-2107-xss]]; inherited. H03 remains the unsafe rendering defect.

### L07: Security events lack dedicated structured context

- **Severity:** Low. **Type:** OWASP Security Logging and Monitoring Failures; CWE-778.
- **File and line:** `internal/server/auth.go:40-66`, `internal/domain/file_manager.go:155-176`, `internal/core/chatter.go:192-205`.
- **Description and impact:** Access logs record statuses, but not authentication reason codes; direct file rejection returns an error without a dedicated event. CLI warnings require external capture for durable auditing. Logging is not wholly absent.
- **Remediation:** Extend existing logging with bounded safe reason codes, request IDs and trusted client metadata; define collection/alerts without keys, bodies or raw model values.
- **Evidence/attribution:** [[pr-2107-logging]]; inherited logging pattern retained by new rejection logic.

### L08: Access logs preserve sensitive URL query values

- **Severity:** Low. **Type:** sensitive information in logs; CWE-532.
- **File and line:** `internal/server/serve.go:37`, `internal/server/ollama.go:228`.
- **Description and impact:** Default Gin access logs retained a synthetic token query even for rejected requests. Clients must put secrets/PII in URLs; header keys were absent and query authentication was not established.
- **Remediation:** Customize the existing logger to omit/redact queries and retain safe route/status metadata; use supported credential headers.
- **Evidence/attribution:** [[pr-2107-logging]]; inherited, distinct from content diagnostics.

### L09: Server-key dotenv loading occurs after flag parsing

- **Severity:** Low. **Type:** OWASP Security Misconfiguration.
- **File and line:** `internal/cli/cli.go:19,40`, `internal/cli/flags.go:86`, `internal/plugins/db/fsdb/db.go:83-88`.
- **Description and impact:** FABRIC_API_KEY set only in Fabric's dotenv file does not populate an already parsed ServeAPIKey in a fresh process. Non-loopback unkeyed startup still fails closed; no bypass was shown.
- **Remediation:** Document process-environment injection before launch, or resolve the key after dotenv loading while preserving explicit flag precedence. Avoid command-line secrets entering history/process arguments.
- **Evidence/attribution:** [[pr-2107-environment]]; inherited.

### L10: Interactive setup can show stored secrets

- **Severity:** Low. **Type:** sensitive-data exposure; CWE-200.
- **File and line:** `internal/plugins/plugin.go:220-226`.
- **Description and impact:** Existing setting values appear as prompt defaults and fmt.Scanln input may echo. Re-running credential setup can disclose secrets to observers/terminal capture; no real credential was used.
- **Remediation:** Add secret metadata to existing settings, mask current defaults and read secret input without echo. Environment injection avoids this prompt path.
- **Evidence/attribution:** [[pr-2107-environment]]; inherited.

## Security Best Practices

Prioritize literal extension arguments and rooted filesystem writes for the PR's stated goals, then repair assistant HTML rendering. Reuse existing executors, storage validation and middleware rather than creating parallel helpers. Add regression tests that require secure behavior; passing characterization probes currently document defects and are not evidence of remediation.

Use high-entropy rotated keys, trusted TLS termination, one shared trust domain per instance, private storage, and bounded safe logs. Keep content diagnostics disabled by default and establish sensitive-history retention/backup policies.

The retained Go scan reports **11 advisory IDs: eight symbol, one package, two module matches**, not 11 confirmed Fabric vulnerabilities. Nine broad Ollama records lack bounded fixes; token-exposure metadata conflicts with a separate GitHub version range. Fabric uses an Ollama API client; matching common symbols does not establish that it embeds the affected daemon. gRPC's xDS warning has package-only evidence (`go.mod:166`), and OpenPGP metadata is module-only (`go.mod:161`) without an imported/called OpenPGP package; Ollama is at `go.mod:24`. Reconcile upstream applicability and versions before severity assignment/upgrades. See [[pr-2107-dependencies]] for every advisory ID and raw scan provenance. Optional Python resolved versions, containers, native libraries and external daemons were not audited.

## No Issues Found

These are scoped negative findings and positive controls, not claims that the entire categories/application are safe:

- No new SQL, NoSQL, LDAP or XPath injection sink in the original diff; the existing SQLite shell example does not establish a separately reproduced SQL exploit ([[pr-2107-injection]]).
- No XML parsing/XXE sink in changed or inspected adjacent paths; XML stayed opaque text ([[pr-2107-xxe]]).
- Typed JSON/schema validation rejected tested invalid file changes before application; no new unsafe deserialization established. Null/duplicate semantics and allocation limits remain hardening considerations without an exploit ([[pr-2107-deserialization]]).
- No confirmed actual hardcoded credential in the original diff or reviewed snapshot candidates: 856 text files scanned, 12 nontext entries and Git history excluded ([[pr-2107-hardcoded-secrets]]).
- Provider credentials use environment-backed settings/external providers. Saved dotenv credentials use owner-only 0600 writes; process variables take precedence ([[pr-2107-environment]]).
- Constant-time fixed-size digest comparison and non-loopback unkeyed startup rejection remain intact. Chat history/provider IDs are not login tokens; no login/logout code was modified ([[pr-2107-auth-logic]]).
- All 26 actual Ollama and 25 reconstructed REST business routes rejected missing/wrong keys; eight rejected mutations left synthetic state unchanged. Swagger is intentionally public, and single-owner shared-key operations are authorized ([[pr-2107-authorization]]).
- Current REST chat sets UpdateChan; four streaming/non-streaming core gate cases confirmed generated-file application is skipped for nonnil channels ([[pr-2107-access-control]]).
- Dependency inputs are unchanged; production-only npm audit has zero warnings in the retained scan snapshot, subject to its tooling/scope limits ([[pr-2107-dependencies]]).

## Validation and limitations

Consolidation reused all 14 completed security-category reports and their retained probes/scans under the authorized Auto Run Working folder. Key filesystem, shell formatting and middleware sources were rechecked through the existing graph index; original source/dependency boundaries were rechecked with Git. A repository mirror is stored at `docs/reviews/pr-2107-security-issues.md` so the findings can be committed with this task; it matches the required Auto Run `SECURITY_ISSUES.md` byte-for-byte.

This task validates report structure, severity counts, cited source line existence, source-report references, report mirror equality, original source/dependency boundaries, and preservation of the playbook checkbox text. Relevant domain/core/server/template/filesystem-db suites are rerun for the consolidation. Prior category probes passed as documented in their reports; this task does not claim to rerun every probe or audit. No production code changes, real secrets, provider requests or destructive payloads are involved. No task images were supplied or analyzed (0).

Full repository/frontend suites, browser execution, cross-user exploits, Windows filesystem behavior, filesystem race exploitation, live proxies/deployments, all provider internals, Git-history secret scanning, and external dependency/runtime inventory remain outside the demonstrated evidence. Task completion means the findings document exists; it is not security remediation, merge approval or human sign-off.
