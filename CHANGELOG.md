# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `shell` command to enter isolated environments for projects
- `list` command to view all running environments
- `delete` command to remove individual environments
- `delete-all` command to remove all environments at once
- Interactive confirmation prompts for destructive operations
- Automatic VM creation and startup when needed
- Environment manager for creating, listing, and deleting isolated environments
- Persistent namespace support using Linux mount namespaces
- User account management for environment isolation
- Namespace entry functionality for running commands in isolated environments
- `make check-fast` target for running validation without network dependencies
- Makefile with build, test, lint, and format targets
- Multi-platform build support (Linux/macOS, ARM64/AMD64)
- GitHub Actions workflow for pull request validation with automatic code fixes
- GitHub Actions workflow for automated releases with changelog extraction
- Pull request template with CHANGELOG.md reminder
- CLAUDE.md with instructions for AI agents
- CHANGELOG.md following Keep a Changelog conventions
- golangci-lint configuration with 15+ linters
- Comprehensive build and development documentation in CONTRIBUTING.md

### Removed
- Test commands (`test-vm`, `test-naming`) - replaced by production CLI commands

## [0.1.0] - TBD

Initial development release (not yet published).

### Added
- Project structure and architecture
- Lima VM integration
- Multi-architecture support (x86_64 + ARM64)
- Environment naming system
- SSH client for VM communication
- Comprehensive documentation

[Unreleased]: https://github.com/middlendian/llima-box/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/middlendian/llima-box/releases/tag/v0.1.0
