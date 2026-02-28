package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/htekdev/gh-hookflow/internal/schema"
)

// TestVersionCommand tests the version command execution
func TestVersionCommand(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute command
	versionCmd.Run(versionCmd, []string{})

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "hookflow version") {
		t.Errorf("Expected version output, got: %s", output)
	}
}

// TestTriggersCommand tests the triggers command execution
func TestTriggersCommand(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	triggersCmd.Run(triggersCmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	expectedTriggers := []string{"hooks", "tool", "file", "commit", "push"}
	for _, trigger := range expectedTriggers {
		if !strings.Contains(output, trigger) {
			t.Errorf("Expected triggers output to contain '%s', got: %s", trigger, output)
		}
	}
}

// TestDiscoverCommand tests the discover command
func TestDiscoverCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-discover-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test with explicit dir flag
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = discoverCmd.Flags().Set("dir", tmpDir)
	err = discoverCmd.RunE(discoverCmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("discoverCmd.RunE returned error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Discovering workflows") {
		t.Errorf("Expected discovering output, got: %s", output)
	}
}

// TestDiscoverCommandDefaultDir tests discover with default directory
func TestDiscoverCommandDefaultDir(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Reset flag to empty for default dir behavior
	_ = discoverCmd.Flags().Set("dir", "")
	err := discoverCmd.RunE(discoverCmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("discoverCmd.RunE returned error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Discovering workflows") {
		t.Errorf("Expected discovering output, got: %s", output)
	}
}

// TestValidateCommand tests validation with a file
func TestValidateCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-validate-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a valid workflow file
	workflowContent := `name: test-workflow
on:
  tool:
    name: edit
steps:
  - name: Test step
    run: echo "test"
`
	workflowFile := filepath.Join(workflowDir, "test.yml")
	if err := os.WriteFile(workflowFile, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = validateCmd.Flags().Set("file", workflowFile)
	_ = validateCmd.Flags().Set("dir", tmpDir)
	err = validateCmd.RunE(validateCmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("validateCmd.RunE returned error: %v, output: %s", err, output)
	}
}

// TestValidateCommandDir tests validation of a directory
func TestValidateCommandDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-validate-dir-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a valid workflow file
	workflowContent := `name: test-workflow
on:
  tool:
    name: edit
steps:
  - name: Test step
    run: echo "test"
`
	if err := os.WriteFile(filepath.Join(workflowDir, "test.yml"), []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = validateCmd.Flags().Set("file", "")
	_ = validateCmd.Flags().Set("dir", tmpDir)
	err = validateCmd.RunE(validateCmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	_ = buf.String()

	if err != nil {
		t.Errorf("validateCmd.RunE for directory returned error: %v", err)
	}
}

// TestRunCommandEmptyEvent tests run command with empty event
func TestRunCommandEmptyEvent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-run-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = runCmd.Flags().Set("event", "")
	_ = runCmd.Flags().Set("workflow", "")
	_ = runCmd.Flags().Set("dir", tmpDir)
	err = runCmd.RunE(runCmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runCmd.RunE returned error: %v", err)
	}

	// Should output allow result
	if !strings.Contains(output, "allow") {
		t.Errorf("Expected allow result, got: %s", output)
	}
}

// TestRunCommandWithEvent tests run command with event JSON
func TestRunCommandWithEvent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-run-event-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	eventJSON := `{"tool":{"name":"edit","args":{"path":"test.go"}}}`

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = runCmd.Flags().Set("event", eventJSON)
	_ = runCmd.Flags().Set("workflow", "")
	_ = runCmd.Flags().Set("dir", tmpDir)
	err = runCmd.RunE(runCmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runCmd.RunE returned error: %v", err)
	}

	if !strings.Contains(output, "allow") {
		t.Errorf("Expected allow result, got: %s", output)
	}
}

// TestRunCommandInvalidJSON tests run command with invalid JSON
func TestRunCommandInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-run-invalid-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	_ = runCmd.Flags().Set("event", "not valid json")
	_ = runCmd.Flags().Set("workflow", "")
	_ = runCmd.Flags().Set("dir", tmpDir)
	err = runCmd.RunE(runCmd, []string{})

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse event JSON") {
		t.Errorf("Expected JSON parse error, got: %v", err)
	}
}

