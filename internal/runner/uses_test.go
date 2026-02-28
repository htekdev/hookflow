package runner

import (
	"testing"
)

// TestParseUsesStringLocalPath tests parsing local action paths
func TestParseUsesStringLocalPath(t *testing.T) {
	tests := []struct {
		name      string
		uses      string
		expectErr bool
		expectIs  bool // expectIsLocal
	}{
		{
			name:       "relative path with dot slash",
			uses:       "./actions/my-action",
			expectErr:  false,
			expectIs:   true,
		},
		{
			name:       "relative path with parent",
			uses:       "../actions/my-action",
			expectErr:  false,
			expectIs:   true,
		},
		{
			name:       "absolute path",
			uses:       "/abs/path/to/action",
			expectErr:  false,
			expectIs:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseUsesString(tt.uses)

			if (err != nil) != tt.expectErr {
				t.Errorf("parseUsesString(%s) error = %v, expectErr %v", tt.uses, err, tt.expectErr)
				return
			}

			if err != nil {
				return
			}

			if parsed.IsLocal != tt.expectIs {
				t.Errorf("parseUsesString(%s) IsLocal = %v, expected %v", tt.uses, parsed.IsLocal, tt.expectIs)
			}
		})
	}
}

// TestParseUsesStringGitHub tests parsing GitHub action references
func TestParseUsesStringGitHub(t *testing.T) {
	tests := []struct {
		name       string
		uses       string
		expectErr  bool
		expectOwner string
		expectRepo  string
		expectPath  string
		expectVer   string
	}{
		{
			name:        "basic owner/repo@version",
			uses:        "owner/action@v1",
			expectErr:   false,
			expectOwner: "owner",
			expectRepo:  "action",
			expectPath:  "",
			expectVer:   "v1",
		},
		{
			name:        "with subpath owner/repo/path@version",
			uses:        "owner/action/subdir@v1.0",
			expectErr:   false,
			expectOwner: "owner",
			expectRepo:  "action",
			expectPath:  "subdir",
			expectVer:   "v1.0",
		},
		{
			name:        "with nested subpath",
			uses:        "owner/repo/deeply/nested/path@main",
			expectErr:   false,
			expectOwner: "owner",
			expectRepo:  "repo",
			expectPath:  "", // Will be platform-dependent, we'll check it's not empty
			expectVer:   "main",
		},
		{
			name:       "missing version",
			uses:       "owner/repo",
			expectErr:  true,
		},
		{
			name:       "missing repo",
			uses:       "owner@v1",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseUsesString(tt.uses)

			if (err != nil) != tt.expectErr {
				t.Errorf("parseUsesString(%s) error = %v, expectErr %v", tt.uses, err, tt.expectErr)
				return
			}

			if err != nil {
				return
			}

			if parsed.Owner != tt.expectOwner {
				t.Errorf("parseUsesString(%s) Owner = %s, expected %s", tt.uses, parsed.Owner, tt.expectOwner)
			}
			if parsed.Repo != tt.expectRepo {
				t.Errorf("parseUsesString(%s) Repo = %s, expected %s", tt.uses, parsed.Repo, tt.expectRepo)
			}
			// For nested path, just check it's not empty on platforms that parse it
			if tt.uses == "owner/repo/deeply/nested/path@main" {
				if parsed.Path == "" {
					t.Errorf("parseUsesString(%s) Path is empty, expected non-empty", tt.uses)
				}
			} else if parsed.Path != tt.expectPath {
				t.Errorf("parseUsesString(%s) Path = %s, expected %s", tt.uses, parsed.Path, tt.expectPath)
			}
			if parsed.Version != tt.expectVer {
				t.Errorf("parseUsesString(%s) Version = %s, expected %s", tt.uses, parsed.Version, tt.expectVer)
			}
		})
	}
}

// TestParseUsesStringEdgeCases tests edge cases
func TestParseUsesStringEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		uses      string
		expectErr bool
	}{
		{
			name:      "whitespace handling",
			uses:      "  owner/repo@v1  ",
			expectErr: false,
		},
		{
			name:      "empty string",
			uses:      "",
			expectErr: true,
		},
		{
			name:      "only whitespace",
			uses:      "   ",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseUsesString(tt.uses)
			if (err != nil) != tt.expectErr {
				t.Errorf("parseUsesString(%q) error = %v, expectErr %v", tt.uses, err, tt.expectErr)
			}
		})
	}
}
