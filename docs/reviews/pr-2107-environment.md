---
type: report
title: PR 2107 Environment Variable Review
created: 2026-10-03
tags:
  - code-review
  - security
  - configuration
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-hardcoded-secrets]]'
  - '[[pr-2107-sensitive-data]]'
  - '[[pr-2107-authentication]]'
  - '[[3_CHECK_SECURITY]]'
---

# PR 2107 environment variable review

Completed only the **Environment variables** checkbox. Sensitive provider configuration is supplied through environment-backed settings, setup-generated `.env` data, or external credential providers rather than identified hardcoded operational credentials. The server key supports the inherited process environment. No new sensitive-configuration regression was established in the original PR. Two inherited configuration hardening concerns are described below. Production code is unchanged; no images were analyzed (0).

## Scope and evidence

Review boundary: base `ddf1aab968caa9adf90137a6536c768d02b237d6` to original head `f32656b529a1eb1c678c793703fe49a0755cd439`. That diff changes only the file-application gate, path validation, and path tests; it changes no credential configuration or dependency inputs. Reused [[pr-2107-hardcoded-secrets]] for literal-secret evidence rather than repeating its scan.

Used the indexed code graph to find shared settings, environment loading, flag definitions, initialization and OAuth configuration, and trace callers of `LoadEnvFile`. The graph conflates same-named Go receiver methods in `plugin.go` and did not return Spotify symbols; direct source inspection supplemented those insufficient results. No tracked `AGENTS.md` or `CLAUDE.md` was found during orientation. Live environment values, local credential files, and provider accounts were not inspected.

| Configuration | Source and evidence |
| --- | --- |
| Shared provider settings | `internal/plugins/plugin.go:41-47` constructs vendor prefixes; `:56-60` adds settings; `:283-290` reads nonempty `os.Getenv` values and validates requirements. Empty required credentials are rejected; constructors do not supply an operational fallback key in inspected paths. |
| API keys | OpenAI (`internal/plugins/ai/openai/openai.go:32`), Gemini (`internal/plugins/ai/gemini/gemini.go:54`), and OpenAI-compatible providers use shared settings. Representative probes confirm `OPENAI_API_KEY` mapping and ingestion without network calls. Optional keys for local providers are intentional, not default credentials. |
| OAuth client secrets/tokens | Copilot declares Client Secret, Access Token, Refresh Token settings (`internal/plugins/ai/copilot/copilot.go:66-72`). Spotify uses stable `SPOTIFY_CLIENT_ID` / `SPOTIFY_CLIENT_SECRET` settings (`internal/tools/spotify/spotify.go:52-53`); its secret mapping and ingestion were probed. |
| Codex OAuth | `internal/plugins/ai/codex/codex.go:83-85` declares environment-backed token/account settings. OAuth results populate them (`:132-134`), and `LoadEnvSettings` (`:174-186`) loads saved values. The constant OAuth client ID (`:36`) is a public application identifier, not a client secret. No separate token-file audit was performed. |
| Bedrock | `internal/plugins/ai/bedrock/bedrock.go:195-202` declares `BEDROCK_API_KEY`, `BEDROCK_AWS_ACCESS_KEY_ID`, and `BEDROCK_AWS_SECRET_ACCESS_KEY` settings. `:378-407` selects bearer credentials, configured static credentials, or the AWS default credential chain. The `BEDROCK_BEARER` pair is a dummy signing input for the bearer transport, not an identified real credential. |
| External cloud authentication | VertexAI uses Application Default Credentials (`internal/plugins/ai/vertexai/vertexai.go:56-65`, `:72`). External SDK credential providers are valid alternatives to literal environment secrets; this task does not certify their runtime configuration. |
| Server API key | `internal/cli/flags.go:86` binds `--api-key` to `FABRIC_API_KEY` with an empty default. A parser probe confirms environment ingestion and explicit flag precedence. REST and Ollama receive this parsed value (`internal/cli/setup_server.go:25,31`). |
| File loading and storage | `internal/cli/initialization.go:22-23` creates the database at `~/.config/fabric`; `internal/plugins/db/fsdb/db.go:83-88` loads its `.env`. Synthetic tests confirm existing process variables take precedence and `SaveEnv` writes with mode `0600`. Plaintext storage limitations are already recorded in [[pr-2107-sensitive-data]]. `.gitignore:103` excludes `.env`; this does not remove previously tracked files. |

## Inherited concerns

### Low: server key in Fabric's dotenv file loads too late for flags

`internal/cli/cli.go:19` parses flags before `initializeFabric` (`:40`) loads Fabric's `.env`. Later server calls use the captured `ServeAPIKey`, so placing `FABRIC_API_KEY` only in that file does not populate the server key for a fresh process. The parser characterization confirms later environment changes do not alter an already parsed struct. This is an initialization/configuration mismatch, not a demonstrated authentication bypass: existing non-loopback unauthenticated binds fail closed, and loopback without a key remains intentional (see [[pr-2107-authentication]]).

Recommendation: explicitly document that the server key must be injected into the process environment before launch; prefer this over a command-line secret that can enter shell history/process arguments. Alternatively resolve the server key after dotenv loading while preserving explicit flag precedence. No implementation change was requested in this review task.

### Low: shared setup prompts can display stored secrets

`internal/plugins/plugin.go:220-221` prints an existing setting value as a prompt default, and `:225-226` reads new input with `fmt.Scanln`. Shared credential setup questions have no sensitivity flag. Re-running setup with a populated key/client secret can expose it in terminal captures or to an observer; ordinary interactive input may also echo. This is conditional local disclosure (OWASP sensitive-data exposure), pre-existing and outside the original PR changes. The second parameter to `AddSetupQuestion` means required, not sensitive.

Recommendation: add explicit secret metadata to shared settings, mask current credential defaults, and read secret input without terminal echo. Environment injection avoids this particular interactive prompt path. Storage and diagnostic disclosures remain cross-referenced rather than counted again.

## Validation and limits

Four temporary Go characterization tests passed: representative provider/Spotify environment ingestion (two subcases), missing required credential rejection, dotenv precedence plus owner-only saved-file permissions, and server flag/environment precedence plus late-loading behavior. Only synthetic values were used and no provider requests were made. Passing characterizations verify observed behavior; they do not fix the inherited concerns.

Formatted test source is retained at `/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2107-codex/Working/review_environment_probe_test.go`. To rerun, temporarily copy it to `internal/plugins/review_environment_probe_test.go`, run `go test ./internal/plugins -run TestReviewEnvironment -v`, then remove the temporary copy. All Go module/build/temp caches must be directed into the authorized Working folder as in prior reviews.

After removing the probes, `go test ./internal/plugins ./internal/plugins/db/fsdb ./internal/cli ./internal/domain ./internal/core ./internal/server` passed. `git diff --check` passed. No permanent application tests or production files changed. This review verifies code wiring, not deployed secret injection, every SDK fallback, credential rotation, Git history, or full runtime confidentiality. Remaining authentication/authorization and consolidated-findings tasks are still pending.
