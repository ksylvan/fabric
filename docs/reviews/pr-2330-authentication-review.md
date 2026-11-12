---
type: analysis
title: PR 2330 Broken Authentication Review
created: 2026-11-11
tags:
  - code-review
  - security
  - authentication
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-security-context]]'
  - '[[pr-2330-error-handling-review]]'
---

# Broken authentication review

Completed the broken-authentication checkbox for fixed base `de0b72766db0436b6c1887bb75768cb3960e67e5` to head `fb296d490cca4fdebbcb78d5107241f134152d19`. The new integration sends an optional outbound service API key; it introduces no application login, password verification, session cookies, or authentication server endpoint. Found one conditional Medium credential-transport issue and reconfirmed the previously recorded Low error-disclosure concern. This review records findings without changing production code or declaring the remaining security tasks complete.

Used the indexed graph to locate Keenable Search and shared plugin configuration/setup methods, then inspected concrete receiver implementations where graph names collided. No repository AGENTS.md or CLAUDE.md files were present; session instructions apply. Confirmed relevant application directories still match the fixed PR head. No images were associated with this task; zero images analyzed.

## Medium: redirects can forward the API key beyond its intended origin

- **Type:** credential disclosure / authentication credential handling (CWE-201).
- **Locations:** `internal/tools/keenable/keenable.go:29` creates a client with only a timeout; `:86` attaches `X-API-Key`; `:90` executes the request with default redirect handling.
- **Behavior:** net/http automatically follows redirects. Its built-in sensitive-header treatment does not recognize this custom header. The key is retained for a redirect to an unrelated hostname, including an HTTPS-to-HTTP redirect. With 307/308, the POST body containing the query is also replayed. This was reproduced using the actual http.Client redirect machinery and an in-process transport; no network destination received anything.
- **Precondition and impact:** the normal initial endpoint is a fixed HTTPS URL, with normal TLS verification. An attacker cannot select a redirect merely through the search string or result URL. Exploitation requires the API endpoint or a trusted intermediary to issue an attacker-directed redirect, for example through compromise or misconfiguration. The redirected recipient can receive the configured API key and potentially query text; an HTTP destination also exposes them in transit. No live Keenable redirect or actual stolen credential was observed. Severity is Medium given this required control over the trusted response, rather than an unauthenticated exploit demonstrated against the service.
- **Recommendation:** reject redirects for this POST API unless the service contract requires them. If redirects are required, use CheckRedirect to permit only the original HTTPS scheme, hostname, and effective port, with a finite hop limit; return an error before issuing any off-origin or downgrade request. Stripping the key alone does not protect query bodies on 307/308. Preserve the existing request timeout. Add tests that assert an untrusted destination is never contacted, keyed and keyless, while explicitly permitted redirects still work. Do not rely on the standard Authorization header policy to cover X-API-Key.

## Low: raw API errors can disclose echoed credentials

`internal/tools/keenable/keenable.go:96-98` copies the non-200 response body into its returned error, which flows through CLI to stderr. As already documented in [[pr-2330-error-handling-review]], a synthetic 401 body containing a fake key and private query retains both. Disclosure requires the service/gateway to echo those values; the formatter itself does not append the key or request body. A 401 containing only a generic invalid-key message does not reveal the key. This is the same finding, not a second independent defect.

Prefer bounded, selected diagnostics or generic HTTP-status errors, with redaction of known request credentials and control-character sanitization. Keep any useful diagnostics free of private query text. No real key or live service response was used.

## Requested checks

| Check | Assessment |
| --- | --- |
| Weak password policies | Not applicable to this change: the CLI accepts a service-issued API key and adds no user password or account-creation policy. Local key length/complexity rules would not establish service-side security. |
| Session management | No session or refresh-token lifecycle is added. The client has no cookie jar, stores no session cookie, and creates a new request per search. Key revocation/expiry are controlled by the external service and were not verified. Redirect credential scope is the Medium finding above. |
| Credential exposure in logs/errors | Shared AddSetting marks API Key sensitive; DisplayValue masks existing values in setup prompts and Setting.Print. Search contains no new direct key logger. Conditional response-body exposure remains the Low finding. Shared readLine uses ordinary stdin; masking a stored default does not establish hidden terminal entry. That existing interactive-input behavior was not changed by this PR or exercised on a physical terminal. |
| Missing auth rate limiting | No inbound authentication endpoint is introduced in Fabric. Remote abuse prevention and public/keyed quotas cannot be assessed from this repository. A rejected key produces an error after one request; no retry loop or fallback to the public endpoint follows 401. The existing 429 test confirms remote rate-limit errors propagate. |
| Optional-key behavior | Existing configuration tests cover empty, whitespace-only, and padded keys. Empty/whitespace uses the public endpoint without an API-key header; a nonempty trimmed key uses the keyed endpoint. Public mode is an intentional feature, not evidence of bypassing a protected Fabric login. |

## Validation

Created `Working/keenable_auth_review_test.go` and `Working/auth-review-overlay.json` under the external Auto Run folder. Reused the existing reviewTransport helper and error-handling overlay instead of introducing another transport helper. Production code and original PR tests remain unchanged.

Ten redirect cases (301/302/303/307/308, each targeting HTTP and HTTPS on an unrelated host) assert current credential forwarding; 307/308 also assert body replay. These are characterization probes of the defect, not the desired security contract. An additional probe verifies sensitive-setting masking, key trimming, generic 401 errors without key disclosure, and absence of retry/public fallback. The reused overlay includes earlier keyless/configuration and synthetic echoed-error cases.

All five complete package suites passed with the race detector on Go 1.27.3 (unchanged packages may use Go test cache):

```sh
go test -race -timeout 90s -overlay /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/auth-review-overlay.json ./internal/tools/keenable ./internal/plugins ./internal/cli ./internal/core ./internal/i18n
```

The overlay and review probes remain in the external Auto Run Working folder, following prior review tasks; reproducing them requires that folder. No live API/model calls, actual credentials, browser activity, or images were involved. The full repository suite was not run. Sensitive-data exposure, access control, dedicated auth-logic review, secret scanning, and consolidated SECURITY_ISSUES.md remain separate unchecked tasks. Carry the Medium redirect finding and existing Low diagnostic finding into that later report.
