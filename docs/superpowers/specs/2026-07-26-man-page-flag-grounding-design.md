# Design: Flag-Level Grounding via Man Pages (Phase 8)

## Problem

`GenerateCommand` produces a full `CommandResult` (tool + args per step) in a single LLM call, and the tool trusts whatever flags come back. LLMs occasionally hallucinate or confuse flags for well-known tools — e.g. asking for a regex search might yield `grep -r '[a-c]'` (recursive) instead of `grep -E '[a-c]'` (extended regex), or invent a flag that doesn't exist at all. There is currently no step that checks generated flags against the real capabilities of the tool being invoked.

## Goal

Add a second, cheap verification pass that checks the specific flags an LLM chose against the actual `man` page for each tool used in the generated pipeline, correcting anything wrong before the command is shown or executed. This targets **well-known tools with real man pages** — it is a correctness/grounding layer, not a knowledge base for tools the LLM has never heard of.

## Non-goals

- Not a general man-page search/autocomplete UI for the user.
- Not a fix for tools with no man page at all — those are skipped, not blocked.
- Not a tool-use/function-calling integration (no mid-generation LLM function calls) — this is a fixed two-pass pipeline.

## Architecture & Data Flow

1. `GenerateCommand` runs exactly as it does today → draft `model.CommandResult` with `Steps []CommandStep`.
2. For each step, the orchestrator (in `cmd/base.go`) determines a **lookup key**:
   - If `step.Args[0]` looks like a subcommand (matches `^[a-zA-Z][a-zA-Z0-9_-]*$`, not a flag or path token), the key is `<tool>-<subcommand>` (e.g. `git-commit`, `docker-run`).
   - Otherwise the key is just `<tool>`.
