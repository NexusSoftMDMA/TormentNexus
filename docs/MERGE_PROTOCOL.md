# 7-Step Merge & Assimilation Protocol

> Standardized procedure for handling upstream forks, submodule updates, and feature integrations in the tormentnexus monorepo.

## The 7 Steps

### Step 1: Merge / Update
Selectively merge feature branches, update submodules, or pull upstream changes.

```bash
# For submodules
git submodule update --remote <path>

# For upstream forks
git fetch upstream
git merge upstream/main --allow-unrelated-histories  # if needed
```

**Resolve conflicts carefully** — prefer the tormentnexus-side implementation when in doubt. If upstream merges fail with "unrelated histories," identify the exact reason and formulate a fallback strategy (cherry-pick, manual merge, or archive).

### Step 2: Reanalyze
Reanalyze the project and history for missing features, regressions, or new capabilities introduced by the merge.

- Run `pnpm run build` in `apps/web` to catch import breaks
- Run `pnpm run test:ci` for test regressions
- Scan for new files, APIs, or patterns that should be integrated

### Step 3: Update Roadmap & Docs
Comprehensively update:
- `ROADMAP.md` — adjust long-term goals based on new capabilities
- `TODO.md` — add/remove tasks based on merged features
- `VISION.md` — update if the merge significantly shifts the project direction

### Step 4: Update Submodule Dashboard
Update `docs/SUBMODULES.md` and run the submodule version checker:

```bash
pnpm submodules:check
node scripts/generate_submodule_index.js
```

Verify the dashboard at `/dashboard/submodules` reflects the updated state.

### Step 5: Version Bump
Bump the version in `VERSION` and add a detailed entry to `CHANGELOG.md`:

```bash
echo "X.Y.Z" > VERSION
```

Update `CHANGELOG.md` with:
- What was merged/updated
- What features were gained
- What conflicts were resolved
- What breaking changes exist (if any)

### Step 6: Commit & Push
Stage all changes and commit with a descriptive message:

```bash
git add -A
git commit -m "merge: assimilate <component> — <summary>"
git push
```

### Step 7: Verify / Redeploy
Verify the deployment or run the release gate:

```bash
pnpm run check:release-gate:ci
```

If deploying to production, follow the instructions in `DEPLOY.md`.

---

## Quick Reference

| Step | Action | Key Command |
|------|--------|-------------|
| 1 | Merge / Update | `git submodule update --remote` |
| 2 | Reanalyze | `pnpm run build && pnpm run test:ci` |
| 3 | Update Docs | Edit `ROADMAP.md`, `TODO.md`, `VISION.md` |
| 4 | Submodule Dashboard | `pnpm submodules:check` |
| 5 | Version Bump | `VERSION` + `CHANGELOG.md` |
| 6 | Commit & Push | `git add -A && git commit && git push` |
| 7 | Verify | `pnpm run check:release-gate:ci` |
