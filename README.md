# llima-box: Secure Multi-Agent Environment Manager

A command-line tool for creating secure, isolated environments for LLM agents within a single Lima VM on macOS. Each agent operates in complete filesystem isolation while sharing CPU and memory resources, with preserved host path structures for seamless development workflows.

## Overview

`llima-box` solves the challenge of running multiple LLM agents securely on the same system. Instead of using Docker containers (which would require docker-in-docker for agents that need Docker access), it leverages Lima VMs with Linux mount namespaces to provide robust isolation while maintaining familiar development environments.

### Key Features

- **Complete Filesystem Isolation**: Each agent sees only its project directory and essential system files
- **Preserved Host Paths**: Agents see `/Users/me/project` exactly as it appears on the host
- **Multi-Shell Support**: Multiple shells for the same environment share the same isolated view
- **Persistent Environments**: User accounts and namespaces persist across sessions
- **SSH Agent Forwarding**: Git operations work seamlessly with host SSH keys
- **Zero Docker Overhead**: Direct VM isolation without container layers
- **Automatic Path Sanitization**: Project paths are automatically converted to valid Linux usernames

## Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────┐
│ macOS Host                                                  │
│ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ │
│ │ Project A       │ │ Project B       │ │ Project C       │ │
│ │ /Users/me/proj-a│ │ /Users/me/proj-b│ │ /Users/me/proj-c│ │
│ └─────────────────┘ └─────────────────┘ └─────────────────┘ │
│           │                   │                   │         │
│           └───────────────────┼───────────────────┘         │
│                               │                             │
│ ┌─────────────────────────────┼───────────────────────────┐ │
│ │ Lima VM (Debian)            │                           │ │
│ │ ┌─────────────────┐ ┌───────┼─────────┐ ┌─────────────┐ │ │
│ │ │ Environment A   │ │ Environment B   │ │Environment C│ │ │
│ │ │ User: proj-a-1a2│ │ User: proj-b-3c4│ │User:proj-c-5e│ │ │
│ │ │ Namespace: NS-A │ │ Namespace: NS-B │ │Namespace:NS-C│ │ │
│ │ │ Sees: /Users/me/│ │ Sees: /Users/me/│ │Sees:/Users/me│ │ │
│ │ │       proj-a    │ │       proj-b    │ │     /proj-c  │ │ │
│ │ │ Isolated FS     │ │ Isolated FS     │ │ Isolated FS │ │ │
│ │ └─────────────────┘ └─────────────────┘ └─────────────┘ │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Isolation Mechanism

Each environment uses a combination of:

1. **Linux Mount Namespaces**: Create isolated filesystem views
2. **Chroot**: Complete root filesystem isolation
3. **Bind Mounts**: Selective exposure of host directories
4. **User Accounts**: Process-level isolation and permissions

### Path Preservation Strategy

Lima transparently mounts host paths at the same location in the guest VM. The isolation system preserves this by:

1. Creating the full parent directory structure in the isolated root
2. Bind mounting only the specific project directory
3. Using `chroot` to make the isolated root appear as `/`
4. Agents see their original host paths without any translation

## Design Decisions

### Why Lima Instead of Docker?

| Aspect | Lima VM | Docker |
|--------|---------|--------|
| **Agent Docker Access** | Native Docker in VM | Requires docker-in-docker |
| **Isolation Level** | Full OS-level isolation | Container-level isolation |
| **Resource Sharing** | Shared VM resources | Individual container overhead |
| **Path Preservation** | Transparent host paths | Volume mount complexity |
| **SSH Agent Forwarding** | Built-in Lima support | Requires socket mounting |

### Environment Naming Strategy

Environment names are generated from project paths using:
- **Sanitized basename**: Lowercase, alphanumeric + hyphens, starts with letter
- **Path hash**: 4-character SHA1 hash of full path for uniqueness
- **Format**: `<sanitized-basename>-<hash>` (e.g., `my-project-a1b2`)

This ensures:
- Valid Linux usernames
- Collision resistance for similar project names
- Deterministic mapping from paths to environments

### Persistent Namespace Architecture

Namespaces are created once during environment setup and persist using:
- **Bind-mounted namespace files**: `/home/<env>/namespace.mnt`
- **Background processes**: `sleep infinity` keeps namespaces alive
- **nsenter for entry**: All shells use `nsenter` to join existing namespaces

## Installation

### Prerequisites

- macOS with Lima installed
- `limactl` command available in PATH

### Setup

1. **Create Lima configuration directory**:
   ```bash
   mkdir -p ~/.lima/agents
   ```

2. **Create `~/.lima/agents/lima.yaml`** with the provided configuration (see Configuration section)

3. **Install `llima-box` script**:
   ```bash
   # Make executable and place in PATH
   chmod +x llima-box
   sudo mv llima-box /usr/local/bin/
   ```

### Lima Configuration

Create `~/.lima/agents/lima.yaml`:

```yaml
images:
- location: "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img"
  arch: "x86_64"

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

## Command Reference

### `llima-box shell [path] [-- <command>]`

Launch an isolated shell environment.

**Parameters:**
- `path` (optional): Project directory path. Defaults to current directory.
- `command` (optional): Command to execute. Defaults to `zsh`.

**Examples:**
```bash
# Launch zsh in current directory
llima-box shell

# Launch in specific directory
llima-box shell /Users/me/my-project

# Execute specific command
llima-box shell -- python script.py

