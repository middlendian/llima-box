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
- golangci-lint configuration with 15+ linters
- Comprehensive build and development documentation
- CONTRIBUTING.md with contributor guidelines
- VM lifecycle management implementation
- SSH client with retry logic and agent forwarding
- Environment naming and sanitization with tests

### Changed
- Made `help` the default Makefile target for better UX

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
