# Design Decisions

This document explains the key design decisions made for llima-box.

## Why Lima Instead of Docker?

| Aspect | Lima VM | Docker |
|--------|---------|--------|
| **Agent Docker Access** | Native Docker in VM | Requires docker-in-docker |
| **Isolation Level** | Full OS-level isolation | Container-level isolation |
| **Resource Sharing** | Shared VM resources | Individual container overhead |
| **Path Preservation** | Transparent host paths | Volume mount complexity |
| **SSH Agent Forwarding** | Built-in Lima support | Requires socket mounting |

### Rationale

LLM agents often need to:
- Run Docker containers as part of their workflow
- Access files at their original host paths
- Use SSH keys for Git operations
- Share resources efficiently

Lima provides all of these capabilities natively, while Docker would require complex workarounds (docker-in-docker, bind mount translation, socket forwarding).

## Go Implementation

### Why Go?

1. **Lima is written in Go**: Direct access to Lima's internal packages
2. **Single binary**: Easy distribution without runtime dependencies
3. **Cross-compilation**: Build for multiple platforms if needed
4. **Performance**: Fast startup and execution
5. **Concurrency**: Built-in support for managing multiple environments

### Lima Integration

We wrap the `limactl` CLI tool for VM management:
- **VM lifecycle**: Create, start, stop, delete operations via `limactl` commands
- **Instance inspection**: Parse JSON output from `limactl list --json`
- **Configuration**: Embed Lima YAML config and pass to `limactl create`

**Why CLI wrapper instead of Go library?**

Initially we explored using Lima's Go packages directly (`pkg/store`, `pkg/instance`), but the CLI wrapper approach offers significant advantages:

1. **Simpler builds**: No CGO dependency required (Lima library needs CGO for VZ support on macOS)
2. **Smaller binaries**: Avoids embedding Lima library and 60+ transitive dependencies
3. **Better compatibility**: Delegates platform-specific VM handling to `limactl`
4. **More maintainable**: Clearer separation between llima-box and Lima internals
5. **Already required**: Users need Lima installed anyway

The wrapper approach provides the same functionality with better reliability and simpler deployment.

## Environment Naming

### Requirements

1. Valid Linux usernames (no special characters, length limits)
2. Unique across different project paths
3. Deterministic (same path = same name)
4. Human-readable for debugging

### Solution

Format: `<sanitized-basename>-<hash>`

Example: `/Users/alice/My Projects/cool-app` → `cool-app-a1b2`

**Algorithm:**
1. Extract basename: `cool-app`
2. Sanitize: lowercase, replace invalid chars with `-`, ensure starts with letter
3. Hash: SHA1 of full path, take first 4 hex chars
4. Combine: `cool-app-a1b2`

**Why SHA1 (not MD5 or SHA256)?**
- SHA1 provides sufficient collision resistance for our use case
- Widely available in standard libraries
- 4 hex chars = 65,536 combinations (good enough for local use)

## Automatic VM Management

### Philosophy

Users shouldn't need to understand Lima internals to use llima-box.

### Implementation

1. **First run**: Automatically create and configure VM
2. **VM stopped**: Automatically start it
3. **VM missing**: Recreate from embedded configuration
4. **VM corrupted**: Detect and offer to recreate

### Trade-offs

**Pros:**
- Better user experience
- No manual setup steps
- Self-healing

**Cons:**
- Slower first run (~30 seconds)
- Hidden complexity
- Harder to debug VM issues

**Decision**: The UX benefits outweigh the cons for our target users (developers running LLM agents).

## Error Handling Strategy

### Guiding Principles

1. **Auto-fix when possible**: Don't bother users with fixable issues
2. **Fail fast on critical errors**: Don't hide data loss risks
3. **Provide clear error messages**: Users should understand what went wrong
4. **Always allow escape hatch**: Advanced users can manually intervene

### Examples

**Auto-fix:**
- VM not running → Start it automatically
- Namespace file missing → Recreate namespace
- User account exists but namespace missing → Recreate namespace

**Fail fast:**
- VM disk full → Error with clear message
- Namespace mount failed → Error with debug info
- SSH connection failed → Error with troubleshooting steps

## Configuration Management

### Lima VM Configuration

**Embedded in binary:**
- Default VM configuration (YAML)
- Namespace setup scripts
- Sandbox entry scripts

**Stored in `~/.lima/agents/`:**
- `lima.yaml`: Generated from embedded default
- `override.yaml`: User customizations (if present)

**Why this approach?**
- Users get working defaults without any setup
- Advanced users can customize by editing `override.yaml`
- We can update defaults in new versions
- Users' customizations persist across updates

### Environment State

Each environment stores:
- User account in VM (`/home/<env>/`)
- Namespace file (`/home/<env>/namespace.mnt`)
- Process keeping namespace alive

**Why not a state database?**
- Linux user accounts ARE the state database
- Namespace files provide persistence
- No need for additional complexity
- Easy to inspect with standard Linux tools

## Multi-Shell Support

### Problem

Multiple shells for the same project should share the same filesystem view.

### Solution

Persistent namespaces using bind-mounted namespace files.

**How it works:**
1. First shell: `unshare --mount=/home/env/namespace.mnt` creates namespace
2. Background process: `sleep infinity` keeps namespace alive
3. Subsequent shells: `nsenter --mount=/home/env/namespace.mnt` joins it

**Alternatives considered:**

| Approach | Pros | Cons |
|----------|------|------|
| **New namespace per shell** | Simple | Different views = confusing |
| **Shared namespace (chosen)** | Consistent view | More complex setup |
| **Container approach** | Familiar to Docker users | Requires docker-in-docker |

## Testing Strategy

### Challenges

- Requires actual Linux VM to test namespaces
- Mount operations need root privileges
- Integration with Lima is complex

### Approach

**Unit Tests:**
- Path sanitization logic
- Environment name generation
- Configuration parsing
- Error handling

**Integration Tests:**
- Require Lima VM on macOS
- Manual testing scenarios
- Documented test cases

**Why no automated integration tests?**
- Requires macOS + Lima + virtualization
- Complex CI/CD setup
- Manual testing is more pragmatic for v1
- Can add automated tests later if needed

### Manual Test Plan

See `docs/TESTING.md` for detailed test scenarios.
