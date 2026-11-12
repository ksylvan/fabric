---
type: analysis
title: PR 2330 Environment Variable Configuration Review
created: 2026-11-11
tags:
  - code-review
  - security
  - configuration
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-hardcoded-secrets-review]]'
  - '[[pr-2330-sensitive-data-review]]'
  - '[[pr-2330-injection-review]]'
---

# Environment variable configuration review

Reviewed the fixed PR snapshot `de0b72766db0436b6c1887bb75768cb3960e67e5` to `fb296d490cca4fdebbcb78d5107241f134152d19`. Sensitive Keenable configuration uses the shared environment/settings mechanism, with no embedded production credential. No additional finding identified. This completes only the environment-variables checkbox; application code is unchanged. Zero images analyzed.

## Configuration evidence

| Boundary | Evidence and assessment |
| --- | --- |
| Credential definition | `internal/tools/keenable/keenable.go:34-49` constructs an empty optional API-key setting. Its stable prefix comes from the literal `Keenable`, independently of the translated display name. `internal/plugins/plugin.go:60-65` constructs the environment name and marks API-key settings sensitive. The resulting name is `KEENABLE_API_KEY`; no key default is embedded. |
| Environment loading | `internal/plugins/plugin.go:304-310` reads `os.Getenv` when Configure runs. The registry creates Keenable at `internal/core/plugin_registry.go:65` and configures it at `:531`. Construction alone does not read the environment. The shared dotenv loader is `internal/plugins/db/fsdb/db.go:93-98`; explicit process values take precedence over dotenv, including an explicitly empty variable. |
| Request selection | `internal/tools/keenable/keenable.go:71-86` trims the configured value, selects the keyed endpoint only for a nonempty key, and sends it in `X-API-Key`. Empty/whitespace values use public search. Both initial endpoints are fixed HTTPS URLs at `:22-23`, without embedded credentials. Service URLs, timeout, result limit, and client title are ordinary constants rather than sensitive configuration. |
| Setup and persistence | Registry setup exposes Keenable as optional (`internal/core/plugin_registry.go:342`); saving calls the inherited serializer at `:181`. Nonempty settings become `KEENABLE_API_KEY=...` through `internal/plugins/plugin.go:198-212`. Shared SaveEnv uses an atomic private file (`internal/plugins/db/fsdb/db.go:106-113`). The key remains plaintext, as already recorded in [[pr-2330-sensitive-data-review]]. |
| Version-control boundary | Root `.gitignore` excludes `.env`; the PR changes no ignore rules or environment files. The prior additions scan found no apparent real credential; see [[pr-2330-hardcoded-secrets-review]]. No actual user environment values or credential files were inspected. |

A fresh client with no environment key has an empty setting and works without authentication. An explicitly empty process variable prevents dotenv from supplying its value, so a fresh client stays keyless. This is not a general credential-reset guarantee: the shared Setting.Configure ignores empty environment values and preserves an already populated in-memory Value. Changing environment variables does not automatically reconfigure existing clients. These are existing shared semantics, not a new hardcoded-key finding.

## Existing limits to carry forward

Environment-variable use does not establish encryption, secrecy from the same user/process, or safe transport after redirects. Preserve the conditional Medium redirect disclosure and Low echoed-error/logging observations from the earlier reviews without duplicating them here. Private file permissions do not encrypt the saved key or protect shared backups.

The shared string serializer writes values verbatim. Locally controlled newlines can create additional dotenv assignments; [[pr-2330-injection-review]] already records this as Low conditional configuration-integrity hardening. Reuse the existing quoted dotenv persistence pattern or validate single-line credentials if this is addressed. No remote response-to-setting path was identified. A new Keenable-only loader/serializer is unnecessary.

## Validation

Used the code graph to locate constructor, search, shared setting configuration, serialization, and dotenv loading. Graph snippets for Configure were ambiguous across receiver types, so inspected the shared file for the exact Setting method. No AGENTS.md or CLAUDE.md files were found in repository discovery; supplied session instructions apply. Relevant implementation files match the fixed head.

Reused prior `Working/keenable_logic_review_test.go` optional-key cases (empty, whitespace, and padded key) and `Working/keenable_sensitive_data_review_test.go` storage/masking checks. Added `Working/keenable_environment_review_test.go` with three dotenv/process-precedence cases. Each new case saves a synthetic key in a temporary private env file, loads it, configures a fresh client, checks the environment name/optional/sensitive metadata and serialization, and verifies the selected request path/header through a loopback server. All use synthetic values and restore the process environment. No live provider/model requests or actual credentials used.

All four complete targeted package suites passed with the race detector (shared plugins and fsdb were cached):

```sh
TMPDIR=/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/tmp go test -race -timeout 90s -overlay /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/environment-review-overlay.json ./internal/tools/keenable ./internal/plugins ./internal/plugins/db/fsdb ./internal/core
```

The overlay composes the existing sensitive-data review fixtures with the new precedence tests. Probes/overlay remain in the external Auto Run Working folder; reproduction requires that folder. The full repository suite was not run. Auth-logic, authorization, and consolidated-report tasks remain for subsequent runs.
