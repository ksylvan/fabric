---
type: analysis
title: PR 2330 Injection Review
created: 2026-11-11
tags:
  - code-review
  - security
  - injection
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-security-context]]'
  - '[[pr-2330-data-flow-review]]'
---

# Injection review

Completed only the first unchecked security task: SQL, command, NoSQL, LDAP, and XPath injection checks. Reviewed fixed base `de0b72766db0436b6c1887bb75768cb3960e67e5` to head `fb296d490cca4fdebbcb78d5107241f134152d19`. Verified all 25 changed working-tree files match the head snapshot. No new exploitable injection vulnerability in these five categories was identified in the Fabric code. This conclusion does not cover Keenable's server implementation.

Used the indexed knowledge graph for shared settings and CLI symbol discovery. The Keenable file search returned no nodes, so direct source/diff inspection supplied the missing implementation; graph omissions were not treated as absence of code. No root AGENTS.md or CLAUDE.md file was present; the session-provided AGENTS.md instructions were followed.

## Reviewed sinks and boundaries

| Category | Evidence | Result |
| --- | --- | --- |
| SQL | `internal/tools/keenable/keenable.go:64` uses json.Marshal; `internal/cli/tools.go:129` passes the flag as one string. Registry persistence delegates to the existing filesystem-backed configuration store. | No new SQL statement construction, query execution, or database schema changes. SQL metacharacters remain query data. |
| Command | `keenable.go:78` constructs an HTTP POST to one of two fixed URLs; `keenable.go:90` executes it through net/http. Changed completions add literal option descriptions and value handling. | No new shell, exec, eval, or command substitution sink consuming the query/results. The caller's shell must still quote CLI arguments normally; expansion before Fabric starts is outside this implementation. |
| NoSQL | `keenable.go:64` marshals query as a string alongside constant integer max_results. | Object/operator-looking input cannot become a JSON object or add request fields. No new NoSQL driver calls. |
| LDAP | Same string-to-JSON boundary. | No LDAP filter construction or LDAP calls added. |
| XPath | Response decoding is encoding/json into typed strings at `keenable.go:103`. | No XML parser or XPath expression construction added. |

Server result fields are formatting arguments to a fixed format string at `keenable.go:117`; percent directives and shell syntax in those fields remain text. Returned URLs are emitted as text and are not fetched by Search. Error bodies are also arguments to repository-controlled localized format strings. Formatting is not a shell execution sink. Third-party Markdown and downstream prompt instructions remain untrusted; XSS and prompt-boundary concerns are handed to later review rather than declared safe by these checks.

## Conditional hardening: shared environment serialization

**Severity: Low (local configuration integrity; pre-existing shared behavior).** The new integration at `internal/core/plugin_registry.go:181` calls SetupFillEnvFileContent for Keenable. The inherited serializer at `internal/plugins/plugin.go:198` writes nonempty string values verbatim. With a locally supplied key `synthetic\nFABRIC_REVIEW_INJECTED=yes`, serialization produces two assignments and godotenv.Unmarshal interprets the second as a separate variable. `internal/plugins/db/fsdb/db.go:94` loads configuration through godotenv rather than shell execution. No automatic command execution was demonstrated.

An attacker must already influence the API-key configuration value before it is saved; the remote search response and search query do not populate that setting. This is not a newly introduced remote vulnerability or a merge blocker. As shared hardening, reject CR/LF in single-line credentials or use the existing godotenv.Marshal-based persistence pattern to quote values and test round trips. Do not add a separate Keenable-only serializer. Record this observation for the later consolidated security report.

## Validation

Created review-only `Working/keenable_injection_review_test.go` and `Working/injection-review-overlay.json` under the Auto Run folder, following the existing overlay convention. Production code and committed application tests are unchanged.

Six hostile query inputs (SQL, shell substitution/pipes, NoSQL operators, LDAP filter syntax, XPath syntax, and quotes/backslashes/control characters/Unicode) ran in both keyless and keyed modes. Tests assert exact decoded query equality, exactly two request fields, fixed max_results, unchanged endpoint paths, no URL query or injected header, literal result formatting, and one handler call per search. The returned loopback URL is only text by source inspection; no OS command or database backend is invoked by the probe. A separate test confirms the conditional dotenv behavior above with synthetic values only.

All complete package suites below, including the new review overlay, passed:

```sh
go test -race -overlay /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/injection-review-overlay.json ./internal/tools/keenable ./internal/plugins ./internal/cli -count=1
```

No live API, credentials, model calls, or images were used. The full repository suite was not run; validation targets the changed search path and shared CLI/settings behavior. Other security checkboxes, including authentication, secret scanning, XSS, and consolidated SECURITY_ISSUES.md creation, remain for subsequent tasks.
