/**
 * /voices routes — CRUD for the voices table.
 *
 * Defaults to ALL voices (enabled + disabled). Fleet/management surfaces
 * need disabled rows so users can re-enable. Template-dropdown contexts
 * pass `?enabled=true` explicitly.
 *
 * See planning/voices.md for design rationale.
 */
import type { FastifyInstance } from 'fastify';
import { z } from 'zod';
import { voices } from '../../lib/db/index.js';
import {
  successResponse,
  errorResponse,
  listEnvelope,
  sendError,
  type ApiResponse,
  type ListEnvelope,
} from '../api-response.js';

const Lineage = z.enum(['anthropic', 'openai', 'google', 'opencode', 'moonshot']);
const Source = z.enum(['cli', 'api']);

const ListQuerySchema = z.object({
  lineage: Lineage.optional(),
  source: Source.optional(),
  provider: z.string().optional(),
  enabled: z
    .enum(['true', 'false'])
    .optional()
    .transform((v) => (v === undefined ? undefined : v === 'true')),
});

// Cost fields: $/Mtok must be a finite, non-negative number. `null` is a
// distinct legitimate state ("unknown / not set"), `0` is also legitimate
// (free tier, opencode-go subscription, plan-priced reviewers). Negative
// or non-finite (NaN, Infinity) costs would corrupt downstream cost
// dashboards and the run-page chips, so we reject them at the boundary
// rather than letting bad values reach the DB.
const Cost = z.number().finite().min(0).nullable().optional();

const PostBodySchema = z.object({
  provider: z.string().min(1),
  model_id: z.string().min(1),
  label: z.string().min(1),
  source: Source.default('api'),
  lineage: Lineage,
  vendor_family: z.string().nullable().optional(),
  input_cost_per_mtok: Cost,
  output_cost_per_mtok: Cost,
  enabled: z.boolean().optional(),
});

const PutBodySchema = z.object({
  label: z.string().min(1).optional(),
  enabled: z.boolean().optional(),
  input_cost_per_mtok: Cost,
  output_cost_per_mtok: Cost,
});

export function registerVoiceRoutes(fastify: FastifyInstance): void {
  // List voices, optionally filtered. Default = all rows (incl. disabled).
  fastify.get<{
    Querystring: {
      lineage?: string;
      source?: string;
      provider?: string;
      enabled?: string;
    };
    Reply: ApiResponse<ListEnvelope<object>>;
  }>('/voices', async (request, reply) => {
    try {
      const parsed = ListQuerySchema.safeParse(request.query);
      if (!parsed.success) {
        return sendError(reply, 'validation', parsed.error.message);
      }
      const items = await voices.list(parsed.data);
      return successResponse(listEnvelope(items));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('db_error', message);
    }
  });

  // Single-row read.
  fastify.get<{
    Params: { id: string };
    Reply: ApiResponse<object>;
  }>('/voices/:id', async (request, reply) => {
    try {
      const v = await voices.getById(request.params.id);
      if (!v) {
        return sendError(reply, 'not_found', `Voice ${request.params.id} not found`);
      }
      return successResponse(v);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('db_error', message);
    }
  });

  // Add a voice (used by OpenRouter inline + by direct API additions).
  fastify.post<{
    Body: unknown;
    Reply: ApiResponse<object>;
  }>('/voices', async (request, reply) => {
    try {
      const parsed = PostBodySchema.safeParse(request.body);
      if (!parsed.success) {
        return sendError(reply, 'validation', parsed.error.message);
      }
      const id = `${parsed.data.provider}:${parsed.data.model_id}`;
      const existing = await voices.getById(id);
      if (existing) {
        return sendError(reply, 'conflict', `Voice ${id} already exists`);
      }
      const row = await voices.upsert({
        id,
        label: parsed.data.label,
        source: parsed.data.source,
        provider: parsed.data.provider,
        model_id: parsed.data.model_id,
        lineage: parsed.data.lineage,
        vendor_family: parsed.data.vendor_family ?? null,
        input_cost_per_mtok: parsed.data.input_cost_per_mtok ?? null,
        output_cost_per_mtok: parsed.data.output_cost_per_mtok ?? null,
        enabled: parsed.data.enabled,
      });
      return successResponse(row);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('db_error', message);
    }
  });

  // Toggle enabled / update label / update costs.
  // source/provider/lineage/vendor_family/model_id are immutable post-create.
  // Note: model_id IS rewritten on cli rows during the seed loop, but only
  // by the seed (via voices.upsert), not via this route.
  fastify.put<{
    Params: { id: string };
    Body: unknown;
    Reply: ApiResponse<object>;
  }>('/voices/:id', async (request, reply) => {
    try {
      const parsed = PutBodySchema.safeParse(request.body);
      if (!parsed.success) {
        return sendError(reply, 'validation', parsed.error.message);
      }
      const existing = await voices.getById(request.params.id);
      if (!existing) {
        return sendError(reply, 'not_found', `Voice ${request.params.id} not found`);
      }
      const row = await voices.update(request.params.id, parsed.data);
      return successResponse(row);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('db_error', message);
    }
  });

  // DELETE allowed for both cli + api source. cli rows auto-heal on next
  // seed if the model is still detected by the gateway (per gem-2 round 1
  // MED — unblocks cleanup of deprecated OpenCode models).
  fastify.delete<{
    Params: { id: string };
    Reply: ApiResponse<object>;
  }>('/voices/:id', async (request) => {
    try {
      await voices.delete(request.params.id);
      return successResponse({ id: request.params.id, deleted: true });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('db_error', message);
    }
  });
}
