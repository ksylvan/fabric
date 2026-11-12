---
type: analysis
title: PR 2330 Vulnerable Dependencies Review
created: 2026-11-11
tags:
  - code-review
  - security
  - dependencies
related:
  - '[[pr-2330-context]]'
  - '[[pr-2330-security-context]]'
  - '[[pr-2330-xss-review]]'
---

# PR 2330 dependency security review

The fixed PR snapshot adds no external dependency, changes no dependency version, and introduces no Keenable SDK. No dependency vulnerability introduced by this PR was identified. Live audits do report existing warnings; this is not a clean bill of health for the repository.

## Scope and evidence

Compared `de0b72766db0436b6c1887bb75768cb3960e67e5` with `fb296d490cca4fdebbcb78d5107241f134152d19`, excluding subsequent review commits. All 25 changed paths were checked: no manifests, lockfiles, vendored libraries, build scripts, or CI workflows change. Byte comparisons also confirmed `go.mod`, `go.sum`, `web/package.json`, `web/package-lock.json`, and `flake.lock` are identical at base, PR head, and review HEAD. Graph discovery and fixed-head source inspection confirm the new client imports only Go standard-library packages and existing internal `i18n`/`plugins` packages.

## Audit results

| Check | Result |
| --- | --- |
| `govulncheck -show verbose ./cmd/... ./internal/...` | Completed; exit 3 for one standard-library symbol warning, plus one module-only warning; 58 root packages. |
| `npm audit --json --ignore-scripts` in `web` | Exit 1: 15 affected package entries (1 critical, 11 moderate, 3 low). Entries include transitive effects, not 15 distinct vulnerabilities. |
| `npm audit --omit=dev --json --ignore-scripts` | Exit 1: DOMPurify only, critical package rating. |
| `go test -race ./internal/tools/keenable ./internal/plugins ./internal/i18n` | All three package suites passed. |
| Review validation probe | Passed manifest identity, audit-count consistency, production warning, and completed Go scan assertions. |

The installed scanner was built with Go 1.25 and failed against Go 1.27.3. Rebuilt `golang.org/x/vuln/cmd/govulncheck@v1.9.0` locally under the authorized Working directory using the installed toolchain, then scanned explicit Go source directories to exclude Go examples inside `web/node_modules`. Initial failure is preserved separately and is not counted as a completed scan.

## Existing warnings and applicability

### Go toolchain and modules

