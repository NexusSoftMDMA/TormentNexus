// Voices API client — talks to /voices on the daemon.
//
// See planning/voices.md for the abstraction's design rationale.

import type { ListEnvelope } from "@/lib/types";
import { fetchFromDaemon } from "./client";

export type VoiceLineage = "anthropic" | "openai" | "google" | "opencode" | "moonshot";
export type VoiceSource = "cli" | "api";

export interface Voice {
  id: string;
  label: string;
  source: VoiceSource;
  provider: string;
  model_id: string;
  lineage: VoiceLineage;
  vendor_family: string | null;
  input_cost_per_mtok: number | null;
  output_cost_per_mtok: number | null;
  enabled: boolean;
  created_at: number;
  updated_at: number;
}

export interface VoiceListFilter {
  lineage?: VoiceLineage;
  source?: VoiceSource;
  provider?: string;
  /** When undefined, returns ALL voices (enabled + disabled). */
  enabled?: boolean;
}

export interface VoiceUpdate {
  label?: string;
  enabled?: boolean;
  input_cost_per_mtok?: number | null;
  output_cost_per_mtok?: number | null;
}

export interface VoiceCreate {
  provider: string;
  model_id: string;
  label: string;
  lineage: VoiceLineage;
  source?: VoiceSource;
  vendor_family?: string | null;
  input_cost_per_mtok?: number | null;
  output_cost_per_mtok?: number | null;
  enabled?: boolean;
}

function buildQuery(filter?: VoiceListFilter): string {
  if (!filter) return "";
  const parts: string[] = [];
  if (filter.lineage) parts.push(`lineage=${encodeURIComponent(filter.lineage)}`);
  if (filter.source) parts.push(`source=${encodeURIComponent(filter.source)}`);
  if (filter.provider) parts.push(`provider=${encodeURIComponent(filter.provider)}`);
  if (filter.enabled !== undefined) parts.push(`enabled=${filter.enabled ? "true" : "false"}`);
  return parts.length === 0 ? "" : `?${parts.join("&")}`;
}

export async function listVoices(filter?: VoiceListFilter): Promise<Voice[]> {
  const env = await fetchFromDaemon<ListEnvelope<Voice>>(`/voices${buildQuery(filter)}`);
  return env.items;
}

export async function updateVoice(id: string, patch: VoiceUpdate): Promise<Voice> {
  return fetchFromDaemon<Voice>(`/voices/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(patch),
  });
}

export async function createVoice(input: VoiceCreate): Promise<Voice> {
  return fetchFromDaemon<Voice>("/voices", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

