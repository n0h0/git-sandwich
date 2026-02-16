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

```bash
go install github.com/n0h0/git-sandwich@latest
```

Or build from source:

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

### Required Flags

| Flag              | Description                       |
| ----------------- | --------------------------------- |
| `--start <regex>` | Regex pattern for BEGIN markers   |
| `--end <regex>`   | Regex pattern for END markers     |

### Optional Flags

| Flag                              | Default       | Description                                      |
| --------------------------------- | ------------- | ------------------------------------------------ |
| `--base`                          | `origin/main` | Base ref for comparison                          |
| `--head`                          | `HEAD`        | Head ref for comparison                          |
| `--allow-nesting`                 | `false`       | Allow nested BEGIN/END blocks                    |
| `--allow-boundary-with-outside`   | `false`       | Allow boundary changes together with outside changes |
| `--json`                          | `false`       | Output results in JSON format                    |

Positional arguments `[paths...]` are passed as path filters to `git diff`.

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
