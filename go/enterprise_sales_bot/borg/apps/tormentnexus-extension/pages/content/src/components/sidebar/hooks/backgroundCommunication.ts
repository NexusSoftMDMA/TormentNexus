/**
 * DEPRECATED: Legacy Background Communication Hook
 *
 * This file previously contained a monolithic React hook (~900 lines) for communicating
 * with the Chrome extension background script. It has been fully replaced by:
 *
 * - `hooks/useMcpCommunication.ts` — Active communication hook
 * - `stores/connection.store.ts`   — Connection state management
 * - `stores/tool-execution.store.ts` — Tool execution pipeline
 * - `services/mcp-client.ts`       — Low-level MCP client
 *
 * The original code was removed in v0.7.2 to reduce codebase noise.
 * If you need to reference the old implementation, check git history:
 *   git log --follow -p -- pages/content/src/components/sidebar/hooks/backgroundCommunication.ts
 *
 * @deprecated Use useMcpCommunication hook instead.
 */

import { createLogger } from '@extension/shared/lib/logger';

const logger = createLogger('useBackgroundCommunication');
