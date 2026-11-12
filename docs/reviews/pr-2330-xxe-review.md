---
type: analysis
title: PR 2330 XML and XXE Review
created: 2026-11-11
tags:
  - code-review
  - security
  - keenable
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2330-security-context]]'
  - '[[pr-2330-sensitive-data-review]]'
---

# XML and XXE review

Reviewed the fixed PR snapshot from `de0b72766db0436b6c1887bb75768cb3960e67e5` to `fb296d490cca4fdebbcb78d5107241f134152d19`. The current Keenable, CLI, and core implementation directories match that head. This completes only the XML/XXE checklist item. No production code changed and zero images were analyzed.

## Assessment

No XML parser is introduced or invoked by the new Keenable search flow. External entity and DTD processing settings are therefore **not applicable** to this flow; no XML parser needs to be configured to disable them. No new XXE vulnerability was identified within this scope.

| Boundary | Evidence and result |
| --- | --- |
| Outbound query | `internal/tools/keenable/keenable.go:64-67` serializes the query using `json.Marshal`; `:82` sets `Content-Type: application/json`. XML syntax in a query stays JSON string data. |
| Successful API response | `internal/tools/keenable/keenable.go:101-105` always uses `encoding/json.NewDecoder(...).Decode`, independent of response Content-Type. XML documents starting with `<`, including documents with DTD/entity declarations, fail JSON decoding. There is no MIME-based XML fallback. |
| Result text and URLs | `internal/tools/keenable/keenable.go:108-119` formats string fields directly into Markdown. XML declarations and entity references embedded in JSON remain literal; result and entity URLs are not fetched by this client. `AppendMessage` in `internal/cli/flags.go:532-539` concatenates text, and the added dispatch in `internal/cli/tools.go:126-139` passes results to output or chat without XML interpretation. |
| Error response | `internal/tools/keenable/keenable.go:96-99` reads non-200 bodies as bytes and includes them in an error. An XML error body is echoed verbatim, without interpreting DTDs/entities. This retains the existing Low diagnostic-disclosure finding in [[pr-2330-sensitive-data-review]]; it is not a separate XXE finding. |
| Registration and supporting changes | Registry additions construct/configure the client and persist settings through existing helpers. Flag/help changes, completion scripts, locale JSON, README, and changelog do not introduce XML parsing or DTD configuration. No dependency manifest changes occur in the fixed PR diff. |

Used the existing codebase knowledge graph to locate/read `Search` and `AppendMessage`, checked the implementation diff and fixed-head correspondence, and searched XML-related import/configuration literals in the relevant source directories. This is a review of the PR's new data path, not an audit of every transitive dependency or all existing tools. XML handling inside Keenable's remote service cannot be established from this repository. Literal markup in output, downstream renderers, and model behavior are separate concerns; this review does not claim XSS or prompt-injection safety.

## Validation

Added review-only probes in the external Auto Run `Working/keenable_xxe_review_test.go`, reusing the existing `reviewTransport` helper and overlay rather than creating a production helper or dependency. `Working/xxe-review-overlay.json` extends the prior sensitive-data overlay. The probes cover:

- Raw XML containing a synthetic file entity, remote external DTD, or nested internal entity declarations. All 36 combinations of keyed/keyless mode, HTTP 200/502, three Content-Type values, and three XML payloads return errors with no result output. Successful-status XML fails JSON decoding; non-200 XML is preserved in the diagnostic.
- XML/DTD/entity text embedded in valid JSON titles and description/snippet fields, even when the response claims `application/xml`. Both field cases preserve the exact literal text and result URL.
- XML-like query text surviving JSON serialization unchanged, plus exactly one request to the fixed API host in each probe. The fake transport intercepts every client request and fails on any additional request, so the probes never contact a remote DTD/entity/result URL or live API. The synthetic file URI is not an actual secret; static inspection of the JSON-only path supplies the evidence against file entity resolution.

All three full targeted package suites, including existing overlay probes, passed with the race detector on Go 1.27.3:

```sh
TMPDIR=/Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/tmp go test -race -timeout 90s -overlay /Users/kayvan/src/worktrees/autorun/fabric/fabric-pr-2330-codex/Working/xxe-review-overlay.json ./internal/tools/keenable ./internal/cli ./internal/core
```

Existing package tests use local test servers; the new XXE probes use an in-process transport. Review probes and overlays remain in the external Working folder, consistent with prior review tasks; reproduction requires those files. The full repository suite was not run. Broken access control and all later checklist items remain unchecked.
