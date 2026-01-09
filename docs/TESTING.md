# Testing Plan

This document describes the testing strategy for llima-box.

## Unit Tests

Unit tests cover isolated logic that doesn't require Lima VMs.

### Test Coverage

#### Path Sanitization
- Convert various path formats to valid environment names
- Handle special characters, spaces, non-ASCII
- Ensure deterministic hash generation
- Validate format: `<basename>-<hash>`

**Test Cases:**
```go
"/Users/alice/my-project"           → "my-project-a1b2"
"/Users/alice/My Cool App"          → "my-cool-app-c3d4"
"/Users/alice/123-invalid"          → "a123-invalid-e5f6"
"/Users/alice/project-α"            → "project-g7h8"
"/Users/bob/my-project"             → "my-project-i9j0"  // Different hash
```

#### Environment Name Generation
- Validate uniqueness across different paths
- Ensure valid Linux usernames
- Handle edge cases (very long names, minimal names)

#### Configuration Parsing
- Parse Lima YAML configuration
- Validate embedded configuration
- Handle user overrides

#### Error Message Formatting
- Clear, actionable error messages
- Include relevant context
- Suggest fixes where possible

## Integration Tests (Manual)

Integration tests require a macOS system with Lima. These are performed manually.

### Test Scenario 1: First-Time Setup

**Goal**: Verify VM creation and first environment setup.

**Steps:**
1. Ensure no Lima VM named "agents" exists
2. Run: `llima-box shell`
3. Wait for VM creation and provisioning
4. Verify shell launches in isolated environment
5. Verify host path is preserved

**Expected Results:**
- VM "agents" is created
- Environment user account created
- Namespace created successfully
- Shell starts in project directory
- Project files are visible
- System directories are read-only

**Verification Commands (in shell):**
```bash
pwd                    # Should show host path
ls -la                 # Should show project files
touch /etc/test        # Should fail (read-only)
touch test.txt         # Should succeed
ls /Users             # Should only show project path
```

### Test Scenario 2: Multiple Shells for Same Project

**Goal**: Verify namespace sharing between shells.

**Steps:**
1. Open first shell: `llima-box shell /Users/me/project`
2. Create file: `touch test-file.txt`
3. Open second shell: `llima-box shell /Users/me/project`
4. Verify file is visible: `ls test-file.txt`
5. Modify file in second shell: `echo "hello" > test-file.txt`
6. Verify change in first shell: `cat test-file.txt`

**Expected Results:**
- Both shells see the same filesystem
- Changes in one shell are immediately visible in the other
- Both shells run as the same user

### Test Scenario 3: Multiple Projects Isolation

**Goal**: Verify filesystem isolation between projects.

**Steps:**
1. Create two project directories
2. Open shell in project A: `llima-box shell ~/project-a`
3. Create file: `touch secret.txt`
4. Open shell in project B: `llima-box shell ~/project-b`
5. Try to access project A: `ls ~/project-a`

**Expected Results:**
- Projects run as different users
- Project B cannot see project A's files
- Each project has isolated filesystem view

### Test Scenario 4: Environment Listing

**Goal**: Verify `list` command shows all environments.

**Steps:**
1. Create environments for 3 different projects
2. Run: `llima-box list`

**Expected Results:**
- All 3 environments are listed
- Environment names match expected format
- Output is readable and formatted

### Test Scenario 5: Environment Deletion

**Goal**: Verify `delete` command removes environment.

**Steps:**
1. Create environment: `llima-box shell ~/test-project`
2. Exit shell
3. Delete environment: `llima-box delete ~/test-project`
4. Verify deletion: `llima-box list`

**Expected Results:**
- User account is removed
- Namespace is cleaned up
- Environment no longer appears in list
- Background processes are terminated

### Test Scenario 6: Delete All Environments

**Goal**: Verify `delete-all` removes all environments.

