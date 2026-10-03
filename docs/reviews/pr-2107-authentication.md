---
type: report
title: PR 2107 Broken Authentication Review
created: 2026-10-03
tags:
  - code-review
  - security
  - authentication
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-security-context]]'
  - '[[pr-2107-injection]]'
  - '[[3_CHECK_SECURITY]]'
---

# PR 2107 broken authentication review

Completed only the **Broken authentication** checkbox, covering credential strength, session management, credential exposure in logs/errors, and authentication throttling. Reviewed original base `ddf1aab968caa9adf90137a6536c768d02b237d6` and PR head `f32656b529a1eb1c678c793703fe49a0755cd439`. No authentication change or new authentication regression was found in the three-file diff. The findings below are pre-existing. This review does not remediate them or certify the full application as secure. No production code changed. No task images were supplied or analyzed (0).

## Assessment

| Check | Result |
| --- | --- |
| Weak password policies | No user/password login in the inspected server flow. Authentication uses a shared configured `X-API-Key`. Any nonempty key is accepted, including a one-character key. |
| Session management | `BuildSession` loads named chat histories, not login sessions. `Send` derives a provider conversation ID from an existing ID/history name or `rand.Text()`; `ChatOptions.SessionID` is excluded from JSON. None of these identifiers grant authentication. No login cookie, auth-session expiry, or logout flow was found in the inspected server/core/domain flow. |
| Credential exposure in logs/errors | Missing/wrong-key errors are generic; normal Gin access logs did not echo header keys in probes. Wire-level chat logging prints message content and can expose credentials supplied in that content. Short vendor-key masking also has a conditional disclosure gap. |
| Missing rate limiting | API-key validation has no attempt limiter, delay, or lockout; both server engines install logging/recovery/auth middleware without a limiter. `docs/rest-api.md:479` explicitly documents no built-in rate limiting. |

`internal/server/auth.go:20` rejects unauthenticated non-loopback binds; loopback operation without a key is intentional. Both `Serve` and `ServeOllama` call this guard before starting. With a key configured, REST and Ollama engines install the shared middleware. `APIKeyMiddleware` hashes both keys with SHA-256 and uses a constant-time comparison of fixed-length digests. The public `/swagger/` exception is documented and serves documentation; no protected business handler beneath that prefix was found in the inspected registration flow. No authentication bypass was established here. External proxy configuration and whether `localhost` actually resolves only to loopback in a deployed environment were not assessed.

The shared key gives no per-user identity and remains valid while the running server retains it. Restarting with a replacement key is the apparent revocation mechanism in this flow; absence of per-user logout is not by itself a defect in this single-key design. Chat-history ownership, filesystem permissions, API/CLI file-application gating, and IDOR remain for the later access-control tasks.

## Medium: weak API keys accepted without validation

- **Type:** authentication credential strength (CWE-521; authentication category).
- **Locations:** `internal/server/auth.go:20-22`, `:40-65`.
- **Evidence:** a one-character key passed the non-loopback bind guard and authenticated a request through the real middleware.
- **Impact/preconditions:** a deployment configured with a guessable shared key can be compromised by guessing it. This is a configuration-dependent risk; no real credential was attacked. SHA-256 comparison does not increase the entropy of the configured key.
- **Recommendation:** generate a high-entropy key during setup and reject obviously weak configured keys with actionable feedback. Document generation and rotation; reuse the existing bind guard and middleware rather than introducing a second authentication path.
- **Attribution:** unchanged by this PR; a pre-existing hardening gap, not a newly introduced merge regression.

## Medium: authentication attempts are not throttled

- **Type:** unrestricted authentication attempts (CWE-307; authentication category).
- **Locations:** `internal/server/auth.go:40-65`, `internal/server/serve.go:35-43`, `internal/server/ollama.go:223-234`; existing deployment note at `docs/rest-api.md:479-481`.
- **Evidence:** 100 consecutive wrong-key requests all returned 401; no throttled response or delay was observed. Source inspection confirms there is no limiter in this authentication path. A finite probe alone would not rule out a higher threshold elsewhere.
- **Impact/preconditions:** remotely exposed deployments without a proxy limiter permit unrestricted key guessing and authentication traffic. Random high-entropy keys greatly reduce guessing feasibility; weak keys amplify the preceding finding. Deployed reverse-proxy controls were not available for review.
- **Recommendation:** enforce attempt/request limits in the shared server middleware or a mandatory deployment proxy, with trusted client-IP handling and bounded state. Preserve fail-closed authentication responses. There is no separate login endpoint to rate-limit in this flow.
- **Attribution:** pre-existing and explicitly documented; the PR neither adds nor removes rate limiting.

## Low: wire diagnostics can record credentials in chat content

- **Type:** sensitive information in logs (CWE-532).
- **Locations:** `internal/core/chatter.go:77-80`, `:120-122`, `:175-176`.
- **Evidence:** source inspection shows unredacted request, streamed content, and response text at wire logging level. Normal header authentication probes showed no key echo in responses or standard Gin access logs.
- **Impact/preconditions:** wire logging must be enabled and a message must contain a credential; readers of diagnostic logs can then recover it. This does not establish automatic logging of the server's configured API key or all vendor keys.
- **Recommendation:** make content logging explicitly opt-in, redact recognizable credentials, and restrict diagnostic retention/access. Keep routine authentication logs free of request headers and message bodies.
- **Attribution:** existing lines in the base, outside the PR modification. No real secrets were used to test logging.

## Low: short vendor credentials are returned unredacted

- **Type:** credential disclosure through configuration response (CWE-200).
- **Locations:** `internal/server/configuration.go:32-37`, `:78-88`.
- **Evidence:** `maskAPIKey("x")` and `maskAPIKey("abcd")` return the complete input; longer input is masked except its last four characters. The configuration handler uses this helper for vendor keys.
- **Impact/preconditions:** a vendor credential of at most four characters must be configured and a caller must reach the configuration endpoint. Typical long provider keys are partially masked. The endpoint requires the shared key when configured; this is not a demonstrated unauthenticated disclosure. It could affect a short credential accepted by a custom provider.
- **Recommendation:** return a redacted sentinel for every nonempty short credential, preserving an empty value only for an unset key. Retain the existing helper and update its corresponding redaction contract.
- **Attribution:** pre-existing; recorded here because it concerns credential exposure, with broader sensitive-data review still pending.

## Validation and limits

Used the existing indexed code graph for authentication, server, chat-session, and configuration discovery/snippets. Cross-checked source line numbers and original commit diff; authentication/server/configuration/session-option files are unchanged by the PR. No root `CLAUDE.md` or `AGENTS.md` exists in this checkout; supplied graph-discovery instructions apply.

Four temporary characterization probes passed: weak-key acceptance, 100 failed attempts without throttling, generic errors/access logs without header-key echo, and short-key redaction behavior. These passing probes confirm observed defects/controls; they are not security-fix regression tests. Retained source: `/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2107-codex/Working/review_authentication_probe_test.go`. To reproduce, copy it temporarily into `internal/server`, run `go test ./internal/server -run TestReviewAuthenticationProbes -v`, then remove it. Its package is `restapi`, matching the existing server package.

Existing bind-guard, REST/Ollama fail-closed, API-key middleware, Ollama middleware wiring, and key-forwarding tests passed alongside the probes. After removing the temporary source copy, `go test ./internal/server ./internal/core ./internal/domain` passed. Go 1.27.1 on darwin/arm64 was used, with test storage/build cache in the authorized Auto Run Working folder. No production source or permanent tests changed. Full repository, deployed proxy/TLS, timing measurements, live credential guessing, and other playbook security categories were not assessed.
