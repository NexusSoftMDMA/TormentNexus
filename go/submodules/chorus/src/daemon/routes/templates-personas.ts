/**
 * /templates and /personas routes.
 *
 * These are pure DB CRUD — no coupling to the runner, the singleton
 * activeRuns map, or the SSE multiplex. Extracted out of index.ts as
 * part of the post-audit refactor that brought the daemon entry file
 * under 700 lines.
 */
import type { FastifyInstance } from 'fastify';
import yaml from 'yaml';
import { templates } from '../../lib/db/index.js';
import { validateTemplateYaml } from '../../lib/template-validation.js';
import {
  successResponse,
  errorResponse,
  listEnvelope,
  type ApiResponse,
  type ListEnvelope,
} from '../api-response.js';

export function registerTemplateRoutes(fastify: FastifyInstance): void {
  // List all templates
  fastify.get<{ Reply: ApiResponse<ListEnvelope<object>> }>('/templates', async () => {
    try {
      const list = await templates.list();
      return successResponse(listEnvelope(list));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('db_error', message);
    }
  });

  // Get one template
  fastify.get<{
    Params: { id: string };
    Reply: ApiResponse<object>;
  }>('/templates/:id', async (request) => {
    try {
      const template = await templates.getById(request.params.id);
      if (!template) {
        return errorResponse('not_found', `Template ${request.params.id} not found`);
      }
      const parsed = yaml.parse(template.yaml);
      return successResponse({ ...template, parsed });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('db_error', message);
    }
  });

  // Save / update template.
  //
  // Source promotion semantics:
  //   - New rows (no existing) → source='user'.
  //   - Edits to a builtin → source='user' AND we promote *only* if the YAML
  //     actually changed. Why: seedBuiltinTemplates() resyncs every builtin
  //     row from prompts/templates/*.yaml on each daemon boot. If we left an
  //     edited row marked 'builtin', the seed loop would clobber the user's
  //     changes on next restart. Promoting to 'user' makes the seed loop
  //     skip the row (it only refreshes source='builtin'), so edits persist.
  //     The user accepts the tradeoff: their fork no longer auto-receives
  //     future builtin updates from chorus upgrades. That's the *expected*
  //     behaviour of "I edited this, it's mine now."
  //   - Edits to an existing user row → stay 'user' (no-op).
  //   - Save with identical YAML on a builtin → stay 'builtin' (we don't
  //     accidentally fork from a no-op resave).
  //
  // Validation runs in two stages so the cockpit can surface the *right*
  // error to the user:
  //   1. yaml.parse — catches syntax errors (indent, missing colon, etc.).
  //   2. TemplateSchema (zod) — catches structural errors (missing phases,
  //      unknown lineage, reviewer.require out of bounds, hybrid review_only,
  //      …). Without this stage a malformed-but-parseable template would
  //      land in the DB and crash the runner the first time it loads it
  //      (getParsedTemplate calls TemplateSchema.parse with no fallback).
  //
  // Zod errors are flattened to {path, message} pairs so the editor can
  // pin each error to its field; the response shape matches the existing
  // ApiResponse `{ error, details: { issues } }` envelope.
  fastify.post<{
    Body: { id: string; yaml: string };
    Reply: ApiResponse<object>;
  }>('/templates', async (request) => {
    try {
      const { id, yaml: yamlContent } = request.body;
      if (!id || !yamlContent) {
        return errorResponse('validation', 'id and yaml are required');
      }

      const validation = validateTemplateYaml(yamlContent);
      if (!validation.valid) {
        const summary = validation.issues
          .slice(0, 3)
          .map((i) => `${i.path}: ${i.message}`)
          .join('; ');
        return errorResponse(
          'validation',
          `Template failed validation: ${summary}${validation.issues.length > 3 ? ` (+${validation.issues.length - 3} more)` : ''}`,
          { issues: validation.issues },
        );
      }

      const existing = await templates.getById(id);
      let source: 'builtin' | 'user';
      if (!existing) {
        source = 'user';
      } else if (existing.source === 'builtin' && existing.yaml === yamlContent) {
        // No-op resave on a builtin — keep it as 'builtin' so the disk
        // resync still works on next boot.
        source = 'builtin';
      } else {
        // Real edit OR an existing user row — both end up as 'user'.
        source = 'user';
      }
      const template = await templates.create(id, yamlContent, source);
      return successResponse({ ...template, sourcePromoted: existing?.source === 'builtin' && source === 'user' });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('db_error', message);
    }
  });

  // Delete a user template. Built-ins refuse because the boot seed in
  // src/daemon/index.ts re-creates them from `templates/*.yaml` on next
  // start — surfacing a 400 here makes the cockpit hide the delete button
  // for builtins instead of silently no-op'ing. Mirrors the persona DELETE
  // route shape.
  fastify.delete<{
    Params: { id: string };
    Reply: ApiResponse<{ id: string }>;
  }>('/templates/:id', async (request) => {
    try {
      const existing = await templates.getById(request.params.id);
      if (!existing) {
        return errorResponse('not_found', `Template ${request.params.id} not found`);
      }
      if (existing.source === 'builtin') {
        return errorResponse(
          'validation',
          'Built-in templates cannot be deleted (the boot seed would recreate them). Edit instead — your changes will be preserved.',
        );
      }
      await templates.delete(request.params.id);
      return successResponse({ id: request.params.id });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('db_error', message);
    }
  });
}

