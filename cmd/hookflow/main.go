package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/htekdev/hookflow/internal/discover"
	"github.com/htekdev/hookflow/internal/event"
	"github.com/htekdev/hookflow/internal/runner"
	"github.com/htekdev/hookflow/internal/schema"
	"github.com/htekdev/hookflow/internal/trigger"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "hookflow",
	Short: "Local workflow engine for agentic DevOps",
	Long: `hookflow is a CLI tool that executes local workflows triggered by
Copilot agent hooks, file changes, commits, and pushes.

Workflows are defined in .github/hooks/*.yml using a GitHub Actions-like syntax.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("hookflow version %s\n", version)
	},
}

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover workflow files in the current directory",
	Long:  `Searches for .github/hooks/*.yml files and lists them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		fmt.Printf("Discovering workflows in: %s\n", dir)

		// Import discover package and call Discover
		workflows, err := discoverWorkflows(dir)
		if err != nil {
			return fmt.Errorf("failed to discover workflows: %w", err)
		}

		if len(workflows) == 0 {
			fmt.Println("No workflows found")
			return nil
		}

		fmt.Printf("Found %d workflow(s):\n", len(workflows))
		for _, wf := range workflows {
			fmt.Printf("  - %s (%s)\n", wf.Name, wf.RelPath)
		}
		return nil
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate workflow files",
	Long:  `Validates workflow YAML files against the schema.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		file, _ := cmd.Flags().GetString("file")

		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		// Validate specific file or directory
		var result *schema.ValidationResult
		if file != "" {
			fmt.Printf("Validating file: %s\n", file)
			result = schema.ValidateWorkflow(file)
		} else {
			fmt.Printf("Validating workflows in: %s\n", dir)
			result = schema.ValidateWorkflowsInDir(dir)
		}

		// Print results
		if result.Valid {
			if file != "" {
				fmt.Printf("✓ File is valid\n")
			} else {
				fmt.Printf("✓ All workflows are valid\n")
			}
			return nil
		}

		// Print errors
		for _, err := range result.Errors {
			fmt.Printf("✗ %s\n", err.File)
			fmt.Printf("  Error: %s\n", err.Message)
			for _, detail := range err.Details {
				fmt.Printf("    - %s\n", detail)
			}
		}

		// Exit with error code
		os.Exit(1)
		return nil
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run workflows for an event",
	Long: `Executes matching workflows based on the provided event payload.

Use --raw to pass raw Copilot hook input (toolName, toolArgs, cwd) and let the CLI
detect the event type automatically. This is the preferred mode for hook scripts.

Use --event to pass a pre-built event JSON (legacy mode).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		eventStr, _ := cmd.Flags().GetString("event")
		workflow, _ := cmd.Flags().GetString("workflow")
		dir, _ := cmd.Flags().GetString("dir")
		raw, _ := cmd.Flags().GetBool("raw")

		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		// If workflow is specified, load and run it
		if workflow != "" {
			return runWorkflow(dir, workflow)
		}

		// If --raw flag is set, use the new event detection
		if raw {
			return runWithRawInput(dir, eventStr)
		}

		// Legacy mode: pre-built event JSON
		return runMatchingWorkflows(dir, eventStr)
	},
}

var triggersCmd = &cobra.Command{
	Use:   "triggers",
	Short: "List available trigger types",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available trigger types:")
		fmt.Println("  hooks    - Agent hook events (preToolUse, postToolUse)")
		fmt.Println("  tool     - Tool-specific triggers with argument filtering")
		fmt.Println("  file     - File create/edit events")
		fmt.Println("  commit   - Git commit events")
		fmt.Println("  push     - Git push events")
	},
}

func init() {
	// Add commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(triggersCmd)

	// discover flags
	discoverCmd.Flags().StringP("dir", "d", "", "Directory to search (default: current directory)")

	// validate flags
	validateCmd.Flags().StringP("dir", "d", "", "Directory to search (default: current directory)")
	validateCmd.Flags().StringP("file", "f", "", "Specific file to validate")

	// run flags
	runCmd.Flags().StringP("event", "e", "", "Event JSON (use '-' for stdin)")
	runCmd.Flags().StringP("workflow", "w", "", "Specific workflow to run")
	runCmd.Flags().StringP("dir", "d", "", "Directory to search (default: current directory)")
	runCmd.Flags().BoolP("raw", "r", false, "Accept raw hook input and auto-detect event type")
}

// runWorkflow loads and executes a specific workflow
func runWorkflow(dir, workflowName string) error {
	// Try to find the workflow file
	path, found := findWorkflowFile(dir, workflowName)
	if !found {
		return fmt.Errorf("workflow '%s' not found", workflowName)
	}

	// Load the workflow
	wf, err := schema.LoadWorkflow(path)
	if err != nil {
		return fmt.Errorf("failed to load workflow: %w", err)
	}

	// Execute the workflow
	ctx := context.Background()
	r := runner.NewRunner(wf, nil, dir)
	result := r.RunWithBlocking(ctx)

	// Output the result as JSON
	return outputWorkflowResult(result)
}

// runWithRawInput handles raw Copilot hook input and auto-detects event type
func runWithRawInput(dir, inputStr string) error {
	// Read from stdin if "-"
	var input []byte
	var err error
	if inputStr == "-" || inputStr == "" {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
	} else {
		input = []byte(inputStr)
	}

	// If empty input, allow by default
	if len(input) == 0 || string(input) == "" {
		result := schema.NewAllowResult()
		return outputWorkflowResult(result)
	}

	// Use the event detector to parse and build the event
	detector := event.NewDetector(nil) // nil = use real git provider
	evt, err := detector.DetectFromRawInput(input)
	if err != nil {
		return fmt.Errorf("failed to detect event: %w", err)
	}

	// Override cwd if dir is specified
	if dir != "" && evt.Cwd == "" {
		evt.Cwd = dir
	}
	if evt.Cwd == "" {
		evt.Cwd = dir
	}

	// Discover and run matching workflows
	return runMatchingWorkflowsWithEvent(dir, evt)
}

// runMatchingWorkflowsWithEvent runs workflows with a pre-built event
func runMatchingWorkflowsWithEvent(dir string, evt *schema.Event) error {
	// Discover workflows
	workflowDir := filepath.Join(dir, ".github", "hooks")
	if _, err := os.Stat(workflowDir); os.IsNotExist(err) {
		// No workflows directory, allow by default
		result := schema.NewAllowResult()
		return outputWorkflowResult(result)
	}

	// Find all workflow files
	var workflowFiles []string
	err := filepath.Walk(workflowDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yml" || ext == ".yaml" {
			workflowFiles = append(workflowFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to scan workflows: %w", err)
	}

	if len(workflowFiles) == 0 {
		// No workflows found, allow by default
		result := schema.NewAllowResult()
		return outputWorkflowResult(result)
	}

	// Load and match workflows
	var matchingWorkflows []*schema.Workflow
	for _, path := range workflowFiles {
		wf, err := schema.LoadWorkflow(path)
		if err != nil {
			// Skip invalid workflows
			continue
		}

		// Check if workflow matches the event
		matcher := trigger.NewMatcher(wf)
		if matcher.Match(evt) {
			matchingWorkflows = append(matchingWorkflows, wf)
		}
	}

	if len(matchingWorkflows) == 0 {
		// No matching workflows, allow by default
		result := schema.NewAllowResult()
		return outputWorkflowResult(result)
	}

	// Run matching workflows
	ctx := context.Background()
	var finalResult *schema.WorkflowResult

	for _, wf := range matchingWorkflows {
		r := runner.NewRunner(wf, evt, dir)
		result := r.RunWithBlocking(ctx)

		// If any workflow denies, the final result is deny
		if result.PermissionDecision == "deny" {
			return outputWorkflowResult(result)
		}

		// Keep the last allow result
		finalResult = result
	}

	if finalResult == nil {
		finalResult = schema.NewAllowResult()
	}

	return outputWorkflowResult(finalResult)
}

// runMatchingWorkflows discovers and runs all matching workflows
func runMatchingWorkflows(dir, eventStr string) error {
	// Parse the event
	var eventData map[string]interface{}
	
	// Handle stdin input
	if eventStr == "-" {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		eventStr = string(input)
	}
	
	if eventStr == "" {
		// No event provided, allow by default
		result := schema.NewAllowResult()
		return outputWorkflowResult(result)
	}
	
	if err := json.Unmarshal([]byte(eventStr), &eventData); err != nil {
		return fmt.Errorf("failed to parse event JSON: %w", err)
	}
	
	// Convert to Event struct
	event := parseEventData(eventData)
	
	// Discover workflows
	workflowDir := filepath.Join(dir, ".github", "hooks")
	if _, err := os.Stat(workflowDir); os.IsNotExist(err) {
		// No workflows directory, allow by default
		result := schema.NewAllowResult()
		return outputWorkflowResult(result)
	}
	
	// Find all workflow files
	var workflowFiles []string
	err := filepath.Walk(workflowDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yml" || ext == ".yaml" {
			workflowFiles = append(workflowFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to scan workflows: %w", err)
	}
	
	if len(workflowFiles) == 0 {
		// No workflows found, allow by default
		result := schema.NewAllowResult()
		return outputWorkflowResult(result)
	}
	
	// Load and match workflows
	var matchingWorkflows []*schema.Workflow
	for _, path := range workflowFiles {
		wf, err := schema.LoadWorkflow(path)
		if err != nil {
			// Skip invalid workflows
			continue
		}
		
		// Check if workflow matches the event
		matcher := trigger.NewMatcher(wf)
		if matcher.Match(event) {
			matchingWorkflows = append(matchingWorkflows, wf)
		}
	}
	
	if len(matchingWorkflows) == 0 {
		// No matching workflows, allow by default
		result := schema.NewAllowResult()
		return outputWorkflowResult(result)
	}
	
	// Run matching workflows
	ctx := context.Background()
	var finalResult *schema.WorkflowResult
	
	for _, wf := range matchingWorkflows {
		r := runner.NewRunner(wf, event, dir)
		result := r.RunWithBlocking(ctx)
		
		// If any workflow denies, the final result is deny
		if result.PermissionDecision == "deny" {
			return outputWorkflowResult(result)
		}
		
		// Keep the last allow result
		finalResult = result
	}
	
	if finalResult == nil {
		finalResult = schema.NewAllowResult()
	}
	
	return outputWorkflowResult(finalResult)
}

// parseEventData converts raw event data to a schema.Event
func parseEventData(data map[string]interface{}) *schema.Event {
	event := &schema.Event{}
	
	// Parse hook event
	if hookData, ok := data["hook"].(map[string]interface{}); ok {
		event.Hook = &schema.HookEvent{}
		if t, ok := hookData["type"].(string); ok {
			event.Hook.Type = t
		}
		if cwd, ok := hookData["cwd"].(string); ok {
			event.Hook.Cwd = cwd
		}
		if toolData, ok := hookData["tool"].(map[string]interface{}); ok {
			event.Hook.Tool = &schema.ToolEvent{}
			if name, ok := toolData["name"].(string); ok {
				event.Hook.Tool.Name = name
			}
			if args, ok := toolData["args"].(map[string]interface{}); ok {
				event.Hook.Tool.Args = args
			}
		}
	}
	
	// Parse tool event
	if toolData, ok := data["tool"].(map[string]interface{}); ok {
		event.Tool = &schema.ToolEvent{}
		if name, ok := toolData["name"].(string); ok {
			event.Tool.Name = name
		}
		if args, ok := toolData["args"].(map[string]interface{}); ok {
			event.Tool.Args = args
		}
		if hookType, ok := toolData["hook_type"].(string); ok {
			event.Tool.HookType = hookType
		}
	}
	
	// Parse file event
	if fileData, ok := data["file"].(map[string]interface{}); ok {
		event.File = &schema.FileEvent{}
		if p, ok := fileData["path"].(string); ok {
			event.File.Path = p
		}
		if a, ok := fileData["action"].(string); ok {
			event.File.Action = a
		}
		if c, ok := fileData["content"].(string); ok {
			event.File.Content = c
		}
	}
	
	// Parse commit event
	if commitData, ok := data["commit"].(map[string]interface{}); ok {
		event.Commit = &schema.CommitEvent{}
		if sha, ok := commitData["sha"].(string); ok {
			event.Commit.SHA = sha
		}
		if msg, ok := commitData["message"].(string); ok {
			event.Commit.Message = msg
		}
		if author, ok := commitData["author"].(string); ok {
			event.Commit.Author = author
		}
		if files, ok := commitData["files"].([]interface{}); ok {
			for _, f := range files {
				if fm, ok := f.(map[string]interface{}); ok {
					fs := schema.FileStatus{}
					if p, ok := fm["path"].(string); ok {
						fs.Path = p
					}
					if s, ok := fm["status"].(string); ok {
						fs.Status = s
					}
					event.Commit.Files = append(event.Commit.Files, fs)
				}
			}
		}
	}
	
	// Parse push event
	if pushData, ok := data["push"].(map[string]interface{}); ok {
		event.Push = &schema.PushEvent{}
		if ref, ok := pushData["ref"].(string); ok {
			event.Push.Ref = ref
		}
		if before, ok := pushData["before"].(string); ok {
			event.Push.Before = before
		}
		if after, ok := pushData["after"].(string); ok {
			event.Push.After = after
		}
	}
	
	// Parse top-level cwd and timestamp
	if cwd, ok := data["cwd"].(string); ok {
		event.Cwd = cwd
	}
	if ts, ok := data["timestamp"].(string); ok {
		event.Timestamp = ts
	}
	
	return event
}

// discoverWorkflows finds all workflow files in a directory
func discoverWorkflows(dir string) ([]discover.WorkflowFile, error) {
	return discover.Discover(dir)
}

// findWorkflowFile finds a workflow file by name
func findWorkflowFile(dir, workflowName string) (string, bool) {
	for _, ext := range []string{".yml", ".yaml"} {
		path := fmt.Sprintf("%s/.github/hooks/%s%s", dir, workflowName, ext)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

// outputWorkflowResult outputs the workflow result as JSON
func outputWorkflowResult(result *schema.WorkflowResult) error {
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}

// Git command detection helpers
//
// These patterns are designed to match git commands at the start of a command line
// or after command separators (&&, ||, ;), but NOT inside quoted strings like echo "git commit"

var gitCommitPattern = regexp.MustCompile(`(?:^|&&|\|\||;|&)\s*git\s+(commit|ci)\b`)
var gitPushPattern = regexp.MustCompile(`(?:^|&&|\|\||;|&)\s*git\s+push\b`)
var commitMessagePattern = regexp.MustCompile(`-m\s+["']([^"']+)["']`)
var tagPushPattern = regexp.MustCompile(`\bgit\s+push\s+\S+\s+(v[\d.]+)`)

// isGitCommitCommand checks if a shell command contains a git commit
// It avoids false positives from git commands inside echo strings
func isGitCommitCommand(command string) bool {
	return gitCommitPattern.MatchString(command)
}

// isGitPushCommand checks if a shell command contains a git push
// It avoids false positives from git commands inside echo strings
func isGitPushCommand(command string) bool {
	return gitPushPattern.MatchString(command)
}

// extractCommitMessage extracts the commit message from a git commit command
func extractCommitMessage(command string) string {
	matches := commitMessagePattern.FindStringSubmatch(command)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// extractPushRef determines the ref being pushed
func extractPushRef(command string, currentBranch string) string {
	// Check if pushing a tag
	matches := tagPushPattern.FindStringSubmatch(command)
	if len(matches) >= 2 {
		return "refs/tags/" + matches[1]
	}
	
	// Default to current branch
	return "refs/heads/" + currentBranch
}
