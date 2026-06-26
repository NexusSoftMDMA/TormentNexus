// Template API endpoints
import yaml from "yaml";
import type { ListEnvelope, Template } from "@/lib/types";
import { fetchFromDaemon } from "./client";

interface RawTemplateRow {
  id: string;
  source: "builtin" | "user";
  yaml: string;
  // Daemon emits 1/0 (legacy SQLite int) or true/false depending on
  // libsql transport — accept either via parseRow.
  is_complete?: boolean | number;
  created_at: number;
  updated_at: number;
}

interface ParsedDoer {
  lineage?: string;
  models?: string[];
  persona?: string;
}

interface ParsedReviewer {
  require?: number;
  crossLineage?: boolean;
  candidates?: Array<{ lineage?: string; models?: string[]; persona?: string }>;
}

interface ParsedPhase {
  id?: string;
  kind?: string;
  title?: string;
  name?: string;
  description?: string;
  doer?: ParsedDoer;
  reviewer?: ParsedReviewer;
  inputs?: { include?: string[]; exclude?: string[] };
  iterate?: { maxRounds?: number; onDisagreement?: string };
  artifact?: { label?: string; hint?: string; maxBytes?: number };
}

interface ParsedTemplateYaml {
  id?: string;
  name?: string;
  description?: string;
  category?: Template["category"];
  author?: string;
  agreementThreshold?: number | Template["agreementThreshold"];
  onThresholdMet?: Template["onThresholdMet"] | string;
  maxRounds?: number;
  yoloDefault?: boolean;
  estimatedBaselineTokens?: number;
  phases?: ParsedPhase[];
}

const KNOWN_LINEAGES: ReadonlySet<string> = new Set([
  "anthropic",
  "openai",
  "google",
  "opencode",
  "moonshot",
  "any",
  // Legacy alias — old templates still use 'xai' for OpenCode.
  "xai",
]);

const UI_LINEAGE_MAP: Record<string, Template["phases"][number]["doer"]["lineage"]> = {
  anthropic: "claude",
  openai: "codex",
  google: "gemini",
  opencode: "opencode",
  moonshot: "kimi",
  any: "claude",
  // Legacy alias.
  xai: "opencode",
};

function mapLineage(raw: string | undefined): Template["phases"][number]["doer"]["lineage"] {
  if (!raw) return "claude";
  return UI_LINEAGE_MAP[raw] ?? "claude";
}

function deriveCategory(parsed: ParsedTemplateYaml, id: string): Template["category"] {
  if (parsed.category) return parsed.category;
  const idLower = id.toLowerCase();
  if (idLower.includes("bug") || idLower.includes("debug") || idLower.includes("diagnose")) return "debug";
  if (idLower.includes("plan") || idLower.includes("architect")) return "plan";
  if (idLower.includes("decide") || idLower.includes("decision")) return "decide";
  return "review";
}

function mapPhase(p: ParsedPhase): Template["phases"][number] {
  const kind = (p.kind ?? "review") as Template["phases"][number]["kind"];
  return {
    id: p.id ?? "phase",
    name: p.title ?? p.name ?? p.id ?? "Phase",
    description: p.description ?? "",
    kind,
    gate: "auto",
    // review_only phases have no doer in the YAML. Synthesise a placeholder
    // so the legacy DoerSlot type is satisfied; isReviewOnlyTemplate gates
    // the UI from rendering it.
    doer: {
      lineage: mapLineage(p.doer?.lineage),
      models: p.doer?.models ?? [],
      ...(p.doer?.persona ? { persona: p.doer.persona } : {}),
    },
    reviewer: {
      require: p.reviewer?.require ?? 1,
      crossLineage: p.reviewer?.crossLineage ?? true,
      // Keep models alongside lineage so the run page can show the model
      // badge ("gpt-5.5") on placeholder reviewer cards. Falls back to a
      // bare-lineage shape via parallel array below — see Template.phases
      // candidatesWithModels for the structured form.
      candidates: (p.reviewer?.candidates ?? [])
        .map((c) => mapLineage(c.lineage))
        .filter((l) => KNOWN_LINEAGES.has(Object.keys(UI_LINEAGE_MAP).find((k) => UI_LINEAGE_MAP[k] === l) ?? "")),
      candidatesWithModels: (p.reviewer?.candidates ?? []).map((c) => ({
        lineage: mapLineage(c.lineage),
        models: c.models ?? [],
        ...(c.persona ? { persona: c.persona } : {}),
      })),
    },
    inputs: {
      include: p.inputs?.include ?? [],
      exclude: p.inputs?.exclude ?? [],
    },
    iterate: {
      max: p.iterate?.maxRounds ?? 2,
      onMax: "ask-user",
    },
    blindSpots: [],
    execution: "parallel",
    builtin: true,
    ...(kind === "review_only"
      ? {
          artifact: {
            label: p.artifact?.label ?? "Artifact to review",
            hint:
              p.artifact?.hint ??
              "Paste a unified diff, a markdown draft, code, or any text blob.",
            maxBytes: p.artifact?.maxBytes ?? 1024 * 1024,
          },
        }
      : {}),
  };
}

