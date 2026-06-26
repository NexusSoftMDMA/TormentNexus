# Hermes Add-ons Area

## Scope
Discover, ingest, and integrate the **top 100 most high-value Hermes-agent add-ons** into the Go codebase.

## Workflow
1. **Search** the web (or curated list) for the top 100 Hermes add-ons.
2. **Clone / Download** each add-on (usually a small TypeScript/Python repo).
3. **Wrap** the add-on as a native Go module (or skill) in `internal/tools/` or `internal/skills/`.
4. **Register** the new module/skill in the Skill Registry (if applicable).
5. **Verify** functionality with a smoke test.
6. **Document** usage in `docs/hermes-addons/`.

## Conventions
- Keep each add-on in its own Go package: `internal/tools/hermes_<addon_name>`.
- Update `go.mod` for any new dependencies.
- If an add-on is already a pure skill, register it in the Skill Registry instead of a full tool.
- Prefer Go-native re-implementation over wrapping an external binary.

## Success Criteria
- All 100 add-ons are available without external process dependencies.
- Each is either a Go tool or a registered skill.
- Documentation is present and searchable.

