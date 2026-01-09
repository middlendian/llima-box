# Lima VM Configuration

This document describes the Lima VM configuration used by llima-box.

## Configuration File

The default Lima configuration is embedded in the llima-box binary and written to `~/.lima/llima-box/lima.yaml` on first run.

## Default Configuration

```yaml
images:
# x86_64 / AMD64 architecture
- location: "https://cloud-images.ubuntu.com/releases/24.04/release/ubuntu-24.04-server-cloudimg-amd64.img"
  arch: "x86_64"
# ARM64 / aarch64 architecture (Apple Silicon)
- location: "https://cloud-images.ubuntu.com/releases/24.04/release/ubuntu-24.04-server-cloudimg-arm64.img"
  arch: "aarch64"

cpus: 4
memory: "8GiB"
disk: "100GiB"

mounts:
- location: "~"
  writable: true
- location: "/tmp/lima"
  writable: true

ssh:
  localPort: 60022
  loadDotSSHPubKeys: true
  forwardAgent: true

provision:
# Install required packages
- mode: system
  script: |
    #!/bin/bash
    set -eux -o pipefail
    export DEBIAN_FRONTEND=noninteractive

    apt-get update
    apt-get install -y zsh build-essential curl git

    # Install mise-en-place
    curl https://mise.run | sh

    chsh -s /bin/zsh

# Create namespace setup script
- mode: system
  script: |
    #!/bin/bash
    set -eux -o pipefail

    cat > /usr/local/bin/create-namespace.sh << 'SCRIPT_EOF'
    #!/bin/bash
    set -e

    ENV_NAME="$1"
    PROJECT_PATH="$2"
    NS_FILE="/home/$ENV_NAME/namespace.mnt"

    if [ -z "$ENV_NAME" ] || [ -z "$PROJECT_PATH" ]; then
        echo "Usage: $0 <env_name> <project_path>"
        exit 1
    fi

    echo "Creating persistent namespace for $ENV_NAME at $PROJECT_PATH"

    unshare --mount="$NS_FILE" --propagation private bash -c '
        mount --make-rprivate /
        mkdir -p /tmp/isolated-root

        # Mount system directories (read-only)
        for dir in bin sbin lib lib64 usr etc; do
            if [ -d "/$dir" ]; then
                mkdir -p "/tmp/isolated-root/$dir"
                mount --bind "/$dir" "/tmp/isolated-root/$dir"
                mount -o remount,bind,ro "/tmp/isolated-root/$dir"
            fi
        done

        # Mount runtime directories
        mkdir -p /tmp/isolated-root/{tmp,var,proc,sys,dev}
        mount --bind /tmp /tmp/isolated-root/tmp
        mount --bind /var /tmp/isolated-root/var
        mount --bind /proc /tmp/isolated-root/proc
        mount --bind /sys /tmp/isolated-root/sys
        mount --bind /dev /tmp/isolated-root/dev

        # Create project path structure
        project_path="'"$PROJECT_PATH"'"
        parent_path="$(dirname "$project_path")"
        mkdir -p "/tmp/isolated-root$parent_path"
        mkdir -p "/tmp/isolated-root$project_path"
        mount --bind "$project_path" "/tmp/isolated-root$project_path"

        # Mount user home
        mkdir -p "/tmp/isolated-root/home/'"$ENV_NAME"'"
        mount --bind "/home/'"$ENV_NAME"'" "/tmp/isolated-root/home/'"$ENV_NAME"'"

        chroot /tmp/isolated-root sleep infinity
    ' &

    sleep 2
    echo "Namespace created successfully"
    SCRIPT_EOF

    chmod +x /usr/local/bin/create-namespace.sh

# Create sandbox entry script
- mode: system
  script: |
    #!/bin/bash
    set -eux -o pipefail

    cat > /usr/local/bin/sandbox.sh << 'SCRIPT_EOF'
    #!/bin/bash
    set -e

    ENV_NAME="$1"
    PROJECT_PATH="$2"
    shift 2

    NS_FILE="/home/$ENV_NAME/namespace.mnt"

    if [ -z "$ENV_NAME" ] || [ -z "$PROJECT_PATH" ]; then
        echo "Usage: $0 <env_name> <project_path> [command...]"
        exit 1
    fi

    if [ $# -eq 0 ]; then
        set -- zsh
    fi

    if [ ! -f "$NS_FILE" ]; then
        echo "Error: Namespace file $NS_FILE not found"
        exit 1
    fi

    nsenter --mount="$NS_FILE" bash -c "cd '$PROJECT_PATH' && exec \"\$@\"" -- "$@"
    SCRIPT_EOF

    chmod +x /usr/local/bin/sandbox.sh

# Configure sudo permissions
- mode: system
  script: |
    #!/bin/bash
    set -eux -o pipefail

    echo "lima ALL=(ALL) NOPASSWD: /usr/sbin/useradd, /usr/sbin/userdel, /bin/mkdir, /bin/chown, /usr/local/bin/create-namespace.sh" > /etc/sudoers.d/lima-environments
    chmod 440 /etc/sudoers.d/lima-environments

probes:
- mode: readiness
  description: "Environment management scripts installed"
  script: |
    #!/bin/bash
    set -eux -o pipefail
    if ! [ -x /usr/local/bin/create-namespace.sh ] || ! [ -x /usr/local/bin/sandbox.sh ]; then
      echo >&2 "Environment scripts not installed"
      exit 1
    fi
```

## Configuration Sections

### Images

Specifies Ubuntu 24.04 LTS cloud images for both x86_64 and ARM64 architectures. Lima automatically selects the appropriate image based on the host architecture.

### Resources

- **CPUs**: 4 cores
- **Memory**: 8 GiB
- **Disk**: 100 GiB

These defaults work well for multiple LLM agent environments. Users can customize by editing `~/.lima/llima-box/lima.yaml` after first run.

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
- `zsh`: Default shell for environments
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
