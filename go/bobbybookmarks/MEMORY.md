# Memory Observations

## Architectural Observations
- The codebase leverages dual-backend structure: Python handles background intelligence harvesting and database enrichment (`deep_research.py`, workers), while a Go backend provides high-speed API endpoints.
- SQLite is the preferred storage format, split across specialized domains (`bookmarks.db`, `atlas.db`, `borg.db`, `tormentnexus.db`).
- All sub-modules should seamlessly share configurations via unified `*.db` formats.
- Due to a data loss event, some intended system parts like `catalog.db` and the `.tormentnexus/skills/` tree were lost from the root workspace and required a documented handoff.
- The `go` implementation was verified to compile safely without vcs checks.

## Codebase Traits
- **Data Hygiene First:** Scripts frequently implement deduplication algorithms (like Jaccard) to prevent database bloat.
- **Fail-Safe Operation:** "Don't Stop the Party" ensures processes are highly resilient and autonomous.
- **Unified Standard:** Important versionings should always source back to `VERSION.md`.

## Design Preferences
- Heavy logging and documentation inside standard markdown formats (`VISION.md`, `ROADMAP.md`, `TODO.md`).
- Scripts should be contained in a dedicated `scripts/` folder or root based on execution context.
