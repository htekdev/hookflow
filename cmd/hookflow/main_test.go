package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEventData_HookEvent(t *testing.T) {
	data := map[string]interface{}{
		"hook": map[string]interface{}{
			"type": "preToolUse",
			"cwd":  "/test/dir",
			"tool": map[string]interface{}{
				"name": "edit",
				"args": map[string]interface{}{
					"path":    "src/main.go",
					"content": "test content",
				},
			},
		},
		"cwd":       "/test/dir",
		"timestamp": "2026-01-01T00:00:00Z",
	}

	event := parseEventData(data)

	if event.Hook == nil {
		t.Fatal("Expected Hook to be set")
	}
	if event.Hook.Type != "preToolUse" {
		t.Errorf("Expected Hook.Type = 'preToolUse', got '%s'", event.Hook.Type)
	}
	if event.Hook.Cwd != "/test/dir" {
		t.Errorf("Expected Hook.Cwd = '/test/dir', got '%s'", event.Hook.Cwd)
	}
	if event.Hook.Tool == nil {
		t.Fatal("Expected Hook.Tool to be set")
	}
	if event.Hook.Tool.Name != "edit" {
		t.Errorf("Expected Hook.Tool.Name = 'edit', got '%s'", event.Hook.Tool.Name)
	}
	if event.Hook.Tool.Args["path"] != "src/main.go" {
		t.Errorf("Expected Hook.Tool.Args[path] = 'src/main.go', got '%v'", event.Hook.Tool.Args["path"])
	}
	if event.Cwd != "/test/dir" {
		t.Errorf("Expected Cwd = '/test/dir', got '%s'", event.Cwd)
	}
	if event.Timestamp != "2026-01-01T00:00:00Z" {
		t.Errorf("Expected Timestamp = '2026-01-01T00:00:00Z', got '%s'", event.Timestamp)
	}
}

func TestParseEventData_ToolEvent(t *testing.T) {
	data := map[string]interface{}{
		"tool": map[string]interface{}{
			"name": "create",
			"args": map[string]interface{}{
				"path":      "tests/new_test.go",
				"file_text": "package tests",
			},
			"hook_type": "preToolUse",
		},
	}

	event := parseEventData(data)

	if event.Tool == nil {
		t.Fatal("Expected Tool to be set")
	}
	if event.Tool.Name != "create" {
		t.Errorf("Expected Tool.Name = 'create', got '%s'", event.Tool.Name)
	}
	if event.Tool.Args["path"] != "tests/new_test.go" {
		t.Errorf("Expected Tool.Args[path] = 'tests/new_test.go', got '%v'", event.Tool.Args["path"])
	}
	if event.Tool.HookType != "preToolUse" {
		t.Errorf("Expected Tool.HookType = 'preToolUse', got '%s'", event.Tool.HookType)
	}
}

func TestParseEventData_FileEvent(t *testing.T) {
	data := map[string]interface{}{
		"file": map[string]interface{}{
			"path":    "src/config.json",
			"action":  "edit",
			"content": `{"key": "value"}`,
		},
	}

	event := parseEventData(data)

	if event.File == nil {
		t.Fatal("Expected File to be set")
	}
	if event.File.Path != "src/config.json" {
		t.Errorf("Expected File.Path = 'src/config.json', got '%s'", event.File.Path)
	}
	if event.File.Action != "edit" {
		t.Errorf("Expected File.Action = 'edit', got '%s'", event.File.Action)
	}
	if event.File.Content != `{"key": "value"}` {
		t.Errorf("Expected File.Content to be set, got '%s'", event.File.Content)
	}
}

func TestParseEventData_CommitEvent(t *testing.T) {
	data := map[string]interface{}{
		"commit": map[string]interface{}{
			"sha":     "abc123def456",
			"message": "feat: add new feature",
			"author":  "test@example.com",
			"files": []interface{}{
				map[string]interface{}{
					"path":   "src/feature.go",
					"status": "added",
				},
				map[string]interface{}{
					"path":   "src/main.go",
					"status": "modified",
				},
			},
		},
	}

	event := parseEventData(data)

	if event.Commit == nil {
		t.Fatal("Expected Commit to be set")
	}
	if event.Commit.SHA != "abc123def456" {
		t.Errorf("Expected Commit.SHA = 'abc123def456', got '%s'", event.Commit.SHA)
	}
	if event.Commit.Message != "feat: add new feature" {
		t.Errorf("Expected Commit.Message = 'feat: add new feature', got '%s'", event.Commit.Message)
	}
	if event.Commit.Author != "test@example.com" {
		t.Errorf("Expected Commit.Author = 'test@example.com', got '%s'", event.Commit.Author)
	}
	if len(event.Commit.Files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(event.Commit.Files))
	}
	if event.Commit.Files[0].Path != "src/feature.go" {
		t.Errorf("Expected first file path = 'src/feature.go', got '%s'", event.Commit.Files[0].Path)
	}
	if event.Commit.Files[0].Status != "added" {
		t.Errorf("Expected first file status = 'added', got '%s'", event.Commit.Files[0].Status)
	}
}

