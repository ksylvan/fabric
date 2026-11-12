---
type: analysis
title: PR 2330 Broken Access Control Review
created: 2026-11-11
tags:
  - code-review
  - security
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-security-context]]'
  - '[[pr-2330-authentication-review]]'
  - '[[pr-2330-sensitive-data-review]]'
---

# Broken access control review

Reviewed the fixed PR snapshot from `de0b72766db0436b6c1887bb75768cb3960e67e5` to `fb296d490cca4fdebbcb78d5107241f134152d19`. Relevant CLI, core, server, and Keenable directories match that head. This completes only the broken-access-control checkbox; later security checks remain pending. Production code is unchanged and zero images were analyzed.

## Assessment

No new broken access control vulnerability was identified in the added local client/CLI flow. The PR introduces outbound API requests, not an inbound HTTP endpoint. The deliberate absence of a CLI `IsConfigured` gate permits public search without a key; configuration readiness is not a user/role authorization policy. Authorization for Keenable's remote data remains the service's responsibility and cannot be verified from this client repository or mocked responses.

| Check | Evidence and result |
| --- | --- |
| Authorization on endpoints | `internal/tools/keenable/keenable.go:20-23,69-84` selects the fixed `/v1/search/public` endpoint for empty/whitespace keys and `/v1/search` with `X-API-Key` for nonempty keys. `:93-97` rejects every non-200 response before decoding search results. There is no retry or public fallback after keyed denial. No inbound search route or middleware exception is added. |
| Direct object references | `internal/tools/keenable/keenable.go:62-65` sends exactly `query` and `max_results` in JSON. There is no tenant/account/user/resource identifier, ownership lookup, or client-controlled API path in this flow. Query text containing traversal syntax remains data. `:105-118` renders result URLs as text without fetching them, so a returned internal-resource URL does not itself access that resource. |
| Role-based access control | `Client` and `searchResponse` (`internal/tools/keenable/keenable.go:29-32,51-58`) contain no roles or permissions. Response fields do not update settings or select privileged operations. This local CLI acts with its invoking process's existing filesystem and environment access; the PR adds no multi-user role model. |
| Privilege escalation / CLI propagation | `internal/cli/tools.go:126-139` returns immediately on a search error, before appending results or writing output. `internal/cli/cli.go:98-101` also returns that error before downstream chat processing. The caller's flags select the output destination; result text does not select filenames, overwrite credentials, or invoke setup/admin operations. No privilege-changing operation is added. |
| Registration and settings | `internal/core/plugin_registry.go:65,159,181,342,531` constructs, stores, persists, and configures the optional client using existing plugin helpers. These additions do not change authorization rules, ownership checks, server-key settings, or server route registration. |
| Existing server boundary | `internal/server/serve.go:43-85,88-132` constructs the secured engine and registers existing handlers. API-key/loopback checks remain in place and unchanged by this PR. The additions do not reference the server's inbound API key or expose the new client via a server handler. Passing existing server suites is regression evidence, not a complete audit of every pre-existing endpoint. |

Started with graph searches and snippets for `Search`, `NewClient`, `handleToolProcessing`, `WriteOutput`, and `Serve`. The graph's inbound trace for `Search` returned no callers despite the explicit CLI call, so corroborated the relationships with literal references across Go/TypeScript/Svelte sources and the fixed PR diff. No repository-local `AGENTS.md` or `CLAUDE.md` was present; followed the supplied graph-discovery guidance.

The previously recorded conditional Medium redirect credential/query disclosure and Low raw-error diagnostic disclosure remain applicable; see [[pr-2330-authentication-review]] and [[pr-2330-sensitive-data-review]]. They are not additional access-control findings. In particular, an error body can still contain remote text even though it is not accepted as successful result output. This assessment does not establish the remote service's tenant isolation, key scope, public-index contents, or authorization behavior, nor downstream model/rendering safety.

## Validation

Added review-only tests in external Auto Run `Working/keenable_access_control_review_test.go`, reusing `dataFlowTransport` and extending the existing overlay as `Working/access-control-review-overlay.json`. No production helpers or dependencies were added. Probes exercise:

- Twelve denial combinations: empty, whitespace-only, or synthetic nonempty key; HTTP 401 or 403; standalone or chat mode. Even a denial body shaped like a successful result returns an error, no result string, no output file, unchanged credentials, and exactly one request to the expected endpoint. No fallback, retry, or result-URL fetch occurs.
- A traversal/role-shaped query remains in the JSON payload to the fixed public endpoint. A successful response containing synthetic `role`, `user_id`, and `api_key` fields leaves credentials unchanged; its loopback/admin result URL remains literal text with no second request.
- Existing overlay probes and repository tests continue to cover normal keyed/public routing, response handling, setup, CLI propagation, and server security behavior.

All four targeted package suites passed with the race detector on Go 1.27.3:

```sh
TMPDIR=/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/tmp go test -race -timeout 90s -overlay /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/access-control-review-overlay.json ./internal/tools/keenable ./internal/cli ./internal/core ./internal/server
```

The added probes intercept requests in process and use synthetic credentials; existing tests use local fixtures/servers. No live Keenable requests or actual credentials were used. Review probes/overlays remain in the external Working folder, as in preceding tasks, and reproduction requires those files. The full repository suite was not run.
