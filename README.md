# hookflow

Local workflow engine for agentic DevOps - run GitHub Actions-like workflows triggered by Copilot agent hooks.

## Overview

`hookflow` lets you run "shift-left" DevOps checks during AI agent editing sessions. Instead of waiting for CI to catch issues on pull requests, you can:

- **Block** dangerous edits in real-time (e.g., .env file modifications)
- **Lint** code as the agent writes it
- **Validate** configurations before commit
- **Run security scans** before code leaves the local machine

## Installation

### npm (Recommended)

```bash
npm install -g hookflow
```

### Go

```bash
go install github.com/htekdev/hookflow/cmd/hookflow@latest
```

### Install Script (Unix)

```bash
curl -sSL https://raw.githubusercontent.com/htekdev/hookflow/main/scripts/install.sh | sh
```

### Install Script (Windows)

```powershell
iwr -useb https://raw.githubusercontent.com/htekdev/hookflow/main/scripts/install.ps1 | iex
```

### Download Binary

Download pre-built binaries from the [Releases](https://github.com/htekdev/hookflow/releases) page.

## Quick Start

```bash
# Initialize hookflow for your repository
cd your-project
hookflow init

# Test a workflow with a mock event
hookflow test --event commit --workflow lint.yml

# Discover workflows in the current directory
hookflow discover
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize hookflow for a repository |
| `discover` | List workflows in the current directory |
| `validate` | Validate workflow YAML files |
| `test` | Test a workflow with a mock event |
| `run` | Run workflows (used by hooks) |
| `triggers` | List available trigger types |
| `version` | Show version information |

## Usage

```bash
# Initialize a repository (creates .github/hooks/ and .copilot/hooks.json)
hookflow init

# Discover workflows in the current directory
hookflow discover

# Validate workflow files
hookflow validate

# Test a workflow with a mock commit event
hookflow test --event commit --path src/app.ts

# Test a workflow with a mock file event
hookflow test --event file --action edit --path src/app.ts

# Run workflows for an event (used by hooks)
hookflow run --raw --dir .
```

## Workflow Syntax

Workflows are defined in `.github/hooks/*.yml`:

```yaml
name: Block Sensitive Files
description: Prevent edits to sensitive files

on:
  tool:
    name: edit
    args:
      path: '**/*.env*'

blocking: true

steps:
  - name: Deny edit
    run: |
      echo "Cannot edit sensitive files"
      exit 1
```

## Event Types

| Trigger | Description |
|---------|-------------|
| `hooks` | Match by hook type (preToolUse, postToolUse) |
| `tool` | Match specific tools with argument patterns |
| `tools` | Match multiple tool configurations |
| `file` | Match file creation/edit events |
| `commit` | Match git commit events |
| `push` | Match git push events |

## Expression Engine

Supports `${{ }}` expressions with GitHub Actions parity:

```yaml
steps:
  - name: Conditional step
    if: ${{ endsWith(event.file.path, '.ts') }}
    run: echo "TypeScript file: ${{ event.file.path }}"
```

### Built-in Functions

- `contains(search, item)` - Check if string/array contains item
- `startsWith(str, value)` - String starts with value
- `endsWith(str, value)` - String ends with value
- `format(str, ...args)` - String formatting
- `join(array, sep)` - Join array to string
- `toJSON(value)` - Convert to JSON string
- `fromJSON(str)` - Parse JSON string
- `always()` - Always true
- `success()` - Previous steps succeeded
- `failure()` - Previous step failed

## Development

```bash
# Build
go build -o bin/hookflow ./cmd/hookflow

# Test
go test ./... -v

# Test with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Related Projects

- [hookflow](https://github.com/htekdev/hookflow) - Copilot plugin that uses this CLI

## License

MIT
