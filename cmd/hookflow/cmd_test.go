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

// TestFileTriggerWithActionsMatches tests that file trigger with 'actions' field matches correctly
func TestFileTriggerWithActionsMatches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-file-actions-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workflow using 'actions' (alias for 'types')
	workflow := `name: Block plugin.json edits
on:
  file:
    paths:
      - 'plugin.json'
    actions:
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

// TestFileTriggerNoMatchWrongAction tests that file trigger doesn't match wrong action
func TestFileTriggerNoMatchWrongAction(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hookflow-file-nomatch-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hooks")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workflow that only matches 'create' action
	workflow := `name: Block creates only
on:
  file:
    paths:
      - '**/*.json'
    actions:
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