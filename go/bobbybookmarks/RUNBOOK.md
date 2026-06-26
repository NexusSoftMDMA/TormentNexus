# Borg Intelligence Harvester (BobbyBookmarks) - RUNBOOK

## Overview
This runbook provides actionable instructions for agents or developers tasked with operating, maintaining, or expanding the Borg Intelligence Harvester system. The system automates the ingestion, LLM-powered enrichment, and data hygiene of AI/developer tools.

For deeper architectural context, please refer to the core governance documents: `VISION.md`, `ROADMAP.md`, `MEMORY.md`, and `CHANGELOG.md`.

---

## 1. Routine Operations & Execution

**The Golden Rule: "Don't Stop the Party"**
Never terminate active background processes (`deep_research.py`, etc.) unless instructed to do so or if diagnosing a critical failure. The system is designed for continuous, autonomous execution.

### Ingestion Pipeline
The ingestion pipeline pulls URLs from `incoming_resources.txt`, feeds them to the LLM-powered parser, and stores them safely in SQLite (`atlas.db`, `borg.db`, etc.).

* **To run the ingest manually:**
  ```bash
  python3 deep_research.py &
  ```
  *(Note: Run in background to preserve autonomy)*

* **To retry failed URLs:**
  Clear the `failed_bookmarks.txt` before restarting the script.

### Data Deduplication & Hygiene
SQLite databases are used extensively (`bookmarks.db`, `atlas.db`, `borg.db`, `tormentnexus.db`, `metamcp.db`).
* Data deduplication is enforced by standard scripts.
* *Example (Prompts)*: If re-indexing external prompts, use Jaccard similarity at a 90% threshold.
  ```bash
  python3 scripts/rebuild_prompts.py
  ```

---

## 2. API / Backend Operations

The Go Backend serves the high-speed logic routing for the extracted tools.

### Building & Testing
Always build and test before committing new changes to ensure zero regressions.

* **Build the Go API:**
  ```bash
  cd backend
  go build -buildvcs=false ./cmd/api
  ```

* **Run Backend Tests:**
  ```bash
  cd backend
  go test -buildvcs=false ./internal/...
  ```

---

## 3. Database Maintenance & Verification

If the system crashes or a new session is resumed, always check the integrity of the core databases.

* **Run Integrity Checks:**
  ```bash
  sqlite3 bookmarks.db "PRAGMA integrity_check;"
  sqlite3 atlas.db "PRAGMA integrity_check;"
  sqlite3 borg.db "PRAGMA integrity_check;"
  sqlite3 tormentnexus.db "PRAGMA integrity_check;"
  ```
  *If a database fails the integrity check or is entirely missing, refer to `HANDOFF.md` for historical data loss context and manually reconstruct what is viable programmatically.*

---

## 4. Deployment

### Static Landing Pages
The frontend sites (`tormentnexus.site`, `hypernexus.site`) are deployed using Cloudflare Pages.
* **Trigger Deployment:** Pushing any changes to the `landing/` directory on the `main` branch will automatically trigger the `.github/workflows/deploy-landing.yml` action.

### Backend Infrastructure
Python environments should be synchronized via:
```bash
pip install -r requirements.txt
```
*(Flask backends require `application = create_app()` for Gunicorn/Render compatibility).*

---

## 5. Session Resumption Protocol

When a new autonomous agent session begins:
1. Parse this `RUNBOOK.md` along with `TODO.md` and `ROADMAP.md` to identify the most impactful open feature.
2. Execute `git pull` recursively on submodules (if applicable) and fetch upstream branches to catch-up cleanly.
3. Validate database integrity and run backend tests (`cd backend && go test...`).
4. Implement the feature.
5. Update `CHANGELOG.md` and bump `VERSION.md`.
6. Submit the patch.
