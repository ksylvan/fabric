---
type: report
title: PR 2107 Sensitive Data Exposure Review
created: 2026-10-03
tags:
  - code-review
  - security
  - sensitive-data
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-security-context]]'
  - '[[pr-2107-authentication]]'
  - '[[pr-2107-data-flow]]'
  - '[[3_CHECK_SECURITY]]'
---

# PR 2107 sensitive data exposure review

Completed only the **Sensitive data exposure** checkbox: secrets in the changed code, protection at rest, PII in logs, and secure transmission. Compared original base `ddf1aab968caa9adf90137a6536c768d02b237d6` with original PR head `f32656b529a1eb1c678c793703fe49a0755cd439`. No hardcoded secret or new transport/encryption implementation appears in that diff. Existing storage, logging, and transport gaps remain; the API file-application gate also retains more response content in named chat history. This is a review, with no production fixes. No images were supplied or analyzed (0).

## Assessment

| Check | Evidence and result |
| --- | --- |
| No secrets/keys in code | Inspected the entire original three-file diff and affected files. No credential literals, private keys, passwords, or connection strings were found there; new test data is synthetic path/content data. Provider configuration reads environment-backed settings, and configuration responses mask typical long keys. This is a scoped result, not a repository/history secret scan; dedicated secret checks remain pending. |
| Protection at rest | Named sessions are JSON containing full message text; `.env` holds plaintext configuration credentials. There is no application encryption in the inspected serialization/write paths. `SaveEnv` atomically writes credentials with `0600`; session writes request `0644`, and storage directories request `0777` before umask. Generated files request `0644` and directories `0755`. A restrictive ancestor directory, umask, encrypted volume, and backup policy can mitigate exposure but are not enforced by these paths. |
| No PII in logs | Not satisfied when wire diagnostics are enabled: request, streamed content, and response text are logged without redaction. Synthetic email markers were present in real `Send` diagnostics at Wire and absent at Off. Normal CLI streaming also prints model content to stdout unless suppressed; that is intended output, but redirected terminal/session capture can retain sensitive text. Standard access-log header-key behavior was already checked in [[pr-2107-authentication]]. |
| Secure transmission | REST and Ollama servers use Gin `Run(address)`, with no built-in TLS listener in these flows. Remote binding requires an API key, but that guard does not encrypt the key or payload. HTTPS termination must be supplied externally for network exposure. Inspected OpenAI/Anthropic defaults are HTTPS; Ollama/LM Studio defaults are loopback HTTP. Configurable provider base URLs and Ollama forwarding accept HTTP. No TLS verification bypass was found in the inspected provider setup. Deployed proxy/SDK internals and every provider were not audited. |

## Medium: sensitive named sessions can be readable by other local users

- **Type:** sensitive-data exposure through filesystem permissions (CWE-732/CWE-200).
- **Locations:** `internal/plugins/db/fsdb/storage.go:25-26`, `:103-108`, `:269-272`; `internal/plugins/db/fsdb/sessions.go:42-43`; `internal/plugins/db/fsdb/db.go:49-50`.
- **Evidence:** under controlled umask `022`, a synthetic email address was readable as plaintext in a newly saved session JSON file with mode `0644`; its session directory had mode `0755`. Source shows serialization without encryption. The probe used a private temporary ancestor and did not expose data to a real second user.
- **Impact/preconditions:** other users can read sensitive histories if all ancestor directories are traversable. Backups or copied configuration trees may also retain plaintext. A private home/configuration ancestor blocks this local read path; actual production permissions were not inspected. Plaintext storage alone does not establish a vulnerability on an appropriately protected local machine.
- **Recommendation:** specialize the existing session storage path to create session directories with `0700` and files with `0600`, and handle existing permissive files/directories. Define retention/deletion and backup protection. Use storage encryption when the deployment's confidentiality requirements demand it, with keys protected separately from the data.
- **Attribution:** pre-existing storage behavior. The PR now skips summary replacement for `create_coding_feature` when `UpdateChan` is nonnil (`chatter.go:192-209`), so the complete response, including generated file content, reaches the existing named-session save (`:212-215`). This is a retention consequence, not a demonstrated new unauthenticated disclosure; the content is already sent to the API client.

## Medium: network-facing HTTP can expose API keys and chat content

- **Type:** cleartext transmission of sensitive information (CWE-319).
- **Locations:** `internal/server/serve.go:79`, `internal/server/ollama.go:214-219`, `:500-521`; configurable provider example at `internal/plugins/ai/openai/openai.go:133-136`.
- **Evidence:** both server entry points use `Run`, not a TLS listener. Ollama URL construction explicitly accepts HTTP and defaults bare network hosts to HTTP. OpenAI client configuration passes the configured base URL without a local HTTPS restriction.
- **Impact/preconditions:** direct HTTP exposure across an untrusted network allows interception of `X-API-Key` and message data; custom remote HTTP provider URLs can similarly expose payloads and provider credentials. A secured loopback connection behind HTTPS termination is a valid mitigating deployment. No running deployment, intercepted credential, or live provider was tested.
- **Recommendation:** require documented TLS termination with restricted backend reachability for remote server access; reject or clearly require explicit opt-in for non-loopback HTTP provider/forwarding URLs. Reuse existing URL/configuration validation and retain local HTTP compatibility.
- **Attribution:** unchanged by the original PR; deployment-dependent existing risk.

