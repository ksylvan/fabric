---
type: report
title: PR 2107 Authentication Logic Review
created: 2026-10-03
tags:
  - code-review
  - security
  - authentication
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-authentication]]'
  - '[[pr-2107-access-control]]'
  - '[[3_CHECK_SECURITY]]'
---

# Authentication logic assessment

Completed only the **Auth logic** task. The conditional requirement “If auth code is modified” is not triggered: the original diff from `ddf1aab968caa9adf90137a6536c768d02b237d6` to `f32656b529a1eb1c678c793703fe49a0755cd439` changes only `internal/core/chatter.go`, `internal/domain/file_manager.go`, and its tests. The chat modification gates generated-file application; it does not alter identity, key comparison, or session authentication. No new authentication regression was established. Production code is unchanged. No images were supplied or analyzed (0).

| Requested check | Assessment |
| --- | --- |
| Login flow | No user/password login flow is changed or present in the inspected REST server authentication path. It accepts a configured shared `X-API-Key`. Existing non-loopback fail-closed binding and middleware checks remain intact. |
| Timing-safe password comparison | Password comparison is not applicable here. API keys are SHA-256 hashed to equal-length digests and compared with `subtle.ConstantTimeCompare` (`internal/server/auth.go:40-65`). This establishes the source-level comparison control, not a measured timing guarantee for the entire HTTP request. |
| Secure session tokens | Named chat histories (`internal/core/chatter.go:220-231`) and internally derived provider conversation IDs (`:71-72`; `internal/domain/domain.go:56-59`) are not authentication tokens. SessionID is excluded from JSON. No new bearer token, login cookie, token generator, or expiry mechanism is added. Key entropy remains a deployment concern recorded in [[pr-2107-authentication]]. |
| Logout invalidates sessions | No per-user authentication session or logout handler exists in the inspected server flow. Deleting chat history does not revoke the shared key. Middleware captures its configured key digest at construction; replacing the engine/server with a new key rejects the old key. An existing engine continues accepting its old key. This does not establish live rotation or revocation across multiple running instances. |

Previously recorded weak-key acceptance and missing throttling remain inherited findings in [[pr-2107-authentication]]; no duplicate issue is introduced by this report. Provider OAuth implementations are outside this unchanged server-authentication task. Endpoint authorization/IDOR is the next unchecked task and was not completed here.

## Evidence and validation

No root CLAUDE.md or AGENTS.md is tracked or present in this checkout. Applied supplied graph-discovery guidance. Used the existing repository graph to discover and read APIKeyMiddleware, BuildSession, and existing API-key tests; checked original Git commit sources/diff and current line references. Graph searching found no Login/Logout symbols under server/core/domain; this is bounded evidence, not a claim about every dependency or provider.

Created four synthetic characterization cases reusing the actual middleware: a history name does not authenticate without a header key; an existing engine accepts its configured key; a replacement engine rejects the previous key; and the replacement key succeeds. All passed with `go test ./internal/server -run TestReviewAuthLogic -count=1 -v`. These tests mount a synthetic protected route and make no network/provider calls; they do not prove production rotation orchestration. Retained formatted source: `Working/review_auth_logic_probe_test.go` in the Auto Run folder. To reproduce, copy it to `internal/server/review_auth_logic_probe_test.go`, run that command, then remove the copy.

The existing server, core, and domain suites passed with `go test ./internal/server ./internal/core ./internal/domain`. The temporary probe was removed afterward. No permanent tests or production source changed. Test storage/build cache used the authorized Auto Run Working folder. Full-repository tests, deployment configuration, distributed revocation, browser logout behavior, and timing measurements were not assessed.
