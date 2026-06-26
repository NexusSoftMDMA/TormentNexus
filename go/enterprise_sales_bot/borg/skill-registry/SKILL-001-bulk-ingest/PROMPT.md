# Task: SKILL-001 - Bulk Ingest 1,489 Agent Skills Into SQLite Registry

## Goal
Load every skill from `C:\Users\hyper\workspace\.agent\skills\` (1,489 directories)
into the SQLite skill registry at `go/internal/tools/skills.db`, with 98% Jaccard
similarity deduplication (near-identical skills merged as revisions, not duplicates).

## Key Rules
- 98%+ Jaccard word similarity = same skill, store as revision (increment version, keep longer content)
- 75-97% = near-duplicate, insert but set canonical_id pointing to most similar existing
- <75% = genuinely different, insert fresh

## Script Location
Write script to: `skill-registry/ingest_skills.py`

## DB Schema
```sql
CREATE TABLE IF NOT EXISTS skills (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT DEFAULT '',
    category TEXT DEFAULT 'general',
    frontmatter TEXT DEFAULT '',      -- first 800 chars, for progressive loading
    content TEXT DEFAULT '',          -- full skill content, loaded on invoke only
    version INTEGER DEFAULT 1,
    similarity_score INTEGER DEFAULT 100,
    canonical_id INTEGER REFERENCES skills(id),
    status TEXT DEFAULT 'active',     -- active|merged|superseded
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Performance Notes
- 1,489 skills x 1,489 comparisons = 2.2M pairs. Too slow.
- Use random sample of 200 existing skills for each new skill's dedup check.
- Commit to DB every 100 records to avoid memory issues.
- Total runtime should be < 5 minutes.

## Expected Output
```
============ SKILL INGESTION COMPLETE ============
Total skill directories: 1489
Inserted (new):          ~1300
Merged (98%+ duplicate): ~150
Skipped (no SKILL.md):   ~30
Errors:                  0
Active skills in DB:     ~1300
```

## After Running
Verify: `python -c "import sqlite3; c=sqlite3.connect('go/internal/tools/skills.db'); print(c.execute('SELECT COUNT(*) FROM skills').fetchone())"`

Commit: `git add go/internal/tools/skills.db skill-registry/ && git commit -m "feat(skills): bulk ingest 1489 skills into SQLite registry [SKILL-001]"`
