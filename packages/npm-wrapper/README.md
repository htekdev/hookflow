# hookflow

Local workflow engine for agentic DevOps - run GitHub Actions-like workflows triggered by Copilot agent hooks.

## Installation

```bash
npm install -g hookflow
```

Or with other package managers:

```bash
# Go
go install github.com/htekdev/hookflow/cmd/hookflow@latest

# Direct download
curl -sSL https://raw.githubusercontent.com/htekdev/hookflow/main/scripts/install.sh | sh
```

## Quick Start

```bash
# Initialize hookflow in your repo
cd your-project
hookflow init

# Create a workflow using AI
hookflow create "block edits to .env files"

# Test a workflow with a mock event
hookflow test --event commit --workflow lint.yml

# List workflows
hookflow discover
```

## What is hookflow?

hookflow lets you run "shift-left" DevOps checks during AI agent editing sessions. Instead of waiting for CI to catch issues on pull requests, you can:

- **Block** dangerous edits in real-time (e.g., .env file modifications)
- **Lint** code as the agent writes it
- **Validate** configurations before commit
- **Run security scans** before code leaves the local machine

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize hookflow for a repository |
| `create <prompt>` | Create a workflow using AI (Copilot SDK) |
| `shift left` | Analyze CI workflows and suggest hooks |
| `shift right` | Generate GitHub Actions from hooks |
| `discover` | List workflows in the current repository |
| `validate` | Validate workflow YAML files |
| `test` | Test a workflow with a mock event |
| `run` | Run workflows (used by hooks) |
| `version` | Show version information |

## Workflow Syntax

Workflows use a GitHub Actions-like syntax:

```yaml
name: Block .env edits
description: Prevent modifications to environment files

on:
  file:
    paths:
      - '**/.env*'
      - '**/secrets/**'

steps:
  - name: Block sensitive file edit
    run: |
      echo "Blocked: Cannot modify sensitive files"
      exit 1
```

## Learn More

- [GitHub Repository](https://github.com/htekdev/hookflow)
- [Workflow Schema Documentation](https://github.com/htekdev/hookflow/blob/main/schema/workflow.schema.json)

## License

MIT
