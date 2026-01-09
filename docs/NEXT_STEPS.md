# Next Steps

This document outlines the remaining work to complete llima-box v1.

## Current Status

✅ **Phase 1: Foundation Complete**
- Project structure established
- Lima integration validated
- VM lifecycle management implemented
- Multi-architecture support (x86_64 + ARM64)
- Ubuntu 24.04 LTS configuration
- Documentation framework in place

✅ **Phase 2: Environment Naming Complete**
- Path-to-environment-name generation implemented
- Basename sanitization for Linux usernames
- 4-character hash generation for uniqueness
- Comprehensive validation and testing
- Location: `pkg/env/naming.go`, `pkg/env/naming_test.go`

✅ **Phase 3: SSH Client Complete**
- SSH client wrapper using Lima's sshutil
- Connection management with auto-connect
- Non-interactive command execution (Exec, ExecContext)
- Interactive shell support with PTY
- Retry logic with exponential backoff
- SSH agent forwarding for Git operations
- Location: `pkg/ssh/client.go`, `pkg/ssh/retry.go`, `pkg/ssh/client_test.go`

## Remaining Work for v1

### Phase 4: Environment Manager (4-5 hours)

**Location**: `pkg/env/manager.go`, `pkg/env/namespace.go`

**Tasks**:
- [ ] Implement user account creation via SSH
- [ ] Create persistent namespace setup
- [ ] Implement namespace entry (nsenter)
- [ ] Add environment listing functionality
- [ ] Add environment deletion with cleanup
- [ ] Handle namespace file persistence
- [ ] Error recovery for corrupted namespaces

**Key Operations**:
```go
func (m *Manager) Create(projectPath string) (*Environment, error)
func (m *Manager) Exists(envName string) (bool, error)
func (m *Manager) List() ([]*Environment, error)
func (m *Manager) Delete(envName string) error
func (m *Manager) EnterNamespace(env *Environment, cmd []string) error
```

### Phase 5: CLI Commands (4-5 hours)

**Location**: `internal/cli/`

**Tasks**:

#### `shell` Command
- [ ] Parse project path (default: current directory)
- [ ] Parse command to execute (default: zsh)
- [ ] Ensure VM is running
- [ ] Create environment if needed
- [ ] Enter namespace
- [ ] Execute command (interactive or non-interactive)

#### `list` Command
- [ ] Connect to VM
- [ ] Query all user accounts
- [ ] Filter system users
- [ ] Format output as table
- [ ] Show environment name and project path

#### `delete` Command
- [ ] Parse project path
- [ ] Generate environment name
- [ ] Prompt for confirmation
- [ ] Kill namespace processes
- [ ] Delete user account
- [ ] Clean up namespace file

#### `delete-all` Command
- [ ] List all environments
- [ ] Show count and names
- [ ] Prompt for confirmation
- [ ] Delete each environment
- [ ] Report success/failures

### Phase 6: Polish & Error Handling (2-3 hours)

**Tasks**:
- [ ] Clear, actionable error messages
- [ ] Auto-recovery for common issues
- [ ] Progress indicators for slow operations
- [ ] Help text and usage examples
- [ ] Handle edge cases (invalid paths, corrupted state)
- [ ] Logging for debugging

### Phase 7: Testing (3-4 hours)

**Tasks**:
- [ ] Unit tests for environment naming
- [ ] Unit tests for path sanitization
- [ ] Manual test scenario 1: First-time setup
- [ ] Manual test scenario 2: Multiple shells
- [ ] Manual test scenario 3: Project isolation
- [ ] Manual test scenario 4-11: (See docs/TESTING.md)
- [ ] Document test results
- [ ] Fix bugs found during testing

## Estimated Timeline

**Total Completed**: ~5-6 hours (Phases 1-3)
**Total Remaining**: ~13-17 hours of development

**Breakdown**:
- ✅ Environment Naming: 2-3 hours (DONE)
- ✅ SSH Client: 2-3 hours (DONE)
- ⏳ Environment Manager: 4-5 hours (NEXT)
- ⏳ CLI Commands: 4-5 hours
- ⏳ Polish: 2-3 hours
- ⏳ Testing: 3-4 hours

## Success Criteria

v1 is complete when:

1. ✅ All four commands work (`shell`, `list`, `delete`, `delete-all`)
2. ✅ VM is created and managed automatically
3. ✅ Environments are properly isolated
4. ✅ Multiple shells can share the same environment
5. ✅ Path preservation works correctly
6. ✅ SSH agent forwarding works for Git
7. ✅ Manual test suite passes (all 14 scenarios)
8. ✅ Error messages are clear and actionable
9. ✅ README reflects actual functionality

## Beyond v1

Ideas for future versions (not blocking v1):

- **Resource Quotas**: CPU/memory limits per environment
- **Network Isolation**: Separate network namespaces
- **Auto Cleanup**: Delete idle environments after N days
- **Multi-VM Support**: Different VMs for different use cases
- **Shell Integration**: Completion scripts, prompt customization
- **Web UI**: Browser-based environment management
- **Monitoring**: Environment metrics and usage tracking
- **Snapshots**: Save/restore environment state
- **Templates**: Pre-configured environment templates

## Getting Started

To continue development:

1. **Next Phase**: Phase 4 (Environment Manager) - namespace operations and user management
2. **Reference**: See `docs/IMPLEMENTATION_PLAN.md` for detailed specs
3. **Testing**: See `docs/TESTING.md` for manual testing procedures
4. **Documentation**: Update README.md and docs as features are implemented

## Questions?

See:
- `docs/ARCHITECTURE.md` - Technical design
- `docs/DESIGN.md` - Design decisions
- `docs/IMPLEMENTATION_PLAN.md` - Detailed implementation guide
- `docs/TESTING.md` - Test scenarios
