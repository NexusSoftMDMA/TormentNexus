// Orchestrator integrations API — used by the /connect page to surface
// "is Claude Code wired up?" status and the one-click "Connect" button.
import type { ListEnvelope } from "@/lib/types";
import { fetchFromDaemon } from "./client";

export type OrchestratorName =
  | "claude"
  | "codex"
  | "gemini"
  | "opencode"
  | "kimi"
  | "cursor"
  | "windsurf";

export interface OrchestratorStatus {
  name: OrchestratorName;
  label: string;
  connected: boolean;
  approvedTools: number;
  totalTools: number;
  note: string;
  supported: boolean;
  firstCallBehavior: "auto" | "prompts_once" | "inherits_global";
}

export interface ConnectResult {
  added: string[];
  alreadyPresent: string[];
  configPath: string;
  status: OrchestratorStatus;
}

export async function listOrchestrators(): Promise<OrchestratorStatus[]> {
  const env = await fetchFromDaemon<ListEnvelope<OrchestratorStatus>>("/orchestrators");
  return env.items;
}

export async function connectOrchestrator(
  name: OrchestratorName,
): Promise<ConnectResult> {
  return fetchFromDaemon<ConnectResult>(`/orchestrators/${name}/connect`, {
    method: "POST",
  });
}

export interface OpencodeModelsResult {
  /** Models grouped by gateway prefix (`opencode-go`, `opencode-zen`, ...). */
  gateways: Record<string, string[]>;
  /** Flat list of all models (`opencode-go/kimi-k2.6`, ...). */
  flat: string[];
  /** Sensible default picks (kimi + deepseek if present in `flat`). */
  defaultPicks: string[];
}

/**
 * List models the local `opencode` CLI knows about. Used by the onboarding
 * flow so the user can tick which subscription models they want chorus to
 * expose as voices.
 */
export async function listOpencodeModels(): Promise<OpencodeModelsResult> {
  return fetchFromDaemon<OpencodeModelsResult>("/orchestrators/opencode/models");
}
