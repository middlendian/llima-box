# Contributing to llima-box

Thank you for your interest in contributing to llima-box! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)

## Code of Conduct

Be respectful, constructive, and professional in all interactions. We're building something useful together.

## Getting Started

### Prerequisites

- Go 1.24.7 or later
- Lima installed: `brew install lima`
- golangci-lint: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- Git

### Setting Up Your Development Environment

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/llima-box.git
   cd llima-box
   ```
3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/middlendian/llima-box.git
   ```
4. Install dependencies:
   ```bash
   make deps
   ```

## Development Workflow

### Creating a Feature Branch

Always create a new branch for your work:

```bash
git checkout -b feature/your-feature-name
```

Branch naming conventions:
- Features: `feature/description`
- Bug fixes: `fix/description`
- Documentation: `docs/description`
- Refactoring: `refactor/description`

### Building and Testing

```bash
# Build the project
make build

# Run tests
make test

# Run tests with coverage
make coverage

# Run all quality checks
make check
```

### Project Structure

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
└── .golangci.yml       # Linter configuration
```

### Continuous Integration

The project uses GitHub Actions for CI/CD:

- **Pull Requests**: Automatically run tests, linting, and build checks on all PRs
- **Releases**: Automatically build and publish binaries for all platforms when a version tag is pushed

Release notes are extracted from [CHANGELOG.md](CHANGELOG.md), which
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) conventions.

To create a new release:

```bash
# 1. Update CHANGELOG.md with version and date
# 2. Commit the changelog
git add CHANGELOG.md
git commit -m "Prepare release v1.0.0"
git push origin main

# 3. Create and push the tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Code Quality

Before committing, run all quality checks:

```bash
# Run all checks (applies formatting/tidying, then validates)
make check
```

**What `make check` does:**
1. Automatically formats code with `gofmt`
2. Updates `go.mod` and `go.sum` with `go mod tidy`
3. Runs `go vet` for static analysis
4. Runs `golangci-lint` with 15+ linters
5. Runs all tests with race detector

If `make check` passes locally, it will pass in CI. If you forget to run it, the CI will automatically apply formatting and module fixes, then commit and push them to your PR branch.

## Coding Standards

### Go Style Guide

- Follow the [official Go style guide](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting (automatically done with `make fmt`)
- Keep functions small and focused
- Write clear, descriptive names
- Add comments for exported functions and types

### Code Organization

```
pkg/
├── env/        # Environment naming and sanitization
├── ssh/        # SSH client for VM communication
└── vm/         # VM lifecycle management
```

- Keep packages focused on a single responsibility
- Use internal packages for implementation details
- Export only what's necessary

### Error Handling

```go
// Good: Clear error messages with context
if err != nil {
    return fmt.Errorf("failed to create VM: %w", err)
}

// Bad: Generic errors without context
if err != nil {
    return err
}
```

### Documentation

- Add package documentation in `doc.go` files
- Document all exported functions, types, and constants
- Include examples for non-trivial functions
- Keep README.md up to date

## Testing

### Writing Tests

- Write tests for all new code
- Place tests in `*_test.go` files alongside the code
- Use table-driven tests for multiple scenarios
- Include edge cases and error conditions

Example:
```go
func TestGenerateName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid path",
            input: "/Users/test/project",
            want:  "project-abc123",
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := GenerateName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("GenerateName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("GenerateName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific test
go test -v ./pkg/env -run TestGenerateName

# Run with race detector
go test -race ./...
```

### Test Coverage

- Aim for 80%+ coverage on new code
- Focus on meaningful tests, not just coverage numbers
- Test error paths and edge cases

## Submitting Changes

### Commit Messages

Write clear, descriptive commit messages:

```
Short (50 chars or less) summary

More detailed explanatory text, if necessary. Wrap it to about 72
characters. The blank line separating the summary from the body is
critical.

- Bullet points are okay
- Use imperative mood ("Add feature" not "Added feature")
- Reference issues: "Fixes #123"
```

Good examples:
- `Add SSH connection retry logic with exponential backoff`
- `Fix race condition in VM status check`
- `Update documentation for environment naming`

### Maintaining the Changelog

We follow [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) conventions. When making changes:

1. **Update CHANGELOG.md** - Add your changes to the `[Unreleased]` section
2. **Use standard categories**:
   - **Added** - New features
   - **Changed** - Changes to existing functionality
   - **Deprecated** - Soon-to-be removed features
   - **Removed** - Removed features
   - **Fixed** - Bug fixes
   - **Security** - Security fixes

Example:
```markdown
## [Unreleased]

### Added
- New command for listing environments
- Support for custom VM configurations

### Fixed
- Race condition in VM status check
```

When a release is created, maintainers will move the unreleased changes to a new version section.

### Pull Request Process

1. Update your branch with latest upstream:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. Ensure all checks pass:
   ```bash
   make check
   ```

3. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

4. Create a pull request on GitHub with:
   - Clear title and description
   - Reference to related issues
   - Screenshots/examples if applicable
   - Notes about testing performed

5. Address review feedback:
   - Make changes in new commits
   - Push updates to the same branch
   - Respond to reviewer comments

6. Once approved, a maintainer will merge your PR

### Pull Request Checklist

- [ ] Code follows project style guidelines
- [ ] All tests pass (`make test`)
- [ ] Linters pass (`make lint`)
- [ ] Code is formatted (`make fmt`)
- [ ] New code has tests
- [ ] Documentation is updated
- [ ] CHANGELOG.md is updated (add to [Unreleased] section)
- [ ] Commit messages are clear
- [ ] Branch is up to date with main

## Release Process

Releases are automated via GitHub Actions and use [CHANGELOG.md](../CHANGELOG.md) for release notes.

### Creating a Release

1. **Update CHANGELOG.md** - Move unreleased changes to a new version section:
   ```markdown
   ## [Unreleased]

   (empty or new unreleased changes)

   ## [1.0.0] - 2024-01-09

   ### Added
   - Your features here
   ```

2. **Commit the changelog update**:
   ```bash
   git add CHANGELOG.md
   git commit -m "Prepare release v1.0.0"
   git push origin main
   ```

3. **Create and push a version tag**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

4. **GitHub Actions automatically**:
   - Runs all tests
   - Builds binaries for all platforms (Linux/macOS, ARM64/AMD64)
   - Creates release archives
   - Generates checksums
   - Extracts release notes from CHANGELOG.md
   - Creates GitHub release with notes and binaries
   - Publishes to Go proxy

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version: Breaking changes
- **MINOR** version: New features, backwards compatible
- **PATCH** version: Bug fixes, backwards compatible

## Getting Help

- Check the [documentation](docs/)
- Read existing issues and pull requests
- Ask questions by creating a new issue with the "question" label

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [GitHub Flow](https://guides.github.com/introduction/flow/)

Thank you for contributing to llima-box!
