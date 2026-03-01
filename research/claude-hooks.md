# Claude Hooks - Comprehensive Reference Guide

> **Research compiled:** 2025-02-25  
> **Sources:** Official Anthropic documentation, Claude Code docs, Model Context Protocol specs

## Table of Contents

1. [Overview](#overview)
2. [Claude Code Hooks](#claude-code-hooks)
   - [Hook Lifecycle Events](#hook-lifecycle-events)
   - [Configuration Format](#configuration-format)
   - [Exit Codes & Flow Control](#exit-codes--flow-control)
   - [Hook Types](#hook-types)
3. [Agent SDK Hooks](#agent-sdk-hooks)
4. [Model Context Protocol (MCP)](#model-context-protocol-mcp)
5. [Tool Use (Function Calling)](#tool-use-function-calling)
6. [Practical Examples](#practical-examples)
7. [Official Documentation Links](#official-documentation-links)

---

## Overview

Anthropic provides several mechanisms for extending and controlling Claude's behavior:

| System | Purpose | Use Case |
|--------|---------|----------|
| **Claude Code Hooks** | Intercept Claude Code CLI events | Workflow automation, file protection, notifications |
| **Agent SDK Hooks** | Intercept agent execution in SDK apps | Custom agents, approval workflows, logging |
| **MCP (Model Context Protocol)** | Connect Claude to external tools/data | Database access, API integrations, custom tools |
| **Tool Use (Function Calling)** | Define callable functions for Claude | Extend Claude's capabilities via API |

---

## Claude Code Hooks

Claude Code hooks are **user-defined shell commands, HTTP endpoints, or LLM prompts** that execute automatically at specific points in Claude Code's lifecycle. They provide deterministic control over Claude Code's behavior.

### Hook Lifecycle Events

| Event | When It Fires | Can Block? |
|-------|---------------|------------|
| `SessionStart` | When a session begins or resumes | No |
| `UserPromptSubmit` | When you submit a prompt, before Claude processes it | Yes |
| `PreToolUse` | Before a tool call executes | **Yes** |
| `PermissionRequest` | When a permission dialog appears | Yes |
| `PostToolUse` | After a tool call succeeds | No |
| `PostToolUseFailure` | After a tool call fails | No |
| `Notification` | When Claude Code sends a notification | No |
| `SubagentStart` | When a subagent is spawned | No |
| `SubagentStop` | When a subagent finishes | No |
| `Stop` | When Claude finishes responding | No |
| `TeammateIdle` | When an agent team teammate is about to go idle | No |
| `TaskCompleted` | When a task is being marked as completed | Yes |
| `ConfigChange` | When a configuration file changes during a session | No |
| `WorktreeCreate` | When a worktree is being created | Yes |
| `WorktreeRemove` | When a worktree is being removed | No |
| `PreCompact` | Before context compaction | No |
| `SessionEnd` | When a session terminates | No |

### Configuration Format

Hooks are configured in `~/.claude/settings.json` or `.claude/settings.json` (project-level):

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash|Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/my-hook.sh",
            "timeout": 60000
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/format-code.sh",
            "async": true,
            "timeout": 300
          }
        ]
      }
    ]
  }
}
```

#### Configuration Options

| Option | Type | Description |
|--------|------|-------------|
| `matcher` | `string` | Regex pattern to match tool names (e.g., `"Bash"`, `"Write\|Edit"`, `"*"`) |
| `type` | `string` | Hook type: `"command"`, `"http"`, or `"prompt"` |
| `command` | `string` | Shell command to execute (for `type: "command"`) |
| `url` | `string` | HTTP endpoint (for `type: "http"`) |
| `prompt` | `string` | LLM prompt (for `type: "prompt"`) |
| `timeout` | `number` | Timeout in milliseconds |
| `async` | `boolean` | Run asynchronously without blocking |

### Exit Codes & Flow Control

Hooks control Claude Code's behavior through exit codes and JSON output:

| Exit Code | Effect |
|-----------|--------|
| `0` | Allow the operation (success) |
| `2` | Block the operation (for PreToolUse) |
| Non-zero | Error, may block depending on hook type |

#### JSON Output Schema

Hooks can return JSON to stdout to provide decisions:

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "deny",
    "permissionDecisionReason": "Destructive command blocked by hook"
  }
}
```

| Field | Values | Description |
|-------|--------|-------------|
| `permissionDecision` | `"allow"`, `"deny"` | Allow or block the operation |
| `permissionDecisionReason` | `string` | Reason shown to Claude |
| `updatedInput` | `object` | Modified tool input (PreToolUse only) |
| `systemMessage` | `string` | Context added to conversation |
| `additionalContext` | `string` | Extra context for Claude |

### Hook Types

#### 1. Command Hooks (Shell Scripts)

Execute shell commands with JSON input on stdin:

```bash
#!/bin/bash
# .claude/hooks/block-rm.sh
COMMAND=$(jq -r '.tool_input.command')

if echo "$COMMAND" | grep -q 'rm -rf'; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: "Destructive command blocked"
    }
  }'
else
  exit 0
fi
```

#### 2. HTTP Hooks

Send JSON to an HTTP endpoint:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "http",
            "url": "https://api.example.com/hooks/audit",
            "timeout": 5000
          }
        ]
      }
    ]
  }
}
```

#### 3. Prompt Hooks (LLM Evaluation)

Use an LLM to evaluate the hook condition:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "prompt",
            "prompt": "Evaluate if this bash command is safe: {{tool_input.command}}"
          }
        ]
      }
    ]
  }
}
```

