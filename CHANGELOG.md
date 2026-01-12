# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Colored, structured logging with INFO/SUCCESS/WARNING/ERROR levels
- Informational messages now output to stderr (allows stdout capture for data)
- Real-time streaming of Lima VM creation and startup output for better visibility
- Real-time streaming of namespace creation script output
- Debug logging for all VM and namespace operations with command execution details

### Changed

- Refactored namespace management to use direct `unshare`/`nsenter` commands instead of embedded shell scripts for better maintainability and debugging
- Simplified VM provisioning by removing unnecessary script generation, keeping only essential package installation and sudoers configuration
- Changed namespace PID file location from `/home/<env>/namespace.pid` to `/envs/<env>/namespace.pid` for cleaner organization

### Fixed

- JSON parsing error when `limactl list --json` returns a single instance object instead of an array
- Namespace creation failing - now stores PID and references namespace via /proc/<pid>/ns/mnt
- Shell command hanging after namespace creation - background process now properly detaches from SSH session
- Shell running as root instead of environment user - now properly switches to environment user
- Sudoers configuration using hardcoded 'lima' user instead of actual VM user
- Error propagation from background namespace process - now properly detects failures
- Usage/help text printing on every error - now shows only error messages
- Error reporting for namespace creation failures - now captures and displays actual command output
- Missing sudo permissions for sandbox.sh, pkill, nsenter, findmnt, mount, mountpoint, and su commands
- Lack of feedback during namespace verification - added debug logging
- Permission denied error when verifying namespace PID file - verification commands now use sudo
- VM provisioning hanging indefinitely - removed non-essential zsh and mise installation that blocked SSH startup
- Shell failing with "No such file or directory" - changed default shell from zsh to bash
- Command arguments incorrectly parsed as paths - fixed handling of `--` separator for commands like `llima-box shell -- bash`
- Interactive shell errors about terminal process group - removed PID namespace entry to avoid terminal control issues

## [0.3.0]

### Changed

- Refactored VM management to use `limactl` CLI instead of Lima Go library for better compatibility
- Simplified build process by removing CGO dependency, enabling fully static builds

### Fixed

- VM creation errors related to VZ driver and guest agent binary discovery
- Warning about non-existent `/tmp/lima` mount path

## [0.2.2]

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
