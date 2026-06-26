---
title: Wiki Maintenance Workflow
type: note
permalink: scarf/ops/wiki-maintenance-workflow
tags:
- docs
- wiki
---

## Observations
- [location] Public docs at https://github.com/awizemann/scarf/wiki — separate git repo cloned to .wiki-worktree/ in repo root (gitignored, sibling to .gh-pages-worktree/) #paths
- [scope] Wiki is public-facing reference; internal dev notes stay in scarf/docs/ #scope
- [workflow] Standard cycle: `./scripts/wiki.sh pull` → edit .wiki-worktree/*.md → `./scripts/wiki.sh commit "docs: …"` → `./scripts/wiki.sh push`. Both commit and push run a secret-scan #workflow
- [security] NEVER commit API keys, tokens, .env files, private keys, or real hostnames/IPs to the wiki. Two-pass secret-scan blocks common patterns + user blocklist at scripts/wiki-blocklist.txt (gitignored). Do not bypass without explicit approval #security #rule
- [update-trigger] Update wiki when: new feature module added → relevant User Guide page; new core service → Core-Services.md; architecture changes → Architecture-Overview.md + sub-page; Hermes version bumps → Hermes-Version-Compatibility.md; non-draft release → bump Home.md latest version + append to Release-Notes-Index.md; keyboard shortcut/sidebar changes → those pages #triggers
- [skip] Skip wiki updates for: bug fixes with no user-observable change, pure refactors, typos, test-only changes, internal cleanups #triggers

## Relations
- complements [[Build and Release Workflow]]