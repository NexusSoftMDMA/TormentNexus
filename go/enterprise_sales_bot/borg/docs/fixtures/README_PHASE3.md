# Evidence Lock Phase 3: Schema Validation & Version Capture

> Status: Phase 2 Complete (2026-03-19) | Phase 3 Starting  
> Focus: Validate golden fixtures and capture runtime versions for L3 promotion

## What's Been Completed

### ✅ Phase 1: Version Pinning Infrastructure (Commit: 25dae9eb)
- `VERSION_PINS.jsonc` — Registry for platform version captures
- `validate-evidence-lock.yml` — CI gate for lock integrity
- Capture procedures documented for all L2 platforms

### ✅ Phase 2: Golden Fixtures (Commit: f5f5f8cd)
- `GOLDEN_FIXTURE_RESPONSES.md` — Concrete I/O examples for all 6 L2 platforms
  - Each platform includes: tool responses, hook decisions, error cases, edge cases
  - Cross-platform tool equivalence matrix showing semantic mapping
- `FIXTURE_SCHEMA.jsonc` — JSON Schema validators for fixture compliance
  - Base patterns for all response types
  - Platform-specific schema definitions
  - Error response contracts

### 📊 Current Lock State
| Level | Count | Platforms |
| --- | --- | --- |
| L3 (Locked) | 1 | OpenCode |
| L2 (First-party docs + fixtures) | 6 | Copilot CLI, Codex, Claude Code, Cursor, Gemini, Kiro |
| L1 (Archive/integration) | 2 | Windsurf, Antigravity |
| L0 (Unlocked) | 1 | VS Code + Copilot IDE Agent |

---

## Phase 3: Schema Validation & Fixture Verification

### Objective
Validate that golden fixtures conform to FIXTURE_SCHEMA and are production-ready.

### Tasks

#### 3.1 Add JSON Schema Validation to CI ⏭️ IMMEDIATE
**File**: `.github/workflows/validate-evidence-lock.yml`  
**Action**: Add AJV step to validate GOLDEN_FIXTURE_RESPONSES against FIXTURE_SCHEMA

```yaml
- name: "🔍 Validate fixtures against schema"
  run: |
    # Extract JSON examples from GOLDEN_FIXTURE_RESPONSES.md
    # Validate each against corresponding schema in FIXTURE_SCHEMA.jsonc
    # Report compliance on PR
```

**Expected outcome**: PR comments showing fixture compliance status

---

#### 3.2 Create Test Harness for Fixture Verification
**New file**: `docs/fixtures/verify-fixtures.mjs` (Node.js script)  
**Purpose**: Parse GOLDEN_FIXTURE_RESPONSES, extract JSON, validate against FIXTURE_SCHEMA, report issues

**Usage**:
```bash
node docs/fixtures/verify-fixtures.mjs
```

**Expected output**:
```
✅ copilot-cli/bash-success: VALID
✅ copilot-cli/edit-conflict: VALID
...
❌ codex-cli/approval-decision: SCHEMA MISMATCH (missing field: timeout_ms)
```

---

#### 3.3 Document Fixture Maintenance Workflow
**New file**: `docs/fixtures/FIXTURE_MAINTENANCE.md`  
**Content**:
- When to update fixtures (on platform release)
- How to compare old vs new response formats
- Drift detection rules
- Sign-off procedure for fixture changes

---

#### 3.4 Plan Version Capture for Each Platform
**Action**: For each L2 platform, create capture example

**Example (Copilot CLI)**:
```bash
# Step 1: Check if copilot installed
copilot --version

# Step 2: If available, extract version
copilot version --json > /tmp/copilot-version.json

# Step 3: Update VERSION_PINS.jsonc
# capturedVersion: "1.x.y" (from output)

# Step 4: Commit
git add docs/fixtures/VERSION_PINS.jsonc
git commit -m "chore: capture copilot-cli v1.x.y"
```

---

## Phase 4: Release Pin Capture & L3 Promotion (NEXT PHASE)

### Trigger: When any L2 platform has a new release

**For each platform release**:
1. Manually or automatically capture version via documented procedure
2. Update `VERSION_PINS.jsonc` with `capturedVersion`
3. CI validates fixtures and version consistency
4. Create PR: "chore: pin {platform} to v{version}"
5. Reviewer checks:
   - Version is official (GitHub/npm/marketplace)
   - Fixtures match latest response format
   - No breaking changes from previous version
6. **Approval**: Move platform from L2 → L3 in BUILTIN_TOOLS_EVIDENCE_LOCK.md
7. Merge + sign-off with reviewer name + date

---

## Quick Reference: Files & Responsibilities

