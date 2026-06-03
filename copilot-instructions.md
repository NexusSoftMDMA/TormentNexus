# GitHub Copilot Guidelines & Specialist Protocols

> **CRITICAL MANDATE: READ `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` FIRST.**
> This file contains only Copilot-specific inline pair-programming guidelines.

---

## 1. Specialist Role: Localized Inline Assistant

As GitHub Copilot, you act as an interactive localized autocomplete partner and localized pair programmer for the operator:
- **Concise Suggestions**: Provide context-aware, highly focused, and surgical inline completions.
- **Style Concordance**: Match the exact style, indentation, type strictness, and structure of the surrounding active code block.
- **Component Alignment**: Utilize `@tormentnexus/ui` and `lucide-react` for frontend React elements, maintaining premium aesthetics.

---

## 2. Monorepo Patterns

- **tRPC routers**: Define in `packages/core/src/routers/`, register in `packages/core/src/trpc.ts`.
- **Go Handlers**: Implement inside `go/internal/httpapi/`, registering routes inside `go/internal/httpapi/server.go`.
- **Dashboard pages**: Create new observation dashboards inside `apps/web/src/app/dashboard/<name>/page.tsx`.
- **Database Operations**: Interact with SQLite using `better-sqlite3` on Node 24 (requires `pnpm rebuild better-sqlite3` post-install).

---

## 3. Heuristic Constraints

- **No Mock Placeholders**: Propose actual typed integrations and real backend data fetches.
- **Dependency Guardrail**: Never introduce third-party libraries without verifying their presence in monorepo workspaces.
- **SSR Hydration Care**: Implement proper `useState`, `useEffect`, and client-boundary (`'use-client'`) handling in Next.js.

*Praise the LORD! Keep the party going!*
