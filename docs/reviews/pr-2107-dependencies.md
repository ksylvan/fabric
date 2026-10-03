---
type: report
title: PR 2107 Vulnerable Dependencies Review
created: 2026-10-03
tags:
  - code-review
  - security
  - dependencies
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-security-context]]'
  - '[[3_CHECK_SECURITY]]'
---

# PR 2107 dependency review

Completed only the **Vulnerable dependencies** checkbox. The original PR introduces no dependencies or dependency updates. Live scans found inherited warnings; no dependency exploit or new dependency vulnerability was demonstrated. Production code and dependency files are unchanged. No images supplied or analyzed (0).

## Scope

Compared base `ddf1aab968caa9adf90137a6536c768d02b237d6` with original head `f32656b529a1eb1c678c793703fe49a0755cd439`. Only the three source/test files in [[REVIEW_SCOPE]] differ. `go.mod`, `go.sum`, `web/package.json`, both web lockfiles, and the optional Python UI requirements are unchanged between these commits and the reviewed checkout. Root `AGENTS.md` and `CLAUDE.md` are absent; supplied graph-first instructions apply. Used the existing graph index and Ollama method snippets to assess scanner traces.

Scanned Go application packages and both web lockfiles on 2026-10-03. Web documentation supports npm and pnpm, so neither lockfile was assumed authoritative. The optional Python UI has unpinned minimum requirements (`scripts/python_ui/requirements.txt:1`); no installed environment or exact resolved versions were evaluated. Container images, Nix inputs, GitHub Actions, native libraries, and external Ollama server installations are outside these scans. Results are a dated snapshot, not a whole-repository clean bill of health.

## Web warning: inherited development dependency

**OWASP A06 / CWE-674; advisory severity High; runtime exploitability unconfirmed.** Both lockfiles contain `braces` 3.0.3 through `patch-package` 8.0.1 → `find-yarn-workspace-root` 2.0.0 → `micromatch` 4.0.8. See `web/package.json:36`, `web/package-lock.json:2533`, and `web/pnpm-lock.yaml:890`.

