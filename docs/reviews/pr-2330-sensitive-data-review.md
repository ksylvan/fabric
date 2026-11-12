---
type: analysis
title: PR 2330 Sensitive Data Exposure Review
created: 2026-11-11
tags:
  - code-review
  - security
  - privacy
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-authentication-review]]'
  - '[[pr-2330-data-flow-review]]'
---

# Sensitive data exposure review

Reviewed the fixed PR snapshot from `de0b72766db0436b6c1887bb75768cb3960e67e5` to `fb296d490cca4fdebbcb78d5107241f134152d19`. Relevant implementation directories match that head. This completes only the sensitive-data-exposure task; production code is unchanged. No images were supplied or analyzed.

## Assessment

| Requested check | Evidence and result |
| --- | --- |
| No secrets/keys in code | Inspected additions across all 25 changed files, including locale JSON, completions, README, and tests. No apparent real credential, private key, connection string, or token was introduced. `secret` in the keyed-request test is an obvious synthetic fixture. Production code obtains the optional key through shared settings/environment configuration; it does not embed one. This is a scoped review of PR additions, not a full-history secret scan or proof of credential validity. |
| Encryption for sensitive data | Both initial API endpoints are fixed HTTPS URLs (`internal/tools/keenable/keenable.go:22-23`); the client uses normal certificate validation. The query is in the JSON POST body and the key in `X-API-Key`, not URL parameters (`:64-86`). Local setup serializes the key as plaintext using the existing dotenv mechanism. `SaveEnv` writes atomically with mode `0600` (`internal/plugins/db/fsdb/db.go:106-113`); configuration directories are created with `0700` (`:51`). Permissions restrict access but do not encrypt data or protect it from the same user, backups, or a compromised host. Existing directory permissions are not tightened by MkdirAll. No new at-rest encryption mechanism is added. |
| No PII in logs | Search adds no direct request/key/query logger. Sensitive setting defaults are masked by `DisplayValue` (`internal/plugins/plugin.go:171-176`). However raw non-200 response bodies enter CLI errors and can expose echoed PII/keys. Also existing opt-in wire logging prints complete model messages containing tool results (`internal/core/chatter.go:102-105`). Therefore this check cannot be marked unconditionally clean. |
| Secure transmission | Untrusted TLS certificates are rejected in a local probe. The default redirect policy still permits off-origin redirects and HTTPS-to-HTTP downgrade. Prior probes reproduce key forwarding and query-body replay on 307/308. HTTPS on the initial request does not protect subsequent redirected requests. |

## Findings to carry into the consolidated report

### Medium: redirects disclose credentials and potentially private queries

**Type:** sensitive data disclosure / OWASP A02 cryptographic failures when downgraded (CWE-201, CWE-319).

**Location:** `internal/tools/keenable/keenable.go:29`, `:86`, `:90`.

This is the same Medium finding in [[pr-2330-authentication-review]], not a second issue. A redirect issued by the trusted service or intermediary can forward `X-API-Key` to an unrelated origin, including HTTP; 307/308 also replay the query body. Control of that redirect response is a required precondition. No live service compromise or actual disclosure was observed. Reject redirects for this API, or restrict them to the original HTTPS origin and port with a finite hop limit. Removing the key alone does not protect the body. Retain tests that ensure rejected destinations are never contacted.

### Low: errors retain sensitive service-echoed content

**Type:** sensitive information in diagnostics (CWE-209).

**Location:** `internal/tools/keenable/keenable.go:96-98`, propagated through `internal/cli/tools.go:129-130` and `internal/cli/cli.go:97-99`.

This duplicates the earlier error/authentication finding. If a service/gateway response echoes credentials or a private query, the complete body survives in the returned error and can reach stderr or captured logs. A generic rejection does not expose the key. Prefer bounded, selected status diagnostics and redact known credentials; do not include private request text. The existing synthetic echoed-error probes reproduce this conditional exposure.

### Low hardening observation: private result text reaches opt-in wire logs

**Type:** sensitive information in logs (CWE-532), existing shared behavior newly reachable with Keenable results.

**Location:** `internal/cli/tools.go:132`, `internal/cli/chat.go:20-23`, `internal/core/chatter.go:102-105`, `internal/log/log.go:28-30`.

Keenable returns titles, URLs, descriptions/snippets without PII redaction. Chat mode appends them to model input, and wire logging writes full model-message content to stderr. If results contain personal/private text, enabling wire logging exposes that text to whoever can read captured diagnostics. Debug logging defaults to Off; this is an opt-in existing behavior, not a new unconditional logger. Document this privacy boundary and prefer metadata-only or explicitly redacted diagnostics where logs are retained/shared. The API key is not deliberately appended to model input. Standalone result output and requested session persistence are intentional destinations for content, not additional independent logging vulnerabilities.

## Privacy and storage boundaries

The CLI sends the exact search query, including any private identifiers a user supplies, to Keenable in both public and keyed mode. Public means no API key is required; it does not establish anonymity or local execution. In chat mode results also become input to the selected model, and existing session/output options can persist content. There is no automatic fetching of each result URL. Provider-side retention, encryption at rest, access policies, and model transport configurations cannot be established from this PR; no claim of end-to-end encryption or server-side deletion is made.

Plaintext local key storage is a pre-existing shared design with private file permissions, recorded as a limitation rather than a newly introduced encryption regression. For a stronger threat model, use an OS credential store and exclude plaintext configuration from shared backups/support bundles. Stored-default masking also does not establish hidden terminal input; shared interactive input behavior was unchanged.

## Validation and scope limits

Reused the existing review transport, redirect, echoed-error, and CLI data-flow probes through `Working/auth-review-overlay.json`. Added `Working/keenable_sensitive_data_review_test.go` and `Working/sensitive-data-review-overlay.json` in the external Auto Run folder. New tests check plaintext serialization plus private permissions when replacing an existing `0644` env file, sensitive-setting masking, rejection of an untrusted TLS certificate before the HTTP handler is called, and successful TLS with an explicitly trusted local test certificate. The success case checks query/body and key/header placement without query URL parameters. Only synthetic credentials and a reserved `.invalid` email were used. TLS requests stay on loopback; redirect/error probes use in-process transports and no external service calls.

All six full targeted package suites passed with the race detector on Go 1.27.3:

```sh
TMPDIR=/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/tmp go test -race -timeout 90s -overlay /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/sensitive-data-review-overlay.json ./internal/tools/keenable ./internal/plugins ./internal/plugins/db/fsdb ./internal/cli ./internal/core ./internal/i18n
```

Review probes and overlay files remain in the external Working folder as in prior tasks; reproducing this command requires that folder. The full repository suite was not run. No production helper or dependency was created. Dedicated secret scanning, XML/XXE, access control, and subsequent checklist items remain unchecked. The later consolidated report should deduplicate the Medium redirect and Low echoed-error issues and include the shared opt-in logging observation with its preconditions.
