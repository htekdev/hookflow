package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/htekdev/gh-hookflow/internal/ai"
	"github.com/htekdev/gh-hookflow/internal/schema"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [prompt]",
	Short: "Create a workflow using AI",
	Long: `Uses GitHub Copilot to generate a workflow from a natural language description.

Examples:
  hookflow create "block edits to .env files"
  hookflow create "run eslint on typescript file edits"
  hookflow create "validate JSON files before commit"

The generated workflow will be saved to .github/hooks/ and validated
before saving.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		output, _ := cmd.Flags().GetString("output")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		prompt := strings.Join(args, " ")
		return runCreate(dir, prompt, output, dryRun)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringP("dir", "d", "", "Directory to create workflow in (default: current directory)")
	createCmd.Flags().StringP("output", "o", "", "Output file name (auto-generated if not specified)")
	createCmd.Flags().Bool("dry-run", false, "Print generated workflow without saving")
}

func runCreate(dir, prompt, outputName string, dryRun bool) error {
	fmt.Printf("🤖 Generating workflow for: %s\n\n", prompt)

	// Initialize AI client
	client := ai.NewClient()
	ctx := context.Background()

	fmt.Println("Starting Copilot client...")
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start AI client: %w\nMake sure you have GitHub Copilot CLI installed and authenticated", err)
	}
	defer client.Stop()

	fmt.Println("Generating workflow...")

	// Generate the workflow
	result, err := client.GenerateWorkflow(ctx, prompt)
	if err != nil {
		return fmt.Errorf("failed to generate workflow: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Workflow generated successfully!")
	fmt.Println()
	fmt.Println("---")
	fmt.Println(result.YAML)
	fmt.Println("---")
	fmt.Println()

	if dryRun {
		fmt.Println("(dry-run mode - not saving)")
		return nil
	}

	// Determine output file name
	if outputName == "" {
		outputName = generateFileName(result.Name)
	}

	// Ensure .github/hooks directory exists
	workflowDir := filepath.Join(dir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}

	// Build output path
	if !strings.HasSuffix(outputName, ".yml") && !strings.HasSuffix(outputName, ".yaml") {
		outputName += ".yml"
	}
	outputPath := filepath.Join(workflowDir, outputName)

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		return fmt.Errorf("file already exists: %s\nUse --output to specify a different name", outputPath)
	}

	// Validate the generated YAML before saving
	fmt.Println("Validating workflow...")
	tempFile := filepath.Join(os.TempDir(), "hookflow-validate.yml")
	if err := os.WriteFile(tempFile, []byte(result.YAML), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	defer os.Remove(tempFile)

	validation := schema.ValidateWorkflow(tempFile)
	if !validation.Valid {
		fmt.Println("⚠ Generated workflow has validation issues:")
		for _, verr := range validation.Errors {
			fmt.Printf("  - %s\n", verr.Message)
		}
		fmt.Println("\nSaving anyway - you may need to fix these issues manually.")
	} else {
		fmt.Println("✓ Workflow is valid")
	}

	// Save the workflow
	if err := os.WriteFile(outputPath, []byte(result.YAML), 0644); err != nil {
		return fmt.Errorf("failed to save workflow: %w", err)
	}

	fmt.Printf("\n✓ Saved to: %s\n", outputPath)
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. Review the workflow: cat %s\n", outputPath)
	fmt.Printf("  2. Test it: hookflow test --event file --workflow %s\n", outputName)
	fmt.Println("  3. Commit the workflow to your repository")

	return nil
}

// generateFileName creates a kebab-case filename from the workflow name
func generateFileName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces and underscores with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Remove special characters
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	name = reg.ReplaceAllString(name, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	name = reg.ReplaceAllString(name, "-")

	// Trim leading/trailing hyphens
	name = strings.Trim(name, "-")

	if name == "" {
		name = "generated-workflow"
	}

	return name
}
