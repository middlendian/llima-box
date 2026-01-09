# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Makefile with build, test, lint, and format targets
- Multi-platform build support (Linux/macOS, ARM64/AMD64)
- GitHub Actions workflow for pull request validation
- GitHub Actions workflow for automated releases
- GitHub Actions check to remind contributors to update CHANGELOG.md
- Pull request template with CHANGELOG.md reminder
- CLAUDE.md with instructions for AI agents
- CHANGELOG.md following Keep a Changelog conventions
- golangci-lint configuration with 15+ linters
- Comprehensive build and development documentation
- CONTRIBUTING.md with contributor guidelines
- VM lifecycle management implementation
- SSH client with retry logic and agent forwarding
- Environment naming and sanitization with tests

### Changed
- Made `help` the default Makefile target for better UX
- Release notes now extracted from CHANGELOG.md instead of git history
- Simplified CI workflow to use `make check` instead of separate lint/vet jobs
- Test target now generates coverage report for CI
- Build target now automatically applies formatting and module tidying
- Check target depends on build to ensure fixes are applied before validation
- CI automatically commits and pushes formatting/module fixes with PR notification

### Fixed
- Code formatting in pkg/env/doc.go and pkg/env/naming_test.go
- Simplified CI workflow to just run 'make check' (was overcomplicated)

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
