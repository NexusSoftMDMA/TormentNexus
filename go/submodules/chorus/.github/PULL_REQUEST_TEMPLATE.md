<!--
Thanks for the PR. Quick checklist:

- One concern per PR. Bug fix vs refactor vs feature should be separate.
- Tests added/updated where relevant.
- pnpm typecheck && pnpm lint && pnpm test all green.
- Conventional commit subject (feat: / fix: / refactor: / docs: / chore:).

If you ran chorus on your own diff (we dogfood) — paste the verdict
into the "Multi-LLM review" section below.
-->

## What changed

<!-- One paragraph. The "why" matters more than the "what" — git diff
shows the what. -->

## Why

<!-- The problem this solves, the constraint it satisfies, or the bug
it fixes. Link the issue if there is one. -->

## How to verify

<!-- Concrete steps a reviewer can run. Tests count, but a manual
repro path is gold for UI changes. -->

## Multi-LLM review (optional but encouraged)

<!-- chorus init && chorus start, run the `code-review` template on
this branch, paste the consensus verdict here. Fan-out diversity
catches blind spots a single reviewer misses. -->

## Checklist

- [ ] Tests added/updated
- [ ] `pnpm typecheck` clean
- [ ] `pnpm lint` clean
- [ ] `pnpm test` green
- [ ] README / docs updated if user-visible
- [ ] No unrelated drive-by changes
