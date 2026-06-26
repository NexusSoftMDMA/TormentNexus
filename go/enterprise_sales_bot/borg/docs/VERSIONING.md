# tormentnexus Versioning Checklist

Use this checklist any time tormentnexus version changes (e.g. `0.9.1` -> `0.9.2`).

## Source of truth

1. `VERSION` (canonical runtime-friendly version string)
2. Root `package.json` -> `version`

Keep these two values identical.

## User-visible/version-coupled files to verify

- `VERSION.md` (release notes/version narrative)
- `CHANGELOG.md` (release entry)
- `README.md` title/version references (if versioned text is present)
- `apps/web/src/app/dashboard/page.tsx`
  - Dashboard version badge is sourced from `VERSION`; no manual edit needed unless plumbing changes.

## Release hygiene checks (required)

1. Search for stale literals before commit:
   - previous version string (example: `0.9.1`)
2. Confirm there are no hardcoded version badges in UI components.
3. Build and typecheck after version bump.

## Quick process

1. Update `VERSION`
2. Update root `package.json` `version`
3. Update `VERSION.md` and `CHANGELOG.md`
4. Run workspace search for old version literal
5. Run build/typecheck and fix any references

## Anti-regression note

If a new UI surface displays version text, source it from `VERSION` (or a single adapter that reads `VERSION`) instead of hardcoding a literal in component markup.
