# Gemini-Specific Instructions

> **Base**: Read [AGENTS.md](./AGENTS.md) first. This file contains Gemini-specific addenda.

## Coding Style
- Same as universal standards: functional React, TypeScript strict, Tailwind CSS, shadcn/ui.
- When generating code, always include proper TypeScript types — never leave parameters untyped.
- Prefer `const` assertions and `as const` for read-only data structures.

## Planning
- When asked to "reanalyze", review all documentation files in `docs/`, the `CHANGELOG.md`, and the `VERSION` file.
- Use task boundaries to communicate progress clearly.
- Research thoroughly before implementing — check existing hooks, stores, and utils to avoid duplication.

## Tool Calling
- Prefer `view_file_outline` for initial file exploration.
- Use `find_by_name` and `grep_search` to locate relevant code before editing.
- Run `pnpm type-check` after making TypeScript changes to verify compilation.

## Commit Style
- Use Conventional Commits: `feat:`, `fix:`, `chore:`, `docs:`.
- Include version bump in commit messages.
- Commit and push after each completed feature; do not batch unrelated changes.
- **Always run `pnpm build` before pushing** to verify the extension compiles cleanly (12/12 tasks).

## Autonomous Operation
- When instructed to "keep going" or "continue", proceed through the roadmap items sequentially.
- Fix errors encountered during development and document them in the CHANGELOG.
- If uncertain about a direction, document the question in the implementation plan rather than guessing.