### Async Hooks

Run hooks in the background without blocking Claude:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/run-tests-async.sh",
            "async": true,
            "timeout": 300
          }
        ]
      }
    ]
  }
}
```

Async hooks cannot return decisions (by the time they complete, the action has proceeded).

---

## Agent SDK Hooks

The **Claude Agent SDK** (Python/TypeScript) provides programmatic hooks for building custom agents:

### Available Events

| Event | Description |
|-------|-------------|
| `PreToolUse` | Before a tool is called |
| `PostToolUse` | After a tool completes |
| `SubagentStart` | When a subagent starts |
| `SubagentStop` | When a subagent completes |
| `Notification` | When a notification is sent |
| `Stop` | When execution stops |
| `TeammateIdle` | When a teammate goes idle |

### Python Example

```python
from claude_agent_sdk import (
    ClaudeSDKClient,
    ClaudeAgentOptions,
    HookMatcher,
)

async def protect_env_files(input_data, tool_use_id, context):
    file_path = input_data["tool_input"].get("file_path", "")
    if file_path.endswith(".env"):
        return {
            "hookSpecificOutput": {
                "hookEventName": input_data["hook_event_name"],
                "permissionDecision": "deny",
                "permissionDecisionReason": "Cannot modify .env files",
            }
        }
    return {}

async def main():
    options = ClaudeAgentOptions(
        hooks={
            "PreToolUse": [
                HookMatcher(
                    matcher="Write|Edit",
                    hooks=[protect_env_files]
                )
            ]
        }
    )
    
    async with ClaudeSDKClient(options=options) as client:
        await client.query("Update the database configuration")
        async for message in client.receive_response():
            print(message)
```

### TypeScript Example

```typescript
import { query, HookCallback, PreToolUseHookInput } from "@anthropic-ai/claude-agent-sdk";

const protectEnvFiles: HookCallback = async (input, toolUseID, { signal }) => {
  const preInput = input as PreToolUseHookInput;
  const toolInput = preInput.tool_input as Record<string, unknown>;
  const filePath = toolInput?.file_path as string;

  if (filePath?.endsWith(".env")) {
    return {
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        permissionDecision: "deny",
        permissionDecisionReason: "Cannot modify .env files",
      },
    };
  }
  return {};
};

for await (const message of query({
  prompt: "Update the configuration",
  options: {
    hooks: {
      PreToolUse: [{ matcher: "Write|Edit", hooks: [protectEnvFiles] }]
    }
  }
})) {
  console.log(message);
}
```

---

## Model Context Protocol (MCP)

MCP is an **open standard** for connecting AI assistants to external tools and data sources.

### Architecture

```
┌─────────────────────────────────────────────────┐
│           MCP Host (AI Application)             │
│  ┌─────────────┐  ┌─────────────┐               │
│  │ MCP Client 1│  │ MCP Client 2│  ...          │
│  └──────┬──────┘  └──────┬──────┘               │
└─────────┼────────────────┼──────────────────────┘
          │                │
          ▼                ▼
