---
type: report
title: PR 2107 Broken Access Control Review
created: 2026-10-03
tags:
  - code-review
  - security
  - access-control
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-security-context]]'
  - '[[pr-2107-authentication]]'
  - '[[pr-2107-correctness]]'
  - '[[3_CHECK_SECURITY]]'
---

# PR 2107 broken access control review

Completed only the **Broken access control** checkbox: endpoint authorization, direct object references, roles, and privilege escalation paths. Reviewed original base `ddf1aab968caa9adf90137a6536c768d02b237d6` and head `f32656b529a1eb1c678c793703fe49a0755cd439`. No newly introduced access-control regression was established. The PR improves remote file-write containment but leaves the pre-existing CLI symlink escape below. This report changes no production code and does not remediate the findings. No images were supplied or analyzed (0).

## Assessment

| Check | Result and evidence |
| --- | --- |
| Authorization on endpoints | REST `Serve` installs the shared API-key middleware before business-handler registration (`internal/server/serve.go:30-43`, `:75-83`); Ollama does likewise (`internal/server/ollama.go:225-258`). All 26 registered Ollama routes rejected missing and wrong keys in probes. REST-only YouTube and strategy routes are registered after the same global middleware. `/swagger/` intentionally bypasses authentication and serves documentation; no business route with this prefix was found. |
| Direct object references | Storage routes validate names/containment, but names have no associated user identity or ownership. The valid shared key can read/delete any named history. No unauthenticated IDOR was demonstrated. This is one shared administrative trust domain, not tenant isolation. |
| Role-based access control | No roles or per-operation scopes in the inspected server flow. Storage mutations and `/config/update` use the same key as reads/chat. This is a deployment constraint if users need different permissions; the PR does not introduce a multi-user contract. |
| Privilege escalation | REST chat always supplies `UpdateChan`; the new conjunction skips generated-file application. Core probes confirmed the gate for streaming and non-streaming calls. CLI writes remain bounded by the process's OS permissions, with symlink escape as detailed below. No remote-to-CLI bypass, OS privilege increase, or new administrative endpoint was demonstrated. |

`Serve` and `ServeOllama` reject a non-loopback bind without a configured key before opening the server. Loopback without a key is intentionally allowed and places trust in local clients; loopback binding is not per-user authorization. Existing tests cover bind guards, REST/Ollama middleware wiring, chat name validation, storage traversal rejection across operations, pattern path rejection, and existing storage symlink rejection. These controls are unchanged by the original PR.

## High: generated file changes can overwrite outside the project through symlinks

- **Type:** broken filesystem access control / link following (CWE-59).
- **Locations:** `internal/domain/file_manager.go:157-169`; invocation at `internal/core/chatter.go:192-205`.
- **Evidence:** both a project-local file symlink to an outside marker and a directory symlink containing that marker passed `filepath.IsLocal` and caused `ApplyFileChanges` to overwrite the marker. All files were synthetic and under disposable test storage.
- **Impact/preconditions:** a symlink must exist in the selected project and the caller must take the CLI-like file-application branch with model output naming it. Writes can affect outside files writable by the process, including configuration or executable content. OS permissions still apply; elevation beyond process privileges was not reproduced. Current REST chat is blocked from this branch by the new gate.
- **Attribution:** pre-existing, still unresolved by the PR's lexical containment checks; not a newly introduced vulnerability. This is the security classification of the same observation in [[pr-2107-correctness]], not a separate duplicate defect.
- **Recommendation:** use operations rooted in a project directory handle that enforce containment during path resolution, including intermediate/final symlinks and races. Reuse existing storage validation patterns where appropriate, but a separate `EvalSymlinks` check followed by ordinary writes does not eliminate check/use races. Add rejection regressions for both link types and retain valid nested-path coverage.

## Conditional Medium: the shared key does not isolate users or read-only clients

- **Type:** object/function authorization limitation (CWE-639/CWE-862 if deployed with distinct user permissions).
- **Locations:** `internal/server/auth.go:40-65`, `internal/server/storage.go:57-64`, `:69-97`, `:113-142`, `internal/server/configuration.go:18-26`, `internal/server/sessions.go:15-18`.
- **Evidence:** a request with the configured key read a synthetic history named `other-user`, then deleted it. Handlers use the name and shared storage directly; authentication yields no user principal or role. Source also shows configuration updates and storage writes share the same middleware.
- **Impact/preconditions:** distributing this key to mutually untrusted users or ostensibly read-only clients gives them access to other histories and administrative mutation functions. In the intended single-owner/shared-trust deployment, these operations are authorized and this is not an established IDOR defect. No anonymous access was shown.
- **Attribution:** pre-existing deployment limitation, not a PR regression. Severity is conditional on a requirement for separate user permissions.
- **Recommendation:** keep one server instance/key per trust domain and explicitly document its full authority. If multi-user access is required, establish principals, scoped administrative permissions, and server-side ownership checks for every named object; opaque/random names alone do not enforce authorization.

## Low: file-write authority is inferred from an output channel

- **Locations:** `internal/core/chatter.go:192`, `internal/server/chat.go:139-151`.
- **Evidence:** identical generated content writes with a nil channel and skips writes with a nonnil channel, independently of streaming mode. Graph caller tracing found REST `HandleChat` and CLI callers; the REST caller sets the channel.
- **Assessment:** this convention correctly protects the current REST call path. A future server caller omitting the channel would enable writes without an explicit authorization decision. The channel is not a request-controlled API permission, and no current bypass was found.
- **Recommendation:** make file-application authority an explicit, default-denied execution option granted by trusted CLI entry points. Keep output transport choices separate from authorization. Treat this as hardening, not a proven current exploit.

## Validation and limits

Used the existing indexed code graph for endpoint registration, middleware, storage resolution, and inbound `Send` caller tracing; cross-checked original commit diff and source line numbers. No root `CLAUDE.md` or `AGENTS.md` exists; supplied discovery guidance applies. Server and filesystem-storage files inspected here are unchanged by the PR.

Three temporary test functions passed: four core gate cases, two symlink escape cases, and a route/shared-key probe (52 rejected requests across 26 routes plus authenticated history read/delete). The first probe run failed because its fresh synthetic database lacked `.env`; creating an empty test-only environment file fixed the fixture, and all probes then passed. No live provider, credentials, network listener, or real user data was used. Passing characterization probes confirm observations, including the residual defect, rather than security remediation.

Retained sources in the authorized Auto Run `Working` folder:

- `review_access_core_probe_test.go`
- `review_access_domain_probe_test.go`
- `review_access_server_probe_test.go`

To reproduce, temporarily copy each source into its matching `internal/core`, `internal/domain`, or `internal/server` package and run `go test ./internal/core ./internal/domain ./internal/server -run TestReviewAccessControl -v`, then remove those copies. The core probe reuses the package's existing mock vendor. Synthetic storage and build cache were placed in the authorized Working folder.

After removing temporary source copies, `go test ./internal/server ./internal/core ./internal/domain ./internal/plugins/db/fsdb` passed on Go 1.27.1, darwin/arm64. No production code or permanent tests changed. REST registration was inspected directly; the route enumeration test used the actual Ollama engine, and gate probes exercised core behavior rather than a live end-to-end REST/provider session. Full-repository tests, Windows behavior, deployed proxies, filesystem race exploitation, and other security checklist categories are outside this task. Remaining playbook checkboxes remain pending.