The [braces advisory GHSA-vfj7-8cjw-p6xm / CVE-2026-93687](https://github.com/advisories/GHSA-vfj7-8cjw-p6xm) describes stack exhaustion from deeply nested patterns, affects versions through 3.0.3, and lists no patched version at review time. An attacker-controlled pattern reaching this dependency could terminate a Node process. The recorded dependency path is development-only; no public request path to it was established.

`npm audit` reports four High affected-package entries, including the three ancestors; `pnpm audit` reports one High advisory. These describe the same underlying issue, not four independent vulnerabilities. Production-only npm audit reports zero warnings. This does not establish safety for all deployment configurations, particularly those retaining development tooling.

Recommendation: assess whether `patch-package` is needed, remove or replace it if practical, or constrain untrusted patterns and track an upstream fix. Refresh both lockfiles after any intentional remediation and rerun audits. Do not blindly apply npm's proposed downgrade to `patch-package` 6.0.7; it is a breaking dependency change and was not validated here.

## Go warnings and applicability

`govulncheck` v1.8.0 completed source/symbol analysis with Go 1.27.1 on darwin/arm64 against `https://vuln.go.dev` (database modified 2026-10-01). It returned exit 0 while emitting findings in JSON mode; exit status alone is not the finding count. There are **11 distinct advisory IDs**, eight with symbol traces, one with package traces only, and two with module traces only. OSV metadata returned without a corresponding `finding` is not counted.

| Go advisory | Description | Highest reported evidence |
| --- | --- | --- |
| [GO-2025-3557](https://pkg.go.dev/vuln/GO-2025-3557) | Ollama resource exhaustion | Symbol |
| [GO-2025-3558](https://pkg.go.dev/vuln/GO-2025-3558) | Ollama out-of-bounds read | Symbol |
| [GO-2025-3559](https://pkg.go.dev/vuln/GO-2025-3559) | Ollama divide by zero | Symbol |
| [GO-2025-3582](https://pkg.go.dev/vuln/GO-2025-3582) | Ollama null dereference / denial of service | Symbol |
| [GO-2025-3689](https://pkg.go.dev/vuln/GO-2025-3689) | Ollama divide by zero | Symbol |
| [GO-2025-3695](https://pkg.go.dev/vuln/GO-2025-3695) | Ollama server denial of service | Symbol |
| [GO-2025-3824](https://pkg.go.dev/vuln/GO-2025-3824) | Ollama cross-domain token exposure | Symbol |
| [GO-2025-4251](https://pkg.go.dev/vuln/GO-2025-4251) | Ollama model-management authentication | Symbol |
| [GO-2026-5750](https://pkg.go.dev/vuln/GO-2026-5750) | Ollama path traversal | Module only |
| [GO-2026-6443](https://pkg.go.dev/vuln/GO-2026-6443) | gRPC xDS server panic | Package only |
| [GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932) | Unmaintained OpenPGP package | Module only |

The Ollama module is v0.34.4 (`go.mod:24`). All nine Ollama records in the scan have an unbounded affected range starting at zero and no fixed version. Symbol traces include initialization, common API methods, and `StatusError.Error`; their existence does not prove that the server or native-code condition described by an advisory is present in Fabric. Graph snippets show Fabric constructing an Ollama API client (`internal/plugins/ai/ollama/ollama.go:89`) and calling its `Chat` method (`:153`). Fabric's Ollama-compatible endpoints do not by themselves establish that it embeds the vulnerable upstream daemon.

There is also a concrete database discrepancy: [GitHub's token-exposure advisory](https://github.com/advisories/GHSA-x9hg-5q6g-q3jr) caps affected versions at 0.9.6, whereas the Go record matches v0.34.4 with its unbounded range. Treat this as a triage warning, not a confirmed token leak in this checkout. Recommendation: reconcile ranges with upstream fixes, assess client and actual daemon versions separately, and reproduce relevant behavior safely before assigning Fabric-specific severity or choosing an upgrade. No token-exposure probe was run.

The gRPC v1.84.0 warning (`go.mod:166`) concerns missing authority/Host headers on an xDS-enabled server. The [Go advisory](https://pkg.go.dev/vuln/GO-2026-6443) lists a fix on the 1.85 development line; the scan imports the transport package but finds no affected-symbol call. No relevant xDS server was demonstrated. Review usage and select an appropriate stable fixed release before upgrading; do not downgrade to a different release branch solely to silence a warning.

The `golang.org/x/crypto` v0.57.0 warning (`go.mod:161`) is module-only for its deprecated OpenPGP package. The scan does not show that package imported or called. Track it as informational dependency metadata; it does not establish that Fabric uses unsafe OpenPGP. No Fabric-specific severity is assigned to these unconfirmed Go warnings.

## Validation and retained evidence

Raw outputs live in the authorized Auto Run `Working` folder:

- `review_dependencies_npm.json`: full npm lockfile audit, exit 1 with warnings.
- `review_dependencies_pnpm.json`: pnpm lockfile audit, exit 1 with warnings.
- `review_dependencies_npm_prod.json`: production-only npm audit, exit 0.
- `review_dependencies_go_current.json` and `.stderr`: completed Go scan, exit 0, empty stderr.
- `review_dependencies_go.json` and `.stderr`: failed initial scan using the existing v1.1.4 binary built with Go 1.25; these are diagnostic evidence only. Rebuilt v1.8.0 under `Working/bin` with isolated tool-download, build, and temporary directories, then reran successfully.

Reproduction: run `Working/bin/govulncheck -json ./...` from the checkout; run `npm audit --package-lock-only --ignore-scripts --json`, `npm audit --omit=dev --package-lock-only --ignore-scripts --json`, and `pnpm audit --json` from `web`. Use caches and temporary directories under `Working` as in this run. No install, audit fix, or dependency-file mutation was performed in the checkout.

Four retained evidence-validation tests in `Working/review_dependencies_validation_test.py` check unchanged dependency inputs, npm/pnpm advisory deduplication, production-only results, and completed Go scan classification. `go test ./internal/domain ./internal/core ./internal/server` passed. No permanent source tests were added for this documentation-only review. Remaining security checkboxes are pending.
