/**
 * OpenRouter inline flow: validate key → save → fetch models → add as
 * voices. The HTTP shim that actually dispatches chat completions to
 * OpenRouter lives in `src/daemon/agents/openrouter.ts`; these routes
 * own the catalogue + secrets-table side.
 */

import type { FastifyInstance } from 'fastify';
import * as openrouter from '../openrouter.js';
import {
  errorResponse,
  successResponse,
  type ApiResponse,
} from '../api-response.js';

export function registerOpenRouterRoutes(fastify: FastifyInstance): void {
  fastify.post<{
    Body: { apiKey?: string };
    Reply: ApiResponse<{ valid: boolean; error?: string }>;
  }>('/openrouter/validate', async (request) => {
    try {
      const apiKey = request.body?.apiKey;
      if (typeof apiKey !== 'string') {
        return errorResponse('validation', 'apiKey (string) is required');
      }
      const result = await openrouter.validateKey(apiKey);
      return successResponse(result);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('openrouter_error', message);
    }
  });

  fastify.post<{
    Body: { apiKey?: string };
    Reply: ApiResponse<{ valid: boolean; error?: string }>;
  }>('/openrouter/save-key', async (request) => {
    try {
      const apiKey = request.body?.apiKey;
      if (typeof apiKey !== 'string') {
        return errorResponse('validation', 'apiKey (string) is required');
      }
      const result = await openrouter.saveKey(apiKey);
      return successResponse(result);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('openrouter_error', message);
    }
  });

  fastify.get<{
    Reply: ApiResponse<{ models: openrouter.OpenRouterModel[] }>;
  }>('/openrouter/models', async () => {
    try {
      const models = await openrouter.listModels();
      return successResponse({ models });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('openrouter_error', message);
    }
  });

  fastify.post<{
    Body: { modelIds?: string[]; apiKey?: string };
    Reply: ApiResponse<{ added: string[]; skipped: string[] }>;
  }>('/openrouter/voices', async (request) => {
    try {
      const modelIds = request.body?.modelIds;
      if (!Array.isArray(modelIds) || !modelIds.every((s) => typeof s === 'string')) {
        return errorResponse('validation', 'modelIds (string[]) is required');
      }
      // Optional apiKey lets a caller bypass the secrets table — useful
      // for one-off automation that already holds the key, and avoids
      // racing a concurrent save-key write.
      const apiKey =
        typeof request.body?.apiKey === 'string' ? request.body.apiKey : undefined;
      const result = await openrouter.addModelsAsVoices(modelIds, apiKey);
      return successResponse(result);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('openrouter_error', message);
    }
  });
}
