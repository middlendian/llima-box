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

### Phase 4: Environment Manager ✅ COMPLETE

**Location**: `pkg/env/manager.go`

**Tasks**:
- [x] Implement user account creation via SSH
- [x] Create persistent namespace setup
- [x] Implement namespace entry (nsenter)
- [x] Add environment listing functionality
- [x] Add environment deletion with cleanup
- [x] Handle namespace file persistence
- [x] Error recovery for corrupted namespaces

**Key Operations**:
```go
func (m *Manager) Create(projectPath string) (*Environment, error)
func (m *Manager) Exists(envName string) (bool, error)
func (m *Manager) List() ([]*Environment, error)
func (m *Manager) Delete(envName string) error
func (m *Manager) EnterNamespace(env *Environment, cmd []string) error
```

### Phase 5: CLI Commands ✅ COMPLETE

**Location**: `internal/cli/`

**Tasks**:

#### `shell` Command ✅
- [x] Parse project path (default: current directory)
- [x] Parse command to execute (default: bash)
- [x] Ensure VM is running
- [x] Create environment if needed
- [x] Enter namespace
- [x] Execute command (interactive or non-interactive)

#### `list` Command ✅
- [x] Connect to VM
- [x] Query all user accounts
- [x] Filter system users
- [x] Format output as table
- [x] Show environment name and project path

#### `delete` Command ✅
- [x] Parse project path
- [x] Generate environment name
- [x] Prompt for confirmation (with --force flag)
- [x] Kill namespace processes
- [x] Delete user account
- [x] Clean up namespace file

#### `delete-all` Command ✅
- [x] List all environments
- [x] Show count and names
- [x] Prompt for confirmation
- [x] Delete each environment
- [x] Report success/failures

### Phase 6: Polish & Error Handling ✅ MOSTLY COMPLETE

**Tasks**:
- [x] Clear, actionable error messages
- [x] Auto-recovery for common issues (VM auto-start, retry logic)
- [x] Progress indicators for slow operations
- [x] Help text and usage examples
- [x] Handle edge cases (invalid paths, corrupted state)
- [ ] End-to-end validation on macOS + Lima (requires manual testing)

### Phase 7: Testing ⏳ IN PROGRESS

**Tasks**:
- [x] Unit tests for environment naming (327 lines in `pkg/env/naming_test.go`)
- [x] Unit tests for SSH client (258 lines in `pkg/ssh/client_test.go`)
- [x] Unit tests for environment manager (95 lines in `pkg/env/manager_test.go`)
- [ ] Manual test scenario 1: First-time setup (requires macOS)
- [ ] Manual test scenario 2: Multiple shells (requires macOS)
- [ ] Manual test scenario 3: Project isolation (requires macOS)
- [ ] Manual test scenarios 4-14: (See docs/TESTING.md)
- [ ] Document test results
- [ ] Fix bugs found during testing

## Estimated Timeline

**Total Completed**: ~18-20 hours (Phases 1-6)
**Total Remaining**: ~2-4 hours of testing and validation

**Breakdown**:
- ✅ Foundation: 1-2 hours (DONE)
- ✅ Environment Naming: 2-3 hours (DONE)
- ✅ SSH Client: 2-3 hours (DONE)
- ✅ Environment Manager: 4-5 hours (DONE)
- ✅ CLI Commands: 4-5 hours (DONE)
- ✅ Polish: 2-3 hours (DONE)
- ⏳ Testing: 2-4 hours (manual testing on macOS remains)

## Success Criteria

v1 is complete when:

1. ✅ All four commands implemented (`shell`, `list`, `delete`, `delete-all`)
2. ✅ VM is created and managed automatically
3. ✅ Environments are properly isolated (implementation complete)
4. ✅ Multiple shells can share the same environment (implementation complete)
5. ✅ Path preservation works correctly (implementation complete)
6. ✅ SSH agent forwarding works for Git
7. ⏳ Manual test suite passes (all 14 scenarios) - **requires macOS testing**
8. ✅ Error messages are clear and actionable
9. ⏳ README reflects actual functionality - **needs update**

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

## Release Strategy

### v0.2.0 Beta (Next - Ready Now)

Implementation is ~95% complete. Ready for beta release to gather real-world feedback.

**Remaining work** (~30-60 minutes):
1. Update CHANGELOG.md with v0.2.0 section
2. Create PR and merge to main
3. Create and push v0.2.0 tag
4. GitHub Actions will build and publish binaries

See `docs/V0.2_RELEASE_CHECKLIST.md` for details.

### v1.0.0 Stable (After Beta Testing)

Release v1.0.0 after real-world validation:

1. **Beta Testing** (1-2 weeks after v0.2.0)
   - Early adopters test on real projects
   - Issues filed and triaged
   - Edge cases identified

2. **Manual Testing on macOS**: Run all 14 test scenarios from `docs/TESTING.md`
   - Requires macOS system with Lima installed
   - Validates end-to-end integration
   - Documents test results

3. **Bug Fixes**: Address critical issues from beta feedback

4. **v1.0.0 Release**:
   - Update CHANGELOG.md with v1.0.0 section
   - Create release tag
   - Build and publish stable binaries

## Questions?

See:
- `docs/ARCHITECTURE.md` - Technical design
- `docs/DESIGN.md` - Design decisions
- `docs/IMPLEMENTATION_PLAN.md` - Detailed implementation guide
- `docs/TESTING.md` - Test scenarios