- **GO-2026-6859, scanner-reported standard-library warning:** `net/http@go1.27.3`, reported fixed in `go1.27.4`; HTTP/1 server crash involving a MaxBytesReader/header race. The scanner reports symbol reachability throughout existing client/server paths. This is local-toolchain dependent; no proof of exploitability in the new outbound search client was established. Existing server request handling warrants separate triage. Rebuild with a patched supported toolchain and rerun the scanner. The report URL could not be loaded through the web tool, so advisory details here are attributed to captured scanner output, not independently verified web content: https://pkg.go.dev/vuln/GO-2026-6859.
- **GO-2026-5932, module-only warning:** existing `golang.org/x/crypto@v0.58.0`; no affected imported package or called symbol detected. Scanner lists a fixed-version boundary of `v0.58.1-0.20261009181718-2f1f834330b2`, but the [official advisory](https://pkg.go.dev/vuln/GO-2026-5932) also warns that legacy OpenPGP remains unsafe and unmaintained. Avoid introducing use of that package; a version bump alone is not a complete OpenPGP remediation. This is not an observed Keenable vulnerability.

### DOMPurify runtime dependency

`web/package.json:75` allows `^3.4.16`; the unchanged lockfile resolves `3.4.16`. npm reports Critical overall, with three underlying advisories. This is an OWASP A06 vulnerable-component warning; application severity depends on enabled features:

- [GHSA-cg2g-xh64-v4q7](https://github.com/cure53/DOMPurify/security/advisories/GHSA-cg2g-xh64-v4q7): upstream Critical; requires SAFE_FOR_TEMPLATES plus DOM output modes and shadow-root content. The current `web/src/lib/utils/sanitize-html.ts:9` accepts a string and uses HTML profile/string output without these flags. No demonstrated application exploit through this wrapper.
- [GHSA-mrv2-3pqr-j5jr](https://github.com/cure53/DOMPurify/security/advisories/GHSA-mrv2-3pqr-j5jr): npm labels High; the maintainer advisory currently labels Low. Requires IN_PLACE and hooks detaching relatives. These are not configured by the current wrapper. Preserve the source-rating discrepancy rather than promoting it to a confirmed High application finding.
- [GHSA-c6gr-qvv4-5v3f](https://github.com/cure53/DOMPurify/security/advisories/GHSA-c6gr-qvv4-5v3f): Low; requires XHTML parsing and SVG animation allowance. Current HTML-only profile does not enable that combination.

All three maintainer advisories list `3.4.17` as patched. Recommend updating the resolved DOMPurify version to a patched release in separate dependency maintenance, then rerunning sanitizer regressions and audit. No automatic audit fix, manifest change, or dependency downgrade was performed. Earlier XSS tests cover selected payloads and do not supersede these dependency advisories.

### Complete npm warning inventory

These are scanner package severity labels, not confirmed application exploit severities. Most remaining entries concern development tooling. Each advisory link identifies its affected range; raw audit JSON includes installed nodes, suggested fixes, and dependency propagation.

| Package | npm severity | Advisory or propagation |
| --- | --- | --- |
| `@sveltejs/kit` | low | [GHSA-pfhg-q3qv-9q6r](https://github.com/advisories/GHSA-pfhg-q3qv-9q6r) (low), [GHSA-27mw-82g8-phjj](https://github.com/advisories/GHSA-27mw-82g8-phjj) (low) |
| `@sveltejs/vite-plugin-svelte` | low | Inherited via deepmerge |
| `@tailwindcss/typography` | moderate | Inherited via postcss-selector-parser |
| `@typescript-eslint/eslint-plugin` | moderate | Inherited via @typescript-eslint/type-utils, @typescript-eslint/utils |
| `@typescript-eslint/parser` | moderate | Inherited via @typescript-eslint/typescript-estree |
| `@typescript-eslint/type-utils` | moderate | Inherited via @typescript-eslint/typescript-estree, @typescript-eslint/utils |
| `@typescript-eslint/typescript-estree` | moderate | Inherited via minimatch |
| `@typescript-eslint/utils` | moderate | Inherited via @typescript-eslint/typescript-estree |
| `@vitest/mocker` | moderate | [GHSA-qxg6-89mg-rqv8](https://github.com/advisories/GHSA-qxg6-89mg-rqv8) (moderate), [GHSA-xp26-3454-mjgp](https://github.com/advisories/GHSA-xp26-3454-mjgp) (moderate) |
| `deepmerge` | low | [GHSA-gcqx-w48m-w2h4](https://github.com/advisories/GHSA-gcqx-w48m-w2h4) (low) |
| `dompurify` | critical | [GHSA-cg2g-xh64-v4q7](https://github.com/advisories/GHSA-cg2g-xh64-v4q7) (critical), [GHSA-mrv2-3pqr-j5jr](https://github.com/advisories/GHSA-mrv2-3pqr-j5jr) (high), [GHSA-c6gr-qvv4-5v3f](https://github.com/advisories/GHSA-c6gr-qvv4-5v3f) (low) |
| `minimatch` | moderate | [GHSA-g3hv-75wg-m249](https://github.com/advisories/GHSA-g3hv-75wg-m249) (moderate) |
| `postcss-selector-parser` | moderate | [GHSA-rj75-hqrm-r3gf](https://github.com/advisories/GHSA-rj75-hqrm-r3gf) (moderate) |
| `typescript-eslint` | moderate | Inherited via @typescript-eslint/eslint-plugin, @typescript-eslint/parser, @typescript-eslint/typescript-estree, @typescript-eslint/utils |
| `vitest` | moderate | Inherited via @vitest/mocker |

Do not apply suggested major-version changes or downgrades blindly. Triage supported upstream patches for Vitest/mocker, minimatch/TypeScript ESLint, selector-parser/typography, deepmerge/Svelte plugin, and SvelteKit. Dev-server and untrusted build-input exposure need separate assessment; this task does not claim those tools are deployed or exploitable.

## Reproduction and limits

Raw outputs and the new assertion probe are in `/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/`:

- `dependencies-govulncheck.txt` (initial incompatible scanner), `dependencies-scanner-build.txt`, `dependencies-govulncheck-retry.txt` (completed scan).
- `dependencies-npm-audit.json`, `dependencies-npm-production-audit.json`, `dependencies-tests.txt`.
- `validate_dependency_review.py`; run from repository root with Python 3.

Audit results reflect databases available during this review, not historical advisories at PR creation. Source scanning used darwin/arm64 Go 1.27.3, default build tags, and excluded test files from vulnerability reachability; other platforms/toolchains and runtime configuration may differ. [Go documents](https://go.dev/doc/security/vuln/) that govulncheck uses call reachability; a scanner result is evidence to triage, not an exploitation test. No live Keenable requests, real credentials, production changes, or images were used. Only the vulnerable-dependencies task is complete; logging and subsequent tasks remain for later runs.
