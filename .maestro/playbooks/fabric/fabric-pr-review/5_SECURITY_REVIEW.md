# Document 5: Security Review

## Context

- **Playbook**: Fabric PR Review
- **Agent**: Fabric-PR-Review
- **Project**: /Users/kayvan/src/fabric
- **Date**: 2026-04-06
- **Working Folder**: /Users/kayvan/src/fabric/.maestro/playbooks

## Purpose

Perform a security-focused review of the code changes, checking for vulnerabilities specific to Fabric's architecture.

## Prerequisites

- `/Users/kayvan/src/fabric/.maestro/playbooks/REVIEW_SCOPE.md` exists from Document 1

## Tasks

### Task 1: Load Context

- [x] **Read scope**: Loaded `/Users/kayvan/src/fabric/.maestro/playbooks/REVIEW_SCOPE.md` and identified the primary security review targets as `internal/tools/youtube/youtube.go`, `internal/cli/cli.go`, `internal/cli/flags.go`, and `internal/cli/help.go`.
  - Note: `internal/tools/youtube/youtube.go` is the highest-risk file because it introduces external process execution (`yt-dlp`, `ffmpeg`, `tesseract`), temp-file handling, and concurrent OCR work.

### Task 2: Check for Secrets and Credentials

- [x] **Hardcoded secrets**: Searched the PR diff and changed files for provider key signatures, bearer tokens, private key headers, and connection-string patterns. No hardcoded credentials were added in runtime code.
  - Note: False positives were limited to synthetic redaction fixtures in `internal/cli/flags_test.go` (`super-secret-server-key`) and `internal/tools/youtube/youtube_logging_test.go` (`super-secret-password`), plus secret-keyword redaction logic in `internal/tools/youtube/youtube.go`.
  - API keys (OpenAI, Anthropic, Gemini, etc.)
  - Passwords or tokens
  - Private keys
  - Connection strings
  - OAuth credentials

- [x] **Config file handling**: Verified that Fabric still sources secrets from `~/.config/fabric/.env` via `fsdb.LoadEnvFile()` and consumes them through environment variables instead of hardcoded runtime values. Confirmed the reviewed PR files only contain synthetic secret fixtures in tests, and verified `.gitignore` still excludes repo-local secret files.
  - Note: Hardened the user config path so `~/.config/fabric` is normalized to `0700` and `~/.config/fabric/.env` to `0600`, including repair of pre-existing insecure permissions during setup and DB initialization.
  - Secrets loaded from `~/.config/fabric/.env`
  - Environment variables used, not hardcoded values
  - No secrets in code comments
  - `.gitignore` excludes sensitive files

### Task 3: Check OWASP Top 10

- [x] **Injection vulnerabilities**: Reviewed the PR's command-execution and request-construction surfaces. No direct shell, SQL, template, LDAP, or XPath injection path was introduced by the YouTube visual extraction changes.
  - Note: `internal/tools/youtube/youtube.go` invokes `yt-dlp`, `ffmpeg`, and `tesseract` through `exec.Command` / `exec.CommandContext` with discrete argv entries, so video IDs, URLs, OCR text, and CLI flags are not interpolated by a shell.
  - Note: Hardened the `--yt-dlp-args` passthrough to always add `--ignore-config` and reject `--exec`, `--exec-before-download`, `--config-locations`, `--plugin-dirs`, and `--alias`, closing the delegated-command-execution path that `yt-dlp` itself could otherwise expose.
  - Note: Added regression coverage for both the shared argument validator and the `GrabVisual()` call path so subprocess-expanding `yt-dlp` options are rejected before any child process starts.
  - Note: No database or server changes in this PR construct SQL, LDAP, or XPath expressions from user-controlled input.
  - Note: The PR does not add any new pattern/template evaluation step; existing pattern-variable handling remains unchanged from the pre-PR code path.
  - Command injection (shell commands with user input)
  - SQL injection (for database plugins)
  - Template injection (pattern variable handling)
  - LDAP/XPath injection (if applicable)

- [x] **Sensitive data exposure**: Reviewed the new logging and error surfaces in YouTube visual extraction, YAML config loading, and chat handling; no remaining API key, signed URL, or internal-path exposure was found after hardening the newly introduced debug/error paths.
  - Note: Hardened `internal/cli/flags.go` so detailed debug logging now reports only configured/applied YAML key names, never the absolute config path or YAML field values, which closes the debug-mode secret leakage path for config-backed settings.
  - Note: Hardened `internal/tools/youtube/youtube.go` so transcript/VTT filesystem failures strip temp-file paths before surfacing errors, preventing transient OCR/transcript failures from revealing Fabric's internal temp-directory layout.
  - Note: Revalidated that yt-dlp trace logging still redacts URLs, cookies, authorization headers, and token-like query parameters via `sanitizeYTArgs()` / `sanitizeYTLogText()`.
  - Note: Reviewed `internal/server/chat.go`; request logging remains debug-level only and logs language/count/model/pattern/context metadata without prompt bodies or secrets.
  - Note: Added regression coverage for both the config-loader log summary and the YAML-application path in `Init()`, plus the existing transcript-file read failure case, then reran `go test ./internal/tools/youtube ./internal/server ./internal/cli ./internal/core ./internal/plugins/db/fsdb`.
  - API keys not logged
  - No PII in debug output
  - Errors don't expose internal paths
  - Responses don't leak system info