┌─────────────────┐  ┌─────────────────┐
│  MCP Server A   │  │  MCP Server B   │
│  (Filesystem)   │  │  (GitHub API)   │
└─────────────────┘  └─────────────────┘
```

### Core Components

| Component | Description |
|-----------|-------------|
| **MCP Host** | AI application (Claude Code, Claude Desktop) |
| **MCP Client** | Maintains connection to MCP server |
| **MCP Server** | Provides tools, resources, and prompts |

### MCP Server Capabilities

MCP servers expose three primitives:

1. **Tools** - Callable functions (like `read_file`, `query_database`)
2. **Resources** - Data sources (files, database records)
3. **Prompts** - Reusable prompt templates

### Configuration Example

Claude Code MCP configuration in `~/.claude/mcp_settings.json`:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/dir"]
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "your-token"
      }
    },
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": {
        "DATABASE_URL": "postgresql://..."
      }
    }
  }
}
```

### Tool Definition Schema

```json
{
  "name": "read_file",
  "description": "Read contents of a file",
  "inputSchema": {
    "type": "object",
    "properties": {
      "path": {
        "type": "string",
        "description": "Path to the file"
      }
    },
    "required": ["path"]
  }
}
```

---

## Tool Use (Function Calling)

Claude's **Tool Use** allows you to define functions Claude can call via the API.

### Basic Tool Definition

```python
import anthropic

client = anthropic.Anthropic()

response = client.messages.create(
    model="claude-sonnet-4-20250514",
    max_tokens=1024,
    tools=[
        {
            "name": "get_weather",
            "description": "Get the current weather in a given location",
            "input_schema": {
                "type": "object",
                "properties": {
                    "location": {
                        "type": "string",
                        "description": "City and state, e.g. San Francisco, CA"
                    }
                },
                "required": ["location"]
            }
        }
    ],
    messages=[{"role": "user", "content": "What's the weather in SF?"}]
)
```

### Tool Use Flow

1. Send request with tool definitions
2. Claude returns `tool_use` content block with tool name and arguments
3. Execute the tool in your code
4. Send `tool_result` back to Claude
5. Claude provides final response

### Response Handling

```python
for block in response.content:
    if block.type == "tool_use":
        tool_name = block.name
        tool_input = block.input
        
        # Execute your tool
        result = execute_tool(tool_name, tool_input)
        
        # Send result back
        response = client.messages.create(
            model="claude-sonnet-4-20250514",
            max_tokens=1024,
            messages=[
                {"role": "user", "content": "What's the weather?"},
                {"role": "assistant", "content": response.content},
                {
                    "role": "user",
                    "content": [
                        {
                            "type": "tool_result",
                            "tool_use_id": block.id,
                            "content": result
                        }
                    ]
                }
            ]
        )
```

---

## Practical Examples

### Example 1: Auto-Format Code After Edits

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "prettier --write \"$CLAUDE_TOOL_INPUT_FILE_PATH\""
          }
        ]
      }
    ]
  }
}
```

### Example 2: Block Dangerous Commands

```bash
#!/bin/bash
# .claude/hooks/safe-bash.sh
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command')

# Block dangerous patterns
DANGEROUS_PATTERNS=(
  "rm -rf /"
  "rm -rf ~"
  "chmod 777"
  "> /dev/sda"
  "mkfs"
)

for pattern in "${DANGEROUS_PATTERNS[@]}"; do
  if [[ "$COMMAND" == *"$pattern"* ]]; then
    jq -n '{
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        permissionDecision: "deny",
        permissionDecisionReason: "Blocked dangerous command pattern"
      }
    }'
    exit 0
  fi
done

exit 0  # Allow
```

### Example 3: Audit Log All Tool Calls

```bash
#!/bin/bash
# .claude/hooks/audit.sh
INPUT=$(cat)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
TOOL=$(echo "$INPUT" | jq -r '.tool_name')
ARGS=$(echo "$INPUT" | jq -c '.tool_input')

