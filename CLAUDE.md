# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
go build ./...                          # Build all packages
go build -o git-sandwich .              # Build binary
go test ./...                           # Run all tests
go test ./... -v                        # Run all tests (verbose)
go test ./internal/sandwich/ -v         # Run tests for a single package
go test ./internal/sandwich/ -run TestIntegration_InsideChange_PASS -v  # Run a single test
```

Integration tests in `internal/sandwich/integration_test.go` create temporary Git repos and run real `git` commands. They require `git` to be available on PATH.

## Architecture

git-sandwich is a CLI tool that validates Git diffs to ensure changes only occur within designated BEGIN/END marker blocks ("sandwich blocks"). It exits 0 on success, 1 on violation.

### Data Flow

```
CLI (cmd/root.go)
  → sandwich.Validate (validator.go)
    → git.GetDiff (git diff -U0 base...head)
    → diff.Parse (extract changed line ranges per file)
    → per file: git.GetFileContent for base & head
    → per file: block.ParseBlocks (find BEGIN/END blocks)
    → per file: classifyLines (inside/boundary/outside)
  → output.FormatText or output.FormatJSON
```

### Key Packages

- **`internal/diff`** — Parses `git diff -U0` output via `sourcegraph/go-diff`. Owns `LineRange` and `FileDiff` types (placed here to avoid import cycle with `sandwich`).
- **`internal/git`** — Thin wrapper around `os/exec` for `git diff` and `git show`.
- **`internal/sandwich`** — Core validation logic. `Block.ContainsLine` treats only lines *between* START and END as "inside" (marker lines themselves are "boundary"). Validation checks both sides: OldRanges against base blocks, NewRanges against head blocks.
- **`internal/output`** — Text and JSON formatters. Depends on both `diff` (for `LineRange`) and `sandwich` (for `Result`).
- **`cmd`** — Cobra CLI setup. Compiles regex from `--start`/`--end` flags, calls `sandwich.Validate`, formats output.

### Import Graph Constraint

`diff` ← `sandwich` ← `output` ← `cmd`. The `diff` package must not import `sandwich` (this would create a cycle). Shared types (`LineRange`, `FileDiff`) live in `diff` for this reason.

### Line Classification Rules

Each changed line is classified against the relevant blocks (base blocks for deletions, head blocks for additions):
- **inside**: `StartLine < line < EndLine` — allowed
- **boundary**: `line == StartLine || line == EndLine` — allowed alone, but fails if combined with outside changes (unless `--allow-boundary-with-outside`)
- **outside**: everything else — always fails if the file has blocks in base
