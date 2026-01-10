# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.1]

### Changed

- Release process now uses GoReleaser for automated builds and releases

## [0.2.0]

**⚠️ Beta Release** - This is an early release for testing and feedback. Not recommended for production use.

This is the first public release of llima-box. All core functionality is implemented and unit-tested. We're releasing as
a beta to gather real-world feedback before committing to v1.0.0.

**Please report issues at**: https://github.com/middlendian/llima-box/issues

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
- VM lifecycle management (create, start, stop, delete)
- SSH client with retry logic and SSH agent forwarding for Git operations
- Environment naming system
- Multi-platform build support (Linux/macOS, ARM64/AMD64)
- `make check-fast` target for running validation without network dependencies
- Makefile with build, test, lint, and format targets
- GitHub Actions workflow for pull request validation with automatic code fixes
- GitHub Actions workflow for automated releases with changelog extraction
- CLAUDE.md with instructions for AI agents
- CHANGELOG.md following Keep a Changelog conventions
- golangci-lint configuration with 15+ linters
- Release archives use mise-compatible structure with `bin/` subdirectory
- Archives use mise-preferred naming: `macos`/`linux` for OS and `x64`/`arm64` for architecture
- Archives include LICENSE, README.md, and CHANGELOG.md files
- Release notes recommend mise installation method

### Known Limitations

These are design choices or planned improvements for future releases:

- **macOS only**: Requires Lima VM (macOS-specific tool)
- **Limited real-world testing**: This is a beta release seeking feedback
- **No resource quotas**: CPU/memory shared across all environments
- **Shared network**: All environments share the VM's network namespace
- **Manual Lima installation**: Users must install Lima separately via Homebrew

[Unreleased]: https://github.com/middlendian/llima-box/compare/v0.2.0...HEAD

[0.2.0]: https://github.com/middlendian/llima-box/releases/tag/v0.2.0