echo "[$TIMESTAMP] $TOOL: $ARGS" >> ~/.claude/audit.log
exit 0
```

### Example 4: Notify on Completion

```bash
#!/bin/bash
# .claude/hooks/notify-complete.sh

# macOS notification
osascript -e 'display notification "Claude has finished" with title "Claude Code"'

# Or cross-platform with ntfy
# curl -d "Claude Code task complete" ntfy.sh/your-topic

exit 0
```

### Example 5: Require Approval for Sensitive Files

```python
# Agent SDK hook for human-in-the-loop approval
async def require_approval_for_config(input_data, tool_use_id, context):
    file_path = input_data["tool_input"].get("file_path", "")
    
    sensitive_patterns = [".env", "config.json", "secrets", "credentials"]
    
    for pattern in sensitive_patterns:
        if pattern in file_path:
            # In production, you'd integrate with your approval system
            print(f"⚠️ Approval required for: {file_path}")
            approved = await get_human_approval(file_path)
            
            if not approved:
                return {
                    "hookSpecificOutput": {
                        "hookEventName": "PreToolUse",
                        "permissionDecision": "deny",
                        "permissionDecisionReason": "Human approval denied",
                    }
                }
    return {}
```

---

## Official Documentation Links

### Claude Code Hooks
- **Reference**: https://docs.anthropic.com/en/docs/claude-code/hooks
- **Guide**: https://docs.anthropic.com/en/docs/claude-code/automate-workflows
- **Blog**: https://claude.com/blog/how-to-configure-hooks

### Agent SDK
- **Python SDK Hooks**: https://platform.claude.com/docs/en/agent-sdk/hooks
- **TypeScript SDK**: https://www.npmjs.com/package/@anthropic-ai/claude-agent-sdk

### Model Context Protocol
- **Official Site**: https://modelcontextprotocol.io
- **Specification**: https://github.com/modelcontextprotocol/specification
- **Server Examples**: https://github.com/modelcontextprotocol/servers
- **Announcement**: https://www.anthropic.com/news/model-context-protocol

### Tool Use (API)
- **Overview**: https://platform.claude.com/docs/en/agents-and-tools/tool-use/overview
- **Programmatic Tool Calling**: https://platform.claude.com/docs/en/agents-and-tools/tool-use/programmatic-tool-calling
- **Advanced Tool Use**: https://www.anthropic.com/engineering/advanced-tool-use

---

## Comparison: Hookflow vs Claude Code Hooks

| Feature | Hookflow | Claude Code Hooks |
|---------|----------|-------------------|
| **Trigger Source** | Copilot hook events | Claude Code tool events |
| **Config Format** | YAML workflows | JSON settings |
| **Event Types** | `preToolUse`, `postToolUse`, file patterns | 15+ events including session lifecycle |
| **Execution** | PowerShell steps | Shell commands, HTTP, LLM prompts |
| **Async Support** | Via `&` background | Native `async: true` |
| **Flow Control** | Exit codes | Exit codes + JSON decisions |

### Event Mapping

| Hookflow Event | Claude Code Equivalent |
|----------------|------------------------|
| `preToolUse:Edit` | `PreToolUse` with `matcher: "Edit"` |
| `postToolUse:Write` | `PostToolUse` with `matcher: "Write"` |
| `on: files: '*.ts'` | `PreToolUse/PostToolUse` with file path check in script |

---

## Summary

Anthropic's Claude ecosystem provides multiple hook mechanisms:

1. **Claude Code Hooks** - Best for automating CLI workflows, file protection, and integrating external tools
2. **Agent SDK Hooks** - Best for building custom agents with programmatic control
3. **MCP** - Best for connecting Claude to external data sources and tools at scale
4. **Tool Use API** - Best for extending Claude's capabilities in custom applications

Choose based on your use case:
- **Local workflow automation** → Claude Code Hooks
- **Custom agent development** → Agent SDK Hooks
- **External tool integration** → MCP Servers
- **API-based applications** → Tool Use
