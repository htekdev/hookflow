package schema

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadWorkflow loads a workflow from a YAML file
func LoadWorkflow(filePath string) (*Workflow, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	// Parse YAML
	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	return &workflow, nil
}

// LoadAndValidateWorkflow loads and validates a workflow using JSON schema
func LoadAndValidateWorkflow(filePath string) (*Workflow, error) {
	// First validate with JSON schema
	result := ValidateWorkflow(filePath)
	if !result.Valid {
		// Return first error
		if len(result.Errors) > 0 {
			return nil, fmt.Errorf("%s", result.Errors[0].Message)
		}
		return nil, fmt.Errorf("workflow validation failed")
	}

	// Then load the workflow
	return LoadWorkflow(filePath)
}

// LoadEvent loads an event from a JSON string
func LoadEvent(jsonStr string) (*Event, error) {
	// For now, we'll just return a nil event
	// In the future, this would parse JSON
	return nil, nil
}
