# VSCode Agent Hooks - Comprehensive Reference

> **Status**: Preview (VS Code 1.109.3+)  
> **Last Updated**: February 2026  
> **Official Documentation**: https://code.visualstudio.com/docs/copilot/customization/hooks

## Overview

VSCode Agent Hooks enable custom shell commands to execute at key lifecycle points during agent sessions. They provide **deterministic, code-driven automation** that:

- Enforces security policies
- Automates code quality workflows
- Creates audit trails
- Injects context into conversations
- Controls tool approvals

Hooks work across agent types: **local agents**, **background agents**, and **cloud agents**.

---

## Hook Lifecycle Events

VS Code supports **eight hook events** that fire at specific points:

| Hook Event | When It Fires | Common Use Cases |
|------------|---------------|------------------|
| `SessionStart` | User submits first prompt of new session | Initialize resources, log session start, validate project state |
| `UserPromptSubmit` | User submits a prompt | Audit user requests, inject system context |
| `PreToolUse` | Before agent invokes any tool | Block dangerous operations, require approval, modify tool input |
| `PostToolUse` | After tool completes successfully | Run formatters, log results, trigger follow-up actions |
| `PreCompact` | Before conversation context is compacted | Export important context, save state before truncation |
| `SubagentStart` | Subagent is spawned | Track nested agent usage, initialize subagent resources |
| `SubagentStop` | Subagent completes | Aggregate results, cleanup subagent resources |
| `Stop` | Agent session ends | Generate reports, cleanup resources, send notifications |

### Lifecycle Diagram

```
┌─────────────────┐
│  SessionStart   │ ← Session begins
└────────┬────────┘
         ▼
┌─────────────────┐
│ UserPromptSubmit│ ← User sends message
└────────┬────────┘
         ▼
    ┌────┴────┐
    │  Agent  │
    │ Thinking│
    └────┬────┘
         ▼
┌─────────────────┐
│   PreToolUse    │ ← Can BLOCK or MODIFY
└────────┬────────┘
         ▼
    ┌────┴────┐
    │  Tool   │
    │Execution│
    └────┬────┘
         ▼
┌─────────────────┐
│  PostToolUse    │ ← Can inject context
└────────┬────────┘
         ▼
    (repeat for each tool)
         │
         ▼
┌─────────────────┐
│     Stop        │ ← Session ends (can block)
└─────────────────┘
```

---

## Configuration

### File Locations

VS Code searches for hook configuration files in these locations (workspace takes precedence):

| Location | Path | Scope |
|----------|------|-------|
| **Workspace** | `.github/hooks/*.json` | Project-specific, shared with team |
| **Workspace** | `.claude/settings.local.json` | Local workspace hooks (not committed) |
| **Workspace** | `.claude/settings.json` | Workspace-level hooks |
| **User** | `~/.claude/settings.json` | Personal hooks, all workspaces |

### Configuration Format

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "type": "command",
        "command": "./scripts/validate-tool.sh",
        "timeout": 15
      }
    ],
    "PostToolUse": [
      {
        "type": "command",
        "command": "npx prettier --write \"$TOOL_INPUT_FILE_PATH\""
      }
    ]
  }
}
```

### Hook Command Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | string | **Required**: Must be `"command"` |
| `command` | string | Default command (cross-platform) |
| `windows` | string | Windows-specific override |
| `linux` | string | Linux-specific override |
| `osx` | string | macOS-specific override |
| `cwd` | string | Working directory (relative to repo root) |
| `env` | object | Additional environment variables |
| `timeout` | number | Timeout in seconds (default: 30) |

### OS-Specific Commands Example

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "type": "command",
        "command": "./scripts/format.sh",
        "windows": "powershell -File scripts\\format.ps1",
        "linux": "./scripts/format-linux.sh",
        "osx": "./scripts/format-mac.sh"
      }
    ]
  }
}
```

---

## Hook Input/Output Protocol

Hooks communicate via **stdin** (JSON input) and **stdout** (JSON output).

### Common Input Fields

Every hook receives:

```json
{
  "timestamp": "2026-02-09T10:30:00.000Z",
  "cwd": "/path/to/workspace",
  "sessionId": "session-identifier",
  "hookEventName": "PreToolUse",
  "transcript_path": "/path/to/transcript.json"
}
```

### Common Output Format

```json
{
  "continue": true,
  "stopReason": "Security policy violation",
  "systemMessage": "Operation blocked by security hook"
}
```

### Exit Codes

| Exit Code | Behavior |
|-----------|----------|
| `0` | Success: parse stdout as JSON |
| `2` | **Blocking error**: stop processing, show error to model |
| Other | Non-blocking warning: show warning, continue |

---

## Hook-Specific Details

### PreToolUse

Fires **before** the agent invokes a tool. Can **block**, **allow**, or **modify** tool execution.

#### Input

```json
{
  "tool_name": "editFiles",
  "tool_input": { "files": ["src/main.ts"] },
  "tool_use_id": "tool-123"
}
```

