# Github Copilot Instructions

> **CRITICAL**: Read `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` first. It contains the mandatory rules for all AI agents working on borg.

## Copilot-Specific Directives

### 1. Role Context
You are acting as an **inline assistant** and localized pair programmer for the Borg operator.

### 2. Code Style
- Use `pnpm` v10 for package management
- In `apps/web`, import shared UI from `@borg/ui`
- Use `lucide-react` for icons
- Prefer type-safe fixes over `any`, `@ts-ignore`, or misleading placeholder adapters
- For UI components, ensure proper SSR hydration handling (Next.js 16)

### 3. Patterns
- **tRPC routers**: Define in `packages/core/src/routers/`, register in `packages/core/src/trpc.ts`
- **REST bridges**: Add in `packages/core/src/orchestrator.ts` after the health endpoint
- **Go handlers**: Add in `go/internal/httpapi/`, register in `go/internal/httpapi/server.go`
- **Dashboard pages**: Add in `apps/web/src/app/dashboard/<name>/page.tsx`

### 4. Database
- SQLite via `better-sqlite3` — requires rebuild after `pnpm install` on Node 24
- Schema in `packages/core/src/db/`
- Repositories follow the `*Repository` pattern

### 5. Do Not
- Override architectural patterns through inline suggestions
- Introduce new dependencies without checking if they exist in the monorepo
- Assume mock data is acceptable — dashboard pages must show real state
1. **Role Context**: You are acting as an inline assistant and localized pair programmer for the borg operator.
2. **Methodology**:
   - Provide highly contextual, concise autocomplete suggestions.
   - Respect the established code style in the currently active file.
   - For UI components, utilize `@borg/ui` and `lucide-react`.
   - For backend files, respect `better-sqlite3` limitations and `TRPC` routing norms.
3. **Synergy**: Assume the operator is navigating a broader plan defined by autonomous agents (Claude/Gemini/GPT). Do not attempt to override large architectural patterns through inline suggestions.

*Keep this file scoped strictly to Copilot inline behaviors. Universal architectural rules belong in `docs/UNIVERSAL_LLM_INSTRUCTIONS.md`.*
