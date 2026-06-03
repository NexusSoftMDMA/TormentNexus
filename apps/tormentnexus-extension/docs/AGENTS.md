# AGENTS — `apps/tormentnexus-extension/docs`
# AGENTS — `apps/tormentnexus-extension/docs`

**Mandatory reading first:**
- `docs/UNIVERSAL_LLM_INSTRUCTIONS.md`
- root `AGENTS.md`

This file contains only guidance for documentation under `apps/tormentnexus-extension/docs`.
This file contains only guidance for documentation under `apps/tormentnexus-extension/docs`.

## Purpose

These docs should describe the browser extension honestly and specifically. They must not overstate implementation maturity or imply parity that the extension does not yet have.

## Documentation rules for this folder

- Prefer concrete behavior over aspirational claims.
- Use the shared maturity labels: **Stable**, **Beta**, **Experimental**, **Vision**.
- Distinguish clearly between:
  - current extension behavior,
  - planned integration work,
  - and broader TormentNexus vision.
  - and broader tormentnexus vision.
- Do not require version bumps or broad project-wide release rituals for routine doc edits.

## What to emphasize

- actual extension capabilities,
- setup and troubleshooting guidance,
- bridge limitations,
- storage and runtime constraints,
- and any user-visible reliability caveats.

## What to avoid

- “ultimate AI operating system” framing as present-day fact,
- vague promises without implementation backing,
- and policy that conflicts with the root stabilization-first instructions.

## Validation

When updating docs in this folder, verify they remain consistent with:
- root `README.md`
- root `VISION.md`
- root `ROADMAP.md`
- root `AGENTS.md`

If extension behavior changed materially, update the relevant canonical docs too.