#### Output

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "deny",
    "permissionDecisionReason": "Destructive command blocked",
    "updatedInput": { "files": ["src/safe.ts"] },
    "additionalContext": "User has read-only access"
  }
}
```

#### Permission Decision Values

| Value | Effect |
|-------|--------|
| `"deny"` | **Most restrictive** - blocks tool execution |
| `"ask"` | Requires user confirmation |
| `"allow"` | Auto-approves execution |

**Priority**: When multiple hooks run, the most restrictive decision wins.

---

### PostToolUse

Fires **after** a tool completes successfully.

#### Input

```json
{
  "tool_name": "editFiles",
  "tool_input": { "files": ["src/main.ts"] },
  "tool_use_id": "tool-123",
  "tool_response": "File edited successfully"
}
```

#### Output

```json
{
  "decision": "block",
  "reason": "Post-processing validation failed",
  "hookSpecificOutput": {
    "hookEventName": "PostToolUse",
    "additionalContext": "The edited file has lint errors"
  }
}
```

---

### SessionStart

Fires when a new agent session begins.

#### Input

```json
{
  "source": "new"
}
```

#### Output

```json
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "Project: my-app v2.1.0 | Branch: main"
  }
}
```

---

### Stop

Fires when the agent session ends. **Can prevent stopping**.

#### Input

```json
{
  "stop_hook_active": false
}
```

#### Output

```json
{
  "hookSpecificOutput": {
    "hookEventName": "Stop",
    "decision": "block",
    "reason": "Run the test suite before finishing"
  }
}
```

> ⚠️ **Warning**: Always check `stop_hook_active` to prevent infinite loops!

---

### SubagentStart / SubagentStop

Track and control subagent spawning.

#### SubagentStart Input

```json
{
  "agent_id": "subagent-456",
  "agent_type": "Plan"
}
```

#### SubagentStop Output (can block)

```json
{
  "decision": "block",
  "reason": "Verify subagent results before completing"
}
```

---

### PreCompact

Fires before conversation context is compacted (truncated for prompt budget).

#### Input

```json
{
  "trigger": "auto"
}
```

---

## Practical Examples

### 1. Block Dangerous Commands

**.github/hooks/security.json**:
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "type": "command",
        "command": "./scripts/block-dangerous.sh",
        "timeout": 5
      }
    ]
  }
}
```

**scripts/block-dangerous.sh**:
```bash
#!/bin/bash
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name')
TOOL_INPUT=$(echo "$INPUT" | jq -r '.tool_input')

if [ "$TOOL_NAME" = "runTerminalCommand" ]; then
  COMMAND=$(echo "$TOOL_INPUT" | jq -r '.command // empty')

  if echo "$COMMAND" | grep -qE '(rm\s+-rf|DROP\s+TABLE|DELETE\s+FROM)'; then
    echo '{"hookSpecificOutput":{"permissionDecision":"deny","permissionDecisionReason":"Destructive command blocked by security policy"}}'
    exit 0
  fi
fi

echo '{"continue":true}'
```

---

### 2. Auto-Format Code After Edits

**.github/hooks/formatting.json**:
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "type": "command",
        "command": "./scripts/format-changed-files.sh",
        "windows": "powershell -File scripts\\format-changed-files.ps1",
        "timeout": 30
      }
    ]
  }
}
```

**scripts/format-changed-files.sh**:
```bash
#!/bin/bash
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name')

if [ "$TOOL_NAME" = "editFiles" ] || [ "$TOOL_NAME" = "createFile" ]; then
  FILES=$(echo "$INPUT" | jq -r '.tool_input.files[]? // .tool_input.path // empty')

  for FILE in $FILES; do
    if [ -f "$FILE" ]; then
      npx prettier --write "$FILE" 2>/dev/null
    fi
  done
fi

echo '{"continue":true}'
```

---

### 3. Audit Trail Logging

**.github/hooks/audit.json**:
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "type": "command",
        "command": "./scripts/log-tool-use.sh",
        "env": {
          "AUDIT_LOG": ".github/hooks/audit.log"
        }
      }
    ]
  }
}
```

**scripts/log-tool-use.sh**:
```bash
#!/bin/bash
INPUT=$(cat)
TIMESTAMP=$(echo "$INPUT" | jq -r '.timestamp')
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name')
SESSION_ID=$(echo "$INPUT" | jq -r '.sessionId')

echo "[$TIMESTAMP] Session: $SESSION_ID, Tool: $TOOL_NAME" >> "${AUDIT_LOG:-audit.log}"
echo '{"continue":true}'
```

---

### 4. Require Approval for Sensitive Tools

**.github/hooks/approval.json**:
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "type": "command",
        "command": "./scripts/require-approval.sh"
      }
    ]
  }
}
```

**scripts/require-approval.sh**:
```bash
#!/bin/bash
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name')

SENSITIVE_TOOLS="runTerminalCommand|deleteFile|pushToGitHub"

if echo "$TOOL_NAME" | grep -qE "^($SENSITIVE_TOOLS)$"; then
  echo '{"hookSpecificOutput":{"permissionDecision":"ask","permissionDecisionReason":"This operation requires manual approval"}}'
