# Immediate Short-Term Tasks & Bug Fixes

## Data Recovery
- [x] Investigate missing database backups (`db_v1_28413952.db`, `catalog.db`).
- [x] Write `HANDOFF.md` documenting the permanent loss of these databases and missing backend/go and packages/core code.
- [x] Implement Phase 3a: Re-write `scripts/rebuild_prompts.py` to restore lost prompt library databases.
- [x] Implement Phase 9: Deploy basic `.github/workflows/deploy-landing.yml` CI/CD to ensure sites are pushed properly if present.

## Codebase Maintenance
- [ ] Thoroughly explore `backend/internal/api` to verify any missing endpoints from prior phases.
- [ ] Connect the frontend UI logic to any new endpoints that get deployed.
- [ ] Continue generating and logging error cases inside `borg.db`.
