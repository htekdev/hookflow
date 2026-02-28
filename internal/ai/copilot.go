package ai

import (
	"context"
	"fmt"
	"strings"
	"sync"

	copilot "github.com/github/copilot-sdk/go"
)

// Client wraps the Copilot SDK for workflow generation
type Client struct {
	client  *copilot.Client
	started bool
	mu      sync.Mutex
}

// NewClient creates a new AI client
func NewClient() *Client {
	return &Client{}
}

// Start initializes the Copilot client
func (c *Client) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return nil
	}

	c.client = copilot.NewClient(&copilot.ClientOptions{
		LogLevel: "error",
	})

	if err := c.client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start Copilot client: %w", err)
	}

	c.started = true
	return nil
}

// Stop shuts down the Copilot client
func (c *Client) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started || c.client == nil {
		return nil
	}

	c.started = false
	return c.client.Stop()
}

// GenerateWorkflowResult contains the generated workflow
type GenerateWorkflowResult struct {
	Name        string
	Description string
	YAML        string
}

// GenerateWorkflow creates a workflow from a natural language prompt
func (c *Client) GenerateWorkflow(ctx context.Context, prompt string) (*GenerateWorkflowResult, error) {
	if !c.started {
		return nil, fmt.Errorf("client not started")
	}

	// Create a session with permission handler (required by SDK)
	session, err := c.client.CreateSession(ctx, &copilot.SessionConfig{
		Model:               "gpt-4o",
		OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer func() { _ = session.Destroy() }()

	// Build the prompt with context
	fullPrompt := buildWorkflowPrompt(prompt)

	// Collect response
	var response strings.Builder
	done := make(chan bool)
	var responseErr error

	session.On(func(event copilot.SessionEvent) {
		switch event.Type {
		case "assistant.message":
			if event.Data.Content != nil {
				response.WriteString(*event.Data.Content)
			}
		case "session.idle":
			close(done)
		case "error":
			if event.Data.Error != nil {
				responseErr = fmt.Errorf("session error: %v", *event.Data.Error)
			}
			close(done)
		}
	})

	// Send the prompt
	_, err = session.Send(ctx, copilot.MessageOptions{
		Prompt: fullPrompt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send prompt: %w", err)
	}

	// Wait for completion
	<-done

	if responseErr != nil {
		return nil, responseErr
	}

	// Parse the response to extract YAML
	yaml := extractYAML(response.String())
	if yaml == "" {
		return nil, fmt.Errorf("no valid YAML found in response")
	}

	// Extract name from YAML
	name := extractWorkflowName(yaml)

	return &GenerateWorkflowResult{
		Name:        name,
		Description: prompt,
		YAML:        yaml,
	}, nil
}

// buildWorkflowPrompt creates the full prompt with schema context
func buildWorkflowPrompt(userPrompt string) string {
	return fmt.Sprintf(`You are an expert at creating hookflow workflow files. These workflows run locally during AI agent editing sessions to enforce governance and quality gates.

Generate a workflow YAML file for the following requirement:
%s

## Workflow Schema

The workflow must follow this structure:

- name: (required) Human-readable workflow name
- description: (optional) What the workflow does
- on: (required) Trigger configuration - can be:
  - hooks: Match hook type (preToolUse, postToolUse)
  - tool: Match specific tool with args patterns
  - file: Match file events with paths and types
  - commit: Match git commit events with paths/message patterns
  - push: Match git push events with branches/tags
- blocking: (optional, default true) Whether to block on failure
- env: (optional) Environment variables
- steps: (required) List of steps to execute
  - name: Step name
  - if: Conditional expression
  - run: Shell command to execute
  - env: Step-specific environment variables

## Trigger Examples

File trigger (use 'types' for actions like edit/create/delete):
on:
  file:
    paths:
      - '**/*.env*'
    types:
      - edit
      - create

Commit trigger:
on:
  commit:
    paths:
      - 'src/**'

Tool trigger:
on:
  tool:
    name: edit
    args:
      path: '**/secrets/**'

## Expression Syntax

Use ${{ }} for expressions:
- ${{ event.file.path }} - File path being edited
- ${{ event.tool.args.path }} - Tool argument (for tool triggers)
- ${{ event.commit.message }} - Commit message
- ${{ contains(event.file.path, '.env') }} - Check if path contains .env
- ${{ endsWith(event.file.path, '.ts') }} - Check file extension

## Output Requirements

1. Output ONLY the YAML workflow file content
2. Start with --- (YAML document separator)
3. Include descriptive name and description
4. Use appropriate triggers for the requirement
5. For file triggers, use 'types' field (not 'actions')
6. Include clear step names
7. Add exit 1 to block/deny the action when needed

Generate the workflow now:`, userPrompt)
}

// extractYAML finds YAML content in the response
func extractYAML(response string) string {
	// Look for YAML code blocks
	if idx := strings.Index(response, "```yaml"); idx != -1 {
		start := idx + 7
		end := strings.Index(response[start:], "```")
		if end != -1 {
			return strings.TrimSpace(response[start : start+end])
		}
	}

	// Look for generic code blocks
	if idx := strings.Index(response, "```"); idx != -1 {
		start := idx + 3
		// Skip language identifier if present
		if newline := strings.Index(response[start:], "\n"); newline != -1 {
			start += newline + 1
		}
		end := strings.Index(response[start:], "```")
		if end != -1 {
			return strings.TrimSpace(response[start : start+end])
		}
	}

	// Look for YAML document separator
	if idx := strings.Index(response, "---"); idx != -1 {
		return strings.TrimSpace(response[idx:])
	}

	// Look for name: as YAML start
	if idx := strings.Index(response, "name:"); idx != -1 {
		return strings.TrimSpace(response[idx:])
	}

	return ""
}

// extractWorkflowName extracts the name from YAML
func extractWorkflowName(yaml string) string {
	lines := strings.Split(yaml, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			name := strings.TrimPrefix(line, "name:")
			name = strings.TrimSpace(name)
			name = strings.Trim(name, "\"'")
			return name
		}
	}
	return "generated-workflow"
}
