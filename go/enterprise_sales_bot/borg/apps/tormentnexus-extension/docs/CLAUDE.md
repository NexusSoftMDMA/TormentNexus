# Claude-Specific Instructions

> **Base**: Read [AGENTS.md](./AGENTS.md) first. This file contains Claude-specific addenda.

## Coding Style
- Prefer functional React components with Hooks.
- Use `lucide-react` for all icons — never import from other icon libraries.
- Use `shadcn/ui` components from `packages/ui` where possible.
- **Strictly** follow TypeScript types; avoid `any`. Use `unknown` with type guards.
- Use `cn()` from `@src/lib/utils` for conditional classNames.

## Planning
- When asked to "reanalyze", read the entire conversation history and all documentation files.
- Break down large tasks into atomic sub-tasks and track them in a task checklist.
- Always create an implementation plan before making changes to more than 3 files.

## Tool Calling
- Use `view_file_outline` before reading large files to understand structure first.
- Use `grep_search` to locate specific patterns rather than reading entire files.
- Make parallel tool calls whenever dependencies allow.

## Commit Style
- Use Conventional Commits: `feat:`, `fix:`, `chore:`, `docs:`.
- Include version bump reference in commit messages when version changes.
- Commit after each logical unit of work, not after every single file edit.
- **Always run `pnpm build` before pushing** to verify the extension compiles cleanly (12/12 tasks).

## Autonomous Operation
- When instructed to "keep going" or "continue", proceed through the roadmap items sequentially.
- Fix errors encountered during development and document them in the CHANGELOG.
- Commit, push, and continue to the next task without pausing when possible.
- If uncertain about a direction, document the question rather than guessing.