/**
 * Daemon stores templates as raw YAML rows. The UI needs structured
 * Template objects, so we parse the YAML field client-side.
 * Falls back to safe defaults for fields the UI doesn't need yet (phases etc).
 */
function parseRow(row: RawTemplateRow): Template {
  let parsed: ParsedTemplateYaml = {};
  try {
    parsed = (yaml.parse(row.yaml) as ParsedTemplateYaml) ?? {};
  } catch {
    // Bad YAML → leave parsed empty; the card will show id only.
  }

  // agreementThreshold: numbers (0..1) ↔ enum strings
  let threshold: Template["agreementThreshold"] = "majority";
  if (typeof parsed.agreementThreshold === "number") {
    if (parsed.agreementThreshold >= 0.99) threshold = "unanimous";
    else if (parsed.agreementThreshold >= 0.5) threshold = "majority";
    else threshold = "any";
  } else if (typeof parsed.agreementThreshold === "string") {
    threshold = parsed.agreementThreshold as Template["agreementThreshold"];
  }

  const phases = (parsed.phases ?? []).map(mapPhase);

  return {
    id: row.id,
    name: parsed.name ?? row.id,
    description: parsed.description ?? "",
    category: deriveCategory(parsed, row.id),
    phases,
    agreementThreshold: threshold,
    onThresholdMet:
      parsed.onThresholdMet === "auto-finalize"
        ? "auto-finalize"
        : "ask-user",
    maxRounds: parsed.maxRounds ?? 3,
    driver: "external",
    driverHandoff: false,
    verificationGate: "auto",
    costCapUsd: 0,
    yoloDefault: parsed.yoloDefault ?? false,
    estimatedBaselineTokens: parsed.estimatedBaselineTokens,
    onError: "ask-user",
    notify: "dashboard-only",
    yaml: row.yaml,
    authorHandle: parsed.author ?? "chorus",
    forks: 0,
    popularity: 0,
    source: row.source,
    isComplete:
      typeof row.is_complete === "number"
        ? row.is_complete !== 0
        : row.is_complete !== false,
  };
}

export async function listTemplates(): Promise<Template[]> {
  const env = await fetchFromDaemon<ListEnvelope<RawTemplateRow>>("/templates");
  return env.items.map(parseRow);
}

export async function getTemplate(id: string): Promise<Template> {
  const row = await fetchFromDaemon<RawTemplateRow>(`/templates/${id}`);
  return parseRow(row);
}

export async function saveTemplate(template: {
  id: string;
  yaml: string;
}): Promise<Template> {
  const row = await fetchFromDaemon<RawTemplateRow>("/templates", {
    method: "POST",
    body: JSON.stringify(template),
  });
  return parseRow(row);
}

/** Hard-delete a user template. The daemon refuses for built-ins (the
 *  boot seed re-creates them from disk on next start) — surface that as a
 *  DaemonError to the cockpit so the delete button is hidden for built-ins.
 */
export async function deleteTemplate(id: string): Promise<void> {
  await fetchFromDaemon<{ id: string }>(`/templates/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
}