const PERSONA_ID_RE = /^[a-z0-9][a-z0-9-]{1,63}$/;
// Ids that would collide with cockpit URL segments or sentinel persona
// values used in template prompts. Keeping this conservative — it's
// easier to lift later than to migrate user data after a clash.
const RESERVED_PERSONA_IDS = new Set([
  'new',
  'edit',
  'create',
  'delete',
  'api',
  'admin',
  'default',
  'system',
  'none',
]);
const VALID_LINEAGES = new Set([
  'anthropic',
  'openai',
  'google',
  'opencode',
  'moonshot',
  'openrouter',
]);

export function registerPersonaRoutes(fastify: FastifyInstance): void {
  fastify.get<{ Reply: ApiResponse<ListEnvelope<object>> }>('/personas', async () => {
    try {
      const { listPersonas } = await import('../../lib/personas.js');
      const items = await listPersonas();
      return successResponse(listEnvelope(items));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('db_error', message);
    }
  });

  fastify.get<{
    Params: { id: string };
    Reply: ApiResponse<object>;
  }>('/personas/:id', async (request) => {
    try {
      const { getPersona } = await import('../../lib/personas.js');
      const row = await getPersona(request.params.id);
      if (!row) {
        return errorResponse('not_found', `Persona ${request.params.id} not found`);
      }
      return successResponse(row);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('db_error', message);
    }
  });

  // Create or update a persona. Edits to a builtin row demote it to
  // builtin=false so the boot-time seed in seedBuiltinPersonas() skips it
  // and the user's edits persist across restarts (same source-promotion
  // tradeoff as templates: the row no longer auto-receives upstream changes).
  fastify.post<{
    Body: {
      id: string;
      label: string;
      one_liner: string;
      system_prompt: string;
      recommended_lineage?: string | null;
      forked_from?: string | null;
    };
    Reply: ApiResponse<object>;
  }>('/personas', async (request) => {
    try {
      const { id, label, one_liner, system_prompt, recommended_lineage, forked_from } = request.body ?? {};
      if (!id || !label || !one_liner || !system_prompt) {
        return errorResponse(
          'validation',
          'id, label, one_liner, and system_prompt are required',
        );
      }
      if (!PERSONA_ID_RE.test(id)) {
        return errorResponse(
          'validation',
          'id must match /^[a-z0-9][a-z0-9-]{1,63}$/ (lowercase, dashes)',
        );
      }
      if (RESERVED_PERSONA_IDS.has(id)) {
        return errorResponse(
          'validation',
          `id "${id}" is reserved (would collide with a UI/API route). Pick a more specific name.`,
        );
      }
      if (
        recommended_lineage !== undefined &&
        recommended_lineage !== null &&
        recommended_lineage !== '' &&
        !VALID_LINEAGES.has(recommended_lineage)
      ) {
        return errorResponse(
          'validation',
          `recommended_lineage must be one of: ${[...VALID_LINEAGES].join(', ')}`,
        );
      }

      const { personas } = await import('../../lib/db/index.js');
      const existing = await personas.getById(id);
      // Provenance precedence: existing row's forked_from wins (so re-saving
      // a row never loses its origin), then any client-supplied forked_from
      // (set by the cockpit's Duplicate handler), else null. Built-ins
      // edited *in place* (id stays the same) intentionally drop forked_from
      // — recording forked_from=existing.id would create a self-cycle that
      // breaks any recursive lineage traversal. Sourcefulness is captured
      // in the sourcePromoted flag returned to the client and the row's
      // builtin=false status; we don't need a row pointing at itself.
      const resolvedForkedFrom: string | null = existing
        ? (existing.forked_from ?? null)
        : (forked_from && forked_from !== id ? forked_from : null);
      const row = await personas.upsert({
        id,
        label,
        one_liner,
        system_prompt,
        recommended_lineage:
          recommended_lineage && recommended_lineage !== '' ? recommended_lineage : null,
        // Any save through the HTTP API marks the row as user-owned so
        // seedBuiltinPersonas() leaves it alone on next boot.
        builtin: false,
        forked_from: resolvedForkedFrom,
      });
      return successResponse({
        ...row,
        sourcePromoted: existing?.builtin === true,
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('db_error', message);
    }
  });

  // Delete a user persona. Builtin rows can't be deleted because the boot
  // seed would just re-create them — surfacing a 400 here makes the cockpit
  // hide the delete button for builtins instead of silently no-op'ing.
  fastify.delete<{
    Params: { id: string };
    Reply: ApiResponse<{ id: string }>;
  }>('/personas/:id', async (request) => {
    try {
      const { personas } = await import('../../lib/db/index.js');
      const existing = await personas.getById(request.params.id);
      if (!existing) {
        return errorResponse('not_found', `Persona ${request.params.id} not found`);
      }
      if (existing.builtin) {
        return errorResponse(
          'validation',
          'Built-in personas cannot be deleted (the boot seed would recreate them). Edit instead — your changes will be preserved.',
        );
      }
      await personas.delete(request.params.id);
      return successResponse({ id: request.params.id });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('db_error', message);
    }
  });
}