- [x] **Broken access control**:
  - Note: Reviewed the main REST server path in `internal/server/serve.go` and revalidated that all non-Swagger routes remain behind `APIKeyMiddleware` whenever an API key is configured; added regression coverage for protected routes vs. the intentional `/swagger/*` exemption.
  - Note: Hardened `internal/plugins/db/fsdb/storage.go` and the REST storage handlers so context/session/pattern names reject traversal-style identifiers (`..`, absolute paths, path separators) and symlinked entries that resolve outside their configured storage roots.
  - Note: Hardened REST pattern resolution so `/patterns/:name/apply` and `/chat` only resolve configured pattern names; HTTP callers can no longer coerce `GetApplyVariables()` into reading arbitrary local files such as `.env`, while the CLI keeps its documented file-path pattern support.
  - Note: Revalidated custom-pattern precedence and tightened the lookup boundary in `internal/plugins/db/fsdb/patterns.go`, so custom patterns can still intentionally override same-name built-ins but cannot escape the configured custom/main pattern directories via crafted names or symlinked directory aliases; invalid custom symlink entries now fall back to the built-in pattern instead of crossing the boundary.
  - Note: Added regression coverage in `internal/server/auth_test.go`, `internal/server/patterns_test.go`, `internal/server/storage_handler_test.go`, `internal/plugins/db/fsdb/storage_test.go`, and `internal/plugins/db/fsdb/patterns_test.go`, then reran `go test ./internal/plugins/db/fsdb ./internal/server`.
  - Server endpoints check authorization
  - No path traversal vulnerabilities
  - Pattern loading respects boundaries
  - Custom patterns isolated from built-in

- [x] **Security misconfiguration**:
  - Note: Revalidated `internal/cli/flags.go` and kept debug mode off by default (`--debug=0`) with no default API key or bootstrap credentials added anywhere in the reviewed server path.
  - Note: Hardened server startup defaults so `--serve` now binds to `127.0.0.1:8080` by default, refuses non-loopback bind addresses unless `--api-key` is provided, and documents the explicit `0.0.0.0` + API key path for remote or Docker exposure.
  - Note: Hardened `internal/server/ollama.go` so Ollama compatibility mode is loopback-only until that surface has its own authentication story, closing the accidental unauthenticated wildcard-bind case.
  - Note: Hardened `internal/server/chat.go` CORS behavior so `/chat` only serves the documented local dev origin (`http://localhost:5173`) and now answers `OPTIONS` preflight requests without reflecting arbitrary origins.
  - Note: Added regression coverage in `internal/server/server_security_test.go`, `internal/server/chat_test.go`, and `internal/cli/flags_test.go`, then reran `go test ./internal/server ./internal/cli`.
  - Debug mode disabled by default
  - No default credentials
  - Secure defaults for server mode
  - CORS properly configured (if applicable)

- [ ] **Insecure deserialization**:
  - JSON/YAML parsing is safe
  - No unsafe unmarshaling of user input
  - Type checking before deserialization

### Task 4: Fabric-Specific Security

- [ ] **API key handling**:
  - Provider API keys stored securely
  - Keys not exposed in URLs or logs
  - OAuth tokens properly managed
  - Key rotation considerations

- [ ] **Pattern security**:
  - Custom patterns can't override system paths
  - Pattern variables are sanitized
  - No code execution via patterns
  - Template extensions are sandboxed

- [ ] **Server mode security**:
  - `/api/` endpoints are protected
  - Rate limiting considered
  - Input validation on all endpoints
  - SSE streams don't leak data

- [ ] **Ollama compatibility mode**:
  - Proper request validation
  - No privilege escalation
  - Secure proxy behavior

### Task 5: Check Third-Party Dependencies

- [ ] **New dependencies**: If `go.mod` or `go.sum` changed:
  - List new dependencies added
  - Check for known vulnerabilities
  - Verify license compatibility
  - Assess dependency maintainability

- [ ] **Dependency updates**: If existing deps updated:
  - Check changelog for security fixes
  - Verify no breaking changes
  - Test compatibility

### Task 6: Review Authentication Code

If auth-related code changed:

- [ ] **OAuth implementation**:
  - Token storage is secure
  - Refresh flow is correct
  - No token exposure in logs
  - Proper session handling

- [ ] **API key validation**:
  - Timing-safe comparison
  - Proper error messages (no info leak)
  - Rate limiting for auth failures

### Task 7: Document Security Issues

- [ ] **Create SECURITY_ISSUES.md**: Write findings to `/Users/kayvan/src/fabric/.maestro/playbooks/SECURITY_ISSUES.md`:

```markdown
# Security Review Findings

## Critical Vulnerabilities
[Immediate security risks - must block merge]

## High Risk Issues
[Significant security concerns]

## Medium Risk Issues
[Security improvements needed]

## Low Risk Issues
[Minor security hardening suggestions]

## Secrets/Credentials Check
- [ ] No hardcoded API keys found
- [ ] No exposed credentials
- [ ] Proper use of environment variables

## OWASP Compliance
- [ ] No injection vulnerabilities
- [ ] No sensitive data exposure
- [ ] Proper access control
- [ ] Secure configuration

## Fabric-Specific Security
- [ ] API key handling secure
- [ ] Pattern system secure
- [ ] Server mode secure

## Dependencies
[New or updated dependencies with security implications]

## Positive Observations
[Good security practices noted]

## No Issues Found
[Areas reviewed with no concerns]
```

For each issue include:
- Vulnerability type
- File and line number
- Description and potential impact
- Remediation recommendation
- Severity: Critical / High / Medium / Low

## Success Criteria

- All changed files reviewed for security
- OWASP categories checked
- Fabric-specific security verified
- No hardcoded secrets found (or flagged)
- SECURITY_ISSUES.md created

## Status

Mark complete when security review document is created.

---

**Next**: Document 6 will run tests and generate changelog.
