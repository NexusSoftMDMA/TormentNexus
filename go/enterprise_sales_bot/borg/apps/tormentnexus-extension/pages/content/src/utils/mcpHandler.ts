/**
 * DEPRECATED: Legacy MCP Handler
 *
 * This file previously contained a monolithic MCP handler (~1000 lines) that has been
 * fully replaced by the new modular architecture:
 *
 * - Connection management → `services/mcp-client.ts` + `stores/connection.store.ts`
 * - Tool execution         → `stores/tool-execution.store.ts`
 * - Message handling       → `stores/app.store.ts` + `hooks/useMcpCommunication.ts`
 * - Reconnection logic     → `stores/connection.store.ts` (with event bus integration)
 *
 * The original code was removed in v0.7.2 to reduce codebase noise.
 * If you need to reference the old implementation, check git history:
 *   git log --follow -p -- pages/content/src/utils/mcpHandler.ts
 *
 * @deprecated Use the modular services/stores architecture instead.
 */
