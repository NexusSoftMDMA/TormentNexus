// OpenRouter API client — talks to /openrouter/* on the daemon.
//
// Purpose: validate the user's OpenRouter API key, fetch the model
// catalog, and multi-add chosen models as voices. The daemon side is
// in src/daemon/openrouter.ts; the chat-completions HTTP shim lives in
// a follow-up.

import { fetchFromDaemon } from "./client";

export interface OpenRouterModel {
  id: string;
  name: string;
  contextLength?: number;
  inputCostPerMtok?: number;
  outputCostPerMtok?: number;
}

export async function saveOpenRouterKey(
  apiKey: string,
): Promise<{ valid: boolean; error?: string }> {
  return fetchFromDaemon("/openrouter/save-key", {
    method: "POST",
    body: JSON.stringify({ apiKey }),
  });
}

export async function listOpenRouterModels(): Promise<{
  models: OpenRouterModel[];
}> {
  return fetchFromDaemon("/openrouter/models");
}

export async function addOpenRouterVoices(
  modelIds: string[],
  apiKey?: string,
): Promise<{ added: string[]; skipped: string[] }> {
  return fetchFromDaemon("/openrouter/voices", {
    method: "POST",
    body: JSON.stringify(apiKey ? { modelIds, apiKey } : { modelIds }),
  });
}
