# Instructions for AI Agents

This document provides guidance for AI agents (like Claude) working on this repository.

## Project Overview

llima-box is a Go project that creates secure, isolated environments for LLM agents using Lima VMs on macOS. The project uses Lima with Linux mount namespaces for filesystem isolation.

## Critical Rules

### 1. ALWAYS Update CHANGELOG.md

**For every pull request that makes code or feature changes, you MUST update CHANGELOG.md.**

- Add entries to the `[Unreleased]` section
- Use the appropriate category:
  - **Added**: New features
  - **Changed**: Changes to existing functionality
  - **Deprecated**: Soon-to-be removed features
  - **Removed**: Removed features
  - **Fixed**: Bug fixes
  - **Security**: Security fixes

**Important Guidelines:**

- **Only document changes from main branch's perspective**: Don't include incremental development fixes made during branch work
- **User-facing, high-level descriptions**: Write for users, not developers. Focus on what changed, not implementation details
- **Don't include**:
  - Fixes to bugs you introduced in the same branch
  - Internal refactoring unless it affects users
  - CI/workflow tweaks unless they affect contributor experience
  - Build process details (e.g., "fixed variable shadowing" or "updated import")

Example:
```markdown
## [Unreleased]

### Added
- New command for listing environments
- Support for custom VM configurations

### Fixed
- Race condition in VM status check (bug that existed in main)
```

**Exceptions:** Documentation-only changes or trivial typo fixes may skip the changelog.

### 2. ALWAYS Run Quality Checks Before Completion

**CRITICAL: Before you consider any work complete, you MUST run:**
```bash
GOPROXY=direct make check
```

**This is mandatory, not optional.** Never tell the user work is done without running checks first.

**Why use `GOPROXY=direct`?**
- Avoids network issues with the default Go proxy (proxy.golang.org)
- Fetches dependencies directly from source repositories
- More reliable in environments with network restrictions

**What `make check` does:**
1. Automatically formats code with `gofmt`
2. Updates `go.mod` and `go.sum` with `go mod tidy`
3. Runs `go vet` for static analysis
4. Runs `golangci-lint` with 15+ linters
5. Runs all tests with race detector

**Alternative if full checks fail:**
- Use `make check-fast` which runs fmt, vet, and test without network dependencies (skips linter)
- **NOTE:** `make check-fast` skips golangci-lint, so CI may still fail if there are linter errors
- Always prefer `GOPROXY=direct make check` over `make check-fast`

**Important:** `make check` applies automatic fixes (formatting, module tidying) before validation. This ensures consistency and reduces manual toil. If you push code without running `make check`, the CI will automatically apply these fixes and push them to your branch.

### 3. Follow Go Best Practices

- Keep functions small and focused
- Use clear, descriptive names
- Add comments for exported functions and types
- Handle errors with context: `fmt.Errorf("failed to X: %w", err)`
- Write table-driven tests

### 4. Makefile Usage

The project uses Make for build automation:
- `make help` - Show available targets (default)
- `GOPROXY=direct make check` - Run all validations (REQUIRED before completion)
- `make check-fast` - Run fast checks without network (fmt, vet, test - skips linter)
- `make build` - Build the binary
- `make test` - Run tests
- `make fmt` - Format code
- `make lint` - Run linters

**Important:** Always use `GOPROXY=direct` when running `make check` to avoid network issues with the default Go proxy.

## Project Structure

```
llima-box/
├── cmd/
│   └── llima-box/      # Main application entry point
├── pkg/
│   ├── env/            # Environment naming and sanitization
│   ├── ssh/            # SSH client for VM communication
│   └── vm/             # VM lifecycle management
├── docs/               # Documentation
├── .github/
│   └── workflows/      # CI/CD workflows
├── Makefile            # Build automation
├── .golangci.yml       # Linter configuration
└── CHANGELOG.md        # Keep a Changelog format
```

## Development Workflow

### Making Changes

1. **Read existing code first** - Never propose changes to code you haven't read
2. **Update CHANGELOG.md** - Add your changes to `[Unreleased]`
3. **Write tests** - All new code needs tests
4. **Run checks** - `GOPROXY=direct make check` must pass BEFORE considering work complete
5. **Commit with clear messages** - Use imperative mood ("Add feature" not "Added feature")