**Steps:**
1. Create 3+ environments
2. Run: `llima-box delete-all`
3. Confirm when prompted
4. Verify: `llima-box list`

**Expected Results:**
- All environments are removed
- List shows no environments
- VM is still running and functional

### Test Scenario 7: VM State Recovery

**Goal**: Verify recovery from VM shutdown.

**Steps:**
1. Create environment and exit
2. Stop VM: `limactl stop agents`
3. Run: `llima-box shell` (same project)

**Expected Results:**
- llima-box detects VM is stopped
- Automatically starts VM
- Connects to existing environment
- Namespace is restored

### Test Scenario 8: SSH Agent Forwarding

**Goal**: Verify Git operations work with SSH keys.

**Steps:**
1. Create environment with a Git repository
2. Run: `llima-box shell`
3. Run: `git fetch` or `git pull`

**Expected Results:**
- SSH agent forwarding works
- Git operations succeed without password prompts
- SSH keys from host are available

### Test Scenario 9: Path Preservation

**Goal**: Verify host paths are preserved in VM.

**Steps:**
1. Create project at: `/Users/alice/Documents/my-app`
2. Run: `llima-box shell /Users/alice/Documents/my-app`
3. Check: `pwd`

**Expected Results:**
- `pwd` shows exact host path: `/Users/alice/Documents/my-app`
- No path translation occurs
- Relative paths work as expected

### Test Scenario 10: Custom Command Execution

**Goal**: Verify command execution without interactive shell.

**Steps:**
1. Run: `llima-box shell ~/project -- ls -la`
2. Run: `llima-box shell ~/project -- python script.py`

**Expected Results:**
- Commands execute in isolated environment
- Output is displayed
- Process exits after command completes
- No interactive shell remains

### Test Scenario 11: Error Handling

**Goal**: Verify graceful error handling.

**Test Cases:**

#### Invalid Path
```bash
llima-box shell /nonexistent/path
```
Expected: Clear error message about invalid path

#### VM Creation Failure
1. Create invalid `~/.lima/agents/lima.yaml`
2. Run: `llima-box shell`

Expected: Error message with troubleshooting steps

#### Namespace Creation Failure
1. Manually corrupt namespace file
2. Run: `llima-box shell`

Expected: Auto-recovery or clear error message

## Performance Tests (Manual)

### Test Scenario 12: First Shell Startup Time

**Measurement**: Time from command execution to shell prompt.

**Steps:**
1. Ensure VM is running
2. Delete environment if exists
3. Run: `time llima-box shell`

**Target**: < 3 seconds for first shell

### Test Scenario 13: Additional Shell Startup Time

**Measurement**: Time for subsequent shells.

**Steps:**
1. First shell already created
2. Run: `time llima-box shell`

**Target**: < 0.5 seconds for additional shells

### Test Scenario 14: VM Startup Time

**Measurement**: Time for VM to become ready.

**Steps:**
1. Delete VM: `limactl delete agents`
2. Run: `time llima-box shell`

**Target**: < 60 seconds for full VM creation and provisioning

## Regression Tests

After code changes, run all integration tests to ensure no regressions.

### Minimum Test Suite (Quick Check)
- Test Scenario 1 (First-time setup)
- Test Scenario 2 (Multiple shells)
- Test Scenario 3 (Isolation)

### Full Test Suite (Before Release)
- All 14 test scenarios
- Performance benchmarks
- Error handling tests

## Automated Testing (Future)

Potential approaches for automated integration testing:

1. **GitHub Actions with macOS runners**
   - Install Lima
   - Run integration tests
   - Expensive but comprehensive

2. **Docker-based unit tests**
   - Test namespace logic in Linux containers
   - Doesn't test Lima integration
   - Good for quick feedback

3. **Mock Lima API**
   - Test llima-box logic without real VMs
   - Requires maintaining mocks
   - Won't catch Lima integration issues

**Decision for v1**: Manual testing is sufficient. Automated tests can be added later if the project grows.
