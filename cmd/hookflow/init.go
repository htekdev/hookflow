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

	fmt.Println("\n✓ hookflow initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Create a workflow: hookflow create \"block edits to .env files\"")
	fmt.Println("  2. Or edit the example workflow in .github/hooks/example.yml")
	fmt.Println("  3. Commit the .github/hooks/ directory to enable for your team")
	fmt.Println("\nNote: Team members need hookflow installed. They can run:")
	fmt.Println("  npm install -g hookflow-cli")

	return nil
}

// generateHooksJSON creates the hooks.json that integrates with Copilot CLI
// This goes in .github/hooks/hooks.json per Copilot CLI documentation
func generateHooksJSON() string {
	return `{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {
        "type": "command",
        "bash": "command -v hookflow >/dev/null 2>&1 || { echo '{\"permissionDecision\":\"deny\",\"permissionDecisionReason\":\"hookflow required. Install: npm i -g hookflow-cli\"}'; exit 0; }; hookflow run --raw --dir \"$PWD\"",
        "powershell": "if (-not (Get-Command hookflow -ErrorAction SilentlyContinue)) { Write-Output '{\"permissionDecision\":\"deny\",\"permissionDecisionReason\":\"hookflow required. Install: npm i -g hookflow-cli\"}'; exit 0 }; hookflow run --raw --dir (Get-Location)",
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
# Learn more: https://github.com/htekdev/hookflow

name: Example Workflow
description: An example workflow that demonstrates hookflow features

# This workflow is disabled by default - rename or modify to enable
on:
  file:
    paths:
      - '**/.env'
      - '**/.env.*'
    actions:
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