else
  echo '{"hookSpecificOutput":{"permissionDecision":"allow"}}'
fi
```

---

### 5. Inject Project Context at Session Start

**.github/hooks/context.json**:
```json
{
  "hooks": {
    "SessionStart": [
      {
        "type": "command",
        "command": "./scripts/inject-context.sh"
      }
    ]
  }
}
```

**scripts/inject-context.sh**:
```bash
#!/bin/bash
PROJECT_INFO=$(cat package.json 2>/dev/null | jq -r '.name + " v" + .version' || echo "Unknown project")
BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")

cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "Project: $PROJECT_INFO | Branch: $BRANCH | Node: $(node -v 2>/dev/null || echo 'not installed')"
  }
}
EOF
```

---

## Cross-Platform Considerations

### Claude Code Compatibility

VS Code parses Claude Code's hook configuration format, including **matcher syntax**. Currently, matchers are ignored (hooks apply to all tools). Claude Code uses `""` as the matcher for all tools.

**Claude Code format (also works in VS Code)**:
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "^(Write|Edit|MultiEdit)$",
        "hooks": [{
          "type": "command",
          "command": "./scripts/validate.sh"
        }]
      }
    ]
  }
}
```

### Copilot CLI Compatibility

VS Code converts Copilot CLI's **lowerCamelCase** event names to **PascalCase**:

| Copilot CLI | VS Code |
|-------------|---------|
| `preToolUse` | `PreToolUse` |
| `postToolUse` | `PostToolUse` |
| `sessionStart` | `SessionStart` |

Both `bash` and `powershell` command formats are supported:

```json
{
  "hooks": {
    "preToolUse": [{
      "type": "command",
      "bash": "./scripts/validate.sh",
      "powershell": "scripts\\validate.ps1"
    }]
  }
}
```

### PowerShell Gotchas

When running PowerShell commands:

1. **Use script files** instead of inline commands to avoid quoting issues
2. **Avoid `powershell.exe` (5.1)** - use `pwsh` (PowerShell Core 7+)
3. **Backslash paths** can be mangled by bash - use forward slashes in scripts or escape properly

---

## Troubleshooting

### View Hook Diagnostics

1. Right-click in the Chat view → **Diagnostics**
2. Look for the hooks section to see loaded hooks and validation errors

### View Hook Output

1. Open the **Output** panel
2. Select **GitHub Copilot Chat Hooks** from the channel list

### Common Issues

| Issue | Solution |
|-------|----------|
| Hook not executing | Verify file is in `.github/hooks/` with `.json` extension. Check `type: "command"` |
| Permission denied | Ensure scripts have execute permissions (`chmod +x script.sh`) |
| Timeout errors | Increase `timeout` value or optimize script (default: 30s) |
| JSON parse errors | Verify hook outputs valid JSON to stdout. Use `jq` to construct output |

### Configure via /hooks Command

Type `/hooks` in chat to configure hooks through an interactive UI:
1. Select a hook event type
2. Choose existing hook or create new
3. Select or create a configuration file

---

## Security Considerations

> ⚠️ **Hooks execute shell commands with the same permissions as VS Code.**

### Best Practices

1. **Review hook scripts** before enabling, especially from untrusted sources
2. **Limit hook permissions** - principle of least privilege
3. **Validate input** - sanitize all input to prevent injection attacks
4. **Secure credentials** - use environment variables, not hardcoded secrets
5. **Protect hook scripts** - use `chat.tools.edits.autoApprove` to prevent agent from modifying hooks

---

## Comparison: Hooks vs Other Customization

| Feature | Purpose | When to Use |
|---------|---------|-------------|
| **Hooks** | Deterministic code execution | Security gates, automation, auditing |
| **Instructions** | Guide agent behavior | Team standards, coding conventions |
| **Prompt Files** | Reusable prompt templates | Named workflows, task templates |
| **Custom Agents** | Specialized agent roles | Domain-specific expertise |
| **MCP Servers** | External tool integration | APIs, databases, services |
| **Skills** | Extended capabilities | Teach agent new abilities |

---

## Official Resources

- **VS Code Hooks Documentation**: https://code.visualstudio.com/docs/copilot/customization/hooks
- **GitHub Copilot Hooks Guide**: https://docs.github.com/en/copilot/how-tos/use-copilot-agents/coding-agent/use-hooks
- **VS Code Enterprise Policies**: https://code.visualstudio.com/docs/enterprise/policies
- **Tool Approval Reference**: https://code.visualstudio.com/docs/copilot/agents/agent-tools

---

## Summary

VSCode Agent Hooks provide **deterministic control** over probabilistic AI agents. Key capabilities:

- **8 lifecycle events** covering session, tool, and subagent phases
- **JSON protocol** for structured input/output
- **Permission control** for PreToolUse (deny, ask, allow)
- **Cross-platform support** with OS-specific command overrides
- **Compatibility** with Claude Code and Copilot CLI formats

Use hooks when you need **guaranteed outcomes** - security enforcement, automated formatting, audit logging, and workflow automation.