# Execute command in specific directory
llima-box shell /path/to/project -- npm start
```

**Behavior:**
- Creates environment if it doesn't exist
- Creates persistent namespace on first run
- Subsequent shells join existing namespace
- All shells share the same isolated filesystem view

### `llima-box delete [path]`

Delete an environment and all associated data.

**Parameters:**
- `path` (optional): Project directory path. Defaults to current directory.

**Examples:**
```bash
# Delete environment for current directory
llima-box delete

# Delete specific environment
llima-box delete /Users/me/old-project
```

**Behavior:**
- Prompts for confirmation
- Removes user account and home directory
- Cleans up persistent namespace
- Terminates any running processes

### `llima-box delete-all`

Delete all environments (useful for testing/cleanup).

**Examples:**
```bash
llima-box delete-all
```

**Behavior:**
- Lists all environments to be deleted
- Prompts for confirmation
- Removes all non-system user accounts
- Cleans up all namespaces and data

### `llima-box list`

List all active environments.

**Examples:**
```bash
llima-box list
```

**Output:**
```
ENVIRONMENT               PROJECT              INFERRED_PATH
-----------               -------              -------------
my-project-a1b2          my-project           <hash: a1b2>
web-app-c3d4             web-app              <hash: c3d4>
```

### `llima-box help`

Display help information and usage examples.

## Usage Examples

### Basic Development Workflow

```bash
# Start working on a project
cd /Users/me/my-ai-project
llima-box shell

# Inside the isolated environment:
pwd                    # Shows: /Users/me/my-ai-project
ls -la                 # Shows only project files
git status             # Works with SSH agent forwarding
npm install            # Installs to project directory
cd /Users/me          # Error: Permission denied (isolated!)

# Open another shell for the same project
# (in another terminal)
cd /Users/me/my-ai-project
llima-box shell        # Joins same namespace, sees same files
```

### Multi-Agent Collaboration

```bash
# Terminal 1: Start main agent
cd /Users/me/collaborative-project
llima-box shell -- python main_agent.py

# Terminal 2: Start helper agent (shares same environment)
cd /Users/me/collaborative-project
llima-box shell -- python helper_agent.py

# Both agents see the same files and can collaborate
# Changes made by one are immediately visible to the other
```

### Project Cleanup

```bash
# Clean up old project
llima-box delete /Users/me/old-project

# Nuclear option: clean everything (for testing)
llima-box delete-all
```

## Security Model

### Isolation Guarantees

| Resource | Isolation Level | Details |
|----------|----------------|---------|
| **Filesystem** | Complete | Only project directory + essential system files visible |
| **Processes** | User-level | Each environment runs as separate user account |
| **Network** | Shared | All environments share VM network (by design) |
| **Memory** | Shared | All environments share VM memory pool |
| **CPU** | Shared | All environments share VM CPU resources |

### What Agents Can Access

**Allowed:**
- Their specific project directory (full read/write)
- Their home directory in the VM (persistent storage)
- Essential system binaries and libraries (read-only)
- Network access for API calls
- SSH agent for Git operations

**Blocked:**
- Other users' project directories
- Other users' home directories
- Host filesystem outside their project
- Root filesystem modifications
- System configuration changes

### Threat Model

**Protects Against:**
- Accidental file access between projects
- Malicious agents reading other projects' data
- Filesystem pollution from agent activities
- Unintended system modifications

**Does Not Protect Against:**
- Network-based attacks between agents
- Resource exhaustion (CPU/memory bombing)
- Privilege escalation within the VM
- Host system compromise (VM-level isolation only)

## Technical Implementation

### Namespace Lifecycle

1. **Creation**: `unshare --mount=<file>` creates persistent namespace
2. **Setup**: Bind mount project directory and essential system paths
3. **Persistence**: Background `sleep infinity` keeps namespace alive
4. **Entry**: `nsenter --mount=<file>` joins existing namespace
5. **Cleanup**: `umount` and `userdel` removes namespace and user

### Mount Structure

Each isolated environment sees:

```
/ (chroot root)
├── bin/          # System binaries (read-only)
├── sbin/         # System binaries (read-only)
├── lib/          # System libraries (read-only)
├── usr/          # User programs (read-only)
├── etc/          # System config (read-only)
├── tmp/          # Temporary files (shared)
├── var/          # Variable data (shared)
├── proc/         # Process info (shared)
├── sys/          # System info (shared)
├── dev/          # Device files (shared)
├── home/
│   └── <env>/    # User home (read-write, persistent)
└── Users/        # Host path structure
    └── me/
        └── project/  # Project directory (read-write)
```

### Performance Characteristics

| Operation | Time | Notes |
|-----------|------|-------|
| **First shell** | ~2-3 seconds | Creates namespace + user |
| **Additional shells** | ~0.1 seconds | Just `nsenter` |
| **Environment deletion** | ~1 second | User cleanup |
| **VM startup** | ~30 seconds | One-time Lima boot |

## Troubleshooting

### Common Issues

**"Lima instance not running"**
```bash
# Check Lima status
limactl list

# Start manually if needed
limactl start agents
```

**"Namespace file not found"**
```bash
# Recreate namespace (automatic on next shell)
llima-box shell
```

**"Permission denied" errors**
```bash
# Check if you're in the right directory
pwd

# Verify environment exists
llima-box list
```

**VM won't start**
```bash
# Check Lima logs
limactl start --log-level debug agents

# Recreate VM if corrupted
limactl delete agents
llima-box shell  # Will recreate
```

### Limitations

1. **No resource quotas**: Agents share VM resources without limits
2. **Network isolation**: All agents share the same network namespace
3. **macOS only**: Designed specifically for Lima on macOS
4. **Single VM**: All environments run in one shared VM
5. **Manual cleanup**: No automatic cleanup of idle environments
