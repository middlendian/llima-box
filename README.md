# llima-box

> **Status**: 🚧 In Development - Not yet functional

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

## Planned Usage

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

## Prerequisites

- macOS (ARM64 or x86_64)
- Lima installed (`brew install lima`)

## Installation

**Not yet available** - Project is in development.

Once ready, installation will be:
```bash
go install github.com/yourusername/llima-box@latest
```

## Development Status

### Completed
- ✅ Architecture design
- ✅ Documentation structure

### In Progress
- 🚧 Go project structure
- 🚧 Lima integration

### Planned
- ⏳ VM lifecycle management
- ⏳ Environment management (create, list, delete)
- ⏳ Shell command implementation
- ⏳ Testing and validation

## How It Works

llima-box creates a Lima VM with Ubuntu 22.04 and uses Linux mount namespaces to isolate each agent environment:

1. **First run**: Creates and provisions a Lima VM named "agents"
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

## Contributing

This project is in early development. Contributions welcome once the core implementation is complete.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Why "llima-box"?

- **llima**: Lima for LLM agents
- **box**: Isolated sandbox environments

Simple, descriptive, easy to type.