| File | Purpose | Owner | Update Frequency |
| --- | --- | --- | --- |
| `BUILTIN_TOOLS_EVIDENCE_LOCK.md` | Platform lock status | Reviewer | Per L2→L3 promotion |
| `VERSION_PINS.jsonc` | Captured versions | Developer | On each release |
| `GOLDEN_FIXTURE_RESPONSES.md` | Response payloads | Developer | On each release |
| `FIXTURE_SCHEMA.jsonc` | Validation rules | Architect | On schema changes |
| `verify-fixtures.mjs` | Validation harness | DevOps | On-demand via CI |
| `.github/workflows/validate-evidence-lock.yml` | CI gate | DevOps | Per PR/commit |

---

## How to Advance Platform to L3

**Prerequisite**: Platform at L2 with golden fixtures (all 6 are ready)

**Steps**:

1. **Capture the version**:
   ```bash
   # Run platform-specific capture procedure from VERSION_PINS.jsonc
   copilot version --json
   ```

2. **Update VERSION_PINS.jsonc**:
   ```jsonc
   "copilot-cli": {
     "level": "L2",
     "versionPin": {
       "capturedVersion": "1.45.0",  // ← NEW
       "lastCaptured": "2026-03-20T14:30:00Z"  // ← NEW
     }
   }
   ```

3. **Run CI validation**:
   - Fixtures pass schema validation ✓
   - Version pin format valid ✓
   - Lock rubric consistent ✓

4. **Create promotion PR**:
   ```
   Title: "docs: promote copilot-cli to L3 (v1.45.0)"
   
   Body:
   - Captured version: 1.45.0
   - Fixtures validated: ✅ bash-success, bash-error, edit-conflict, preToolUse-hook
   - Schema compliance: ✅ All fixtures pass FIXTURE_SCHEMA
   - Blocking gaps resolved: ✅ version pin + fixture set complete
   ```

5. **Review & sign-off**:
   - Reviewer verifies version is official
   - Reviewer confirms fixtures match observed behavior
   - Reviewer updates BUILTIN_TOOLS_EVIDENCE_LOCK.md:
     ```md
     ## GitHub Copilot CLI — ✅ Locked (L3, first-party + fixtures)
     - Level: L3
     - Version pin: 1.45.0 (captured 2026-03-20)
     - Fixtures: ✅ All categories complete
     - Reviewed by: @reviewer-name on 2026-03-20
     ```

6. **Merge & done**:
   ```bash
   git push && merge PR
   # Platform now L3!
   ```

---

## Success Metrics

✅ **Phase 3 Complete When**:
- [ ] JSON Schema validator integrated into CI
- [ ] Test harness created and runs green
- [ ] Fixture maintenance workflow documented
- [ ] Version capture examples documented for all L2 platforms

✅ **First L3 Promotion When**:
- [ ] Any L2 platform has captured version
- [ ] All fixtures for that platform pass schema validation
- [ ] PR reviewed and merged with sign-off

✅ **Full L3 Coverage When**:
- [ ] All 6 L2 platforms have captured versions
- [ ] All fixtures validated in CI
- [ ] BUILTIN_TOOLS_EVIDENCE_LOCK.md updated to show 7 L3 platforms

---

## Related Documentation

- [BUILTIN_TOOLS_EVIDENCE_LOCK.md](../BUILTIN_TOOLS_EVIDENCE_LOCK.md) — Lock state and platform details
- [L3_PROMOTION_STRATEGY.md](./L3_PROMOTION_STRATEGY.md) — Detailed promotion criteria and phases
- [GOLDEN_FIXTURE_RESPONSES.md](./GOLDEN_FIXTURE_RESPONSES.md) — Actual response payloads
- [FIXTURE_SCHEMA.jsonc](./FIXTURE_SCHEMA.jsonc) — Validation schemas
- [VERSION_PINS.jsonc](./VERSION_PINS.jsonc) — Version capture registry

---

## Quick Start: Run Validation Now

```bash
# CD to repo
cd c:\Users\hyper\workspace\tormentnexus

# Check fixture structure (manual)
cat docs/fixtures/GOLDEN_FIXTURE_RESPONSES.md | grep -E "^### Tool:|^## Platform:"

# Validate JSONC syntax
python3 -c "import json, re; content=open('docs/fixtures/VERSION_PINS.jsonc').read(); json.loads(re.sub(r'//.*','',content,flags=re.M)); print('✅ VERSION_PINS.jsonc is valid JSON')"

# Next: Create verify-fixtures.mjs to automate schema validation
```

---

## What's Next (Recommended Order)

1. ⏭️ **Create verify-fixtures.mjs** — Automate fixture validation harness
2. ⏭️ **Add schema validation to CI** — Integrate AJV into validate-evidence-lock.yml
3. ⏭️ **Capture first version** — Run version pin capture for any L2 platform
4. ⏭️ **Promote first platform to L3** — Follow sign-off procedure above
5. ✅ **Repeat for all L2 platforms** — Systematic L3 promotion

**Timeline: Phase 3 → L3 ready (1-2 weeks)**

