package env

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateName(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantPrefix  string // Expected prefix (before hash)
		wantErr     bool
		description string
	}{
		{
			name:        "simple path",
			path:        "/Users/alice/my-project",
			wantPrefix:  "my-project",
			wantErr:     false,
			description: "Basic project path",
		},
		{
			name:        "path with spaces",
			path:        "/Users/alice/My Cool App",
			wantPrefix:  "my-cool-app",
			wantErr:     false,
			description: "Spaces converted to hyphens",
		},
		{
			name:        "path with special chars",
			path:        "/Users/alice/project@2024!",
			wantPrefix:  "project-2024",
			wantErr:     false,
			description: "Special characters converted to hyphens",
		},
		{
			name:        "path starting with number",
			path:        "/Users/alice/123-project",
			wantPrefix:  "env-123-project",
			wantErr:     false,
			description: "Number prefix gets 'env-' prepended",
		},
		{
			name:        "path with unicode",
			path:        "/Users/alice/project-α-β",
			wantPrefix:  "project",
			wantErr:     false,
			description: "Unicode characters handled",
		},
		{
			name:        "very long name",
			path:        "/Users/alice/" + strings.Repeat("a", 100),
			wantPrefix:  strings.Repeat("a", 27),
			wantErr:     false,
			description: "Long names truncated",
		},
		{
			name:        "path with underscores",
			path:        "/Users/alice/my_project_name",
			wantPrefix:  "my-project-name",
			wantErr:     false,
			description: "Underscores converted to hyphens",
		},
		{
			name:        "path with consecutive hyphens",
			path:        "/Users/alice/my---project",
			wantPrefix:  "my-project",
			wantErr:     false,
			description: "Consecutive hyphens collapsed",
		},
		{
			name:        "path ending with hyphen",
			path:        "/Users/alice/my-project-",
			wantPrefix:  "my-project",
			wantErr:     false,
			description: "Trailing hyphens removed",
		},
		// Note: "." gets converted to absolute path by filepath.Abs,
		// so it's actually valid (will be the basename of cwd)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateName(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Check prefix
			parts := strings.Split(got, "-")
			if len(parts) < 2 {
				t.Errorf("GenerateName() = %v, expected format '<name>-<hash>'", got)
				return
			}

			// Get prefix (everything except last part which is hash)
			prefix := strings.Join(parts[:len(parts)-1], "-")
			hash := parts[len(parts)-1]

			if prefix != tt.wantPrefix {
				t.Errorf("GenerateName() prefix = %v, want %v", prefix, tt.wantPrefix)
			}

			// Check hash is 4 characters
			if len(hash) != 4 {
				t.Errorf("GenerateName() hash length = %v, want 4", len(hash))
			}

			// Check result is valid
			if !IsValidName(got) {
				t.Errorf("GenerateName() = %v, is not a valid name", got)
			}

			// Check length constraint
			if len(got) > 32 {
				t.Errorf("GenerateName() length = %v, exceeds 32 characters", len(got))
			}
		})
	}
}

func TestGenerateNameUniqueness(t *testing.T) {
	// Same basename, different paths should generate different names
	path1 := "/Users/alice/my-project"
	path2 := "/Users/bob/my-project"
	path3 := "/Users/alice/projects/my-project"

	name1, err1 := GenerateName(path1)
	name2, err2 := GenerateName(path2)
	name3, err3 := GenerateName(path3)

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("GenerateName() errors: %v, %v, %v", err1, err2, err3)
	}

	// All should have same prefix
	prefix1 := strings.TrimSuffix(name1, name1[len(name1)-5:])
	prefix2 := strings.TrimSuffix(name2, name2[len(name2)-5:])
	prefix3 := strings.TrimSuffix(name3, name3[len(name3)-5:])

	if prefix1 != prefix2 || prefix2 != prefix3 {
		t.Errorf("Prefixes should match: %v, %v, %v", prefix1, prefix2, prefix3)
	}

	// But hashes should differ
	if name1 == name2 || name2 == name3 || name1 == name3 {
		t.Errorf("Names should be unique: %v, %v, %v", name1, name2, name3)
	}
}

func TestGenerateNameDeterminism(t *testing.T) {
	// Same path should always generate same name
	path := "/Users/alice/my-project"

	name1, _ := GenerateName(path)
	name2, _ := GenerateName(path)
	name3, _ := GenerateName(path)

	if name1 != name2 || name2 != name3 {
		t.Errorf("Same path should generate same name: %v, %v, %v", name1, name2, name3)
	}
}

