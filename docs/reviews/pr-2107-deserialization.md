---
type: report
title: PR 2107 Insecure Deserialization Review
created: 2026-10-03
tags:
  - code-review
  - security
  - deserialization
related:
  - '[[REVIEW_SCOPE]]'
  - '[[pr-2107-security-context]]'
  - '[[3_CHECK_SECURITY]]'
---

# PR 2107 insecure deserialization review

Completed only the **Insecure deserialization** checkbox. No new insecure-deserialization vulnerability was established in the original PR diff or inspected adjacent decoders. No production code changed. No images were supplied or analyzed (0).

## Scope and evidence

Compared original base `ddf1aab968caa9adf90137a6536c768d02b237d6` with original head `f32656b529a1eb1c678c793703fe49a0755cd439`. The three-file diff adds chat file-write gating and lexical path validation; it adds no serialization library or custom decoder. Root `CLAUDE.md` and `AGENTS.md` are absent; supplied graph-first instructions apply. Used the existing project index, graph searches, source snippets, and inbound caller traces. Current inspected parser and chat decoder match the original head.

| Boundary | Evidence and assessment |
| --- | --- |
| `internal/domain/file_manager.go:23-27,30-95` | Model output is decoded into `[]FileChange`, whose operation, path, and content are strings. No input-selected class lookup or custom deserialization hook was found for this type. Object, numeric, boolean, and nested-array type mismatches return errors; invalid operation/path and oversize content return no changes. |
| `internal/core/chatter.go:190-209` | The graph identifies `Chatter.Send` as the production parser caller. It applies changes only after successful parsing and under the pattern/channel gate. Parsing itself does not write files or execute content. Existing symlink/write-authority concerns remain covered by [[pr-2107-access-control]]. |
| `internal/server/chat.go:24-40,71-79` | Request binding uses a concrete request struct, including typed prompt fields and a string-to-string variables map. Binding errors return HTTP 400 before chat processing. This is source inspection; no new HTTP probe was run for this task. |
| `internal/chat/chat.go:101-134` | The custom message decoder tries two concrete schemas for string versus multipart content. Both branches call JSON decoding into declared fields and return decoding errors; neither dispatches an input-selected constructor or executes tool calls. Downstream tool behavior is outside this decoder check. |
| `internal/plugins/db/fsdb/storage.go:280-290` | Session/storage loading calls `json.Unmarshal` into a caller-supplied destination, rather than constructing a type named by serialized data. The graph identifies `Get` and `PrintSession` callers. No execution hook was found in the inspected message decoder. |
| `internal/plugins/template/extension_registry.go:18-30,98-143,184-215` | Local extension YAML populates a declared definition; generic config values remain data. Registration validates the definition and hashes configuration/executable files; loading checks those hashes. Configured commands execute later by design. The independent shell-injection finding is in [[pr-2107-injection]], not a newly demonstrated YAML object-construction exploit. |

Graph-backed searches across application Go and web sources found no `encoding/gob`, Python pickle/jsonpickle, `eval`, `new Function`, or custom YAML unmarshal hook matching the search. The custom chat JSON decoder above was inspected. This search is not an audit of every transitive library or an assertion that JSON parsing makes all downstream data use safe.

## Type checking and hardening considerations

Types are declared **before** decoding, and the decoder checks their compatibility **during** decoding. Operation, path locality, and per-file content length are checked **after** decoding and before the caller applies changes. A separate pre-decoding type inspection is not present and was not required to demonstrate these rejection properties.

Unknown JSON fields, including `$type` and `__proto__`, are ignored by this parser. Null or omitted content becomes an empty string; null operation/path fails semantic validation. Duplicate paths use the last decoded value, and that final value is validated. These are permissive schema semantics, not demonstrated execution or containment bypasses. Consider explicit required-field and duplicate/unknown-field policy if strict schema enforcement is desired.

The per-file size limit is enforced after allocation; no aggregate input or file-count cap exists in this parser. Consider bounding input before decoding and bounding aggregate changes. No resource-exhaustion exploit or severity finding was established here. The bracket extractor and invalid-escape repair are existing compatibility behavior, not a strict JSON envelope validator; this review does not certify them for every malformed input.

## Validation

Four temporary characterization tests passed on darwin/arm64 with Go 1.27.1:

- Twelve invalid cases covered scalar/null/nested-array elements, wrong field types, null required fields, invalid operations, a duplicate traversal path, and a valid change followed by an invalid change. Every case returned an error and zero changes.
- `$type` and `__proto__` metadata remained inert. A shell/Python-like synthetic content string survived parsing and writing byte-for-byte; the temporary root contained only the intended file. No shell or Python interpreter was invoked by the probe.
- Three cases characterized null/missing content and last-value duplicate-path behavior.
- Content exceeding `MaxFileSize` was rejected with zero changes.

Formatted probe source is retained at `Working/review_deserialization_probe_test.go` in the authorized Auto Run folder. To reproduce, temporarily copy it into `internal/domain`, run `go test ./internal/domain -run TestReviewDeserialization -v`, and remove the copy. Probe temporary files and Go build caches were placed under `Working`.

After removing the probe copy, `go test ./internal/domain ./internal/core ./internal/chat ./internal/server ./internal/plugins/template ./internal/plugins/db/fsdb` passed; `internal/chat` reports no test files. No permanent source/test changes were needed. Remaining playbook checkboxes are pending; dependency vulnerability scanning belongs to the next task.
