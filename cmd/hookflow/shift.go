package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/htekdev/gh-hookflow/internal/ai"
	"github.com/spf13/cobra"
)

var shiftCmd = &cobra.Command{
	Use:   "shift",
	Short: "Shift protections between CI and agent workflows",
	Long: `Analyze and shift protections between GitHub Actions (CI) and hooks.

Subcommands:
  left   - Analyze CI workflows and suggest hooks (shift checks earlier)
  right  - Analyze hooks and generate GitHub Actions (defense in depth)`,
}

var shiftLeftCmd = &cobra.Command{
	Use:   "left",
	Short: "Shift CI checks left to agent workflows",
	Long: `Analyzes your .github/workflows/ (CI) and suggests hooks that can
run the same checks earlier, during the agent editing phase.

This "shifts left" your quality gates - catching issues before code is committed
rather than waiting for CI to run on pull requests.

Examples of shifts:
  CI: ESLint runs on PR        → Agent: ESLint runs on file edit
  CI: Type checking on PR      → Agent: tsc runs on TypeScript edit
  CI: Security scan on PR      → Agent: Block edits to sensitive files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		return runShiftLeft(dir, dryRun)
	},
}

var shiftRightCmd = &cobra.Command{
	Use:   "right",
	Short: "Generate GitHub Actions from agent workflows",
	Long: `Analyzes your .github/hooks/ and generates GitHub Actions that
provide the same protections at PR time.

This creates "defense in depth" - protections exist in both:
  1. Agent layer (real-time during editing)
  2. CI layer (verification on pull request)

Examples of shifts:
  Agent: Block .env edits      → CI: Verify no .env changes in PR
  Agent: Lint on file edit     → CI: Lint check on PR
  Agent: Validate JSON         → CI: JSON schema validation on PR`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		return runShiftRight(dir, dryRun)
	},
}

func init() {
	rootCmd.AddCommand(shiftCmd)
	shiftCmd.AddCommand(shiftLeftCmd)
	shiftCmd.AddCommand(shiftRightCmd)

	// Common flags
	shiftLeftCmd.Flags().StringP("dir", "d", "", "Directory to search (default: current directory)")
	shiftLeftCmd.Flags().Bool("dry-run", false, "Print suggestions without creating files")

	shiftRightCmd.Flags().StringP("dir", "d", "", "Directory to search (default: current directory)")
	shiftRightCmd.Flags().Bool("dry-run", false, "Print suggestions without creating files")
}

func runShiftLeft(dir string, dryRun bool) error {
	fmt.Println("🔄 Analyzing CI workflows to suggest hooks...")
	fmt.Println()

	// Find CI workflow files
	ciDir := filepath.Join(dir, ".github", "workflows")
	if _, err := os.Stat(ciDir); os.IsNotExist(err) {
		return fmt.Errorf("no .github/workflows/ directory found")
	}

	// Read all workflow files
	var workflows []string
	err := filepath.Walk(ciDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yml" || ext == ".yaml" {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil // Skip unreadable files
			}
			workflows = append(workflows, fmt.Sprintf("# File: %s\n%s", filepath.Base(path), string(content)))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to scan workflows: %w", err)
	}

	if len(workflows) == 0 {
		fmt.Println("No CI workflows found in .github/workflows/")
		return nil
	}

	fmt.Printf("Found %d CI workflow(s)\n", len(workflows))

	// Initialize AI client
	client := ai.NewClient()
	ctx := context.Background()

	fmt.Println("Starting Copilot client...")
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start AI client: %w\nMake sure you have GitHub Copilot CLI installed and authenticated", err)
	}
	defer client.Stop()

	// Build prompt
	prompt := buildShiftLeftPrompt(workflows)

	fmt.Println("Analyzing with AI...")

	// Generate suggestions
	result, err := client.GenerateWorkflow(ctx, prompt)
	if err != nil {
		return fmt.Errorf("failed to generate suggestions: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Analysis complete!")
	fmt.Println()
	fmt.Println("Suggested agent-workflow(s):")
	fmt.Println("---")
	fmt.Println(result.YAML)
	fmt.Println("---")
	fmt.Println()

	if dryRun {
		fmt.Println("(dry-run mode - not saving)")
		return nil
	}

	// Save the workflow
	workflowDir := filepath.Join(dir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}

	outputPath := filepath.Join(workflowDir, "shifted-from-ci.yml")
	if err := os.WriteFile(outputPath, []byte(result.YAML), 0644); err != nil {
		return fmt.Errorf("failed to save workflow: %w", err)
	}

	fmt.Printf("✓ Saved to: %s\n", outputPath)
	return nil
}

func runShiftRight(dir string, dryRun bool) error {
	fmt.Println("🔄 Analyzing hooks to generate GitHub Actions...")
	fmt.Println()

	// Find agent workflow files
	agentDir := filepath.Join(dir, ".github", "hooks")
	if _, err := os.Stat(agentDir); os.IsNotExist(err) {
		return fmt.Errorf("no .github/hooks/ directory found")
	}

	// Read all workflow files
	var workflows []string
	err := filepath.Walk(agentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yml" || ext == ".yaml" {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			workflows = append(workflows, fmt.Sprintf("# File: %s\n%s", filepath.Base(path), string(content)))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to scan workflows: %w", err)
	}

	if len(workflows) == 0 {
		fmt.Println("No hooks found in .github/hooks/")
		return nil
	}

	fmt.Printf("Found %d agent-workflow(s)\n", len(workflows))

	// Initialize AI client
	client := ai.NewClient()
	ctx := context.Background()

	fmt.Println("Starting Copilot client...")
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start AI client: %w\nMake sure you have GitHub Copilot CLI installed and authenticated", err)
	}
	defer client.Stop()

	// Build prompt
	prompt := buildShiftRightPrompt(workflows)

	fmt.Println("Generating with AI...")

	// This will generate a GitHub Actions workflow, not an agent workflow
	// So we need to handle the response differently
	result, err := client.GenerateWorkflow(ctx, prompt)
	if err != nil {
		return fmt.Errorf("failed to generate GitHub Action: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Generation complete!")
	fmt.Println()
	fmt.Println("Generated GitHub Action:")
	fmt.Println("---")
	fmt.Println(result.YAML)
	fmt.Println("---")
	fmt.Println()

	if dryRun {
		fmt.Println("(dry-run mode - not saving)")
		return nil
	}

	// Save to .github/workflows/
	ciDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(ciDir, 0755); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}

	outputPath := filepath.Join(ciDir, "agent-protections.yml")
	if err := os.WriteFile(outputPath, []byte(result.YAML), 0644); err != nil {
		return fmt.Errorf("failed to save workflow: %w", err)
	}

	fmt.Printf("✓ Saved to: %s\n", outputPath)
	return nil
}

func buildShiftLeftPrompt(ciWorkflows []string) string {
	return fmt.Sprintf(`Analyze these GitHub Actions CI workflows and suggest hooks that can shift these checks LEFT (earlier in the development process).

The goal is to run checks DURING agent editing rather than waiting for CI on pull requests.

## CI Workflows to Analyze

%s

## What to Look For

1. Linting/formatting checks → agent-workflow on file edit
2. Type checking → agent-workflow on file edit
3. Security scans → agent-workflow to block sensitive file edits
4. Test runs → agent-workflow to run related tests on edit
5. Validation checks → agent-workflow to validate on save

## Output Requirements

Generate ONE consolidated agent-workflow YAML that implements the most valuable shifted checks.
Focus on checks that will catch issues early without being too disruptive.

For file-based checks, use file triggers:
on:
  file:
    paths:
      - '**/*.ts'
    actions:
      - edit

For commit-based checks, use commit triggers:
on:
  commit:
    paths:
      - 'src/**'`, strings.Join(ciWorkflows, "\n\n"))
}

func buildShiftRightPrompt(agentWorkflows []string) string {
	return fmt.Sprintf(`Analyze these hooks and generate a GitHub Actions workflow that provides the same protections at PR time.

This creates "defense in depth" - the same checks run both:
1. During editing (hooks) - catches issues immediately
2. On PR (GitHub Actions) - verifies nothing slipped through

## Agent Workflows to Analyze

%s

## Output Requirements

Generate a GitHub Actions workflow (not an agent-workflow) that:
1. Runs on: pull_request
2. Implements equivalent checks to the hooks
3. Fails the PR if protections would be violated
4. Uses standard GitHub Actions syntax

Example structure:
name: Agent Protections CI
on:
  pull_request:
    branches: [main]

jobs:
  verify-protections:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Check for blocked files
        run: |
          # Script to verify protected files weren't modified`, strings.Join(agentWorkflows, "\n\n"))
}
