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

	err = runMatchingWorkflows(tmpDir, eventJSON, "pre")

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

	err = runMatchingWorkflows(tmpDir, eventJSON, "pre")

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

	err = runMatchingWorkflows(tmpDir, eventJSON, "pre")

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

// TestIsHookflowSelfRepair tests the self-repair detection function
func TestIsHookflowSelfRepair(t *testing.T) {
	tests := []struct {
		name     string
		event    *schema.Event
		expected bool
	}{
		{
			name: "edit to hookflow workflow yml",
			event: &schema.Event{
				File: &schema.FileEvent{
					Path:   ".github/hooks/my-workflow.yml",
					Action: "edit",
				},
			},
			expected: true,
		},
		{
			name: "create hookflow workflow yaml",
			event: &schema.Event{
				File: &schema.FileEvent{
					Path:   ".github/hooks/new-workflow.yaml",
					Action: "create",
				},
			},
			expected: true,
		},
		{
			name: "edit to non-hookflow file",
			event: &schema.Event{
				File: &schema.FileEvent{
					Path:   "src/main.go",
					Action: "edit",
				},
			},
			expected: false,
		},
		{
			name: "edit to hookflow but not yml",
			event: &schema.Event{
				File: &schema.FileEvent{
					Path:   ".github/hooks/README.md",
					Action: "edit",
				},
			},
			expected: false,
		},
		{
			name: "delete hookflow workflow - not allowed for self-repair",
			event: &schema.Event{
				File: &schema.FileEvent{
					Path:   ".github/hooks/my-workflow.yml",
					Action: "delete",
				},
			},
			expected: false,
		},
		{
			name: "no file event",
			event: &schema.Event{
				Tool: &schema.ToolEvent{
					Name: "edit",
				},
			},
			expected: false,
		},
		{
			name: "nested path in hooks",
			event: &schema.Event{
				File: &schema.FileEvent{
					Path:   ".github/hooks/subdir/workflow.yml",
					Action: "edit",
				},
			},
			expected: true,
		},
		{
			name: "windows-style path",
			event: &schema.Event{
				File: &schema.FileEvent{
					Path:   ".github\\hooks\\workflow.yml",
					Action: "edit",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHookflowSelfRepair(tt.event, "/test/dir")
			if result != tt.expected {
				t.Errorf("isHookflowSelfRepair() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestInvalidWorkflowDeniesNonHookflowEdits tests that invalid workflows deny non-hookflow edits
func TestInvalidWorkflowDeniesNonHookflowEdits(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-invalid-deny-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create an invalid workflow (unknown field)
	invalidWorkflow := `name: invalid
on:
  file:
    unknown_field: true
steps:
  - run: echo test
`
	if err := os.WriteFile(filepath.Join(workflowDir, "invalid.yml"), []byte(invalidWorkflow), 0644); err != nil {
		t.Fatal(err)
	}

	// Use runMatchingWorkflowsWithEvent directly with a file event for non-hookflow file
	evt := &schema.Event{
		File: &schema.FileEvent{
			Path:   "src/main.go",
			Action: "edit",
		},
		Cwd: tmpDir,
	}

	oldStdout := os.Stdout
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW

	_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

	_ = stdoutW.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(stdoutR)
	output := buf.String()

	// Should deny because workflow is invalid
	if !strings.Contains(output, "deny") {
		t.Errorf("Expected deny for invalid workflow, got: %s", output)
	}
	if !strings.Contains(output, "Invalid workflow") {
		t.Errorf("Expected 'Invalid workflow' in reason, got: %s", output)
	}
}

// TestInvalidWorkflowAllowsSelfRepair tests that invalid workflows allow edits to .github/hooks/
func TestInvalidWorkflowAllowsSelfRepair(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-self-repair-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create an invalid workflow
	invalidWorkflow := `name: invalid
on:
  file:
    unknown_field: true
steps:
  - run: echo test
`
	if err := os.WriteFile(filepath.Join(workflowDir, "invalid.yml"), []byte(invalidWorkflow), 0644); err != nil {
		t.Fatal(err)
	}

	// Use runMatchingWorkflowsWithEvent directly - editing the hookflow workflow itself
	evt := &schema.Event{
		File: &schema.FileEvent{
			Path:   ".github/hooks/invalid.yml",
			Action: "edit",
		},
		Cwd: tmpDir,
	}

	oldStdout := os.Stdout
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW

	_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

	_ = stdoutW.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(stdoutR)
	output := buf.String()

	// Should allow because it's a self-repair edit to .github/hooks/
	if !strings.Contains(output, "allow") {
		t.Errorf("Expected allow for self-repair edit, got: %s", output)
	}
	if !strings.Contains(output, "self-repair") {
		t.Errorf("Expected 'self-repair' in reason, got: %s", output)
	}
}

// TestFileTriggerWithTypesMatches tests that file trigger with 'types' field matches correctly
func TestFileTriggerWithTypesMatches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-file-types-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workflow using 'types' for file events
	workflow := `name: Block plugin.json edits
on:
  file:
    paths:
      - 'plugin.json'
    types:
      - edit
blocking: true
steps:
  - name: Block edit
    run: |
      echo "Blocking edit to plugin.json"
      exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "block-plugin.yml"), []byte(workflow), 0644); err != nil {
		t.Fatal(err)
	}

	// Create event for editing plugin.json
	evt := &schema.Event{
		File: &schema.FileEvent{
			Path:   "plugin.json",
			Action: "edit",
		},
		Cwd: tmpDir,
	}

	oldStdout := os.Stdout
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW

	_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

	_ = stdoutW.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(stdoutR)
	output := buf.String()

	// Should deny because workflow matches and step exits 1
	if !strings.Contains(output, "deny") {
		t.Errorf("Expected deny when workflow matches file edit, got: %s", output)
	}
	if !strings.Contains(output, "Block plugin.json edits") {
		t.Errorf("Expected workflow name in output, got: %s", output)
	}
}

// TestFileTriggerWithMultipleTypes tests that file trigger matches multiple types
func TestFileTriggerWithMultipleTypes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-file-types-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workflow using 'types' (original field name)
	workflow := `name: Block config edits
on:
  file:
    paths:
      - 'config.json'
    types:
      - edit
      - create
blocking: true
steps:
  - name: Block
    run: exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "block-config.yml"), []byte(workflow), 0644); err != nil {
		t.Fatal(err)
	}

	// Create event
	evt := &schema.Event{
		File: &schema.FileEvent{
			Path:   "config.json",
			Action: "edit",
		},
		Cwd: tmpDir,
	}

	oldStdout := os.Stdout
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW

	_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

	_ = stdoutW.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(stdoutR)
	output := buf.String()

	// Should deny
	if !strings.Contains(output, "deny") {
		t.Errorf("Expected deny when workflow matches, got: %s", output)
	}
}

// TestFileTriggerNoMatchWrongType tests that file trigger doesn't match wrong type
func TestFileTriggerNoMatchWrongType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-file-nomatch-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workflow that only matches 'create' type
	workflow := `name: Block creates only
on:
  file:
    paths:
      - '**/*.json'
    types:
      - create
blocking: true
steps:
  - name: Block
    run: exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "block-create.yml"), []byte(workflow), 0644); err != nil {
		t.Fatal(err)
	}

	// Create event with 'edit' action - should NOT match workflow that only wants 'create'
	evt := &schema.Event{
		File: &schema.FileEvent{
			Path:   "test.json",
			Action: "edit",
		},
		Cwd: tmpDir,
	}

	oldStdout := os.Stdout
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW

	_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

	_ = stdoutW.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(stdoutR)
	output := buf.String()

	// Should allow because action doesn't match (edit vs create)
	if !strings.Contains(output, "allow") {
		t.Errorf("Expected allow when action doesn't match, got: %s", output)
	}
}

// TestConditionalStepWithToolArgsNewStr tests that a step with if condition on event.tool.args.new_str works
func TestConditionalStepWithToolArgsNewStr(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-conditional-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workflow with conditional step based on new_str content
	workflow := `name: Block password in code
on:
  file:
    paths:
      - '**/*.js'
    types:
      - edit
blocking: true
steps:
  - name: Check for password
    if: contains(event.tool.args.new_str, 'password')
    run: |
      echo "Found password in code!"
      exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "block-password.yml"), []byte(workflow), 0644); err != nil {
		t.Fatal(err)
	}

	// Test case 1: new_str contains "password" - should deny
	t.Run("matches when new_str contains password", func(t *testing.T) {
		evt := &schema.Event{
			File: &schema.FileEvent{
				Path:   "src/auth.js",
				Action: "edit",
			},
			Tool: &schema.ToolEvent{
				Name: "edit",
				Args: map[string]interface{}{
					"path":    "src/auth.js",
					"new_str": "const password = 'secret123';",
				},
			},
			Cwd: tmpDir,
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		if !strings.Contains(output, "deny") {
			t.Errorf("Expected deny when new_str contains 'password', got: %s", output)
		}
	})

	// Test case 2: new_str does NOT contain "password" - should allow (step skipped)
	t.Run("allows when new_str does not contain password", func(t *testing.T) {
		evt := &schema.Event{
			File: &schema.FileEvent{
				Path:   "src/utils.js",
				Action: "edit",
			},
			Tool: &schema.ToolEvent{
				Name: "edit",
				Args: map[string]interface{}{
					"path":    "src/utils.js",
					"new_str": "const username = 'john';",
				},
			},
			Cwd: tmpDir,
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		// When step is skipped due to if condition, workflow should allow
		if strings.Contains(output, "deny") {
			t.Errorf("Expected allow when new_str doesn't contain 'password', got: %s", output)
		}
	})
}

// TestConditionalStepWithFileContent tests conditional step using event.file.content
func TestConditionalStepWithFileContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-file-content-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workflow with conditional step based on file content
	workflow := `name: Block API keys in new files
on:
  file:
    paths:
      - '**/*.env'
    types:
      - create
blocking: true
steps:
  - name: Check for API key
    if: contains(event.file.content, 'API_KEY')
    run: |
      echo "Found API_KEY in file!"
      exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "block-apikey.yml"), []byte(workflow), 0644); err != nil {
		t.Fatal(err)
	}

	// Test case 1: content contains "API_KEY" - should deny
	t.Run("denies when file content contains API_KEY", func(t *testing.T) {
		evt := &schema.Event{
			File: &schema.FileEvent{
				Path:    "config/.env",
				Action:  "create",
				Content: "DATABASE_URL=postgres://localhost\nAPI_KEY=sk-12345\n",
			},
			Tool: &schema.ToolEvent{
				Name: "create",
				Args: map[string]interface{}{
					"path":      "config/.env",
					"file_text": "DATABASE_URL=postgres://localhost\nAPI_KEY=sk-12345\n",
				},
			},
			Cwd: tmpDir,
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		if !strings.Contains(output, "deny") {
			t.Errorf("Expected deny when file content contains 'API_KEY', got: %s", output)
		}
	})

	// Test case 2: content does NOT contain "API_KEY" - should allow
	t.Run("allows when file content does not contain API_KEY", func(t *testing.T) {
		evt := &schema.Event{
			File: &schema.FileEvent{
				Path:    "config/.env",
				Action:  "create",
				Content: "DATABASE_URL=postgres://localhost\nDEBUG=true\n",
			},
			Tool: &schema.ToolEvent{
				Name: "create",
				Args: map[string]interface{}{
					"path":      "config/.env",
					"file_text": "DATABASE_URL=postgres://localhost\nDEBUG=true\n",
				},
			},
			Cwd: tmpDir,
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		if strings.Contains(output, "deny") {
			t.Errorf("Expected allow when file content doesn't contain 'API_KEY', got: %s", output)
		}
	})
}

// TestConditionalStepFileEventWithoutToolContext tests what happens when file trigger
// tries to access event.tool.args.new_str but Tool context is not set
func TestConditionalStepFileEventWithoutToolContext(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-file-no-tool-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workflow with conditional step based on new_str content
	workflow := `name: Block password in code
on:
  file:
    paths:
      - '**/*.js'
    types:
      - edit
blocking: true
steps:
  - name: Check for password
    if: contains(event.tool.args.new_str, 'password')
    run: |
      echo "Found password in code!"
      exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "block-password.yml"), []byte(workflow), 0644); err != nil {
		t.Fatal(err)
	}

	// File event WITHOUT Tool context - simulates pure file trigger scenario
	t.Run("file event without tool context", func(t *testing.T) {
		evt := &schema.Event{
			File: &schema.FileEvent{
				Path:   "src/auth.js",
				Action: "edit",
			},
			// Tool is nil - no tool context available
			Cwd: tmpDir,
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		// When Tool is nil, event.tool.args.new_str should evaluate to empty/nil
		// The step should be skipped (if condition false) and workflow should allow
		t.Logf("Output: %s", output)
		// Document current behavior - what actually happens?
		if strings.Contains(output, "deny") {
			t.Logf("BEHAVIOR: Denies when Tool context is nil")
		} else if strings.Contains(output, "allow") {
			t.Logf("BEHAVIOR: Allows when Tool context is nil (step skipped)")
		} else {
			t.Logf("BEHAVIOR: Unknown - output: %s", output)
		}
	})
}

// ============================================================================
// LIFECYCLE TESTS
// ============================================================================

// TestLifecyclePreMatchesPre tests that pre lifecycle triggers match pre events
func TestLifecyclePreMatchesPre(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-lifecycle-pre-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Workflow with explicit lifecycle: pre
	workflow := `name: Pre-file validation
on:
  file:
    lifecycle: pre
    paths:
      - '**/*.js'
    types:
      - edit
blocking: true
steps:
  - name: Block
    run: exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "pre-check.yml"), []byte(workflow), 0644); err != nil {
		t.Fatal(err)
	}

	// Pre event should match
	evt := &schema.Event{
		File: &schema.FileEvent{
			Path:   "src/app.js",
			Action: "edit",
		},
		Cwd:       tmpDir,
		Lifecycle: "pre",
	}

	oldStdout := os.Stdout
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW

	_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

	_ = stdoutW.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(stdoutR)
	output := buf.String()

	if !strings.Contains(output, "deny") {
		t.Errorf("Expected deny for pre event matching pre workflow, got: %s", output)
	}
}

// TestLifecyclePostMatchesPost tests that post lifecycle triggers match post events
func TestLifecyclePostMatchesPost(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-lifecycle-post-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Workflow with lifecycle: post
	workflow := `name: Post-file linting
on:
  file:
    lifecycle: post
    paths:
      - '**/*.js'
    types:
      - edit
blocking: true
steps:
  - name: Lint
    run: exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "post-lint.yml"), []byte(workflow), 0644); err != nil {
		t.Fatal(err)
	}

	// Post event should match
	evt := &schema.Event{
		File: &schema.FileEvent{
			Path:   "src/app.js",
			Action: "edit",
		},
		Cwd:       tmpDir,
		Lifecycle: "post",
	}

	oldStdout := os.Stdout
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW

	_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

	_ = stdoutW.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(stdoutR)
	output := buf.String()

	if !strings.Contains(output, "deny") {
		t.Errorf("Expected deny for post event matching post workflow, got: %s", output)
	}
}

