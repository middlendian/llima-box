package env

import (
	"crypto/sha1"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// GenerateName generates a valid environment name from a project path.
// The format is: <sanitized-basename>-<hash>
// Example: "/Users/alice/my-project" -> "my-project-a1b2"
func GenerateName(projectPath string) (string, error) {
	// Get absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Extract basename
	basename := filepath.Base(absPath)
	if basename == "." || basename == "/" {
		return "", fmt.Errorf("invalid project path: %s", projectPath)
	}

	// Sanitize basename
	sanitized := sanitizeBasename(basename)
	if sanitized == "" {
		return "", fmt.Errorf("project path basename '%s' results in empty name after sanitization", basename)
	}

	// Generate hash from full path
	hash := pathHash(absPath)

	// Combine
	name := fmt.Sprintf("%s-%s", sanitized, hash)

	// Validate result
	if !IsValidName(name) {
		return "", fmt.Errorf("generated name '%s' is not valid", name)
	}

	return name, nil
}

// sanitizeBasename converts a basename to a valid Linux username component.
// Rules:
// - Convert to lowercase
// - Replace spaces and invalid characters with hyphens
// - Remove consecutive hyphens
// - Ensure it starts with a letter
// - Trim to reasonable length
func sanitizeBasename(basename string) string {
	// Convert to lowercase
	s := strings.ToLower(basename)

	// Replace spaces and invalid characters with hyphens
	// Only keep ASCII letters, digits, hyphens, and underscores
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		} else if r == '-' || r == '_' {
			result.WriteRune(r)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			// Non-ASCII letter/digit: replace with hyphen
			result.WriteRune('-')
		} else {
			result.WriteRune('-')
		}
	}
	s = result.String()

	// Replace underscores with hyphens for consistency
	s = strings.ReplaceAll(s, "_", "-")

	// Remove consecutive hyphens
	re := regexp.MustCompile(`-+`)
	s = re.ReplaceAllString(s, "-")

	// Trim hyphens from start and end
	s = strings.Trim(s, "-")

	// If it doesn't start with a letter, prepend 'env'
	if len(s) > 0 && !unicode.IsLetter(rune(s[0])) {
		s = "env-" + s
	}

	// If still empty, use default
	if s == "" {
		s = "env"
	}

	// Limit length (leave room for -XXXX suffix)
	const maxBaseLength = 27 // 32 total - 5 for "-XXXX"
	if len(s) > maxBaseLength {
		s = s[:maxBaseLength]
		s = strings.TrimRight(s, "-")
	}

	return s
}

// pathHash generates a 4-character hash from a path for uniqueness.
func pathHash(path string) string {
	h := sha1.New()
	h.Write([]byte(path))
	sum := h.Sum(nil)
	// Take first 2 bytes (4 hex characters)
	return fmt.Sprintf("%x", sum[:2])
}

// IsValidName checks if a name meets Linux username requirements.
func IsValidName(name string) bool {
	// Check length
	if len(name) == 0 || len(name) > 32 {
		return false
	}

	// Must start with a letter or underscore
	firstChar := rune(name[0])
	if !unicode.IsLetter(firstChar) && firstChar != '_' {
		return false
	}

	// Can only contain letters, digits, underscores, and hyphens
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return false
		}
	}

	return true
}