3. `manpage.GetFlags(key)` resolves flags for that key:
   - Check `~/.cmdfy/man_cache/<key>.json`. If present, return cached flags.
   - Otherwise run `man <key>`. If that fails and the key contains a subcommand (`tool-subcommand`), retry with the top-level `tool` alone (git has `git-commit(1)`, but many tools don't split man pages per subcommand).
   - Parse the output (see Parsing below). Cache the result (even if empty) so repeat lookups are free.
   - Any failure (no man page, parse yields zero flags) is treated as "no docs available" for that tool — not an error that propagates.
4. Steps whose tool produced no docs are left out of `manDocs`. If `manDocs` ends up empty across all steps, verification is skipped entirely and the draft is returned as final — no second LLM call, no added cost.
5. If `manDocs` is non-empty and `--no-verify-flags` was not passed, call `Provider.VerifyFlags(ctx, draft, manDocs)`:
   - Prompt: draft pipeline (as JSON) + real flag list per tool/subcommand key → asks the LLM to return a corrected `CommandResult` in the same schema (steps, explanation, dangerous), fixing only what's wrong.
   - On success, the corrected result replaces the draft.
   - On any error (network, bad JSON), log a warning to stderr and keep the original draft untouched. Verification can only improve the result, never fail the generation.
6. Proceed to `printAndExecute` with whichever result is final.

In `--compare` (benchmark) mode, this same per-provider flow (draft → verify) runs inside each provider's goroutine before its result is sent to the TUI, so compared results are equally grounded.

## Components

### `pkg/manpage` (new package)

```go
type Flag struct {
    Name        string   // e.g. "-E"
    Aliases     []string // e.g. ["--extended-regexp"]
    TakesArg    bool
    ArgName     string   // e.g. "PATTERN", empty if TakesArg is false
    Description string
}

// GetFlags resolves flags for a man page key ("tool" or "tool-subcommand"),
// using the on-disk cache when available and parsing `man <key>` otherwise.
func GetFlags(key string) ([]Flag, error)

// RefreshCache deletes cached entries for the given keys, or all entries if none given.
func RefreshCache(keys ...string) error
```

- Cache location: `~/.cmdfy/man_cache/<key>.json`, one file per resolved key. Written once, kept indefinitely (no TTL/version key) — man pages for installed tools rarely change; staleness is handled by the manual refresh command below.
- Parsing: shell out to `man <key> | col -bx` to strip terminal formatting, then regex-scan line by line for flag headers (`-x`, `-x, --long`, `--long ARG`, etc.) followed by indented description lines, accumulating description text until the next flag header or a section header (all-caps line) is hit. This needs to tolerate both GNU-style (long descriptions, `--long-flag` common) and BSD-style (macOS, shorter inline descriptions) formatting — both get fixture-based tests since this repo runs on macOS but Linux is a supported target.
- Any error at any stage (man not found, non-zero exit, zero flags parsed) returns `(nil, err)`; callers treat this as "skip this tool," never as fatal.

### New CLI: `cmdfy man refresh [key...]`

Clears specific cache entries (`cmdfy man refresh git-commit grep`) or, with no args, the entire `~/.cmdfy/man_cache/` directory.

### `pkg/llm` — `Provider` interface addition

```go
type Provider interface {
    GenerateCommand(ctx context.Context, query string, meta SystemMetadata) (*model.CommandResult, error)
    VerifyFlags(ctx context.Context, draft *model.CommandResult, manDocs map[string][]manpage.Flag) (*model.CommandResult, error)
}
```

Implemented in all four providers (`gemini`, `openai`, `anthropic`, `ollama`), following the same request/response shape each already uses for `GenerateCommand` (build a prompt, call the API, sanitize markdown fences, unmarshal into `model.CommandResult`). The prompt embeds the draft's JSON and, per tool/subcommand key present in `manDocs`, its flag list (name, aliases, takes-arg, description) — asking the LLM to return the same schema with only incorrect flags/args changed.

### `cmd/base.go` orchestration

- New flag: `--no-verify-flags` (bool, default `false`, so verification is on by default).
- After `GenerateCommand`, build `manDocs` per the lookup-key logic above, then conditionally call `VerifyFlags` as described in the data flow.
- Metrics: verification latency is tracked separately and added to the displayed metrics (e.g. `METRICS: 1.1s generate + 0.4s verify, 512 tokens`) rather than silently folded into the existing `Latency` field.

## Error Handling Summary

| Failure | Behavior |
|---|---|
| `man` not installed / key not found | Tool excluded from `manDocs`, no error surfaced |
| Man page found but parses to 0 flags | Same as above — excluded, not an error |
| Subcommand man page missing (e.g. no `git-foo`) | Falls back to plain `tool` key automatically |
| `VerifyFlags` LLM call fails or returns bad JSON | Warning on stderr, draft result kept as final |
| All steps' tools have no docs | Verification call skipped entirely (zero extra cost) |

## Testing Plan

- `pkg/manpage`: table-driven parser tests using fixture text for GNU-style and BSD-style `man` output (e.g. captured `grep`, `tar` pages from both a Linux box and macOS), asserting correct `Flag` extraction.
- Cache round-trip test: write → read → matches; `RefreshCache` removes the right file(s).
- Subcommand fallback test: fake a resolver where `man tool-sub` errors, confirm fallback to `man tool` is attempted and used.
- Per-provider `VerifyFlags` tests mirroring existing `*_test.go` patterns: prompt construction includes the right man docs, JSON response parses into `CommandResult`, and a simulated API failure leaves the draft unchanged upstream.
- Manual/integration check: run a query that reliably produces a wrong flag on at least one provider, confirm the verification pass corrects it end-to-end.

## Open Items for Future Phases (explicitly out of scope here)

- On-demand tool-calling (LLM-invoked `lookup_man_page` function) as a more flexible alternative to the fixed two-pass flow.
- Structured flag data reused for other features (e.g. flag autocomplete in the TUI).
- TTL/version-based cache invalidation, if manual refresh proves insufficient in practice.