// TestLifecyclePreDoesNotMatchPost tests that pre workflow doesn't match post event
func TestLifecyclePreDoesNotMatchPost(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-lifecycle-mismatch-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Workflow with lifecycle: pre (explicit)
	workflow := `name: Pre-file validation
on:
  file:
    lifecycle: pre
    paths:
      - '**/*.js'
    types:
      - edit
blocking: true
steps:
  - name: Block
    run: exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "pre-check.yml"), []byte(workflow), 0644); err != nil {
		t.Fatal(err)
	}

	// Post event should NOT match pre workflow
	evt := &schema.Event{
		File: &schema.FileEvent{
			Path:   "src/app.js",
			Action: "edit",
		},
		Cwd:       tmpDir,
		Lifecycle: "post",
	}

	oldStdout := os.Stdout
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW

	_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

	_ = stdoutW.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(stdoutR)
	output := buf.String()

	if strings.Contains(output, "deny") {
		t.Errorf("Expected allow for post event with pre workflow (no match), got: %s", output)
	}
}

// TestLifecyclePostDoesNotMatchPre tests that post workflow doesn't match pre event
func TestLifecyclePostDoesNotMatchPre(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-lifecycle-mismatch2-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Workflow with lifecycle: post
	workflow := `name: Post-file linting
on:
  file:
    lifecycle: post
    paths:
      - '**/*.js'
    types:
      - edit
blocking: true
steps:
  - name: Lint
    run: exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "post-lint.yml"), []byte(workflow), 0644); err != nil {
		t.Fatal(err)
	}

	// Pre event should NOT match post workflow
	evt := &schema.Event{
		File: &schema.FileEvent{
			Path:   "src/app.js",
			Action: "edit",
		},
		Cwd:       tmpDir,
		Lifecycle: "pre",
	}

	oldStdout := os.Stdout
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW

	_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

	_ = stdoutW.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(stdoutR)
	output := buf.String()

	if strings.Contains(output, "deny") {
		t.Errorf("Expected allow for pre event with post workflow (no match), got: %s", output)
	}
}

