# llima-box

> **Status**: 🧪 Beta - Implementation complete, testing in progress

A command-line tool for creating secure, isolated environments for LLM agents within a single Lima VM on macOS.

## What is llima-box?

`llima-box` solves the challenge of running multiple LLM agents securely on the same macOS system. Each agent operates in complete filesystem isolation while sharing CPU and memory resources, with preserved host path structures for seamless development workflows.

Instead of using Docker containers (which would require docker-in-docker for agents that need Docker access), llima-box leverages Lima VMs with Linux mount namespaces to provide robust isolation while maintaining familiar development environments.

## Key Features

- **Complete Filesystem Isolation**: Each agent sees only its project directory and essential system files
- **Preserved Host Paths**: Agents see `/Users/me/project` exactly as it appears on the host
- **Multi-Shell Support**: Multiple shells for the same environment share the same isolated view
- **Persistent Environments**: User accounts and namespaces persist across sessions
- **SSH Agent Forwarding**: Git operations work seamlessly with host SSH keys
- **Zero Docker Overhead**: Direct VM isolation without container layers
- **Automatic Management**: VM creation, startup, and configuration handled automatically

## Usage

```bash
# Launch isolated shell in current directory
llima-box shell

# Launch shell in specific directory
llima-box shell /path/to/project

# Execute command in isolated environment
llima-box shell -- python script.py

# List all environments
llima-box list

# Delete environment
llima-box delete /path/to/project

# Delete all environments
llima-box delete-all
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - Technical architecture and isolation mechanisms
- [Design Decisions](docs/DESIGN.md) - Why we made specific choices
- [Lima Configuration](docs/LIMA_CONFIG.md) - VM configuration details
- [Testing Plan](docs/TESTING.md) - Manual and automated testing approach
- [Implementation Plan](docs/IMPLEMENTATION_PLAN.md) - Step-by-step implementation phases
- [POC Status](docs/POC_STATUS.md) - Lima integration proof-of-concept validation
- [Next Steps](docs/NEXT_STEPS.md) - Release strategy (v0.2.0 beta, then v1.0.0 stable)

## Prerequisites

- macOS (ARM64 or x86_64)
- Lima installed (`brew install lima`)

## Installation

### From Release (Recommended)

Download the latest release for your platform from the [releases page](https://github.com/middlendian/llima-box/releases).

**macOS ARM64 (Apple Silicon):**
```bash
curl -LO https://github.com/middlendian/llima-box/releases/latest/download/llima-box-darwin-arm64.tar.gz
tar -xzf llima-box-darwin-arm64.tar.gz
chmod +x llima-box-darwin-arm64
sudo mv llima-box-darwin-arm64 /usr/local/bin/llima-box
```

**macOS AMD64 (Intel):**
```bash
curl -LO https://github.com/middlendian/llima-box/releases/latest/download/llima-box-darwin-amd64.tar.gz
tar -xzf llima-box-darwin-amd64.tar.gz
chmod +x llima-box-darwin-amd64
sudo mv llima-box-darwin-amd64 /usr/local/bin/llima-box
```

### From Source

```bash
go install github.com/middlendian/llima-box@latest
```

Or clone and build:
```bash
git clone https://github.com/middlendian/llima-box.git
cd llima-box
make build
sudo mv bin/llima-box /usr/local/bin/
```

## Development Status

### ✅ Implementation Complete
- ✅ Architecture design
- ✅ Documentation structure
- ✅ Go project structure
- ✅ Lima integration validated (POC)
- ✅ VM lifecycle management (create, start, stop, delete)
- ✅ Multi-architecture support (x86_64 + ARM64)
- ✅ Environment naming and sanitization (with 327 lines of tests)
- ✅ SSH client for VM communication (with retry logic and agent forwarding)
- ✅ Environment manager (namespace operations, user management)
- ✅ CLI commands:
  - `shell` - Launch isolated shell or execute commands
  - `list` - View all environments
  - `delete` - Remove specific environment
  - `delete-all` - Remove all environments

### 🎯 Release Plan
- **v0.2.0 Beta** (Ready now): Implementation complete, seeking real-world testing feedback
- **v1.0.0 Stable** (After beta): Manual testing complete, bugs fixed, production-ready

See [docs/V0.2_RELEASE_CHECKLIST.md](docs/V0.2_RELEASE_CHECKLIST.md) for beta release details.

## How It Works

llima-box creates a Lima VM with Ubuntu 24.04 LTS and uses Linux mount namespaces to isolate each agent environment:

1. **First run**: Creates and provisions a Lima VM named "llima-box"
2. **Environment creation**: Creates a Linux user account and persistent mount namespace
3. **Isolation**: Bind mounts only the project directory and essential system files
4. **Shell access**: Uses `nsenter` to join the existing namespace
5. **Persistence**: Background processes keep namespaces alive between shell sessions

See [Architecture](docs/ARCHITECTURE.md) for detailed technical design.

## Security Model

**Isolates:**
- ✅ Filesystem access (per-project isolation)
- ✅ User permissions (separate accounts)

**Shares:**
- ⚠️ Network (all environments share VM network)
- ⚠️ CPU/Memory (no resource quotas)

llima-box is designed for development environments, not for running untrusted code. See [Architecture](docs/ARCHITECTURE.md#security-model) for threat model details.

## Development

### Prerequisites

- Go 1.24.7 or later
- golangci-lint (for linting): `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

### Building

The project uses a Makefile for common development tasks:

```bash
# Build for current platform
make build

# Build for all platforms (Linux and macOS, ARM64 and AMD64)
make build-all

# Run tests
make test

# Format code
make fmt

# Run linters
make lint

# Run all checks (formatting, vetting, linting, tests)
make check

# Clean build artifacts
make clean

# Show all available targets
make help
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

### Running Tests

```bash
# Run all tests
make test

# Run tests with race detector and coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

### Code Quality

Before submitting a PR, ensure your code passes all checks:

```bash
make check
```

This will:
1. Check code formatting
2. Run `go vet`
3. Run golangci-lint
4. Run all tests

### Continuous Integration

The project uses GitHub Actions for CI/CD:

- **Pull Requests**: Automatically run tests, linting, and build checks on all PRs
- **Releases**: Automatically build and publish binaries for all platforms when a version tag is pushed

Release notes are extracted from [CHANGELOG.md](CHANGELOG.md), which follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) conventions.

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

## Contributing

This project is in early development. Contributions welcome once the core implementation is complete.

### Contribution Guidelines

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run `make check` to ensure code quality
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

GPLv3 License - see [LICENSE](LICENSE) for details.

## Why "llima-box"?

- **llima**: Lima for LLM agents
- **box**: Isolated sandbox environments

Simple, descriptive, easy to type.
