# v1 Release Checklist

This document outlines the final steps required to release llima-box v1.0.0.

## Current Status: 95% Complete

All implementation is done. Remaining work is testing, bug fixes, and release preparation.

## Pre-Release Checklist

### 1. Manual Testing (Required - macOS + Lima)

**Prerequisite**: macOS system with Lima installed

Run all 14 test scenarios from `docs/TESTING.md`:

- [ ] Scenario 1: First-time VM setup
- [ ] Scenario 2: Multiple shells for same project
- [ ] Scenario 3: Multiple projects (isolation)
- [ ] Scenario 4: Path preservation
- [ ] Scenario 5: Command execution (`--`)
- [ ] Scenario 6: VM stop/start persistence
- [ ] Scenario 7: Environment deletion
- [ ] Scenario 8: Delete all environments
- [ ] Scenario 9: SSH agent forwarding (Git operations)
- [ ] Scenario 10: Invalid paths
- [ ] Scenario 11: Permission errors
- [ ] Scenario 12: VM in bad state
- [ ] Scenario 13: Corrupted namespace
- [ ] Scenario 14: Resource usage (CPU/memory)

**Action Items**:
- Document test results (create `docs/TEST_RESULTS.md`)
- File issues for any bugs discovered
- Prioritize and fix critical bugs

### 2. Bug Fixes

After manual testing, address any issues:

- [ ] Fix critical bugs (blocking release)
- [ ] Document known issues (non-blocking) in release notes
- [ ] Update tests if new edge cases discovered

### 3. Documentation Review

- [x] README.md reflects actual functionality
- [x] CHANGELOG.md structure in place
- [ ] Verify all usage examples work as documented
- [ ] Check that all documentation links are valid
- [ ] Review CONTRIBUTING.md for accuracy

### 4. Code Quality

- [ ] Run `GOPROXY=direct make check` and ensure all checks pass
- [ ] Review golangci-lint output for any warnings
- [ ] Verify all unit tests pass with race detector

### 5. Release Preparation

- [ ] Update CHANGELOG.md with v1.0.0 section
  - Move all `[Unreleased]` entries to `[1.0.0] - YYYY-MM-DD`
  - Include sections: Added, Changed, Fixed (as applicable)
  - Write clear, user-facing descriptions
- [ ] Review and finalize version number (v1.0.0 or v1.0.0-beta.1)
- [ ] Ensure LICENSE is correct and up-to-date

### 6. Build Verification

- [ ] Build for all platforms: `make build-all`
- [ ] Test built binaries on macOS (both ARM64 and AMD64 if possible)
- [ ] Verify binary size is reasonable
- [ ] Check that embedded Lima config is included correctly

### 7. Release Process

```bash
# 1. Update CHANGELOG.md with version and date
# Example: ## [1.0.0] - 2025-01-10

# 2. Commit the changelog
git add CHANGELOG.md docs/
git commit -m "Prepare release v1.0.0"

# 3. Push to branch
git push origin claude/v1-docs-update-q5M0z

# 4. Create pull request and merge to main

# 5. After merge to main, create and push tag
git checkout main
git pull origin main
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### 8. Post-Release

- [ ] Verify GitHub Actions release workflow completes successfully
- [ ] Check that release binaries are uploaded to GitHub Releases
- [ ] Test downloading and running a release binary
- [ ] Announce release (update README.md status if needed)

## Known Limitations (Document in Release Notes)

These are design choices, not bugs:

1. **macOS only**: Lima-based solution requires macOS host
2. **Network isolation**: All environments share VM network (by design)
3. **No resource quotas**: CPU/memory are shared across all environments
4. **Manual Lima installation**: User must install Lima separately

## Success Criteria for v1.0.0

Release v1.0.0 when:

- ✅ All 4 CLI commands implemented and working
- ⏳ All 14 manual test scenarios pass
- ⏳ No critical bugs remaining
- ⏳ Documentation is accurate and complete
- ⏳ Release builds successfully for all platforms

## Timeline Estimate

**2-4 hours** of work remaining (primarily manual testing on macOS)

- Testing: 1-2 hours
- Bug fixes: 0-1 hour (depends on findings)
- Release prep: 0.5-1 hour

## Next Steps

1. Find a macOS system with Lima installed
2. Run the 14 manual test scenarios
3. Document results and file issues
4. Fix critical bugs
5. Follow release process above

## Questions?

- For testing procedures, see `docs/TESTING.md`
- For implementation details, see `docs/ARCHITECTURE.md`
- For development setup, see `CONTRIBUTING.md`
