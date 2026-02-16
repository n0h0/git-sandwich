# git-sandwich

> A Git diff validator that rejects changes outside BEGIN/END blocks.

## Overview

`git-sandwich` is a CLI tool that enforces **editable regions** in source code. You define sandbox blocks using comment markers, and the tool analyzes Git diffs to **detect and reject any changes outside those blocks**.

Use cases:

- Protecting framework-generated code from accidental edits
- Safe customization of config templates (e.g. Rails initializers)
- Preventing breakage in shared templates
- Coexistence of generated and hand-written code

## Installation

### Homebrew (macOS / Linux)

```bash
brew install n0h0/tap/git-sandwich
```

### Go

```bash
go install github.com/n0h0/git-sandwich@latest
```

### Build from source

```bash
git clone https://github.com/n0h0/git-sandwich.git
cd git-sandwich
go build -o git-sandwich .
```

## Quick Start

Given a file with sandwich blocks:

```ruby
# frozen_string_literal: true

class ApplicationConfig
  # CUSTOM START
  config.time_zone = "Tokyo"
  # CUSTOM END
end
```

Run the validator:

```bash
git-sandwich --start '# CUSTOM START' --end '# CUSTOM END'
```

- Changes **inside** the block (between START and END) are allowed.
- Changes **outside** the block are rejected.

## Usage

```
git-sandwich [options] [paths...]
```

### Required Options

`--start` and `--end` must be provided via CLI flags or a config file (see [Configuration](#configuration)).

| Flag              | Description                       |
| ----------------- | --------------------------------- |
| `--start <regex>` | Regex pattern for BEGIN markers   |
| `--end <regex>`   | Regex pattern for END markers     |

### Optional Flags

| Flag                              | Default                | Description                                      |
| --------------------------------- | ---------------------- | ------------------------------------------------ |
| `--base`                          | `origin/main`          | Base ref for comparison                          |
| `--head`                          | `HEAD`                 | Head ref for comparison                          |
| `--allow-nesting`                 | `false`                | Allow nested BEGIN/END blocks                    |
| `--allow-boundary-with-outside`   | `false`                | Allow boundary changes together with outside changes |
| `--json`                          | `false`                | Output results in JSON format                    |
| `--include <glob>`                |                        | Glob pattern for files to include (repeatable)   |
| `--exclude <glob>`                |                        | Glob pattern for files to exclude (repeatable)   |
| `--config <path>`                 | `.git-sandwich.yml`    | Path to config file                              |

Positional arguments `[paths...]` are passed as path filters to `git diff`.

### Configuration

Create a `.git-sandwich.yml` file in your project root to avoid passing flags every time:

```yaml
# .git-sandwich.yml
start: "# CUSTOM START"
end: "# CUSTOM END"
base: "origin/main"
head: "HEAD"
allow_nesting: false
allow_boundary_with_outside: false
json: false
include:
  - "*.go"
exclude:
  - "vendor/**"
```

All fields are optional. However, `start` and `end` must be provided either in the config file or via CLI flags.

**Priority**: CLI flags > config file > default values.

- If `--config` is explicitly specified, the file must exist (error if missing).
- If `--config` is not specified, `.git-sandwich.yml` is loaded from the current directory if it exists, otherwise silently skipped.
- CLI flags always override config file values.

```bash
# Use default config file (.git-sandwich.yml)
git-sandwich

# Use a custom config file
git-sandwich --config path/to/custom.yml

# Override config file values with CLI flags
git-sandwich --start "OTHER_BEGIN" --end "OTHER_END"
```

### File Filtering (`--include` / `--exclude`)

Use `--include` and `--exclude` to control which files are validated using glob patterns. Patterns support `**` for recursive directory matching.

- **`--include` only**: Only files matching at least one pattern are validated.
- **`--exclude` only**: Files matching any pattern are skipped.
- **Both specified**: `--include` is applied first, then `--exclude` removes from that set.
- **Directory shorthand**: A plain directory name like `vendor` is automatically treated as `vendor/**`.

Filtered files are completely excluded from the results (they do not appear as skipped).

```bash
# Only validate Go files
git-sandwich --start '# START' --end '# END' --include '**/*.go'

# Exclude vendor and generated directories
git-sandwich --start '# START' --end '# END' --exclude vendor --exclude generated

# Combine: validate Go files except tests
git-sandwich --start '# START' --end '# END' --include '**/*.go' --exclude '**/*_test.go'
```

### Exit Codes

- `0` — All changes are within sandwich blocks (or no protected files were modified).
- `1` — Changes detected outside sandwich blocks.

## How It Works

1. Runs `git diff -U0 base...head` to get changed line ranges.
2. For each modified file, retrieves the file content at both base and head refs.
3. Parses BEGIN/END markers to identify sandwich blocks.
4. Classifies every changed line as **inside**, **boundary**, or **outside**:
   - **inside** (between START+1 and END-1) — allowed.
   - **boundary** (the START or END marker lines themselves) — allowed, unless combined with outside changes.
   - **outside** (everything else) — rejected.
5. Validates both sides of the diff: deletions against base blocks, additions against head blocks. This prevents gaming the system by shifting boundaries to reclassify outside changes as inside.

### Boundary Change Rules

Boundary changes (adding, removing, or modifying BEGIN/END markers) are permitted on their own. However, combining boundary changes with outside changes in the same diff is rejected by default — this prevents disguising outside edits under the cover of a boundary shift.

Use `--allow-boundary-with-outside` to override this behavior, or separate boundary changes into their own commits.

### Block Structure Validation

The following are unconditionally rejected:

- Unmatched BEGIN (no corresponding END)
- Unmatched END (no corresponding BEGIN)
- Reversed order
- Invalid nesting (unless `--allow-nesting` is set)

## Output

### Text (default)

```
FAIL config/application.rb
  outside(base): lines 10-12
  outside(head): lines 25
  note: boundary changed
```

### JSON (`--json`)

```json
{
  "success": false,
  "files": [
    {
      "path": "config/application.rb",
      "success": false,
      "outside_base": [{ "Start": 10, "End": 12 }],
      "outside_head": [{ "Start": 25, "End": 25 }],
      "boundary_changed": true
    }
  ]
}
```

## Examples

### Basic usage

```bash
git-sandwich \
  --start '# EDITABLE START' \
  --end '# EDITABLE END'
```

### With custom base/head refs

```bash
git-sandwich \
  --start '# EDITABLE START' \
  --end '# EDITABLE END' \
  --base origin/develop \
  --head feature-branch
```

### Validate specific paths

```bash
git-sandwich \
  --start '# EDITABLE START' \
  --end '# EDITABLE END' \
  config/ lib/templates/
```

### CI integration (GitHub Actions)

```yaml
- name: Validate sandwich blocks
  run: |
    git-sandwich \
      --start '# CUSTOM START' \
      --end '# CUSTOM END' \
      --base origin/main \
      --head ${{ github.sha }}
```

## Development

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Build
go build -o git-sandwich .
```

## License

See [LICENSE](LICENSE).
