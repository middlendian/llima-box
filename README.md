# llima-box

> **Status**: 🧪 Beta - Implementation complete, testing in progress

A command-line tool for creating secure, isolated environments for LLM agents within a single Lima VM on macOS.

## What is llima-box?

`llima-box` solves the challenge of running multiple LLM agents securely on the same macOS system. Each agent operates
in complete filesystem isolation while sharing CPU and memory resources, with preserved host path structures for
seamless development workflows.

Instead of using Docker containers (which would require docker-in-docker for agents that need Docker access), llima-box
leverages a Lima VM with Linux mount namespaces to provide robust isolation while maintaining familiar development
environments.

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
- [Next Steps](docs/NEXT_STEPS.md) - Completed and planned work
- [Release Checklist](./docs/RELEASE_CHECKLIST.md) -

## Prerequisites

- macOS (ARM64 or x86_64)
- Lima installed (`brew install lima`)

## Installation

### Using mise (Recommended)

The easiest way to install llima-box is using [mise](https://mise.jdx.dev):

```bash
mise use -g github:middlendian/llima-box
```

This automatically downloads and installs the correct version for your platform.

### Manual Installation

Download the latest release for your platform from
the [releases page](https://github.com/middlendian/llima-box/releases).

**macOS ARM64 (Apple Silicon):**

```bash
curl -LO https://github.com/middlendian/llima-box/releases/latest/download/llima-box-macos-arm64.tar.gz
tar -xzf llima-box-macos-arm64.tar.gz
sudo mv llima-box-macos-arm64/bin/llima-box /usr/local/bin/
```

**macOS x64 (Intel):**

```bash
curl -LO https://github.com/middlendian/llima-box/releases/latest/download/llima-box-macos-x64.tar.gz
tar -xzf llima-box-macos-x64.tar.gz
sudo mv llima-box-macos-x64/bin/llima-box /usr/local/bin/
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

## Development Status: Beta

We're working on real-world testing and validation before a v1.0.0 release.

## How It Works

llima-box creates a Lima VM with Ubuntu 24.04 LTS and uses Linux mount namespaces to isolate each agent environment:

1. **First run**: Creates and provisions a Lima VM named "llima-box"
2. **Environment creation**: Creates a Linux user account and persistent mount namespace
3. **Isolation**: Bind mounts only the project directory and essential system files
4. **Shell access**: Uses `nsenter` to join the existing namespace
5. **Persistence**: Background processes keep namespaces alive between shell sessions

See [Architecture](docs/ARCHITECTURE.md) for detailed technical design.

## Security Model

For each project, llima-box sets up an isolated user and namespace within a single shared VM.

**Isolated:**

- Filesystem access (per-project isolation)
- User permissions (separate accounts)

**Shared:**

- Network (all environments share VM network)
- CPU/Memory (no resource quotas)

llima-box is designed for development environments, not for running untrusted code.
See [Architecture](docs/ARCHITECTURE.md#security-model) for threat model details.

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

## Contributing

This project is in early development. Contributions welcome once the core implementation is complete.

## License

GPLv3 License - see [LICENSE](LICENSE) for details.

## Why "llima-box"?

- **llima**: Lima for LLM agents
- **box**: Isolated sandbox environments

Simple, descriptive, easy to type.
