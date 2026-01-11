# Lima VM Configuration

This document describes the Lima VM configuration used by llima-box.

## Configuration File

The default Lima configuration is embedded in the llima-box binary and written to `~/.lima/llima-box/lima.yaml` on first
run. See the details in [pkg/vm/lima.yaml](../pkg/vm/lima.yaml).

## Configuration Sections

### Images

Specifies Ubuntu 24.04 LTS cloud images for both x86_64 and ARM64 architectures. Lima automatically selects the
appropriate image based on the host architecture.

### Resources

- **CPUs**: 4 cores
- **Memory**: 8 GiB
- **Disk**: 100 GiB

These defaults work well for multiple LLM agent environments. Users can customize by editing
`~/.lima/llima-box/lima.yaml` after first run.

### Mounts

- **Home directory**: Full read/write access to macOS home directory
- **Temporary directory**: `/tmp/lima` for transient files

Lima preserves host path structures in the VM, which is crucial for our isolation strategy.

### SSH

- **Port**: 60022 (avoids conflicts with other SSH services)
- **SSH Keys**: Automatically loads public keys from `~/.ssh/`
- **Agent Forwarding**: Enabled for Git operations

### Provisioning Scripts

#### Package Installation

Installs essential development tools:

- `build-essential`: Compilation tools
- `curl`, `git`: Standard development utilities
- `mise-en-place`: Modern development environment manager

#### Namespace Setup Script

Creates `/usr/local/bin/create-namespace.sh` that:

1. Creates a new mount namespace
2. Sets up bind mounts for system directories (read-only)
3. Bind mounts the project directory (read-write)
4. Starts a background process to keep the namespace alive

#### Sandbox Entry Script

Creates `/usr/local/bin/sandbox.sh` that:

1. Joins an existing namespace using `nsenter`
2. Changes to the project directory
3. Executes the specified command (or zsh by default)

#### Sudo Configuration

Grants the `lima` user passwordless sudo access for:

- `useradd`, `userdel`: User account management
- `mkdir`, `chown`: Directory setup
- `create-namespace.sh`: Namespace creation

This allows llima-box to manage environments without prompting for passwords.

## Customization

Users can customize the VM configuration by:

1. **Before first run**: Create `~/.lima/llima-box/lima.yaml` manually
2. **After first run**: Edit `~/.lima/llima-box/lima.yaml` and recreate VM

Changes require VM recreation:

```bash
llima-box delete-all
limactl delete llima-box
llima-box shell  # Recreates VM with new config
```

## Readiness Probes

The configuration includes a readiness probe that verifies:

- Namespace setup script is installed and executable
- Sandbox entry script is installed and executable

This ensures the VM is fully ready before llima-box attempts to create environments.
