# GitHub Copilot CLI Hooks - Comprehensive Reference

> **Research Document** - Created for the hookflow project  
> **Last Updated:** 2025-07-02  
> **Sources:** Official GitHub Documentation + Local hookflow Implementation

## Table of Contents

1. [Overview](#overview)
2. [Hook Types](#hook-types)
3. [Configuration Format](#configuration-format)
4. [Event Context](#event-context)
5. [Permission Decisions](#permission-decisions)
6. [Input/Output Formats](#inputoutput-formats)
7. [hookflow Implementation](#hookflow-implementation)
8. [Practical Examples](#practical-examples)
9. [Best Practices](#best-practices)
10. [Official Documentation Links](#official-documentation-links)

---

## Overview

### What Are Copilot CLI Hooks?

Hooks are custom scripts that execute at specific points during a GitHub Copilot CLI agent session. They enable you to:

- **Extend behavior** - Execute custom shell commands at key lifecycle points
- **Enforce policies** - Block dangerous operations before they execute
- **Audit actions** - Log prompts and tool usage for compliance
- **Integrate systems** - Connect to external validation or notification services
- **Control execution** - Approve or deny tool calls in real-time

Hooks run **deterministically** and can control agent behavior, including blocking tool execution or injecting context into the conversation.

### Key Capabilities

| Capability | Description |
|------------|-------------|
| **Blocking** | Deny tool execution before it happens |
| **Logging** | Record prompts, tool calls, and results |
| **Validation** | Run security scans or lint checks |
| **Notification** | Alert teams on specific actions |
| **Context Injection** | Provide additional information to the agent |

---

## Hook Types

GitHub Copilot CLI supports **six hook types** that fire at different points in the agent session lifecycle:

### 1. Session Start Hook (`sessionStart`)

**When:** Fires when a new agent session begins or when resuming an existing session.

**Use Cases:**
- Display policy banners
- Initialize logging
- Set up session-specific environment

**Input Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | number | Unix timestamp in milliseconds |
| `cwd` | string | Current working directory |
| `source` | string | `"new"`, `"resume"`, or `"startup"` |
| `initialPrompt` | string | The user's initial prompt (if provided) |

**Output:** Ignored (no return value processed)

### 2. Session End Hook (`sessionEnd`)

**When:** Fires when the agent session completes or is terminated.

**Use Cases:**
- Cleanup temporary files
- Finalize audit logs
- Send session summary notifications

**Input Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | number | Unix timestamp in milliseconds |
| `cwd` | string | Current working directory |
| `reason` | string | `"complete"`, `"error"`, `"abort"`, `"timeout"`, or `"user_exit"` |

**Output:** Ignored

### 3. User Prompt Submitted Hook (`userPromptSubmitted`)

**When:** Fires when the user submits a prompt to the agent.

**Use Cases:**
- Log all user prompts for audit
- Detect sensitive keywords
- Track usage patterns

**Input Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | number | Unix timestamp in milliseconds |
| `cwd` | string | Current working directory |
| `prompt` | string | The exact text the user submitted |

**Output:** Ignored (prompt modification not currently supported)

### 4. Pre-Tool Use Hook (`preToolUse`) ⭐

**When:** Fires **before** the agent uses any tool (edit, create, bash, view, etc.)

**This is the most powerful hook** - it can approve or deny tool executions.

**Use Cases:**
- Block dangerous commands (`rm -rf`, `sudo`, etc.)
- Prevent edits to sensitive files (`.env`, secrets)
- Enforce security policies
- Validate operations before execution

**Input Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | number | Unix timestamp in milliseconds |
| `cwd` | string | Current working directory |
| `toolName` | string | Name of the tool (`"bash"`, `"edit"`, `"view"`, `"create"`) |
| `toolArgs` | string | JSON string containing the tool's arguments |

**Output Fields (optional JSON):**
| Field | Type | Description |
|-------|------|-------------|
| `permissionDecision` | string | `"allow"`, `"deny"`, or `"ask"` |
| `permissionDecisionReason` | string | Human-readable explanation |

> **Note:** Currently only `"deny"` is actively processed. Omitting output or returning `"allow"` allows execution.

### 5. Post-Tool Use Hook (`postToolUse`)

**When:** Fires **after** the agent uses a tool.

**Use Cases:**
- Log tool results for audit
- Track execution metrics
- Analyze patterns of tool usage
- Send notifications on specific outcomes

**Input Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | number | Unix timestamp in milliseconds |
| `cwd` | string | Current working directory |
| `toolName` | string | Name of the tool that was executed |
| `toolArgs` | string | JSON string of tool arguments |
| `toolResult` | object | Contains `resultType` and result data |

**Output:** Ignored

### 6. Error Occurred Hook (`errorOccurred`)

**When:** Fires when an error occurs during agent execution.

**Use Cases:**
- Log errors for debugging
- Send alerts to monitoring systems
- Track error patterns

**Input Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | number | Unix timestamp in milliseconds |
| `cwd` | string | Current working directory |
| `error` | object | Contains `message` and error details |

**Output:** Ignored

---

## Configuration Format

### File Location

Hooks are configured in JSON files located at:

```
.github/hooks/*.json
```

For **Copilot CLI**, hooks are loaded from the current working directory's `.github/hooks/` folder.

### Schema

```json
{
  "version": 1,
  "hooks": {
    "sessionStart": [...],
    "sessionEnd": [...],
    "userPromptSubmitted": [...],
    "preToolUse": [...],
    "postToolUse": [...],
    "errorOccurred": [...]
  }
}
```

### Hook Entry Format

Each hook entry in an array follows this format:

```json
{
  "type": "command",
  "bash": "./scripts/my-hook.sh",
  "powershell": "./scripts/my-hook.ps1",
  "cwd": ".github/hooks",
  "timeoutSec": 30,
  "env": {
    "LOG_LEVEL": "INFO"
  },
  "comment": "Description of what this hook does"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | Always `"command"` |
| `bash` | string | Conditional | Script/command for Unix systems |
| `powershell` | string | Conditional | Script/command for Windows |
| `cwd` | string | No | Working directory for the script |
| `timeoutSec` | number | No | Timeout in seconds (default: 30) |
| `env` | object | No | Environment variables for the script |
| `comment` | string | No | Documentation comment |

> **Note:** At least one of `bash` or `powershell` is required.

### Multiple Hooks

You can define multiple hooks for the same event. They execute in order:

```json
{
  "preToolUse": [
    { "type": "command", "bash": "./security-check.sh", "comment": "Runs first" },
    { "type": "command", "bash": "./audit-log.sh", "comment": "Runs second" },
    { "type": "command", "bash": "./metrics.sh", "comment": "Runs third" }
  ]
}
```

---

## Event Context

### Tool Names

Common tools that can be intercepted:

| Tool Name | Description |
|-----------|-------------|
| `bash` | Shell command execution |
| `powershell` | PowerShell command execution |
| `edit` | File editing |
| `create` | File creation |
| `view` | File viewing |
| `terminal` | Terminal operations |
| `shell` | Generic shell operations |

### Tool Arguments by Type

#### bash/powershell
```json
{
  "command": "npm install",
  "description": "Install dependencies"
}
```

#### edit
```json
{
  "path": "src/utils/helper.js",
  "old_str": "function old() {}",
  "new_str": "function new() {}"
}
```

#### create
```json
{
  "path": "src/new-file.ts",
  "file_text": "// New file content"
}
```

### Example Input JSON (preToolUse)

```json
{
  "timestamp": 1704614600000,
  "cwd": "/path/to/project",
  "toolName": "bash",
  "toolArgs": "{\"command\":\"rm -rf dist\",\"description\":\"Clean build directory\"}"
}
```

---

## Permission Decisions

### Decision Types

| Decision | Effect |
|----------|--------|
| `"allow"` | Tool execution proceeds (default if no output) |
| `"deny"` | Tool execution is blocked |
| `"ask"` | Prompt user for confirmation (not fully supported) |

### Output Format

```json
{
  "permissionDecision": "deny",
  "permissionDecisionReason": "Destructive operations require approval"
}
```

### Decision Logic

1. **No output** → Allow (implicit)
2. **Output with `"allow"`** → Allow
3. **Output with `"deny"`** → Block with reason displayed to agent
4. **Script exits non-zero** → Allow (error handling, not denial)

---

## Input/Output Formats

### Reading Input (Bash)

```bash
#!/bin/bash
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')
TOOL_ARGS=$(echo "$INPUT" | jq -r '.toolArgs')
COMMAND=$(echo "$TOOL_ARGS" | jq -r '.command')
```

### Reading Input (PowerShell)

```powershell
$input = [Console]::In.ReadToEnd() | ConvertFrom-Json
$toolName = $input.toolName
$toolArgs = $input.toolArgs | ConvertFrom-Json
$command = $toolArgs.command
```

### Writing Output (Bash)

```bash
# Deny
echo '{"permissionDecision":"deny","permissionDecisionReason":"Dangerous command detected"}'

# Allow (explicit)
echo '{"permissionDecision":"allow"}'

# Allow (implicit - no output)
exit 0
```

### Writing Output (PowerShell)

```powershell
$output = @{
    permissionDecision = "deny"
    permissionDecisionReason = "Dangerous command detected"
}
$output | ConvertTo-Json -Compress
```

---

## hookflow Implementation

### Overview

The `hookflow` project extends Copilot CLI hooks with a **GitHub Actions-like workflow engine**. Instead of raw shell scripts, you define YAML workflows that are triggered by hook events.

### Architecture

```
.github/hooks/
├── hooks.json          # Copilot hook config (calls hookflow)
└── workflows/          # hookflow workflow definitions
    ├── lint.yml
    ├── security.yml
    └── block-secrets.yml
```

### Workflow Schema

```yaml
name: Block Sensitive Files
description: Prevent edits to sensitive files
blocking: true  # Default: true - failures deny the operation

on:
  hooks:
    types: [preToolUse]
    tools: [edit, create]
  
  tool:
    name: edit
    args:
      path: '**/*.env*'  # Glob pattern matching

concurrency:
  group: validation-${{ event.cwd }}
  max-parallel: 2

env:
  STRICT_MODE: true

steps:
  - name: Deny edit
    run: |
      echo "Cannot edit sensitive files"
      exit 1  # Non-zero = deny (if blocking: true)
```

### Trigger Types

hookflow supports six trigger types (OR logic - any match triggers):

| Trigger | Description | Example |
|---------|-------------|---------|
| `hooks` | Match by hook type | `types: [preToolUse]` |
| `tool` | Match single tool + args | `name: edit, args.path: '**/*.env*'` |
| `tools` | Match multiple tool configs | Array of tool triggers |
| `file` | Match file events | `types: [create, edit], paths: ['src/**']` |
| `commit` | Match git commits | `paths: ['src/**'], branches: [main]` |
| `push` | Match git pushes | `tags: ['v*'], branches: [main]` |

### Expression Engine

hookflow supports `${{ }}` expressions with GitHub Actions parity:

```yaml
steps:
  - name: TypeScript only
    if: ${{ endsWith(event.file.path, '.ts') }}
    run: npm run lint -- "${{ event.file.path }}"
```

**Available Functions:**
- `contains(search, item)` - Check if string/array contains item
- `startsWith(str, value)` - String starts with value
- `endsWith(str, value)` - String ends with value
- `format(str, ...args)` - String formatting
- `join(array, sep)` - Join array to string
- `toJSON(value)` / `fromJSON(str)` - JSON conversion
- `always()` - Always true
- `success()` - Previous steps succeeded
- `failure()` - Previous step failed

### Event Context in hookflow

The `event` object provides rich context:

```yaml
# Hook context
event.hook.type        # "preToolUse" or "postToolUse"
event.hook.tool.name   # "edit", "create", "bash", etc.
event.hook.cwd         # Working directory

# Tool context
event.tool.name        # Tool name
event.tool.args        # Tool arguments (object)
event.tool.args.path   # For file operations
event.tool.args.command # For shell operations

# File context
event.file.path        # File path
event.file.action      # "create" or "edit"
event.file.content     # File content (if available)

# Git context
event.commit.sha       # Commit SHA
event.commit.message   # Commit message
event.commit.files     # Array of files

# Global
event.cwd              # Current working directory
event.timestamp        # ISO timestamp
event.lifecycle        # "pre" or "post"
```

### Blocking vs Non-Blocking

| `blocking` Value | Step Failure | Result |
|------------------|--------------|--------|
| `true` (default) | Any step fails | **DENY** operation |
| `false` | Any step fails | **ALLOW** (with warning) |

---

## Practical Examples

### Example 1: Block Dangerous Shell Commands

**hooks.json:**
```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {
        "type": "command",
        "bash": "./scripts/block-dangerous.sh",
        "powershell": "./scripts/block-dangerous.ps1",
        "cwd": ".github/hooks",
        "timeoutSec": 10
      }
    ]
  }
}
```

**block-dangerous.sh:**
```bash
#!/bin/bash
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')

if [ "$TOOL_NAME" != "bash" ]; then
  exit 0  # Allow non-bash tools
fi

COMMAND=$(echo "$INPUT" | jq -r '.toolArgs' | jq -r '.command')

# Check for dangerous patterns
if echo "$COMMAND" | grep -qE "rm -rf /|sudo|mkfs|dd if="; then
  echo '{"permissionDecision":"deny","permissionDecisionReason":"Dangerous system command blocked"}'
  exit 0
fi

# Allow by default
exit 0
```

### Example 2: Block .env File Edits (hookflow)

**.github/hooks/block-env.yml:**
```yaml
name: Block Environment Files
description: Prevent AI from editing .env files
blocking: true

on:
  hooks:
    types: [preToolUse]
  tool:
    name: edit
    args:
      path: '**/*.env*'

steps:
  - name: Deny edit
    run: |
      echo "Error: Cannot edit environment files"
      echo "File: ${{ event.file.path }}"
      exit 1
```

### Example 3: Lint TypeScript Files

**.github/hooks/lint-ts.yml:**
```yaml
name: Lint TypeScript
description: Run ESLint on TypeScript file changes
blocking: false  # Don't block on lint failures

on:
  file:
    types: [create, edit]
    paths: ['**/*.ts', '**/*.tsx']
    paths-ignore: ['**/*.test.ts', '**/*.spec.ts']

steps:
  - name: Run ESLint
    if: ${{ endsWith(event.file.path, '.ts') }}
    run: npx eslint "${{ event.file.path }}" --fix
    continue-on-error: true
```

### Example 4: Audit All Tool Usage

**.github/hooks/audit-tools.yml:**
```yaml
name: Audit Tool Usage
description: Log all tool executions
blocking: false

on:
  hooks:
    types: [preToolUse, postToolUse]

steps:
  - name: Log to audit file
    run: |
      $timestamp = Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ"
      $logEntry = @{
        timestamp = $timestamp
        hookType = "${{ event.hook.type }}"
        tool = "${{ event.tool.name }}"
        cwd = "${{ event.cwd }}"
      } | ConvertTo-Json -Compress
      Add-Content -Path ".github/hooks/logs/audit.jsonl" -Value $logEntry
```

### Example 5: Comprehensive Validation Workflow

**.github/hooks/validate.yml:**
```yaml
name: Comprehensive Validation
description: Multi-step validation workflow
blocking: true

on:
  hooks:
    types: [preToolUse]
    tools: [edit, create]
  file:
    types: [create, edit]
    paths: ['src/**']
    paths-ignore: ['**/*.test.*']

concurrency:
  group: validate-${{ event.cwd }}
  max-parallel: 1

env:
  NODE_ENV: development

steps:
  - name: Check file size
    if: ${{ event.file.action == 'create' }}
    run: |
      # Block files larger than 1MB
      if (Test-Path "${{ event.file.path }}") {
        $size = (Get-Item "${{ event.file.path }}").Length
        if ($size -gt 1MB) {
          Write-Error "File too large: $size bytes"
          exit 1
        }
      }

  - name: Lint code
    if: ${{ endsWith(event.file.path, '.ts') }}
    run: npx eslint "${{ event.file.path }}"
    continue-on-error: true

  - name: Security scan
    run: npm run security:check -- --file "${{ event.file.path }}"
    timeout: 60

  - name: Always log
    if: ${{ always() }}
    run: echo "Validation completed for ${{ event.file.path }}"
```

---

## Best Practices

### Script Development

1. **Always handle JSON input properly**
   - Use `jq` (Bash) or `ConvertFrom-Json` (PowerShell)
   - Handle missing or null fields gracefully

2. **Output valid JSON**
   - Use `jq -c` or `ConvertTo-Json -Compress` for single-line output
   - Validate output format during development

3. **Use appropriate timeouts**
   - Default is 30 seconds
   - Increase for slower operations (builds, tests)
   - Keep hooks fast for good UX

4. **Handle errors gracefully**
   - Set `set -e` (Bash) or `$ErrorActionPreference = "Stop"` (PowerShell)
   - Log errors before exiting

### Policy Design

1. **Start with logging-only**
   - Deploy hooks that log but don't block
   - Analyze patterns before enforcing

2. **Be specific, not broad**
   - Target specific dangerous patterns
   - Avoid overly broad rules that frustrate users

3. **Provide clear reasons**
   - Always include `permissionDecisionReason`
   - Help users understand why an action was blocked

4. **Test thoroughly**
   - Test with sample inputs before deployment
   - Verify both allow and deny paths

### Security Considerations

1. **Never log secrets**
   - Redact sensitive data from prompts/commands
   - Use environment variables for credentials

2. **Validate input**
   - Don't trust tool arguments blindly
   - Sanitize paths and commands

3. **Limit permissions**
   - Run hooks with minimal privileges
   - Avoid `sudo` in hook scripts

---

## Official Documentation Links

### Primary References

- **Hooks Configuration Reference**: https://docs.github.com/en/copilot/reference/hooks-configuration
- **Using Hooks with Copilot CLI**: https://docs.github.com/en/copilot/how-tos/copilot-cli/use-hooks
- **Hooks Tutorial**: https://docs.github.com/en/copilot/tutorials/copilot-cli-hooks
- **About GitHub Copilot CLI**: https://docs.github.com/en/copilot/concepts/agents/about-copilot-cli

### VS Code Integration

- **Agent Hooks in VS Code**: https://code.visualstudio.com/docs/copilot/customization/hooks

### Related Documentation

- **GitHub Actions Expressions**: https://docs.github.com/en/actions/learn-github-actions/expressions
- **Glob Pattern Syntax**: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet

---

## Appendix: hookflow vs Native Hooks

| Feature | Native Hooks | hookflow |
|---------|--------------|----------|
| Configuration | JSON + shell scripts | YAML workflows |
| Trigger Types | 6 hook events | 6 triggers + file/commit/push |
| Expression Engine | Manual in scripts | Built-in `${{ }}` |
| Blocking Control | Script exit code | `blocking:` field |
| Concurrency | Manual | Built-in groups |
| Step Composition | Single script | Multi-step workflows |
| Conditionals | Manual in scripts | `if:` expressions |
| Learning Curve | Lower | Higher (but familiar to GitHub Actions users) |

hookflow is designed for teams already familiar with GitHub Actions who want more sophisticated validation logic without writing complex shell scripts.

---

*This document was created as part of the hookflow project. For updates, see the [hookflow repository](https://github.com/htekdev/hookflow).*