// TestRunCommandNonexistentWorkflow tests run with nonexistent workflow
func TestRunCommandNonexistentWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-run-noworkflow-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	_ = runCmd.Flags().Set("event", "")
	_ = runCmd.Flags().Set("workflow", "nonexistent")
	_ = runCmd.Flags().Set("dir", tmpDir)
	err = runCmd.RunE(runCmd, []string{})

	if err == nil {
		t.Error("Expected error for nonexistent workflow")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// TestOutputWorkflowResult tests JSON output
func TestOutputWorkflowResult(t *testing.T) {
	result := &schema.WorkflowResult{
		PermissionDecision:       "allow",
		PermissionDecisionReason: "test reason",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputWorkflowResult(result)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("outputWorkflowResult returned error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify it's valid JSON
	var parsed schema.WorkflowResult
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	if parsed.PermissionDecision != "allow" {
		t.Errorf("Expected allow, got: %s", parsed.PermissionDecision)
	}
}

// TestFindWorkflowFileYAML tests finding .yaml extension
func TestFindWorkflowFileYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-find-yaml-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a .yaml file (not .yml)
	workflowContent := `name: test
on:
  tool:
    name: edit
steps:
  - run: echo test
`
	if err := os.WriteFile(filepath.Join(workflowDir, "myworkflow.yaml"), []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}

	path, found := findWorkflowFile(tmpDir, "myworkflow")
	if !found {
		t.Error("Expected to find myworkflow.yaml")
	}
	if !strings.Contains(path, "myworkflow.yaml") {
		t.Errorf("Expected path to contain myworkflow.yaml, got: %s", path)
	}
}

// TestRunWorkflowFound tests running a specific workflow
func TestRunWorkflowFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-run-workflow-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	workflowContent := `name: test-workflow
on:
  tool:
    name: edit
steps:
  - name: Test step
    run: echo "test"
`
	if err := os.WriteFile(filepath.Join(workflowDir, "test.yml"), []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runWorkflow(tmpDir, "test")

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runWorkflow returned error: %v", err)
	}

	if !strings.Contains(output, "permissionDecision") {
		t.Errorf("Expected permissionDecision in output, got: %s", output)
	}
}

// TestRunMatchingWorkflowsWithMatchingWorkflow tests workflow matching
func TestRunMatchingWorkflowsWithMatchingWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-matching-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workflow that matches edit tool
	workflowContent := `name: edit-checker
on:
  tool:
    name: edit
steps:
  - name: Check
    run: echo "checking edit"
`
	if err := os.WriteFile(filepath.Join(workflowDir, "edit-check.yml"), []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Event that matches the workflow
	eventJSON := `{"tool":{"name":"edit","args":{"path":"test.go"}}}`

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runMatchingWorkflows(tmpDir, eventJSON)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runMatchingWorkflows returned error: %v", err)
	}

	if !strings.Contains(output, "permissionDecision") {
		t.Errorf("Expected permissionDecision in output, got: %s", output)
	}
}

// TestRunMatchingWorkflowsNoMatch tests when no workflows match
func TestRunMatchingWorkflowsNoMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-nomatch-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workflow that only matches create tool
	workflowContent := `name: create-checker
on:
  tool:
    name: create
steps:
  - name: Check
    run: echo "checking create"
`
	if err := os.WriteFile(filepath.Join(workflowDir, "create-check.yml"), []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Event for edit tool (won't match)
	eventJSON := `{"tool":{"name":"edit","args":{"path":"test.go"}}}`

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runMatchingWorkflows(tmpDir, eventJSON)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runMatchingWorkflows returned error: %v", err)
	}

	// Should default to allow when no match
	if !strings.Contains(output, "allow") {
		t.Errorf("Expected allow result when no match, got: %s", output)
	}
}

