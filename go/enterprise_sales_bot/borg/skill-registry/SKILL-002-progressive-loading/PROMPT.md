# Task: SKILL-002 - Implement 3-Tier Progressive Skill Loading

## Goal
Add deferred/progressive loading to `go/internal/tools/skill_registry.go` so the system
doesn't dump all 1,489 skill contents into context. Load skills in three tiers:

## Tier Architecture

### Tier 1 — Name Manifest (always loaded at startup)
- Just: `id`, `name`, `category`
- Loaded once, cached 5 minutes
- Tool: `skill_manifest` → returns JSON array of all skill names

### Tier 2 — Frontmatter (loaded on context relevance match)
- Fields: `id`, `name`, `description`, `frontmatter` (first 800 chars), `category`
- Triggered when conversation keywords match skill name/description
- Tool: `skill_search` → takes `query` string, returns top 20 matching summaries

### Tier 3 — Full Content (loaded only on explicit invocation)
- Complete `content` field
- Only loaded when model explicitly calls `skill_get` with a skill name
- Tool: `skill_get` → takes `name` or `id`, returns full skill content

## Implementation Requirements

Add to `skill_registry.go`:
1. `SkillManifest` struct — Tier 1 data
2. `SkillSummary` struct — Tier 2 data (embeds SkillManifest)
3. `HandleSkillManifest(ctx, args)` — returns all Tier 1 data
4. `HandleSkillSearch(ctx, args)` — keyword search returning Tier 2 data, limit 20
5. `HandleSkillGet(ctx, args)` — returns full Tier 3 data for named skill

In-memory cache for Tier 1 manifest (5 minute TTL using `time.Time`).

## Keyword Relevance Scoring
In `HandleSkillSearch`, rank results by:
- Exact name match: +100
- Query word in name: +50 per word
- Query word in description: +20 per word
- Query word in category: +10

## Register in registry.go
```go
r.handlers["skill_manifest"] = HandleSkillManifest
r.handlers["skill_search"]   = HandleSkillSearch
r.handlers["skill_get"]      = HandleSkillGet
```

## Tests Required
Add to `skill_registry_test.go`:
- TestSkillManifest: verifies all skills returned without content
- TestSkillSearch: verifies keyword matching returns correct results
- TestSkillGet: verifies full content returned for valid name

## Compile & Test
```
cd C:\Users\hyper\workspace\tormentnexus
go build ./go/...
go test ./go/internal/tools/... -run TestSkill -v
```

## Commit
`git add -A && git commit -m "feat(skills): 3-tier progressive skill loading system [SKILL-002]"`
`git push origin main`