func TestSanitizeBasename(t *testing.T) {
	tests := []struct {
		name     string
		basename string
		want     string
	}{
		{"simple", "my-project", "my-project"},
		{"uppercase", "MyProject", "myproject"},
		{"spaces", "My Project", "my-project"},
		{"special chars", "project@2024!", "project-2024"},
		{"underscores", "my_project_name", "my-project-name"},
		{"consecutive hyphens", "my---project", "my-project"},
		{"leading hyphen", "-my-project", "my-project"},
		{"trailing hyphen", "my-project-", "my-project"},
		{"starts with number", "123project", "env-123project"},
		{"only numbers", "123456", "env-123456"},
		{"unicode", "project-α-β-γ", "project"},
		{"mixed unicode", "café-project", "caf-project"},
		{"empty after sanitize", "@@@", "env"},
		{"very long", strings.Repeat("a", 50), strings.Repeat("a", 27)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeBasename(tt.basename)
			if got != tt.want {
				t.Errorf("sanitizeBasename(%q) = %q, want %q", tt.basename, got, tt.want)
			}
		})
	}
}

func TestPathHash(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"simple", "/Users/alice/project"},
		{"long", "/Users/alice/very/long/path/to/project"},
		{"unicode", "/Users/alice/проект"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := pathHash(tt.path)
			if len(hash) != 4 {
				t.Errorf("pathHash() length = %v, want 4", len(hash))
			}
			// Check it's all hex
			for _, r := range hash {
				if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
					t.Errorf("pathHash() = %v, contains non-hex character", hash)
				}
			}
		})
	}
}

func TestPathHashUniqueness(t *testing.T) {
	paths := []string{
		"/Users/alice/project",
		"/Users/bob/project",
		"/Users/alice/other-project",
		"/home/alice/project",
	}

	hashes := make(map[string]string)
	for _, path := range paths {
		hash := pathHash(path)
		if existingPath, exists := hashes[hash]; exists {
			t.Errorf("Hash collision: %s and %s both produce %s", path, existingPath, hash)
		}
		hashes[hash] = path
	}
}

func TestPathHashDeterminism(t *testing.T) {
	path := "/Users/alice/project"
	hash1 := pathHash(path)
	hash2 := pathHash(path)
	hash3 := pathHash(path)

	if hash1 != hash2 || hash2 != hash3 {
		t.Errorf("pathHash() not deterministic: %s, %s, %s", hash1, hash2, hash3)
	}
}

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid simple", "my-project", true},
		{"valid with underscore", "my_project", true},
		{"valid with numbers", "project123", true},
		{"valid starts with underscore", "_project", true},
		{"valid max length", strings.Repeat("a", 32), true},
		{"invalid empty", "", false},
		{"invalid too long", strings.Repeat("a", 33), false},
		{"invalid starts with hyphen", "-project", false},
		{"invalid starts with number", "1project", false},
		{"invalid special char", "project@123", false},
		{"invalid space", "my project", false},
		{"invalid dot", "my.project", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidName(tt.input)
			if got != tt.want {
				t.Errorf("IsValidName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateNameRealWorldExamples(t *testing.T) {
	// Test with realistic project paths
	tests := []struct {
		path       string
		wantPrefix string
	}{
		{"/Users/alice/Documents/my-ai-agent", "my-ai-agent"},
		{"/Users/bob/Projects/Web App 2024", "web-app-2024"},
		{"/home/charlie/dev/project_name", "project-name"},
		{"/Users/dave/Desktop/NEW PROJECT!!!", "new-project"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			// Convert to absolute path if not already
			absPath, err := filepath.Abs(tt.path)
			if err != nil {
				absPath = tt.path
			}

			got, err := GenerateName(absPath)
			if err != nil {
				t.Errorf("GenerateName() error = %v", err)
				return
			}

			// Extract prefix
			parts := strings.Split(got, "-")
			prefix := strings.Join(parts[:len(parts)-1], "-")

			if prefix != tt.wantPrefix {
				t.Errorf("GenerateName() prefix = %v, want %v (full name: %v)", prefix, tt.wantPrefix, got)
			}

			// Ensure valid
			if !IsValidName(got) {
				t.Errorf("GenerateName() = %v is invalid", got)
			}
		})
	}
}
