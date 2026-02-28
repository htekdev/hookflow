package trigger

import (
	"path/filepath"
	"strings"

	"github.com/htekdev/gh-hookflow/internal/schema"
)

// Matcher determines if a workflow should be triggered by an event
type Matcher struct {
	workflow *schema.Workflow
}

// NewMatcher creates a new trigger matcher for a workflow
func NewMatcher(workflow *schema.Workflow) *Matcher {
	return &Matcher{workflow: workflow}
}

// Match checks if the event matches any of the workflow's triggers
func (m *Matcher) Match(event *schema.Event) bool {
	on := m.workflow.On

	// Check tool trigger (most specific)
	if on.Tool != nil && event.Tool != nil {
		if m.matchToolTrigger(on.Tool, event.Tool) {
			return true
		}
	}

	// Check tools array
	if len(on.Tools) > 0 && event.Tool != nil {
		for _, toolTrigger := range on.Tools {
			if m.matchToolTrigger(&toolTrigger, event.Tool) {
				return true
			}
		}
	}

	// Check hooks trigger
	if on.Hooks != nil && event.Hook != nil {
		if m.matchHooksTrigger(on.Hooks, event.Hook) {
			return true
		}
	}

	// Check file trigger
	if on.File != nil && event.File != nil {
		if m.matchFileTrigger(on.File, event.File, event.GetLifecycle()) {
			return true
		}
	}

	// Check commit trigger
	if on.Commit != nil && event.Commit != nil {
		if m.matchCommitTrigger(on.Commit, event.Commit, event.GetLifecycle()) {
			return true
		}
	}

	// Check push trigger
	if on.Push != nil && event.Push != nil {
		if m.matchPushTrigger(on.Push, event.Push, event.GetLifecycle()) {
			return true
		}
	}

	return false
}

// matchToolTrigger checks if a tool event matches a tool trigger
func (m *Matcher) matchToolTrigger(trigger *schema.ToolTrigger, event *schema.ToolEvent) bool {
	// Check tool name
	if trigger.Name != event.Name {
		return false
	}

	// Check args patterns
	for argName, pattern := range trigger.Args {
		argValue, ok := event.Args[argName]
		if !ok {
			return false
		}
		argStr, _ := argValue.(string)
		if !matchGlob(pattern, argStr) {
			return false
		}
	}

	// Note: trigger.If expression is evaluated separately by expression engine
	return true
}

