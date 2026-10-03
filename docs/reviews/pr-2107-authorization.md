---
type: report
title: PR 2107 Authorization Checks Review
created: 2026-10-03
tags:
  - code-review
  - security
  - authorization
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-access-control]]'
  - '[[pr-2107-auth-logic]]'
  - '[[3_CHECK_SECURITY]]'
---

# Authorization assessment

Completed only the **Authorization checks** task: protected endpoints, direct object references, and administrative functions. No new authorization regression was established in original PR base `ddf1aab968caa9adf90137a6536c768d02b237d6` to head `f32656b529a1eb1c678c793703fe49a0755cd439`. That diff changes only chatter and file-manager code/tests; server authorization is unchanged. Production code is unchanged. No images were supplied or analyzed (0).

| Requirement | Assessment |
| --- | --- |
| Protected endpoints check permissions | REST `Serve` installs APIKeyMiddleware before all business registrations (`internal/server/serve.go:30-43`, `:75-82`). Ollama installs the same middleware before its handlers (`internal/server/ollama.go:225-255`). Probes rejected missing/wrong keys across 26 actual Ollama routes and 25 REST business routes reconstructed with the same constructors. `/swagger/` is the documented public exception (`internal/server/auth.go:45-49`); no business handler is registered there. Loopback without a configured key intentionally permits local clients; production entry points reject unkeyed non-loopback binds. |
| No IDOR | The middleware authenticates one shared trust domain, not a user principal. Named histories have no ownership check (`internal/server/sessions.go:15-18`, `internal/server/storage.go:57-64`, `:69-97`). Anonymous read/mutation requests were rejected; the valid key read and deleted a synthetic history called `other-user`. This demonstrates absence of user isolation, not an established unauthorized operation in a single-owner deployment. Existing traversal rejection is separate from ownership enforcement. |
| Administrative functions protected | `/config/update`, history delete/save/rename, and pattern/context mutations inherit the same key check. Eight rejected mutation requests left the original history, name, process setting, and `.env` contents intact. A valid-key configuration update succeeded. There is no distinct administrator/read-only role; distributing the key grants mutation authority. |

## Conditional Medium: shared-key object and administrative access

This is the same inherited deployment limitation recorded in [[pr-2107-access-control]], not a second finding. Type: broken access control (CWE-639/CWE-862 when distinct-user/scoped permissions are required). Locations: `internal/server/auth.go:40-65`, `internal/server/storage.go:57-64`, `internal/server/configuration.go:18-26`, `internal/server/sessions.go:15-18`.

If mutually untrusted users or clients intended to be read-only share the key, they can read/delete other histories and change provider configuration. No anonymous IDOR, separate role bypass, or new privilege escalation was demonstrated. Use one instance/key per trust domain and document the key's full authority. If multi-user service is required, introduce authenticated principals, object ownership checks, and administrative scopes on every applicable operation; random object names are insufficient.

The channel-based generated-file gate and inherited symlink escape remain covered in [[pr-2107-access-control]]. This task does not remediate those findings or create the later consolidated SECURITY_ISSUES.md artifact.

## Evidence and validation

No CLAUDE.md or AGENTS.md was found in this checkout; supplied graph discovery instructions were applied. Used the existing indexed graph to discover/read Serve, newOllamaEngine, APIKeyMiddleware, NewStorageHandler, NewConfigHandler, and storage/configuration handlers. Verified original Git diff and that current server/chatter sources match the original PR head.

Reused and extended `Working/review_access_server_probe_test.go` rather than duplicating production helpers. Two temporary characterization tests passed using `go test ./internal/server -run TestReviewAuthorization -count=1 -v`: 102 missing/wrong-key route requests, authenticated history read/delete, eight rejected mutations with side-effect checks, and a successful authorized configuration update. An initial compile error from treating the typed Session result as bytes was corrected by checking its actual HTTP representation before the passing run.

Formatted probe source remains at `/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2107-codex/Working/review_authorization_probe_test.go`. To reproduce, copy it into `internal/server/review_authorization_probe_test.go`, run the command above, then remove it. The temporary package copy was removed. Subsequently `go test ./internal/server ./internal/core ./internal/domain ./internal/plugins/db/fsdb` passed. Temporary storage/build cache stayed in the authorized Working folder.

Limits: REST engine registrations were reconstructed from Serve; no live listener or provider was used. Tests used synthetic keys and data. Public Swagger implementation was source-inspected, not browser-tested. Unknown routes, deployed proxies, web application endpoints, multi-user identity implementations, and full-repository tests are outside this original-PR task. The consolidated findings checkbox remains pending.
