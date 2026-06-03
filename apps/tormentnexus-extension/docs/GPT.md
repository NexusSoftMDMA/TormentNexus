# GPT-Specific Instructions

> **Base**: Read [AGENTS.md](./AGENTS.md) first. This file contains GPT/OpenAI-specific addenda.
> **Codex 5.3**: These instructions also apply to you. Follow the same coding style and commit conventions.

## Coding Style
- Same as universal standards: functional React, TypeScript strict, Tailwind CSS, shadcn/ui.
- Always provide complete function signatures with return types.
- Use descriptive variable names — avoid single-letter variables except in trivial lambdas.

## Planning
- When asked to "reanalyze", read all docs in `docs/`, `CHANGELOG.md`, and the `VERSION` file.
- Break tasks into small, verifiable steps.
- Always describe what you're about to do before doing it.

## Tool Usage
- ChatGPT / GPT models are often used via the TormentNexus Extension extension itself.
- ChatGPT / GPT models are often used via the tormentnexus Extension extension itself.
- When writing adapter code for OpenAI/ChatGPT, reference `pages/content/src/plugins/adapters/` for patterns.

## Commit Style
- Use Conventional Commits: `feat:`, `fix:`, `chore:`, `docs:`.
- Include version bump in commit messages.
- **Always run `pnpm build` before pushing** to verify the extension compiles cleanly (12/12 tasks).

## Autonomous Operation
- When instructed to "keep going" or "continue", proceed through the roadmap items sequentially.
- Fix errors encountered during development and document them in the CHANGELOG.
- Commit, push, and continue to the next task without pausing when possible.

## Context Window Management
- For large files (>500 lines), focus on the specific function or section being modified.
- Reference line numbers when describing changes.
- Avoid rewriting entire files when only a small section needs updating.