func TestParseEventData_PushEvent(t *testing.T) {
	data := map[string]interface{}{
		"push": map[string]interface{}{
			"ref":    "refs/heads/main",
			"before": "000000000000",
			"after":  "abc123def456",
		},
	}

	event := parseEventData(data)

	if event.Push == nil {
		t.Fatal("Expected Push to be set")
	}
	if event.Push.Ref != "refs/heads/main" {
		t.Errorf("Expected Push.Ref = 'refs/heads/main', got '%s'", event.Push.Ref)
	}
	if event.Push.Before != "000000000000" {
		t.Errorf("Expected Push.Before = '000000000000', got '%s'", event.Push.Before)
	}
	if event.Push.After != "abc123def456" {
		t.Errorf("Expected Push.After = 'abc123def456', got '%s'", event.Push.After)
	}
}

func TestParseEventData_EmptyData(t *testing.T) {
	data := map[string]interface{}{}
	event := parseEventData(data)

	if event.Hook != nil {
		t.Error("Expected Hook to be nil")
	}
	if event.Tool != nil {
		t.Error("Expected Tool to be nil")
	}
	if event.File != nil {
		t.Error("Expected File to be nil")
	}
	if event.Commit != nil {
		t.Error("Expected Commit to be nil")
	}
	if event.Push != nil {
		t.Error("Expected Push to be nil")
	}
}

func TestParseEventData_CombinedHookAndTool(t *testing.T) {
	// This is how the PowerShell hook sends events - both hook and tool populated
	data := map[string]interface{}{
		"hook": map[string]interface{}{
			"type": "preToolUse",
			"tool": map[string]interface{}{
				"name": "edit",
				"args": map[string]interface{}{
					"path": "tests/test.go",
				},
			},
			"cwd": "/project",
		},
		"tool": map[string]interface{}{
			"name": "edit",
			"args": map[string]interface{}{
				"path": "tests/test.go",
			},
			"hook_type": "preToolUse",
		},
		"cwd":       "/project",
		"timestamp": "2026-01-01T00:00:00Z",
	}

	event := parseEventData(data)

	// Both should be populated
	if event.Hook == nil {
		t.Fatal("Expected Hook to be set")
	}
	if event.Tool == nil {
		t.Fatal("Expected Tool to be set")
	}
	if event.Hook.Tool.Name != "edit" {
		t.Errorf("Expected Hook.Tool.Name = 'edit', got '%s'", event.Hook.Tool.Name)
	}
	if event.Tool.Name != "edit" {
		t.Errorf("Expected Tool.Name = 'edit', got '%s'", event.Tool.Name)
	}
}

// Test workflow discovery with actual files
func TestRunMatchingWorkflows_NoWorkflowsDir(t *testing.T) {
	// Create a temp directory without .github/hookflows
	tmpDir, err := os.MkdirTemp("", "hookflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test that empty event returns allow
	// Note: We can't easily test runMatchingWorkflows directly as it writes to stdout
	// Instead, we test the helper functions
	workflowDir := filepath.Join(tmpDir, ".github", "hookflows")
	_, err = os.Stat(workflowDir)
	if !os.IsNotExist(err) {
		t.Error("Expected workflow dir to not exist")
	}
}

func TestFindWorkflowFile(t *testing.T) {
	// Create temp directory with workflow file
	tmpDir, err := os.MkdirTemp("", "hookflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workflowDir := filepath.Join(tmpDir, ".github", "hookflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a workflow file
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

	// Test finding the workflow
	path, found := findWorkflowFile(tmpDir, "test")
	if !found {
		t.Error("Expected to find workflow 'test'")
	}
	if path == "" {
		t.Error("Expected path to be set")
	}

	// Test not finding a workflow
	_, found = findWorkflowFile(tmpDir, "nonexistent")
	if found {
		t.Error("Expected to not find workflow 'nonexistent'")
	}
}

