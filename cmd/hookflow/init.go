package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize hookflow for a repository",
	Long: `Creates the necessary directory structure and hook configuration
for hookflow in the current repository.

This command creates:
- .github/hooks/ directory for your workflow files
- .github/hooks/hooks.json to integrate with Copilot CLI hooks
- .github/skills/hookflow/SKILL.md for AI agent guidance

After running init, you can create workflows using 'hookflow create'
or by manually creating YAML files in .github/hooks/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		force, _ := cmd.Flags().GetBool("force")

		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		return runInit(dir, force)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("dir", "d", "", "Directory to initialize (default: current directory)")
	initCmd.Flags().BoolP("force", "f", false, "Overwrite existing configuration")
}

func runInit(dir string, force bool) error {
	fmt.Printf("Initializing hookflow in %s\n", dir)

	// Create .github/hooks directory
	hooksDir := filepath.Join(dir, ".github", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}
	fmt.Printf("✓ Created %s\n", hooksDir)

	// Create hooks.json in .github/hooks/ (the standard location)
	hooksFile := filepath.Join(hooksDir, "hooks.json")
	if _, err := os.Stat(hooksFile); err == nil && !force {
		fmt.Printf("⚠ %s already exists (use --force to overwrite)\n", hooksFile)
	} else {
		hooksContent := generateHooksJSON()
		if err := os.WriteFile(hooksFile, []byte(hooksContent), 0644); err != nil {
			return fmt.Errorf("failed to create hooks.json: %w", err)
		}
		fmt.Printf("✓ Created %s\n", hooksFile)
	}

	// Create example workflow
	exampleWorkflow := filepath.Join(hooksDir, "example.yml")
	if _, err := os.Stat(exampleWorkflow); os.IsNotExist(err) {
		exampleContent := generateExampleWorkflow()
		if err := os.WriteFile(exampleWorkflow, []byte(exampleContent), 0644); err != nil {
			fmt.Printf("⚠ Could not create example workflow: %v\n", err)
		} else {
			fmt.Printf("✓ Created %s\n", exampleWorkflow)
		}
	}

	// Create skill directory and SKILL.md
	skillDir := filepath.Join(dir, ".github", "skills", "hookflow")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		fmt.Printf("⚠ Could not create skill directory: %v\n", err)
	} else {
		skillFile := filepath.Join(skillDir, "SKILL.md")
		if _, err := os.Stat(skillFile); err == nil && !force {
			fmt.Printf("⚠ %s already exists (use --force to overwrite)\n", skillFile)
		} else {
			skillContent := generateSkillMD()
			if err := os.WriteFile(skillFile, []byte(skillContent), 0644); err != nil {
				fmt.Printf("⚠ Could not create SKILL.md: %v\n", err)
			} else {
				fmt.Printf("✓ Created %s\n", skillFile)
			}
		}
	}

	fmt.Println("\n✓ hookflow initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Create a workflow: hookflow create \"block edits to .env files\"")
	fmt.Println("  2. Or edit the example workflow in .github/hooks/example.yml")
	fmt.Println("  3. Commit the .github/ directory to enable for your team")
	fmt.Println("\nNote: Team members need hookflow installed. They can run:")
	fmt.Println("  gh extension install htekdev/gh-hookflow")

	return nil
}

// generateHooksJSON creates the hooks.json that integrates with Copilot CLI
// This goes in .github/hooks/hooks.json per Copilot CLI documentation
func generateHooksJSON() string {
	// The hook checks for hookflow in PATH first, then falls back to gh hookflow
	return `{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {
        "type": "command",
        "bash": "if command -v hookflow >/dev/null 2>&1; then hookflow run --raw --dir \"$PWD\"; elif command -v gh >/dev/null 2>&1 && gh extension list 2>/dev/null | grep -q hookflow; then gh hookflow run --raw --dir \"$PWD\"; else echo '{\"permissionDecision\":\"deny\",\"permissionDecisionReason\":\"hookflow required. Install: gh extension install htekdev/gh-hookflow\"}'; fi",
        "powershell": "$hf = Get-Command hookflow -ErrorAction SilentlyContinue; if ($hf) { hookflow run --raw --dir (Get-Location) } elseif ((Get-Command gh -ErrorAction SilentlyContinue) -and ((gh extension list 2>$null) -match 'hookflow')) { gh hookflow run --raw --dir (Get-Location) } else { Write-Output '{\"permissionDecision\":\"deny\",\"permissionDecisionReason\":\"hookflow required. Install: gh extension install htekdev/gh-hookflow\"}' }",
        "timeoutSec": 60
      }
    ]
  }
}
`
}

// generateExampleWorkflow creates an example workflow file
func generateExampleWorkflow() string {
	return `# Example hookflow workflow
# Learn more: https://github.com/htekdev/gh-hookflow

name: Example Workflow
description: An example workflow that demonstrates hookflow features

# This workflow is disabled by default - rename or modify to enable
on:
  file:
    paths:
      - '**/.env'
      - '**/.env.*'
    types:
      - edit
      - create

blocking: true

steps:
  - name: Block sensitive file edits
    run: |
      echo "⚠️ Editing environment files requires review"
      echo "File: ${{ event.file.path }}"
      # Uncomment the next line to actually block:
      # exit 1
`
}

// generateSkillMD creates the SKILL.md file for AI agent guidance
func generateSkillMD() string {
	return `---
name: hookflow
description: Create and manage hookflow workflows for agent governance. Use this skill when creating, editing, or troubleshooting workflow files in .github/hooks/. Trigger phrases include "create workflow", "block file edits", "add validation", "hookflow", "agent gate".
---

