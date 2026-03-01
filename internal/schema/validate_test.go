package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateWorkflow_Valid(t *testing.T) {
	// Test validating a valid workflow
	result := ValidateWorkflow("../../testdata/workflows/valid/simple.yml")
	if !result.Valid {
		t.Errorf("Expected valid workflow, but got errors: %v", result.Errors)
	}
}

func TestValidateWorkflow_ValidComplex(t *testing.T) {
	// Test validating a more complex valid workflow
	result := ValidateWorkflow("../../testdata/workflows/valid/all-triggers.yml")
	if !result.Valid {
		t.Errorf("Expected valid workflow, but got errors: %v", result.Errors)
	}
}

func TestValidateWorkflow_ValidExpressions(t *testing.T) {
	// Test validating a workflow with expressions
	result := ValidateWorkflow("../../testdata/workflows/valid/expressions.yml")
	if !result.Valid {
		t.Errorf("Expected valid workflow, but got errors: %v", result.Errors)
	}
}

func TestValidateWorkflow_InvalidMissingRequired(t *testing.T) {
	// Test validating a workflow with missing required fields
	result := ValidateWorkflow("../../testdata/workflows/invalid/missing-required.yml")
	if result.Valid {
		t.Errorf("Expected invalid workflow, but validation passed")
	}
	if len(result.Errors) == 0 {
		t.Errorf("Expected validation errors, but got none")
	}
}

func TestValidateWorkflow_InvalidSyntax(t *testing.T) {
	// Test validating a workflow with bad YAML syntax
	result := ValidateWorkflow("../../testdata/workflows/invalid/bad-syntax.yml")
	if result.Valid {
		t.Errorf("Expected invalid workflow, but validation passed")
	}
	if len(result.Errors) == 0 {
		t.Errorf("Expected validation errors, but got none")
	}
}

func TestValidateWorkflow_FileNotFound(t *testing.T) {
	// Test validating a non-existent file
	result := ValidateWorkflow("../../testdata/workflows/nonexistent.yml")
	if result.Valid {
		t.Errorf("Expected invalid result, but validation passed")
	}
	if len(result.Errors) == 0 {
		t.Errorf("Expected validation errors, but got none")
	}
}

func TestValidateWorkflowsInDir(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "hookflow-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .github/hookflows directory
	workflowDir := filepath.Join(tmpDir, ".github", "hookflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatalf("Failed to create workflow directory: %v", err)
	}

	// Copy a valid workflow
	sourceValid := "../../testdata/workflows/valid/simple.yml"
	destValid := filepath.Join(workflowDir, "simple.yml")

	content, err := os.ReadFile(sourceValid)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}

	if err := os.WriteFile(destValid, content, 0644); err != nil {
		t.Fatalf("Failed to write dest file: %v", err)
	}

	// Validate the directory
	result := ValidateWorkflowsInDir(tmpDir)
	if !result.Valid {
		t.Errorf("Expected valid workflows in directory, but got errors: %v", result.Errors)
	}
}

func TestValidateWorkflowsInDir_WithErrors(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "hookflow-test-invalid")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .github/hookflows directory
	workflowDir := filepath.Join(tmpDir, ".github", "hookflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatalf("Failed to create workflow directory: %v", err)
	}

	// Copy an invalid workflow
	sourceInvalid := "../../testdata/workflows/invalid/missing-required.yml"
	destInvalid := filepath.Join(workflowDir, "invalid.yml")

	content, err := os.ReadFile(sourceInvalid)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}

	if err := os.WriteFile(destInvalid, content, 0644); err != nil {
		t.Fatalf("Failed to write dest file: %v", err)
	}

	// Validate the directory
	result := ValidateWorkflowsInDir(tmpDir)
	if result.Valid {
		t.Errorf("Expected invalid workflows in directory, but validation passed")
	}
	if len(result.Errors) == 0 {
		t.Errorf("Expected validation errors, but got none")
	}
}

func TestValidateWorkflowsInDir_NoWorkflowDir(t *testing.T) {
	// Create a temporary directory without workflows
	tmpDir, err := os.MkdirTemp("", "hookflow-test-empty")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Validate the directory (should be valid - no workflows is not an error)
	result := ValidateWorkflowsInDir(tmpDir)
	if !result.Valid {
		t.Errorf("Expected valid result for directory with no workflows, but got errors: %v", result.Errors)
	}
}

func TestValidationError_Details(t *testing.T) {
	// Ensure validation errors contain details
	result := ValidateWorkflow("../../testdata/workflows/invalid/missing-required.yml")
	if !result.Valid {
		if len(result.Errors) > 0 {
			err := result.Errors[0]
			if err.File == "" {
				t.Errorf("Expected file path in error, got empty")
			}
			if err.Message == "" {
				t.Errorf("Expected message in error, got empty")
			}
		}
	}
}

