package event

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/htekdev/gh-hookflow/internal/schema"
)

// RealGitProvider executes actual git commands to gather context
type RealGitProvider struct{}

// GetBranch returns the current git branch
func (g *RealGitProvider) GetBranch(cwd string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GetAuthor returns the git user email
func (g *RealGitProvider) GetAuthor(cwd string) string {
	cmd := exec.Command("git", "config", "user.email")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GetStagedFiles returns files currently staged for commit
func (g *RealGitProvider) GetStagedFiles(cwd string) []schema.FileStatus {
	cmd := exec.Command("git", "diff", "--cached", "--name-status")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	return parseGitStatus(string(out))
}

// GetPendingFiles returns files that would be affected by a git add command
// This handles the case where "git add . && git commit" is run - the files
// aren't staged yet when the hook fires
func (g *RealGitProvider) GetPendingFiles(cwd string, command string) []schema.FileStatus {
	// Get all modified/untracked files
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	allFiles := parsePorcelainStatus(string(out))

	// If command has specific files in git add, filter to those
	addFiles := ExtractGitAddFiles(command)
	if len(addFiles) == 0 {
		// git add . or git add -A - return all files
		return allFiles
	}

	// Filter to files matching the git add patterns
	var filtered []schema.FileStatus
	for _, f := range allFiles {
		for _, pattern := range addFiles {
			if matchGitAddPattern(f.Path, pattern, cwd) {
				filtered = append(filtered, f)
				break
			}
		}
	}
	return filtered
}

// GetRemote returns the default remote (usually "origin")
func (g *RealGitProvider) GetRemote(cwd string) string {
	cmd := exec.Command("git", "remote")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return "origin"
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 0 && lines[0] != "" {
		return lines[0]
	}
	return "origin"
}

// GetAheadBehind returns how many commits ahead/behind the remote
func (g *RealGitProvider) GetAheadBehind(cwd string) (ahead, behind int) {
	cmd := exec.Command("git", "rev-list", "--left-right", "--count", "@{upstream}...HEAD")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) >= 2 {
		// First number is behind, second is ahead
		if b, err := parseCount(parts[0]); err == nil {
			behind = b
		}
		if a, err := parseCount(parts[1]); err == nil {
			ahead = a
		}
	}
	return ahead, behind
}

// parseGitStatus parses git diff --name-status output
func parseGitStatus(output string) []schema.FileStatus {
	var files []schema.FileStatus
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			status := "modified"
			switch parts[0] {
			case "A":
				status = "added"
			case "M":
				status = "modified"
			case "D":
				status = "deleted"
			case "R":
				status = "renamed"
			case "C":
				status = "copied"
			}
			files = append(files, schema.FileStatus{
				Path:   parts[1],
				Status: status,
			})
		}
	}
	return files
}

// parsePorcelainStatus parses git status --porcelain output
func parsePorcelainStatus(output string) []schema.FileStatus {
	var files []schema.FileStatus
	// Don't trim the output - leading spaces are significant in porcelain format!
	// Just trim trailing newlines
	output = strings.TrimRight(output, "\n\r")
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		// Porcelain format: XY PATH where XY is 2 chars, then space, then path
		// Example: " M file.txt" or "M  file.txt" or "?? file.txt"
		indexStatus := line[0]
		workTreeStatus := line[1]
		// The path starts after "XY " (3 characters)
		path := line[3:]

		// Handle renamed files (format: R  old -> new)
		if strings.Contains(path, " -> ") {
			parts := strings.Split(path, " -> ")
			if len(parts) == 2 {
				path = parts[1]
			}
		}

		status := "modified"
		// Determine status based on index and work tree
		if indexStatus == '?' || workTreeStatus == '?' {
			status = "added" // Untracked = will be added
		} else if indexStatus == 'A' {
			status = "added"
		} else if indexStatus == 'D' || workTreeStatus == 'D' {
			status = "deleted"
		} else if indexStatus == 'R' {
			status = "renamed"
		} else if indexStatus == 'M' || workTreeStatus == 'M' {
			status = "modified"
		}

		files = append(files, schema.FileStatus{
			Path:   path,
			Status: status,
		})
	}
	return files
}

// matchGitAddPattern checks if a file path matches a git add pattern
func matchGitAddPattern(filePath, pattern, cwd string) bool {
	// Handle "." - matches everything
	if pattern == "." {
		return true
	}

	// Handle "-A" or "--all" - matches everything
	if pattern == "-A" || pattern == "--all" {
		return true
	}

	// Handle glob patterns
	if strings.Contains(pattern, "*") {
		matched, err := filepath.Match(pattern, filePath)
		if err == nil && matched {
			return true
		}
		// Also try matching just the filename
		matched, err = filepath.Match(pattern, filepath.Base(filePath))
		if err == nil && matched {
			return true
		}
	}

	// Handle directory patterns
	if strings.HasSuffix(pattern, "/") || !strings.Contains(pattern, ".") {
		// Might be a directory
		if strings.HasPrefix(filePath, pattern) || strings.HasPrefix(filePath, strings.TrimSuffix(pattern, "/")) {
			return true
		}
	}

	// Exact match
	if filePath == pattern || filepath.Base(filePath) == pattern {
		return true
	}

	// Path contains pattern
	if strings.Contains(filePath, pattern) {
		return true
	}

	return false
}

// parseCount parses a string to int, returning 0 on error
func parseCount(s string) (int, error) {
	var n int
	_, err := strings.NewReader(s).Read([]byte{})
	if err != nil {
		return 0, err
	}
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n, nil
}

// MockGitProvider provides predetermined values for testing
type MockGitProvider struct {
	Branch       string
	Author       string
	StagedFiles  []schema.FileStatus
	PendingFiles []schema.FileStatus
	Remote       string
	Ahead        int
	Behind       int
}

func (m *MockGitProvider) GetBranch(cwd string) string {
	return m.Branch
}

func (m *MockGitProvider) GetAuthor(cwd string) string {
	return m.Author
}

func (m *MockGitProvider) GetStagedFiles(cwd string) []schema.FileStatus {
	return m.StagedFiles
}

func (m *MockGitProvider) GetPendingFiles(cwd string, command string) []schema.FileStatus {
	return m.PendingFiles
}

func (m *MockGitProvider) GetRemote(cwd string) string {
	if m.Remote == "" {
		return "origin"
	}
	return m.Remote
}

func (m *MockGitProvider) GetAheadBehind(cwd string) (ahead, behind int) {
	return m.Ahead, m.Behind
}