// matchHooksTrigger checks if a hook event matches a hooks trigger
func (m *Matcher) matchHooksTrigger(trigger *schema.HooksTrigger, event *schema.HookEvent) bool {
	// Check hook types
	if len(trigger.Types) > 0 {
		found := false
		for _, t := range trigger.Types {
			if t == event.Type {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check tools filter
	if len(trigger.Tools) > 0 && event.Tool != nil {
		found := false
		for _, t := range trigger.Tools {
			if t == event.Tool.Name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// matchFileTrigger checks if a file event matches a file trigger
func (m *Matcher) matchFileTrigger(trigger *schema.FileTrigger, event *schema.FileEvent, eventLifecycle string) bool {
	// Check lifecycle first
	if trigger.GetLifecycle() != eventLifecycle {
		return false
	}

	// Check file types
	if len(trigger.Types) > 0 {
		found := false
		for _, t := range trigger.Types {
			if t == event.Action {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check paths-ignore first
	if len(trigger.PathsIgnore) > 0 {
		for _, pattern := range trigger.PathsIgnore {
			if matchGlob(pattern, event.Path) {
				return false
			}
		}
	}

	// Check paths
	if len(trigger.Paths) > 0 {
		matched := false
		for _, pattern := range trigger.Paths {
			// Handle negation
			if strings.HasPrefix(pattern, "!") {
				if matchGlob(pattern[1:], event.Path) {
					matched = false
				}
			} else if matchGlob(pattern, event.Path) {
				matched = true
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// matchCommitTrigger checks if a commit event matches a commit trigger
func (m *Matcher) matchCommitTrigger(trigger *schema.CommitTrigger, event *schema.CommitEvent, eventLifecycle string) bool {
	// Check lifecycle first
	if trigger.GetLifecycle() != eventLifecycle {
		return false
	}

	// Check branches - would need branch info from context
	// For now, focus on path matching

	// Check paths-ignore
	if len(trigger.PathsIgnore) > 0 {
		allIgnored := true
		for _, file := range event.Files {
			ignored := false
			for _, pattern := range trigger.PathsIgnore {
				if matchGlob(pattern, file.Path) {
					ignored = true
					break
				}
			}
			if !ignored {
				allIgnored = false
				break
			}
		}
		if allIgnored {
			return false
		}
	}

	// Check paths
	if len(trigger.Paths) > 0 {
		matched := false
		for _, file := range event.Files {
			for _, pattern := range trigger.Paths {
				if strings.HasPrefix(pattern, "!") {
					continue
				}
				if matchGlob(pattern, file.Path) {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// matchPushTrigger checks if a push event matches a push trigger
func (m *Matcher) matchPushTrigger(trigger *schema.PushTrigger, event *schema.PushEvent, eventLifecycle string) bool {
	// Check lifecycle first
	if trigger.GetLifecycle() != eventLifecycle {
		return false
	}

	// Check branches
	if len(trigger.Branches) > 0 {
		branch := extractBranch(event.Ref)
		if branch != "" {
			matched := false
			for _, pattern := range trigger.Branches {
				if strings.HasPrefix(pattern, "!") {
					if matchGlob(pattern[1:], branch) {
						matched = false
					}
				} else if matchGlob(pattern, branch) {
					matched = true
				}
			}
			if !matched {
				return false
			}
		}
	}

	// Check branches-ignore
	if len(trigger.BranchesIgnore) > 0 {
		branch := extractBranch(event.Ref)
		if branch != "" {
			for _, pattern := range trigger.BranchesIgnore {
				if matchGlob(pattern, branch) {
					return false
				}
			}
		}
	}

	// Check tags
	if len(trigger.Tags) > 0 {
		tag := extractTag(event.Ref)
		if tag == "" {
			return false
		}
		matched := false
		for _, pattern := range trigger.Tags {
			if strings.HasPrefix(pattern, "!") {
				if matchGlob(pattern[1:], tag) {
					matched = false
				}
			} else if matchGlob(pattern, tag) {
				matched = true
			}
		}
		if !matched {
			return false
		}
	}

	// Check tags-ignore
	if len(trigger.TagsIgnore) > 0 {
		tag := extractTag(event.Ref)
		if tag != "" {
			for _, pattern := range trigger.TagsIgnore {
				if matchGlob(pattern, tag) {
					return false
				}
			}
		}
	}

	return true
}

// matchGlob performs glob pattern matching
func matchGlob(pattern, path string) bool {
	// Normalize path separators
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	// Handle ** patterns
	if strings.Contains(pattern, "**") {
		return matchDoubleGlob(pattern, path)
	}

	// Use filepath.Match for simple patterns
	matched, _ := filepath.Match(pattern, path)
	return matched
}

// matchDoubleGlob handles ** patterns that match across directories
func matchDoubleGlob(pattern, path string) bool {
	parts := strings.Split(pattern, "**")
	if len(parts) == 1 {
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	// For patterns like **/*.js
	if parts[0] == "" {
		suffix := strings.TrimPrefix(parts[1], "/")
		// Match suffix against any path segment
		pathParts := strings.Split(path, "/")
		for i := range pathParts {
			subpath := strings.Join(pathParts[i:], "/")
			if matched, _ := filepath.Match(suffix, subpath); matched {
				return true
			}
		}
		// Also try matching just the filename
		if matched, _ := filepath.Match(suffix, filepath.Base(path)); matched {
			return true
		}
		return false
	}

	// For patterns like src/**/test.js
	prefix := strings.TrimSuffix(parts[0], "/")
	suffix := strings.TrimPrefix(parts[1], "/")

	if !strings.HasPrefix(path, prefix) {
		return false
	}

	remaining := strings.TrimPrefix(path, prefix)
	remaining = strings.TrimPrefix(remaining, "/")

	if suffix == "" {
		return true
	}

	// Match suffix against remaining path
	pathParts := strings.Split(remaining, "/")
	for i := range pathParts {
		subpath := strings.Join(pathParts[i:], "/")
		if matched, _ := filepath.Match(suffix, subpath); matched {
			return true
		}
	}

	return false
}

// extractBranch extracts branch name from a ref
func extractBranch(ref string) string {
	const prefix = "refs/heads/"
	if strings.HasPrefix(ref, prefix) {
		return ref[len(prefix):]
	}
	return ""
}

// extractTag extracts tag name from a ref
func extractTag(ref string) string {
	const prefix = "refs/tags/"
	if strings.HasPrefix(ref, prefix) {
		return ref[len(prefix):]
	}
	return ""
}
