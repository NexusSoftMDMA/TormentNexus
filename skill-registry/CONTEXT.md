# Skill Registry Area

## Scope
Build and maintain an internal skill registry database that:
1. **Stores** all skills as JSON objects.
2. **Dedupes** skills that are ≥ 90 % similar (content‑based comparison).
3. **Implements** progressive loading: only load a skill *when* it is requested.
4. **Integrates** with the Taskplane worker and supervisor.

## Workflow
- **Scan** existing `.skill.md` files in the repository or a designated `skills/` folder.
- **Hash** content with SHA‑256; compare against existing hashes, merge if similarity > 90 %.
- **Persist** the registry in a lightweight SQLite (or JSON) file in `.pi/db/skills.db`.
- **Expose** a Go API at `internal/skillregistry` that can retrieve a skill’s metadata or body.
- **Hook** into Taskplane’s loading logic to lazily load skill content based on the tool‑prediction or history.

## Success Criteria
- All pre‑existing skills are in the registry.
- Duplicate skills are unified.
- Lazy loading works; a new skill is only read from disk when first invoked.

