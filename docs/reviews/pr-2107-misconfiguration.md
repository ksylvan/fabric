---
type: report
title: PR 2107 Security Misconfiguration Review
created: 2026-10-03
tags:
  - code-review
  - security
  - configuration
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-security-context]]'
  - '[[pr-2107-authentication]]'
  - '[[pr-2107-sensitive-data]]'
  - '[[3_CHECK_SECURITY]]'
---

# PR 2107 security misconfiguration review

Completed only the **Security misconfiguration** checkbox: production debug mode, default credentials, unnecessary features, and security headers. Compared original base `ddf1aab968caa9adf90137a6536c768d02b237d6` with original head `f32656b529a1eb1c678c793703fe49a0755cd439`. The three-file PR changes no server, logging, CLI configuration, Docker, or Svelte configuration. No new configuration regression was established. Findings below are pre-existing; no production code changed. No task images were supplied or analyzed (0).

## Assessment

| Check | Result |
| --- | --- |
| Debug mode | Fabric content diagnostics default to Off (`internal/log/log.go:28`, `internal/cli/flags.go:117,183`). Gin is separate: constructors do not set release mode, and pinned Gin v1.12.0 defaults to debug outside tests when `GIN_MODE` is empty. Debug construction prints route/handler metadata; explicit release mode is retained. |
| Default credentials | CLI API key defaults to empty and reads `FABRIC_API_KEY` or `--api-key` (`internal/cli/flags.go:85-86`). Default bind is loopback. Both server entry points reject non-loopback binds without a key (`internal/server/auth.go:20-34`, `serve.go:31-33`, `ollama.go:214-219`). Documentation example keys are placeholders, not built-in credentials. Weak nonempty keys are the existing finding in [[pr-2107-authentication]]. |
| Unnecessary features | Servers are opt-in CLI flags. REST always registers public Swagger (`serve.go:46-66`, `auth.go:44-49`); this documented discovery surface has no production disable option in the inspected flow. No profiler/debug route or wildcard CORS middleware was found. Ollama registers 26 routes, including storage/configuration and chat, protected by its configured key; expanded functionality is intentional, not an established bypass. |
| Security headers | No common header middleware is installed by either server constructor. Real Ollama 401 and 200 responses lacked HSTS, nosniff, framing protection, and CSP. A storage-name validation 400 had HSTS but lacked the other three. REST uses the same auth/storage paths. Swagger HTML has no application-level CSP/frame policy in the inspected server setup; deployed proxy headers and dependency-served HTML were not probed. |

The Docker runtime runs as `appuser` (`scripts/docker/Dockerfile:52-56`) and sets no `GIN_MODE`. Chat sets a fixed local development CORS origin (`internal/server/chat.go:96-100`); this is not a wildcard or demonstrated arbitrary-origin disclosure. No new configuration helper or parallel authentication path is warranted by this review.

## Low: production servers inherit framework debug mode

- **Type:** OWASP security misconfiguration; diagnostic metadata exposure (CWE-489).
- **Locations:** `internal/server/serve.go:35-38`, `internal/server/ollama.go:225-229`; pinned dependency `github.com/gin-gonic/gin@v1.12.0/mode.go:52-65`; Docker runtime configuration in `scripts/docker/Dockerfile`.
- **Evidence:** dependency initialization reads `GIN_MODE`; with an empty value outside a test binary it chooses DebugMode. Source contains no production SetMode override. A characterization probe constructed the real Ollama engine in explicit debug and release modes: both persisted, and debug emitted all route/handler registrations.
- **Impact/preconditions:** production deployment leaves framework mode unset and retains accessible diagnostic stdout. This exposes route/handler metadata to log readers, not a demonstrated remote debugger, stack-trace response, or automatic credential leak. Fabric Wire logging remains independently opt-in.
- **Recommendation:** set `GIN_MODE=release` in production deployment examples/runtime configuration, or provide a shared explicit server mode policy while retaining deliberate development diagnostics.
- **Attribution:** unchanged by this PR; Low hardening, not a newly introduced merge blocker.

## Low: browser security headers are missing or inconsistent

- **Type:** OWASP security misconfiguration; missing defense-in-depth response policy.
- **Locations:** `internal/server/serve.go:35-46`, `internal/server/ollama.go:225-234`, `internal/server/storage.go:20-24,33-35,51-53`; public Swagger registration at `serve.go:46-66`.
- **Evidence:** authenticated version 200 and unauthenticated 401 lacked `Strict-Transport-Security`, `X-Content-Type-Options`, `X-Frame-Options`, and `Content-Security-Policy`. Storage validation 400 alone contained HSTS. No application middleware for CSP, nosniff, or frame policy was found by graph-backed source search; Svelte configuration has no explicit CSP policy.
- **Impact/preconditions:** browser-facing deployments without proxy policy lose MIME-sniffing/framing/script defenses. JSON API responses have narrower exposure than Swagger/UI HTML; no exploit was demonstrated. HSTS on plain HTTP is ignored by browsers and is not a substitute for TLS. Existing remote cleartext transport risk belongs to [[pr-2107-sensitive-data]], not a second finding here.
- **Recommendation:** centralize appropriate headers for success/error responses in existing server middleware or the deployment proxy; apply nosniff and framing restrictions, choose/test a CSP compatible with served HTML, and emit HSTS consistently at an HTTPS termination point. Review `includeSubDomains` before deployment. An optional production Swagger switch would reduce public documentation surface, but public docs alone are not a confirmed vulnerability.
- **Attribution:** pre-existing conditional hardening gap, outside original changed lines.

## Validation and limits

Used the indexed code graph for symbol discovery/snippets and header/mode searches, then inspected exact source lines, deployment configuration, and pinned local Gin source. No root `CLAUDE.md` or `AGENTS.md` exists; supplied graph guidance applies.

Three temporary characterization tests passed: separate framework/content logging modes; three real header/status cases; and 26-route diagnostic-surface enumeration with three unauthenticated non-loopback bind rejections. These tests confirm observed gaps and controls rather than fixing them. Formatted source is retained at `/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2107-codex/Working/review_misconfiguration_probe_test.go`. To reproduce, copy temporarily to `internal/server`, run `go test ./internal/server -run TestReviewMisconfiguration -v`, and remove the copy. The constructor probe selects modes explicitly because Gin defaults to TestMode in Go test binaries; production default is established from dependency initialization source, not inferred from that test default.

After removing temporary probes, `go test ./internal/server ./internal/core ./internal/domain ./internal/log` passed on Go 1.27.1, darwin/arm64, with build cache and temporary storage in the authorized Working folder. No permanent source/tests changed. Full repository tests, live deployment/proxy configuration, browser rendering, all extension/provider defaults, and later playbook categories were not assessed. External TLS/header configuration can mitigate the reported deployment gaps; none was available here.