# Hookflow Workflow Creation

This skill helps you create hookflow workflow files that enforce governance during AI agent sessions.

## When to Use This Skill

- Creating new workflow files in ` + "`" + `.github/hooks/` + "`" + `
- Editing existing hookflow workflows
- Troubleshooting workflow triggers or validation
- Understanding the hookflow schema

## Workflow Schema

### Required Fields

` + "```yaml" + `
name: string          # Human-readable workflow name (required)
on: object            # Trigger configuration (required)
steps: array          # Steps to execute (required)
` + "```" + `

### Optional Fields

` + "```yaml" + `
description: string   # What the workflow does
blocking: boolean     # Block on failure (default: true)
env: object          # Environment variables
concurrency: string   # Concurrency group name
` + "```" + `

## Trigger Types

### File Trigger

Matches file create/edit/delete operations.

` + "```yaml" + `
on:
  file:
    paths:              # File patterns to match (glob supported)
      - '**/*.env'
      - 'secrets/**'
    paths-ignore:       # Patterns to exclude
      - '**/*.md'
    types:              # Event types: create, edit, delete
      - edit
      - create
` + "```" + `

### Tool Trigger

Matches specific tool calls with argument patterns.

` + "```yaml" + `
on:
  tool:
    name: edit          # Tool name: edit, create, powershell, bash, etc.
    args:
      path: '**/secrets/**'  # Glob pattern for argument values
` + "```" + `

### Commit Trigger

Matches git commit events.

` + "```yaml" + `
on:
  commit:
    paths:              # Files that must be in the commit
      - 'src/**'
    paths-ignore:
      - '**/*.md'
    message: 'feat:*'   # Commit message pattern
` + "```" + `

### Push Trigger

Matches git push events.

` + "```yaml" + `
on:
  push:
    branches:
      - main
      - 'release/*'
    tags:
      - 'v*'
` + "```" + `

## Expression Syntax

Use ` + "`${{ }}`" + ` for dynamic values:

| Expression | Description |
|------------|-------------|
| ` + "`${{ event.file.path }}`" + ` | Path of file being edited |
| ` + "`${{ event.file.action }}`" + ` | Action: edit, create, delete |
| ` + "`${{ event.tool.name }}`" + ` | Tool name being called |
| ` + "`${{ event.tool.args.path }}`" + ` | Tool argument value |
| ` + "`${{ event.commit.message }}`" + ` | Commit message |
| ` + "`${{ event.commit.sha }}`" + ` | Commit SHA |
| ` + "`${{ env.MY_VAR }}`" + ` | Environment variable |

### Functions

| Function | Example |
|----------|---------|
| ` + "`contains(str, substr)`" + ` | ` + "`${{ contains(event.file.path, '.env') }}`" + ` |
| ` + "`startsWith(str, prefix)`" + ` | ` + "`${{ startsWith(event.file.path, 'src/') }}`" + ` |
| ` + "`endsWith(str, suffix)`" + ` | ` + "`${{ endsWith(event.file.path, '.ts') }}`" + ` |
| ` + "`format(fmt, ...args)`" + ` | ` + "`${{ format('File: {0}', event.file.path) }}`" + ` |

## Step Configuration

` + "```yaml" + `
steps:
  - name: Step name        # Human-readable name (required)
    if: ${{ condition }}   # Conditional execution (optional)
    run: |                 # Shell command (required)
      echo "Running step"
      # exit 1 to deny/block
    env:                   # Step-specific env vars (optional)
      MY_VAR: value
    shell: bash            # Shell: bash, sh, pwsh (optional)
    timeout: 60            # Timeout in seconds (optional)
` + "```" + `

## Common Patterns

### Block Sensitive Files

` + "```yaml" + `
name: Block Sensitive Files
on:
  file:
    paths:
      - '**/.env*'
      - '**/secrets/**'
      - '**/*.pem'
      - '**/*.key'
    types: [edit, create]
blocking: true
steps:
  - name: Deny sensitive file access
    run: |
      echo "❌ Cannot modify sensitive file: ${{ event.file.path }}"
      exit 1
` + "```" + `

### Validate JSON Files

` + "```yaml" + `
name: Validate JSON
on:
  file:
    paths: ['**/*.json']
    types: [edit, create]
blocking: true
steps:
  - name: Check JSON syntax
    run: |
      cat "${{ event.file.path }}" | jq . > /dev/null
      echo "✓ Valid JSON"
` + "```" + `

### Require Tests for Source Changes

` + "```yaml" + `
name: Require Tests
on:
  commit:
    paths: ['src/**']
    paths-ignore: ['src/**/*.test.*']
blocking: true
steps:
  - name: Check for test files
    run: |
      if ! echo "${{ event.commit.files }}" | grep -q '\.test\.'; then
        echo "❌ Source changes require accompanying tests"
        exit 1
      fi
` + "```" + `

## Troubleshooting

### Workflow Not Triggering

1. Check trigger type matches event (file vs tool vs commit)
2. Verify path patterns use correct glob syntax
3. Ensure ` + "`types`" + ` field matches the action (edit/create/delete)

### Validation Errors

Run ` + "`hookflow validate`" + ` to check workflow syntax:

` + "```bash" + `
hookflow validate --file .github/hooks/my-workflow.yml
` + "```" + `

### Testing Workflows

Use ` + "`hookflow test`" + ` to simulate events:

` + "```bash" + `
hookflow test --workflow my-workflow --event file --path "test.env"
` + "```" + `
`
}
