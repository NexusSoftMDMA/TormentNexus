# L3 Promotion Strategy

## Overview

This document outlines the systematic path from L2 (first-party docs captured) to L3 (fully locked with version pins and response payload fixtures).

## L3 Blocking Gaps (per platform)

### 1. GitHub Copilot CLI → L3

**Requirement 1: Runtime Version Pin**
```bash
# Capture command
copilot version --json
# Expected output format
{
  "version": "1.x.y",
  "buildDate": "2026-XX-XX",
  "commit": "xxxxxxxx"
}
```

**Requirement 2: Response Payload Fixture Set**
- [ ] `bash` tool call response (success + error variants)
- [ ] `edit` tool response (conflict detection)
- [ ] `view` tool response (file encoding handling)
- [ ] `powershell` tool response (exit code mapping)
- [ ] Hook decision contract (preToolUse)

**Requirement 3: CI Validation Script**
- Verify `copilot version` is pinned in lock document
- Validate hook contract structure against schema
- Check permission pattern syntax matches documented rules

---

### 2. Codex CLI → L3

**Requirement 1: Runtime Version Pin**
```bash
# Capture command
codex features --json
# Expected: version in features list response
```

**Requirement 2: Response Payload Fixture Set**
- [ ] Sandbox transition response (read-only → workspace-write)
- [ ] Approval JSON contract (`{ permissionDecision, approvalReason }`)
- [ ] Tool execution response with exit code semantics
- [ ] `/permissions` interactive command response contract

**Requirement 3: CI Validation**
- Verify `--ask-for-approval` behavior matches documented modes
- Validate sandbox preset combinations are exhaustive
- Check JSON output formats match spec

---

### 3. Claude Code → L3

**Requirement 1: Release Pin**
```json
{
  "platform": "Claude Code",
  "versionSource": "claudeCode.extensionVersion",
  "lastVerified": "2026-03-19",
  "capturedVersion": "x.y.z"
}
```

**Requirement 2: Response Payload Fixture Set**
- [ ] PreToolUse hook decision contract
- [ ] Permission rule evaluation result
- [ ] Bash execution response with exit code
- [ ] MCP tool invocation response

**Requirement 3: CI Validation**
- Verify hook exit code map (0=allow, 2=deny)
- Validate tool name matchers exhaustive (Shell, Read, Write, etc.)
- Check permission rule precedence (deny > ask > allow)

---

### 4. Cursor → L3

**Requirement 1: Version Pin**
```bash
# Capture from cursor CLI or extension version
# Store in fixture alongside hook schema version
```

**Requirement 2: Response Payload Fixture Set**
- [ ] Hook event payload (all categories: Shell, Read, Write, Grep, Delete, Task, MCP)
- [ ] Permission decision response
- [ ] Tool execution result (exit code semantics)

**Requirement 3: CI Validation**
- Verify hook matchers match documented categories
- Validate decision responses align with Cursor behavior
- Check MCP tool invocation patterns

---

### 5. Gemini CLI → L3

**Requirement 1: Version Pin**
```bash
# Capture from CLI version output
gemini --version
```

**Requirement 2: Response Payload Fixture Set**
- [ ] Slash command response contracts
- [ ] Built-in tool mutation confirmation semantics
- [ ] Context switching response
- [ ] Tool execution output format

**Requirement 3: CI Validation**
- Verify slash command inventory completeness
- Validate safety confirmation fields present on mutations
- Check MCP integration contract

---

### 6. Kiro → L3

**Requirement 1: Version Pin**
```bash
# Capture from hooks.json schema version or CLI
kiro --version
```

**Requirement 2: Response Payload Fixture Set**
- [ ] Hook event contracts (all event types)
- [ ] Tool execution response
- [ ] MCP server response format
- [ ] Permission decision semantics

**Requirement 3: CI Validation**
- Verify hook event types completeness
- Validate MCP response structure
- Check version pin strategy doc

---

## Implementation Timeline

### Phase 1: Version Pinning Infrastructure ✅ COMPLETED (2026-03-19)
- [x] Created `docs/fixtures/VERSION_PINS.jsonc` with structure for all L2 platforms
- [x] Documented capture methods for each platform (CLI command or config file location)
- [x] Added GitHub Actions workflow to validate version pins on CI
- **Commit**: `25dae9eb` — "docs: add Phase 1 evidence lock fixtures and CI validation"

### Phase 2: Golden Fixtures (Response Payloads) ✅ COMPLETED (2026-03-19)
- [x] Created `docs/fixtures/GOLDEN_FIXTURE_RESPONSES.md` with concrete I/O examples for all L2 platforms
- [x] Included response payloads for all tool categories (bash, edit, view, hooks, sandbox, approval flows)
- [x] Added error/edge cases (conflicts, encoding, transitions)
- [x] Created `docs/fixtures/FIXTURE_SCHEMA.jsonc` — JSON Schema validators for fixture compliance
- [x] Documented cross-platform tool equivalence matrix
- **Next step**: Run CI validation on these fixtures

### Phase 3: Tool Equivalence Mapping ⏭️ NEXT
- [ ] Add schema validation to CI workflow (validate GOLDEN_FIXTURE_RESPONSES against FIXTURE_SCHEMA)
- [ ] Create test harness to verify fixtures match documented contracts
- [ ] Document behavioral differences requiring special handling

### Phase 4: Release Pin Capture & Promotion
- [ ] Capture actual runtime versions for each L2 platform (populate VERSION_PINS.jsonc)
- [ ] Run CI validation with live version pins
- [ ] Promote platforms from L2 → L3 as versions are pinned and fixtures validated

---

## Success Criteria for L3 Promotion

Platform moves to L3 when:
- ✅ Version pin strategy defined and documented in lock file
- ✅ Runtime version capture validated in CI on each release
- ✅ Response payload fixtures created for all major tool categories
- ✅ Fixture schema validated in CI (JSON Schema or YAML schema)
- ✅ Tool equivalence mapped to standardization document
- ✅ Reviewer sign-off on lock line item (name + date)

---

## Executive Summary

**Current State (2026-03-19):**
- 6 platforms at L2 (first-party docs captured)
- 3 platforms at L1 (archive/integration level)
- 1 platform at L0 (no evidence)
- 1 platform at L3 (OpenCode, baseline)

**Phase 1 Completed**: Version pin infrastructure established            
**Phase 2 Completed**: Golden fixture responses and schemas created    
**Phase 3 In Progress**: Schema validation in CI and fixture completeness checks  
**Timeline**: Phase 2 fixtures now ready for CI validation and version capture

