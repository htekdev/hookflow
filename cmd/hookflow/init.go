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
	// Direct call to gh hookflow - requires gh extension to be installed
	return `{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {
        "type": "command",
        "bash": "gh hookflow run --raw --event-type preToolUse --dir \"$PWD\"",
        "powershell": "gh hookflow run --raw --event-type preToolUse --dir (Get-Location)",
        "timeoutSec": 60
      }
    ],
    "postToolUse": [
      {
        "type": "command",
        "bash": "gh hookflow run --raw --event-type postToolUse --dir \"$PWD\"",
        "powershell": "gh hookflow run --raw --event-type postToolUse --dir (Get-Location)",
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

### Lifecycle (pre vs post)

All triggers support a ` + "`lifecycle`" + ` field to control when workflows run:

- **pre** (default): Runs BEFORE the action - can block/deny the operation
- **post**: Runs AFTER the action - for validation, linting, notifications

` + "```yaml" + `
# Block before file is created (pre)
on:
  file:
    lifecycle: pre
    paths: ['**/*.env']
    types: [create]

# Lint after file is edited (post)
on:
  file:
    lifecycle: post
    paths: ['**/*.ts']
    types: [edit]
` + "```" + `

### File Trigger

Matches file create/edit/delete operations.

` + "```yaml" + `
on:
  file:
    lifecycle: pre        # pre (default) or post
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
    lifecycle: pre        # pre (default) or post
    paths:              # Files that must be in the commit
      - 'src/**'
    paths-ignore:
      - '**/*.md'
` + "```" + `

### Push Trigger

Matches git push events.

` + "```yaml" + `
on:
  push:
    lifecycle: pre        # pre (default) or post
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
| ` + "`${{ event.file.content }}`" + ` | File content (for create) |
| ` + "`${{ event.tool.name }}`" + ` | Tool name being called |
| ` + "`${{ event.tool.args.path }}`" + ` | Tool argument value |
| ` + "`${{ event.tool.args.new_str }}`" + ` | New content (for edit, pre only) |
| ` + "`${{ event.commit.message }}`" + ` | Commit message |
| ` + "`${{ event.commit.sha }}`" + ` | Commit SHA |
| ` + "`${{ event.lifecycle }}`" + ` | Hook lifecycle: pre or post |
| ` + "`${{ env.MY_VAR }}`" + ` | Environment variable |

**Note:** ` + "`event.tool.args.new_str`" + ` is only available during **pre** lifecycle for edit operations. 
For **post** lifecycle, use shell commands to read the actual file from disk.

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

### Post-Edit Linting (TypeScript)

` + "```yaml" + `
name: Post-Edit TypeScript Lint
on:
  file:
    lifecycle: post        # Run AFTER the edit
    paths: ['**/*.ts', '**/*.tsx']
    types: [edit]
blocking: false            # Non-blocking - just report
steps:
  - name: Run ESLint
    run: |
      npx eslint "${{ event.file.path }}" --fix
      echo "✓ Linting complete"
` + "```" + `

### Block Password Strings (Pre-Edit)

` + "```yaml" + `
name: Block Hardcoded Passwords
on:
  file:
    lifecycle: pre
    paths: ['**/*.js', '**/*.ts', '**/*.py']
    types: [edit, create]
blocking: true
steps:
  - name: Check for passwords
    if: contains(event.tool.args.new_str, 'password')
    run: |
      echo "❌ Hardcoded password detected in edit"
      exit 1
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
4. Check ` + "`lifecycle`" + ` matches hook type (pre = preToolUse, post = postToolUse)

### Pre vs Post Confusion

- **pre** workflows run in ` + "`preToolUse`" + ` hook - can block actions
- **post** workflows run in ` + "`postToolUse`" + ` hook - run after action completes
- Default is ` + "`pre`" + ` if not specified

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