**CRITICAL:** Step 4 is mandatory. Never commit, push, or tell the user work is complete without running validation checks first. Always use `GOPROXY=direct make check` to avoid network issues.

### Commit Message Format

```
Short (50 chars or less) summary

More detailed explanatory text, if necessary. Wrap it to about 72
characters. The blank line separating the summary from the body is
critical.

- Bullet points are okay
- Reference issues: "Fixes #123"
```

### Creating Pull Requests

The PR template includes a prominent CHANGELOG.md checkbox. The automated workflow will warn if CHANGELOG.md wasn't updated.

## CI/CD

### Pull Requests
- Automated checks for formatting, linting, and tests
- Builds for all platforms (Linux/macOS, ARM64/AMD64)
- Warning if CHANGELOG.md wasn't updated
- All checks must pass

### Releases
- Triggered by version tags (e.g., `v1.0.0`)
- Extracts release notes from CHANGELOG.md
- Builds multi-arch binaries
- Publishes to GitHub Releases
- Updates Go proxy

## Code Quality Standards

### Linting
The project uses golangci-lint with these linters enabled:
- errcheck, gosimple, govet, staticcheck
- gofmt, goimports, misspell
- gosec (security), gocritic, revive
- bodyclose, unconvert, unparam

### Testing
- Write table-driven tests for multiple scenarios
- Include edge cases and error conditions
- Aim for 80%+ coverage on new code
- Use race detector: `go test -race ./...`

### Security
- Avoid command injection, XSS, SQL injection
- Validate at system boundaries only
- Don't add unnecessary error handling
- Trust internal code and framework guarantees

## Anti-Patterns to Avoid

### Over-Engineering
- Don't add features beyond what was asked
- Don't refactor surrounding code unnecessarily
- Don't add docstrings to code you didn't change
- Don't add error handling for impossible scenarios
- Don't create abstractions for one-time operations

### Breaking Changes
- Avoid backwards-compatibility hacks
- If something is unused, delete it completely
- Don't rename unused variables or add `// removed` comments

## Common Tasks

### Building for Multiple Platforms
```bash
make build-all
```
This builds:
- Linux AMD64 and ARM64
- macOS AMD64 and ARM64

### Running Tests with Coverage
```bash
make coverage
# Opens coverage.html in browser
```

### Formatting Code
```bash
make fmt
```

### Full Validation
```bash
GOPROXY=direct make check
```

## Release Process

For maintainers creating releases:

1. **Update CHANGELOG.md**:
   ```markdown
   ## [Unreleased]

   ## [1.0.0] - 2024-01-09

   ### Added
   - Your features here
   ```

2. **Commit and push**:
   ```bash
   git add CHANGELOG.md
   git commit -m "Prepare release v1.0.0"
   git push origin main
   ```

3. **Create and push tag**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

4. GitHub Actions automatically:
   - Runs all tests
   - Builds all platform binaries
   - Extracts changelog section for this version
   - Creates GitHub release with binaries

## Documentation

- **README.md** - User-facing documentation
- **CONTRIBUTING.md** - Contributor guidelines
- **CHANGELOG.md** - Version history (Keep a Changelog format)
- **docs/** - Architecture, design, and implementation docs
- **CLAUDE.md** - This file (AI agent instructions)

## Key Files

- **go.mod** - Go 1.24.7, uses Lima VM library
- **Makefile** - Build automation, default target is `help`
- **.golangci.yml** - Linter configuration
- **.github/workflows/pr.yml** - PR validation workflow
- **.github/workflows/release.yml** - Release automation
- **.github/pull_request_template.md** - PR template with changelog reminder

## Questions?

- Check existing documentation in `docs/`
- Review `CONTRIBUTING.md` for detailed guidelines
- Look at existing code for examples
- The codebase follows standard Go conventions

## Remember

**The most important rule: Update CHANGELOG.md with every meaningful change!**

This ensures proper release notes and keeps the project history clear.
