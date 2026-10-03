---
type: report
title: PR 2107 Injection Vulnerability Review
created: 2026-10-03
tags:
  - code-review
  - security
  - injection
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-security-context]]'
  - '[[pr-2107-correctness]]'
  - '[[3_CHECK_SECURITY]]'
---

# PR 2107 injection review

Completed only the **Injection vulnerabilities** checkbox. Reviewed the original base `ddf1aab968caa9adf90137a6536c768d02b237d6` and PR head `f32656b529a1eb1c678c793703fe49a0755cd439`. The three changed files are `internal/core/chatter.go`, `internal/domain/file_manager.go`, and `internal/domain/file_manager_test.go`. Also inspected the template-extension execution path because the PR title explicitly claims shell-injection remediation. Production code is unchanged by this review. No task images were supplied or analyzed (0).

## Category assessment

| Category | Result within the review boundary |
| --- | --- |
| SQL injection | No SQL query construction or database execution is added or modified in the three-file diff. The existing `Examples/sqlite3_demo.yaml:17` interpolates a value into a SQL command inside shell double quotes; this is an example configuration outside the diff and illustrates the shell-context finding below. SQL injection against that example was not separately reproduced; shell escaping is not a substitute for SQL parameter binding. |
| Command injection | Confirmed residual High risk in template-extension execution when an installed command template places a user-controlled placeholder inside double quotes. Unquoted standalone placeholders resisted the tested payloads. |
| NoSQL injection | No NoSQL query construction or execution in the changed files; no new sink identified in the inspected flow. |
| LDAP injection | No LDAP filter construction or execution in the changed files; not applicable to this diff. |
| XPath injection | No XPath expression construction or evaluation in the changed files; not applicable to this diff. |

These conclusions concern the PR and its relevant template execution path, not a repository-wide injection audit. Model-generated file writes and lexical path validation are filesystem boundaries; their containment problems are documented in [[pr-2107-correctness]] and remain for the access-control review.

## High: shell escaping fails inside double-quoted template arguments

- **Type:** command injection (OWASP Injection; CWE-78).
- **Locations:** `internal/plugins/template/extension_executor.go:52`, `:78`, `:83`, and `:86`; representative existing configuration at `internal/plugins/template/Examples/sqlite3_demo.yaml:17`.
- **Attribution:** pre-existing in the base and unchanged by the PR. The original diff contains no template execution changes, so it does not substantiate the claim that shell-based execution was removed.

`formatCommand` wraps `value` and numbered values in shell single quotes, substitutes them into an arbitrary command template through `ApplyTemplate`, and passes the final string to `exec.Command("sh", "-c", cmdStr)`. This quoting protects standalone unquoted placeholders, but cannot provide protection independent of the surrounding template syntax. Within shell double quotes, the inserted single quotes are ordinary characters; command substitution still runs.

The following harmless reproduction uses the real registry and executor, with `/usr/bin/printf` as the registered executable:

```text
Command template: {{executable}} '%s' "{{value}}"
User value:       $(printf REVIEW_INJECTED)
Formatted shell:  /usr/bin/printf '%s' "'$(printf REVIEW_INJECTED)'"
Observed output:  'REVIEW_INJECTED'
```

A literal output would contain `$(printf REVIEW_INJECTED)`. Instead, the inner `printf` executed. The same result was reproduced for `{{1}}`, and backtick command substitution executed for `{{value}}`. With standalone unquoted `{{value}}` and `{{1}}`, substitution syntax remained literal, including a semicolon payload in the value control case.

**Preconditions and impact:** an installed extension must use a vulnerable quoting context, and an attacker must control a value reaching that extension. The shipped SQLite example already uses that context. Such a value can execute commands with the Fabric process's privileges. This is a conditional execution vulnerability, not evidence that every installation or every remote request is exploitable. Graph-backed traces establish `BuildSession` → `ApplyTemplate` → `ProcessExtension` → `Execute`; `Send` builds the session before the new automatic-file-application gate. That gate does not guard template execution. No remote end-to-end exploit was attempted.

**Remediation:** represent executable and arguments separately and pass expanded values as literal argument elements to `exec.Command`, without `sh -c`. Reuse the extension registry/executor and migrate command-template configuration deliberately. Do not use `strings.Fields` as an argument parser: it mishandles quoted values and is currently only an empty-command check. If shell templates must remain temporarily supported, reject unsupported placeholder contexts or confine dynamic data to a separate literal-data channel rather than interpolating it into shell source. Treat executable/template definitions as trusted configuration. Database extensions should bind SQL parameters independently of shell handling.

**Regression coverage needed for a fix:** assert literal argument preservation and absence of command execution for standalone, double-quoted, and single-quoted placeholders; numbered values; spaces; apostrophes; dollar substitutions; backticks; semicolons; and newlines. The existing `TestExtensionExecutor/ShellInjectionBlocked` covers a semicolon in an unquoted placeholder, which does not detect this context-dependent bypass.

## Validation

- Used the existing indexed codebase graph for symbol discovery, source snippets, injection-sink search, and inbound call tracing. No root `CLAUDE.md` or `AGENTS.md` was present; session-supplied graph-discovery instructions apply.
- Confirmed no production source changes between original PR head and this checkout. Original base/head diff also confirms the template files are unchanged.
- Created and ran five temporary characterization tests using the existing registry and executor. `go test ./internal/plugins/template -run TestReviewInjectionProbes -v` passed all five: two literal controls and three observed command-substitution cases. Passing characterization tests establish the defect; they do not certify secure behavior.
- Retained probe source at `/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2107-codex/Working/review_injection_probe_test.go`, removed its temporary copy from the source tree, then ran `go test ./internal/domain ./internal/core ./internal/plugins/template`; all three packages passed. No permanent test encodes vulnerable behavior as expected behavior.
- To reproduce, temporarily copy the retained probe into `internal/plugins/template`, run the focused test, then remove the copy. This runner used Go 1.27.1 on darwin/arm64; test temporary storage and Go build cache were under the authorized Auto Run Working folder. No network payload or destructive command was used.
- Full repository, Windows, remote API, and database exploit tests were not run. Authentication, sensitive-data exposure, XML, access control, and the other remaining playbook checkboxes were not completed by this task.
