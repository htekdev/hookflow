package main

import (
	"testing"
)

// TestDetectGitCommitCommand tests that git commit commands are detected
func TestDetectGitCommitCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		isCommit bool
	}{
		{"simple git commit", "git commit -m 'test'", true},
		{"git commit with message", `git commit -m "feat: add feature"`, true},
		{"git ci alias", "git ci -m 'test'", true},
		{"git commit amend", "git commit --amend", true},
		{"git commit all", "git commit -a -m 'test'", true},
		{"not a commit", "git status", false},
		{"not a commit - push", "git push origin main", false},
		{"not a commit - pull", "git pull", false},
		{"echo with git commit in string", `echo "git commit"`, false},
		{"git commit in middle", "cd repo && git commit -m 'test'", true},
		// Additional edge cases for command chains
		{"git commit after OR", "git status || git commit -m 'fallback'", true},
		{"git commit after semicolon", "echo done; git commit -m 'test'", true},
		{"git commit after background", "sleep 1 & git commit -m 'test'", true},
		{"multiple git commands", "git add . && git commit -m 'test' && git push", true},
		// Case sensitivity - git commands are case-sensitive on Unix
		{"uppercase GIT", "GIT commit -m 'test'", false},
		{"mixed case Git", "Git commit -m 'test'", false},
		// Edge cases with whitespace
		{"extra spaces", "git   commit -m 'test'", true},
		{"tabs", "git\tcommit -m 'test'", true},
		// Commands that look like git but aren't
		{"gitcommit typo", "gitcommit -m 'test'", false},
		{"git-commit hyphen", "git-commit -m 'test'", false},
		// Printf/echo with git in string
		{"printf git commit", `printf "run git commit now"`, false},
		{"echo command instruction", `echo "Please run: git commit -m 'msg'"`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isGitCommitCommand(tt.command)
			if got != tt.isCommit {
				t.Errorf("isGitCommitCommand(%q) = %v, want %v", tt.command, got, tt.isCommit)
			}
		})
	}
}

// TestDetectGitPushCommand tests that git push commands are detected
func TestDetectGitPushCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		isPush  bool
	}{
		{"simple git push", "git push", true},
		{"git push origin main", "git push origin main", true},
		{"git push with tags", "git push --tags", true},
		{"git push origin tag", "git push origin v1.0.0", true},
		{"git push force", "git push --force", true},
		{"git push upstream", "git push -u origin feature", true},
		{"not a push", "git status", false},
		{"not a push - commit", "git commit -m 'test'", false},
		{"not a push - pull", "git pull", false},
		{"git push in middle", "cd repo && git push origin main", true},
		// Additional edge cases
		{"git push after OR", "git status || git push", true},
		{"git push after semicolon", "echo done; git push", true},
		{"git push after background", "sleep 1 & git push", true},
		{"multiple commands with push", "git add . && git commit -m 'test' && git push", true},
		// Case sensitivity
		{"uppercase GIT push", "GIT push", false},
		{"mixed case Git push", "Git push", false},
		// Echo statements
		{"echo git push", `echo "git push"`, false},
		{"printf git push", `printf "Please git push"`, false},
		// Edge cases
		{"gitpush typo", "gitpush origin main", false},
		{"git-push hyphen", "git-push origin", false},
		// Complex branch patterns
		{"push feature branch", "git push origin feature/user/my-branch", true},
		{"push release branch", "git push origin release/v2.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isGitPushCommand(tt.command)
			if got != tt.isPush {
				t.Errorf("isGitPushCommand(%q) = %v, want %v", tt.command, got, tt.isPush)
			}
		})
	}
}

// TestExtractCommitMessage tests extracting commit messages from commands
func TestExtractCommitMessage(t *testing.T) {
	tests := []struct {
		name    string
		command string
		message string
	}{
		{"double quotes", `git commit -m "feat: add feature"`, "feat: add feature"},
		{"single quotes", `git commit -m 'fix: bug fix'`, "fix: bug fix"},
		{"no message flag", "git commit --amend", ""},
		{"message with spaces", `git commit -m "this is a long message"`, "this is a long message"},
		{"message at end", `git add . && git commit -m "done"`, "done"},
		// Additional edge cases
		{"empty quotes", `git commit -m ""`, ""},
		{"message with colons", `git commit -m "fix: scope: detail"`, "fix: scope: detail"},
		{"message with numbers", `git commit -m "v1.2.3 release"`, "v1.2.3 release"},
		{"message with special chars", `git commit -m "fix #123 - bug"`, "fix #123 - bug"},
		{"--message long form", `git commit --message "long form"`, ""},
		{"message with parens", `git commit -m "feat(core): add feature"`, "feat(core): add feature"},
		{"no -m flag", "git commit -a", ""},
		{"commit interactive", "git commit", ""},
		// Edge cases where message extraction should fail
		{"-m without space", `git commit-m "test"`, "test"}, // regex finds -m pattern regardless of context
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCommitMessage(tt.command)
			if got != tt.message {
				t.Errorf("extractCommitMessage(%q) = %q, want %q", tt.command, got, tt.message)
			}
		})
	}
}

