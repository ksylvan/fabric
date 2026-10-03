---
type: report
title: PR 2107 Hardcoded Secrets Review
created: 2026-10-03
tags:
  - code-review
  - security
  - secrets
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-security-context]]'
  - '[[pr-2107-sensitive-data]]'
  - '[[3_CHECK_SECURITY]]'
---

# PR 2107 hardcoded secrets review

Completed only the **Hardcoded secrets** checkbox, covering API keys, passwords, private keys, connection strings, and tokens. No actual hardcoded credential was identified in the original PR diff or the reviewed text-snapshot candidates. This finding does not certify the entire repository or its history as secret-free. No production code changed. No images were analyzed (0).

## Scope and method

Compared base `ddf1aab968caa9adf90137a6536c768d02b237d6` with original head `f32656b529a1eb1c678c793703fe49a0755cd439`. Manually inspected the full three-file diff: changes introduce a channel gate, lexical path validation, and synthetic path/content test literals, with no credentials. Later review documents were excluded from the original snapshot.

The indexed code graph was available and used to orient credential/token symbols. Literal scanning used local Git blobs, which is appropriate for strings and configuration values. No root or tracked nested `CLAUDE.md`/`AGENTS.md` was found. Neither `gitleaks` nor `trufflehog` was installed, so a retained Python heuristic scan was used instead; it is not an equivalent comprehensive secret scanner.

The scan examined all 868 tracked entries at the original head. It decoded 856 UTF-8 text files and skipped 12 nontext files (two SQLite databases and ten images). It searched provider-key formats, PEM private-key headers, credential-bearing URLs, JWTs, quoted credential assignments, and literal bearer tokens. Results contain file paths, line numbers, and rule names, never matched values. All 300 candidates lie outside the three changed files.

## Candidate assessment

| Category | Matches | Assessment |
| --- | ---: | --- |
| Translation messages | 240 | Credential-related localization keys and human-readable setup/help text, rather than stored credential values. |
| Test fixtures | 41 | Synthetic keys, short fixture words, token-rotation markers, mocked bearer headers, and pagination tokens. The AWS-shaped access/secret pair in `internal/plugins/ai/bedrock/bedrock_test.go:88-89` and `:115-116` exactly matches the standard published example pair, not an identified operational credential. |
| Documentation/pattern examples | 6 | Placeholder API keys and instructional snippets. `docs/rest-api.md:312-313` uses ellipsis placeholders. `data/patterns/create_golden_rules/system.md:102` illustrates a forbidden hardcoding practice. `data/patterns/write_nuclei_template_rule/system.md:414` includes an example JWT in token-generation documentation, not a Fabric authentication credential. |
| Production noncredential literals | 8 | Help text (`internal/cli/help.go:71`), OAuth endpoint/question/token-type constants (`internal/plugins/ai/copilot/copilot.go:42,67,70,73,122`), the header name (`internal/server/auth.go:15`), and an OAuth endpoint (`internal/tools/spotify/spotify.go:31`). |
| Package metadata | 5 | Token-related package names and associated lockfile metadata in `web/package-lock.json`; not credentials. |

By detector, there were 296 quoted-assignment matches, two AWS-shaped provider-key matches, one JWT match, and one bearer-literal match. No private-key header or credential-bearing URL matched. Candidate context and recognized fixture/placeholder values were checked before classifying them; test-file location alone is not evidence that a credential is harmless. No live credential verification or external provider request was performed.

## Validation and limits

Three retained Python tests passed: required detector categories (seven synthetic examples), environment-reference avoidance plus the original PR diff, and redaction/line-number behavior. The scanner and annotated evidence are retained in the authorized Auto Run folder as `Working/review_secrets_scan.py` and `Working/review_secrets_scan.json`. Run the detector checks with `python3 Working/review_secrets_scan.py --test` from that folder; run without `--test` from the repository to rescan the original head (this regenerates the raw evidence and removes manual classification annotations).

The scan is heuristic: it can miss arbitrary or split strings, unquoted configuration assignments, unfamiliar provider formats, encoded/encrypted credentials, and binary contents. Deleted secrets in Git history, untracked/local environment files, live credentials, deployment state, and the two binary databases were not audited. Nothing in this task warrants credential rotation based on an identified real leak. Sensitive storage/logging concerns remain in [[pr-2107-sensitive-data]]. Environment-variable configuration verification is a separate pending task.

The existing domain, core, and server suites also passed with `go test ./internal/domain ./internal/core ./internal/server`. Module, build, and temporary caches were directed into the authorized Auto Run Working folder. No permanent source or application test files changed; scanner tests are review evidence. `git diff --check` passed.