// TestRunMatchingWorkflowsEmptyDir tests when workflow dir has no workflows
func TestRunMatchingWorkflowsEmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-empty-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	eventJSON := `{"tool":{"name":"edit","args":{"path":"test.go"}}}`

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runMatchingWorkflows(tmpDir, eventJSON)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("runMatchingWorkflows returned error: %v", err)
	}

	if !strings.Contains(output, "allow") {
		t.Errorf("Expected allow result, got: %s", output)
	}
}

// TestParseEventDataPartialHookData tests parsing with partial hook data
func TestParseEventDataPartialHookData(t *testing.T) {
	// Hook with no tool
	data := map[string]interface{}{
		"hook": map[string]interface{}{
			"type": "postToolUse",
		},
	}

	event := parseEventData(data)

	if event.Hook == nil {
		t.Fatal("Expected Hook to be set")
	}
	if event.Hook.Type != "postToolUse" {
		t.Errorf("Expected type postToolUse, got: %s", event.Hook.Type)
	}
	if event.Hook.Tool != nil {
		t.Error("Expected Tool to be nil when not provided")
	}
}

// TestParseEventDataPartialCommit tests parsing commit with partial data
func TestParseEventDataPartialCommit(t *testing.T) {
	// Commit with only sha
	data := map[string]interface{}{
		"commit": map[string]interface{}{
			"sha": "abc123",
		},
	}

	event := parseEventData(data)

	if event.Commit == nil {
		t.Fatal("Expected Commit to be set")
	}
	if event.Commit.SHA != "abc123" {
		t.Errorf("Expected sha abc123, got: %s", event.Commit.SHA)
	}
	if event.Commit.Message != "" {
		t.Errorf("Expected empty message, got: %s", event.Commit.Message)
	}
}

// TestParseEventDataInvalidFileInCommit tests commit with malformed files
func TestParseEventDataInvalidFileInCommit(t *testing.T) {
	data := map[string]interface{}{
		"commit": map[string]interface{}{
			"sha": "abc123",
			"files": []interface{}{
				"not a map", // Invalid - should be skipped
				map[string]interface{}{
					"path":   "valid.go",
					"status": "added",
				},
			},
		},
	}

	event := parseEventData(data)

	if event.Commit == nil {
		t.Fatal("Expected Commit to be set")
	}
	// Should only have 1 valid file
	if len(event.Commit.Files) != 1 {
		t.Errorf("Expected 1 file (invalid skipped), got: %d", len(event.Commit.Files))
	}
	if event.Commit.Files[0].Path != "valid.go" {
		t.Errorf("Expected path valid.go, got: %s", event.Commit.Files[0].Path)
	}
}

// TestRootCmdInit tests that root command is initialized properly
func TestRootCmdInit(t *testing.T) {
	// Verify commands are registered
	commands := rootCmd.Commands()
	expectedCmds := []string{"version", "discover", "validate", "run", "triggers"}

	for _, expected := range expectedCmds {
		found := false
		for _, cmd := range commands {
			if cmd.Use == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command '%s' not found in root commands", expected)
		}
	}
}

// TestValidateCommandDefaultDir tests validate with default directory
func TestValidateCommandDefaultDir(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = validateCmd.Flags().Set("file", "")
	_ = validateCmd.Flags().Set("dir", "")
	_ = validateCmd.RunE(validateCmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	// Just verify it runs without panic - actual validation depends on cwd content
}

// TestRunCommandDefaultDir tests run with default directory
func TestRunCommandDefaultDir(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = runCmd.Flags().Set("event", "")
	_ = runCmd.Flags().Set("workflow", "")
	_ = runCmd.Flags().Set("dir", "")
	_ = runCmd.RunE(runCmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	// Just verify it runs - results depend on cwd content
}

