// Package env provides environment management for llima-box, including
// environment naming, namespace operations, and user account management.
//
// # Environment Naming
//
// Each isolated environment is uniquely identified by a name generated from
// the project path. The naming algorithm ensures:
//
//   - Valid Linux usernames (required for user account creation)
//   - Uniqueness across different project paths
//   - Deterministic mapping (same path always produces same name)
//   - Human-readable format for debugging
//
// Format: <sanitized-basename>-<hash>
//
// Example:
//
//	"/Users/alice/my-project"     -> "my-project-a1b2"
//	"/Users/alice/My Cool App"    -> "my-cool-app-c3d4"
//	"/Users/bob/my-project"       -> "my-project-i9j0"  (different hash)
//
// Sanitization Rules:
//
//   - Convert to lowercase
//   - Keep only ASCII letters, digits, hyphens, underscores
//   - Replace non-ASCII characters with hyphens
//   - Collapse consecutive hyphens
//   - Ensure starts with a letter (prepend "env-" if needed)
//   - Limit to 27 characters (leaves room for "-XXXX" hash suffix)
//
// Hash Generation:
//
//   - SHA1 hash of full absolute path
//   - First 4 hex characters (2 bytes)
//   - Provides 65,536 unique combinations
//   - Sufficient for local development use cases
//
// Linux Username Requirements:
//
//   - 1-32 characters
//   - Start with letter or underscore
//   - Contain only letters, digits, underscores, hyphens
//   - Conventionally lowercase
package env
