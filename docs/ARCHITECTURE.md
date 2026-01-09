# llima-box Architecture

This document describes the technical architecture of llima-box, a secure multi-agent environment manager for LLM agents using Lima VMs.

## Overview

`llima-box` provides complete filesystem isolation for LLM agents within a single Lima VM on macOS. Each agent operates in a separate mount namespace while sharing CPU and memory resources.

## High-Level Design

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

## Isolation Mechanism

Each environment uses a combination of:

1. **Linux Mount Namespaces**: Create isolated filesystem views
2. **Chroot**: Complete root filesystem isolation
3. **Bind Mounts**: Selective exposure of host directories
4. **User Accounts**: Process-level isolation and permissions

## Path Preservation Strategy

Lima transparently mounts host paths at the same location in the guest VM. The isolation system preserves this by:

1. Creating the full parent directory structure in the isolated root
2. Bind mounting only the specific project directory
3. Using `chroot` to make the isolated root appear as `/`
4. Agents see their original host paths without any translation

## Environment Naming Strategy

Environment names are generated from project paths using:
- **Sanitized basename**: Lowercase, alphanumeric + hyphens, starts with letter
- **Path hash**: 4-character SHA1 hash of full path for uniqueness
- **Format**: `<sanitized-basename>-<hash>` (e.g., `my-project-a1b2`)

This ensures:
- Valid Linux usernames
- Collision resistance for similar project names
- Deterministic mapping from paths to environments

## Persistent Namespace Architecture

Namespaces are created once during environment setup and persist using:
- **Bind-mounted namespace files**: `/home/<env>/namespace.mnt`
- **Background processes**: `sleep infinity` keeps namespaces alive
- **nsenter for entry**: All shells use `nsenter` to join existing namespaces

## Namespace Lifecycle

1. **Creation**: `unshare --mount=<file>` creates persistent namespace
2. **Setup**: Bind mount project directory and essential system paths
3. **Persistence**: Background `sleep infinity` keeps namespace alive
4. **Entry**: `nsenter --mount=<file>` joins existing namespace
5. **Cleanup**: `umount` and `userdel` removes namespace and user

## Mount Structure

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

## Performance Characteristics

| Operation | Time | Notes |
|-----------|------|-------|
| **First shell** | ~2-3 seconds | Creates namespace + user |
| **Additional shells** | ~0.1 seconds | Just `nsenter` |
| **Environment deletion** | ~1 second | User cleanup |
| **VM startup** | ~30 seconds | One-time Lima boot |

## Limitations

1. **No resource quotas**: Agents share VM resources without limits
2. **Network isolation**: All agents share the same network namespace
3. **macOS only**: Designed specifically for Lima on macOS
4. **Single VM**: All environments run in one shared VM
5. **Manual cleanup**: No automatic cleanup of idle environments