## Low: wire diagnostics retain unredacted PII and credentials

- **Type:** sensitive information in logs (CWE-532).
- **Locations:** `internal/core/chatter.go:77-80`, `:120-121`, `:175-176`.
- **Evidence:** a probe invoked the real non-streaming `Send` with the existing mock vendor and synthetic request/response email markers. Both markers appeared at Wire and neither appeared at Off. Streaming logging was inspected in source rather than separately probed. Logging defaults to Off in `internal/log/log.go:28`.
- **Impact/preconditions:** wire logging must be enabled and messages must contain sensitive data; anyone with access to captured stderr can recover it. `%q` quoting is formatting, not redaction. This extends the credential-in-content observation in [[pr-2107-authentication]] to PII and should be consolidated as the same finding rather than counted twice.
- **Recommendation:** explicitly opt in to full content diagnostics, redact sensitive fields/recognizable identifiers, and restrict log permissions and retention. Metadata-only diagnostics should remain available.
- **Attribution:** pre-existing and outside the changed lines.

## Low: newly generated sensitive files inherit general source-file permissions

- **Type:** permissive permissions for potentially sensitive output (CWE-732).
- **Location:** `internal/domain/file_manager.go:164-168`.
- **Evidence:** a probe confirmed a new synthetic-content file becomes `0644` under umask `022`. Updating an existing `0600` file preserves that mode. There is no content-aware confidential-file policy in this writer.
- **Impact/preconditions:** generated content must include secrets or PII and the project ancestors must be accessible to another user. Ordinary source code may intentionally be shared; these permissions are not a universal defect. The PR's lexical path check does not address content confidentiality or encryption.
- **Recommendation:** provide an explicit confidential-output policy using the existing writer, with restrictive modes for secret-bearing outputs and user review before applying them. Preserve restrictive permissions on updates.
- **Attribution:** unchanged by this PR. Symlink overwrite/containment remains covered in [[pr-2107-correctness]] and is not relabeled as a demonstrated file-read vulnerability here.

## Existing credential protection and remaining limitations

`SaveEnv` uses the existing atomic helper (`db.go:97-104`, `:167-190`), creates a temporary file, applies `0600`, writes/syncs, and renames it. The probe confirmed the final mode and plaintext contents using a placeholder only. Owner-only access is a positive control, though it is not encryption and cannot protect against an attacker with the process owner's privileges. Existing environment files loaded without a save are not automatically permission-hardened by `LoadEnvFile` (`:83-87`).

Short credential masking at `internal/server/configuration.go:32-37` remains the conditional Low finding from [[pr-2107-authentication]]: keys of at most four characters are returned intact. Typical long keys expose only their last four characters. No real credential was loaded, printed, or sent during this review.

## Validation and limits

Used the existing indexed graph for symbol discovery, source snippets, and inbound session-save tracing; inspected exact source lines where graph results merged same-named storage methods or where literals/configuration values were relevant. No root `CLAUDE.md` or `AGENTS.md` exists; supplied graph instructions apply. Confirmed that storage, logging, server, and provider files are unchanged in the original PR range.

Three temporary characterization probes passed: session plaintext/permissions and `.env` protection; request/response PII in real `Send` wire logging; generated-file permissions and preservation of existing restrictive permissions. These tests characterize observed gaps and controls; passing does not mean the gaps were fixed. Retained formatted sources in the authorized Auto Run `Working` folder:

- `review_sensitive_storage_probe_test.go` — copy temporarily to `internal/plugins/db/fsdb`.
- `review_sensitive_logging_probe_test.go` — copy temporarily to `internal/core`; reuses the existing `mockVendor`.
- `review_sensitive_files_probe_test.go` — copy temporarily to `internal/domain`.

Run `go test ./internal/plugins/db/fsdb ./internal/core ./internal/domain -run TestReviewSensitive -v`, then remove the temporary copies. Permission probes use Unix `syscall.Umask` and were run on darwin/arm64 with Go 1.27.1; they are not Windows-compatible tests. Test data, temporary storage, and build cache stayed in the authorized Auto Run Working directory. After removing the copies, the existing filesystem-db, core, domain, and server suites passed. No permanent source/tests changed. Full repository/history secret scanning, all providers, production TLS/proxy configuration, storage encryption, backup access, and later security categories remain unassessed.
