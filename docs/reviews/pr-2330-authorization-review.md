---
type: analysis
title: PR 2330 Authorization Checks Review
created: 2026-11-11
tags:
  - code-review
  - security
  - authorization
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-access-control-review]]'
  - '[[pr-2330-auth-logic-review]]'
---

# Authorization checks review

Reviewed fixed PR base `de0b72766db0436b6c1887bb75768cb3960e67e5` to head `fb296d490cca4fdebbcb78d5107241f134152d19`. No additional authorization vulnerability identified. The PR adds an outbound search client and local CLI integration, with no inbound route, role, ownership policy, or admin function. Production code remains unchanged. Only the authorization checkbox is completed; consolidated security findings remain a separate task.

## Evidence and limits

| Requested check | Assessment |
| --- | --- |
| Protected endpoint permissions | `internal/tools/keenable/keenable.go:69-84` selects the fixed public endpoint for an empty/whitespace key or the keyed endpoint with `X-API-Key`. At `:93-97`, non-200 responses return an error before result decoding. `internal/cli/tools.go:126-139` propagates denial before result output/chat augmentation. No public retry follows a keyed denial. Public search is intentional; configuration readiness is not an authorization policy. |
| No IDOR | `keenable.go:62-65` sends query text and a result count, without tenant/user/object IDs or caller-selected endpoint paths. At `:105-118`, returned URLs become text, without fetching resources. Existing access-control probes verify traversal/role-shaped input stays in the JSON query and response credential/role fields cannot mutate the key or trigger a resource fetch. Remote search-result entitlement and tenant isolation cannot be established with local mocks. |
| Admin functions protected | No admin operation is added. `internal/core/plugin_registry.go` adds optional client construction, configuration and persistence through existing local setup helpers. These run with the invoking process's permissions. Existing server registration at `internal/server/serve.go:88-124` remains behind `newSecuredEngine` at `:43-84`. Keyed mode applies `APIKeyMiddleware` before handler execution; keyless mode applies loopback Host/Origin checks. Swagger documentation is deliberately public in keyed mode. No new exception is added. |

The relevant Keenable, CLI, core and server sources match the fixed PR head. The fixed base/head diff contains no server changes. Graph discovery and snippets located Search and the server engine builders; graph code search found no Keenable reference under `internal/server`. No repository-local AGENTS.md or CLAUDE.md was found; supplied session guidance applies.

The existing server has a shared API key, not per-user identities or differentiated administrator roles. Rejecting unauthenticated requests across registered routes does not prove ownership isolation between holders of that shared key, nor constitute a complete audit of pre-existing server handlers. Prior conditional Medium redirect credential/query disclosure and Low raw-error disclosure remain recorded in [[pr-2330-authentication-review]]; they are not new authorization findings.

## Validation

Reused the earlier `Working/access-control-review-overlay.json` and its denied-search/resource-selection probes. Added external review-only `Working/server_authorization_review_test.go`, composing `Working/authorization-review-overlay.json`. The new probe enumerates the actual standard server's registered routes, excludes intentionally public Swagger documentation, substitutes synthetic object references, and verifies each route rejects missing/wrong keys with 401 and a foreign Host in keyless mode with 403. Each request uses a fresh engine/limiter so rate limiting cannot mask a missed permission check. Existing server tests supply successful authorized-request controls and Origin coverage. These checks exercise routing and middleware, including settings and resource mutation handlers, without reaching a live provider.

The initial fixture failed to compile because it used the directory name as the package name and an incorrect fsdb import. Corrected it to the existing `restapi` package and `internal/plugins/db/fsdb`. All four complete targeted package suites then passed with the race detector:

```sh
TMPDIR=/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/tmp go test -race -timeout 90s -overlay /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/authorization-review-overlay.json ./internal/tools/keenable ./internal/cli ./internal/core ./internal/server
```

Review fixtures and overlays remain in the external Auto Run Working folder; reproduction requires those files. The full repository suite was not run. No live provider/model calls or actual credentials were used. Zero images analyzed.