// TestLifecycleDefaultIsPre tests that workflow without lifecycle defaults to pre
func TestLifecycleDefaultIsPre(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-lifecycle-default-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Workflow WITHOUT lifecycle (should default to pre)
	workflow := `name: Default lifecycle validation
on:
  file:
    paths:
      - '**/*.js'
    types:
      - edit
blocking: true
steps:
  - name: Block
    run: exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "default-check.yml"), []byte(workflow), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("matches pre event (default)", func(t *testing.T) {
		evt := &schema.Event{
			File: &schema.FileEvent{
				Path:   "src/app.js",
				Action: "edit",
			},
			Cwd:       tmpDir,
			Lifecycle: "pre",
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		if !strings.Contains(output, "deny") {
			t.Errorf("Expected deny for pre event with default workflow, got: %s", output)
		}
	})

	t.Run("does not match post event", func(t *testing.T) {
		evt := &schema.Event{
			File: &schema.FileEvent{
				Path:   "src/app.js",
				Action: "edit",
			},
			Cwd:       tmpDir,
			Lifecycle: "post",
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		if strings.Contains(output, "deny") {
			t.Errorf("Expected allow for post event with default (pre) workflow, got: %s", output)
		}
	})

	t.Run("matches empty lifecycle (treated as pre)", func(t *testing.T) {
		evt := &schema.Event{
			File: &schema.FileEvent{
				Path:   "src/app.js",
				Action: "edit",
			},
			Cwd:       tmpDir,
			Lifecycle: "", // Empty defaults to pre
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		if !strings.Contains(output, "deny") {
			t.Errorf("Expected deny for empty lifecycle event with default workflow, got: %s", output)
		}
	})
}

// TestLifecycleCommitTrigger tests lifecycle on commit triggers
func TestLifecycleCommitTrigger(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-commit-lifecycle-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Pre-commit workflow
	preWorkflow := `name: Pre-commit check
on:
  commit:
    lifecycle: pre
    paths:
      - 'src/**'
blocking: true
steps:
  - name: Check
    run: exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "pre-commit.yml"), []byte(preWorkflow), 0644); err != nil {
		t.Fatal(err)
	}

	// Post-commit workflow
	postWorkflow := `name: Post-commit notification
on:
  commit:
    lifecycle: post
    paths:
      - 'src/**'
blocking: false
steps:
  - name: Notify
    run: echo "Commit done"
`
	if err := os.WriteFile(filepath.Join(workflowDir, "post-commit.yml"), []byte(postWorkflow), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("pre commit event matches pre workflow", func(t *testing.T) {
		evt := &schema.Event{
			Commit: &schema.CommitEvent{
				SHA:     "abc123",
				Message: "test commit",
				Files:   []schema.FileStatus{{Path: "src/main.go", Status: "modified"}},
			},
			Cwd:       tmpDir,
			Lifecycle: "pre",
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		if !strings.Contains(output, "deny") {
			t.Errorf("Expected deny for pre commit event, got: %s", output)
		}
	})

	t.Run("post commit event matches post workflow", func(t *testing.T) {
		evt := &schema.Event{
			Commit: &schema.CommitEvent{
				SHA:     "abc123",
				Message: "test commit",
				Files:   []schema.FileStatus{{Path: "src/main.go", Status: "modified"}},
			},
			Cwd:       tmpDir,
			Lifecycle: "post",
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		// Post workflow is non-blocking and step succeeds
		if strings.Contains(output, "deny") {
			t.Errorf("Expected allow for post commit event (non-blocking), got: %s", output)
		}
	})
}

// TestLifecyclePushTrigger tests lifecycle on push triggers
func TestLifecyclePushTrigger(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-push-lifecycle-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Pre-push workflow
	workflow := `name: Pre-push check
on:
  push:
    lifecycle: pre
    branches:
      - main
blocking: true
steps:
  - name: Check
    run: exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "pre-push.yml"), []byte(workflow), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("pre push matches pre workflow", func(t *testing.T) {
		evt := &schema.Event{
			Push: &schema.PushEvent{
				Ref: "refs/heads/main",
			},
			Cwd:       tmpDir,
			Lifecycle: "pre",
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		if !strings.Contains(output, "deny") {
			t.Errorf("Expected deny for pre push event, got: %s", output)
		}
	})

	t.Run("post push does not match pre workflow", func(t *testing.T) {
		evt := &schema.Event{
			Push: &schema.PushEvent{
				Ref: "refs/heads/main",
			},
			Cwd:       tmpDir,
			Lifecycle: "post",
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		if strings.Contains(output, "deny") {
			t.Errorf("Expected allow for post push with pre workflow, got: %s", output)
		}
	})
}

// TestLifecycleMixedWorkflows tests having both pre and post workflows
func TestLifecycleMixedWorkflows(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-mixed-lifecycle-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Pre-edit workflow (blocks)
	preWorkflow := `name: Pre-edit validation
on:
  file:
    lifecycle: pre
    paths:
      - '**/*.ts'
    types:
      - edit
blocking: true
steps:
  - name: Validate
    run: exit 1
`
	if err := os.WriteFile(filepath.Join(workflowDir, "pre-edit.yml"), []byte(preWorkflow), 0644); err != nil {
		t.Fatal(err)
	}

	// Post-edit workflow (runs after)
	postWorkflow := `name: Post-edit linting
on:
  file:
    lifecycle: post
    paths:
      - '**/*.ts'
    types:
      - edit
blocking: false
steps:
  - name: Lint
    run: echo "Linting completed"
`
	if err := os.WriteFile(filepath.Join(workflowDir, "post-edit.yml"), []byte(postWorkflow), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("pre event only runs pre workflow", func(t *testing.T) {
		evt := &schema.Event{
			File: &schema.FileEvent{
				Path:   "src/index.ts",
				Action: "edit",
			},
			Cwd:       tmpDir,
			Lifecycle: "pre",
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		// Should deny because pre workflow blocks
		if !strings.Contains(output, "deny") {
			t.Errorf("Expected deny for pre event, got: %s", output)
		}
		// Should mention pre-edit workflow
		if !strings.Contains(output, "Pre-edit") {
			t.Errorf("Expected Pre-edit workflow name in output, got: %s", output)
		}
	})

	t.Run("post event only runs post workflow", func(t *testing.T) {
		evt := &schema.Event{
			File: &schema.FileEvent{
				Path:   "src/index.ts",
				Action: "edit",
			},
			Cwd:       tmpDir,
			Lifecycle: "post",
		}

		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		_ = runMatchingWorkflowsWithEvent(tmpDir, evt)

		_ = stdoutW.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(stdoutR)
		output := buf.String()

		// Should allow because post workflow is non-blocking
		if strings.Contains(output, "deny") {
			t.Errorf("Expected allow for post event (non-blocking), got: %s", output)
		}
	})
}

// ============================================================================
// EVENT TYPE CONVERSION TESTS
// ============================================================================

// TestEventTypeToLifecycle tests the conversion from Copilot hook types to lifecycle
func TestEventTypeToLifecycle(t *testing.T) {
	tests := []struct {
		eventType string
		expected  string
	}{
		// Standard Copilot hook types
		{"preToolUse", "pre"},
		{"postToolUse", "post"},
		// Short forms
		{"pre", "pre"},
		{"post", "post"},
		// Unknown defaults to pre
		{"", "pre"},
		{"unknown", "pre"},
		{"something", "pre"},
		// Future hook types can be added here
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			result := eventTypeToLifecycle(tt.eventType)
			if result != tt.expected {
				t.Errorf("eventTypeToLifecycle(%q) = %q, want %q", tt.eventType, result, tt.expected)
			}
		})
	}
}