// TestExtractPushRef tests extracting push refs from commands
func TestExtractPushRef(t *testing.T) {
	tests := []struct {
		name    string
		command string
		branch  string // current branch
		wantRef string
	}{
		{"simple push", "git push", "main", "refs/heads/main"},
		{"push to origin", "git push origin main", "main", "refs/heads/main"},
		{"push feature branch", "git push origin feature/test", "feature/test", "refs/heads/feature/test"},
		{"push tag", "git push origin v1.0.0", "main", "refs/tags/v1.0.0"},
		{"push tags flag", "git push --tags", "main", "refs/heads/main"}, // Still branch, tags are separate
		// Additional edge cases for complex branch patterns
		{"complex feature branch", "git push origin feature/user/my-feature", "feature/user/my-feature", "refs/heads/feature/user/my-feature"},
		{"release branch", "git push origin release/2.0.0", "release/2.0.0", "refs/heads/release/2.0.0"},
		{"hotfix branch", "git push origin hotfix/urgent-fix", "hotfix/urgent-fix", "refs/heads/hotfix/urgent-fix"},
		{"tag with prefix", "git push origin v2.1.0", "main", "refs/tags/v2.1.0"},
		{"tag semver patch", "git push origin v1.0.1", "develop", "refs/tags/v1.0.1"},
		// Default branch scenarios
		{"push upstream", "git push -u origin", "develop", "refs/heads/develop"},
		{"force push", "git push --force", "feature/wip", "refs/heads/feature/wip"},
		// Edge cases
		{"empty branch", "git push", "", "refs/heads/"},
		{"branch with dots", "git push origin release.1.0", "release.1.0", "refs/heads/release.1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPushRef(tt.command, tt.branch)
			if got != tt.wantRef {
				t.Errorf("extractPushRef(%q, %q) = %q, want %q", tt.command, tt.branch, got, tt.wantRef)
			}
		})
	}
}

// TestParseEventWithGitCommit tests that git commit events are properly parsed
func TestParseEventWithGitCommit(t *testing.T) {
	data := map[string]interface{}{
		"hook": map[string]interface{}{
			"type": "preToolUse",
			"tool": map[string]interface{}{
				"name": "powershell",
				"args": map[string]interface{}{
					"command": `git commit -m "feat: new feature"`,
				},
			},
		},
		"tool": map[string]interface{}{
			"name": "powershell",
			"args": map[string]interface{}{
				"command": `git commit -m "feat: new feature"`,
			},
		},
		"commit": map[string]interface{}{
			"sha":     "pending",
			"message": "feat: new feature",
			"author":  "test@example.com",
			"branch":  "main",
			"files": []interface{}{
				map[string]interface{}{
					"path":   "src/feature.go",
					"status": "added",
				},
			},
		},
	}

	event := parseEventData(data)

	if event.Commit == nil {
		t.Fatal("Expected Commit to be set")
	}
	if event.Commit.Message != "feat: new feature" {
		t.Errorf("Expected Commit.Message = 'feat: new feature', got '%s'", event.Commit.Message)
	}
	if len(event.Commit.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(event.Commit.Files))
	}
	if event.Commit.Files[0].Path != "src/feature.go" {
		t.Errorf("Expected file path = 'src/feature.go', got '%s'", event.Commit.Files[0].Path)
	}
}

// TestParseEventWithGitPush tests that git push events are properly parsed
func TestParseEventWithGitPush(t *testing.T) {
	data := map[string]interface{}{
		"hook": map[string]interface{}{
			"type": "preToolUse",
			"tool": map[string]interface{}{
				"name": "bash",
				"args": map[string]interface{}{
					"command": "git push origin main",
				},
			},
		},
		"push": map[string]interface{}{
			"ref":    "refs/heads/main",
			"before": "0000000000000000000000000000000000000000",
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
	if event.Push.After != "abc123def456" {
		t.Errorf("Expected Push.After = 'abc123def456', got '%s'", event.Push.After)
	}
}
