import { TRPCError } from '@trpc/server';
import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../../lib/trpc-core.js';
import { sessionRouter } from './sessionRouter.js';
import { councilRouter as baseCouncilRouter } from './councilRouter.js';
import { quotaRouter } from './quotaRouter.js';
import { historyRouter } from './historyRouter.js';
import { visualRouter } from './visualRouter.js';
import { smartPilotRouter } from './smartPilotRouter.js';
import { evolutionRouter } from './evolutionRouter.js';
import { hooksRouter } from './hooksRouter.js';
import { fineTuneRouter } from './fineTuneRouter.js';
import { ideRouter } from './ideRouter.js';
import { rotationRouter } from './rotationRouter.js';
import { readFileSync, writeFileSync, existsSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

/** Schema for a council member entry from config/council.json */
const memberSchema = z.object({
  name: z.string(),
  provider: z.string(),
  modelId: z.string(),
  systemPrompt: z.string(),
});

/**
 * Resolves the path to config/council.json relative to this module.
 * Path: packages/core/config/council.json
 */
function getCouncilConfigPath(): string {
  return join(
    dirname(fileURLToPath(import.meta.url)),
    '..', '..', '..', 'config', 'council.json'
  );
}

/**
 * Council router — the central tRPC namespace for the TormentNexus Council of Agents.
 *
 * Top-level procedures:
 *   - `members` — reads agent definitions from config/council.json
 *   - `updateMembers` — persists reordered agent hierarchy to config/council.json
 *
 * Namespaced sub-routers:
 *   - `sessions` — PTY session management for council supervisors
 *   - `base` — council orchestrator (debate, status, config, supervisors)
 *   - `quota` — quota-aware routing and budget controls
 *   - `history` — historical session list
 *   - `visual` — visualization data for debate UI
 *   - `smartPilot` — autonomous pilot integration
 *   - `evolution` — Darwin-style mutation experiments
 *   - `hooks` — event hook registrations
 *   - `fineTuning` — fine-tuning experiment management
 *   - `ide` — IDE extension hooks
 *   - `rotation` — shared-context rotating multi-model rooms
 */
export const councilRouter = t.router({
  /**
   * members — Returns council agent definitions from config/council.json.
   * Used by the providers/routing dashboard page to render the fallback hierarchy.
   */
  members: publicProcedure
    .output(z.array(memberSchema))
    .query(() => {
      try {
        const configPath = getCouncilConfigPath();
        const raw = readFileSync(configPath, 'utf-8');
        const parsed = JSON.parse(raw);
        return parsed.members || [];
      } catch (error) {
        const errorCode = (error as NodeJS.ErrnoException | undefined)?.code;
        if (errorCode === 'ENOENT') {
          return [];
        }

        const detail = error instanceof SyntaxError
          ? 'council.json contains invalid JSON.'
          : error instanceof Error
            ? error.message
            : String(error);

        throw new TRPCError({
          code: 'INTERNAL_SERVER_ERROR',
          message: `Council configuration is unavailable: ${detail}`,
        });
      }
    }),

  /**
   * updateMembers — Persists the drag-and-drop reordered provider hierarchy.
   * Called from providers/routing page when the operator saves route priorities.
   */
  updateMembers: adminProcedure
    .input(z.array(memberSchema))
    .mutation(async ({ input }) => {
      const configPath = getCouncilConfigPath();
      let config: any = { members: [] };
      try {
        if (existsSync(configPath)) {
          config = JSON.parse(readFileSync(configPath, 'utf-8'));
        }
      } catch (e) {
        console.warn('[councilRouter] Could not parse existing council.json:', e);
      }
      config.members = input;
      writeFileSync(configPath, JSON.stringify(config, null, 2));
      return { success: true };
    }),

  // --- Namespaced sub-routers ---
  sessions: sessionRouter,
  base: baseCouncilRouter,
  quota: quotaRouter,
  history: historyRouter,
  visual: visualRouter,
  smartPilot: smartPilotRouter,
  evolution: evolutionRouter,
  hooks: hooksRouter,
  fineTuning: fineTuneRouter,
  ide: ideRouter,
  rotation: rotationRouter,
});
