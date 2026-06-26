import yaml from "yaml";
import type {
  AgreementThreshold,
  ReviewerLineage,
  Template,
  TemplatePhase,
} from "@/lib/cockpit-types";
import {
  DAEMON_DEFAULT_MODEL,
  DAEMON_TO_COCKPIT,
  DEFAULT_FORM,
} from "./constants";
import { actionFromDaemon, thresholdFromNumber } from "./emit";
import type {
  FallbackVoice,
  ParsedDaemonTemplate,
  ParseResult,
} from "./types";

function deriveCategory(id: string): Template["category"] {
  const i = id.toLowerCase();
  if (i.includes("bug") || i.includes("debug") || i.includes("diagnose"))
    return "debug";
  if (i.includes("plan") || i.includes("architect")) return "plan";
  if (i.includes("decide") || i.includes("decision")) return "decide";
  return "review";
}

/** Flatten daemon-shape fallback rows ({lineage, models[]}) into one form
 *  row per (lineage, model) tuple. The form's add/remove UI is row-oriented. */
function flattenFallbackList(
  list:
    | Array<{ lineage?: string; models?: string[]; persona?: string }>
    | undefined,
): FallbackVoice[] {
  if (!Array.isArray(list)) return [];
  return list.flatMap((fb) => {
    const cockpitLineage = DAEMON_TO_COCKPIT[fb.lineage ?? ""] ?? "claude";
    const models = (fb.models ?? []).filter(
      (m) => typeof m === "string" && m.trim().length > 0,
    );
    if (models.length === 0) return [];
    return models.map((model) => ({
      lineage: cockpitLineage,
      model,
      ...(fb.persona ? { persona: fb.persona } : {}),
    }));
  });
}

const KNOWN_PHASE_KINDS = [
  "plan",
  "spec",
  "tests",
  "implement",
  "review",
  "verify",
  "divergence",
  "review_only",
] as const;

interface ParsedPhase {
  phase: TemplatePhase;
  reasons: string[];
}

function parsePhase(p: NonNullable<ParsedDaemonTemplate["phases"]>[number]): ParsedPhase {
  const reasons: string[] = [];
  if (p.timeoutMs !== undefined) {
    reasons.push(
      `Phase ${p.id ?? "?"} has \`timeoutMs\` — not exposed in form mode.`,
    );
  }
  const isReviewOnly = p.kind === "review_only";
  // review_only has no doer in YAML; pick a sensible placeholder so the
  // FormState shape stays uniform. The form hides the doer panel for
  // review_only and the YAML emitter drops it.
  const doerLineage = isReviewOnly
    ? "claude"
    : DAEMON_TO_COCKPIT[p.doer?.lineage ?? ""] ?? "claude";

  // Accumulate per-lineage models from YAML. The form's chip row needs
  // ONE chip per lineage even when YAML has duplicate-lineage entries
  // (e.g. codex gpt-5.5 + codex gpt-5.5-pro), and the model picker for
  // that lineage shows both as separate rows.
  const candidateModels: Partial<Record<ReviewerLineage, string[]>> = {};
  const candidatePersonas: Partial<
    Record<ReviewerLineage, Record<string, string | undefined>>
  > = {};
  const seenLineages = new Set<ReviewerLineage>();
  const candidates: ReviewerLineage[] = [];
  // Track when a (lineage, model) tuple appears in two different YAML
  // candidate blocks with conflicting personas. The form's
  // one-persona-per-tuple representation can't faithfully store both,
  // so we mark lossy and let the user fix it in YAML mode.
  const overlapConflicts: string[] = [];
  for (const c of p.reviewer?.candidates ?? []) {
    const cockpitLineage = DAEMON_TO_COCKPIT[c.lineage ?? ""];
    if (!cockpitLineage) continue;
    if (!seenLineages.has(cockpitLineage)) {
      seenLineages.add(cockpitLineage);
      candidates.push(cockpitLineage);
    }
    // YAML candidates may declare persona without an explicit model;
    // keying by model allows round-trip of either case ("" for the
    // lineage-only slot, model id for the rest).
    const trimmedModels = (c.models ?? [])
      .map((m) => (m ?? "").trim())
      .filter((m) => m.length > 0);
    const yamlPersona =
      c.persona && c.persona.trim().length > 0 ? c.persona : undefined;
    const writePersona = (modelKey: string): void => {
      if (!yamlPersona) return;
      const map = (candidatePersonas[cockpitLineage] ??= {});
      const existing = map[modelKey];
      if (existing !== undefined && existing !== yamlPersona) {
        overlapConflicts.push(
          `${cockpitLineage}/${modelKey || "(no model)"}: "${existing}" vs "${yamlPersona}"`,
        );
      }
      map[modelKey] = yamlPersona;
    };
    if (trimmedModels.length === 0) {
      writePersona("");
    } else {
      for (const m of trimmedModels) {
        (candidateModels[cockpitLineage] ??= []).push(m);
        writePersona(m);
      }
    }
  }
  if (overlapConflicts.length > 0) {
    reasons.push(
      `Phase ${p.id ?? "?"} has overlapping reviewer candidates with conflicting personas (${overlapConflicts.join("; ")}) — form mode would silently last-write-wins; edit in YAML to disambiguate.`,
    );
  }

  const phaseKind = KNOWN_PHASE_KINDS.includes(p.kind as never)
    ? (p.kind as TemplatePhase["kind"])
    : "review";
  if (p.kind && !KNOWN_PHASE_KINDS.includes(p.kind as never)) {
    reasons.push(
      `Phase ${p.id ?? "?"} has unknown kind \`${p.kind}\` — form mode would coerce it to \`review\`.`,
    );
  }

  const phase: TemplatePhase = {
    id: p.id ?? "phase",
    name: p.title ?? p.name ?? p.id ?? "Phase",
    description: p.description ?? "",
    kind: phaseKind,
    gate: "auto",
    doer: {
      lineage: doerLineage,
      models:
        p.doer?.models && p.doer.models.length > 0
          ? p.doer.models
          : [DAEMON_DEFAULT_MODEL[doerLineage] ?? "claude-opus-4-7"],
      ...(p.doer?.persona && p.doer.persona.trim().length > 0
        ? { persona: p.doer.persona.trim() }
        : {}),
    },
    reviewer: {
      require: p.reviewer?.require ?? 1,
      crossLineage: p.reviewer?.crossLineage ?? true,
      candidates,
      candidateModels,
      ...(Object.keys(candidatePersonas).length > 0
        ? { candidatePersonas }
        : {}),
    },
    inputs: {
      include: p.inputs?.include ?? [],
      exclude: p.inputs?.exclude ?? [],
    },
    iterate: {
      // maxRounds + onDisagreement defaults must mirror IterateSchema
      // (template-schema.ts) so a YAML with omitted iterate fields
      // round-trips identically. Earlier defaults (3 + 'ask-user') drifted
      // from schema defaults (2 + 'continue'): opening a builtin template,
      // making any unrelated edit, then saving silently rewrote the
      // phase's iteration policy from "loop with revisions" to "halt and
      // ask the user", and bumped maxRounds 2→3.
      max: p.iterate?.maxRounds ?? 2,
      // Inverse of formToDaemonShape: 'continue'/'escalate'/'accept-doer'
      // → 'loopback'/'ask-user'/'fail'. The undefined coalesce to
      // 'continue' is critical — schema default is 'continue', so an
      // omitted onDisagreement must round-trip as 'loopback' (form) →
      // 'continue' (YAML).
      onMax:
        (p.iterate?.onDisagreement ?? "continue") === "continue"
          ? "loopback"
          : p.iterate?.onDisagreement === "accept-doer"
            ? "fail"
            : "ask-user",
    },
    blindSpots: [],
    execution: "parallel",
    builtin: false,
    // Preserve artifact on review_only so the form can edit it and the
    // emitter can re-emit it.
    ...(isReviewOnly && p.artifact
      ? {
          artifact: {
            label: p.artifact.label ?? "Artifact to review",
            hint:
              p.artifact.hint ??
              "Paste a unified diff, a markdown draft, code, or any text blob.",
            maxBytes: p.artifact.maxBytes ?? 1024 * 1024,
          },
        }
      : isReviewOnly
        ? {
            artifact: {
              label: "Artifact to review",
              hint:
                "Paste a unified diff, a markdown draft, code, or any text blob.",
              maxBytes: 1024 * 1024,
            },
          }
        : {}),
  };

  return { phase, reasons };
}

export function parseYamlToForm(
  yamlText: string,
  existingId: string,
): ParseResult {
  const reasons: string[] = [];
  let parsed: ParsedDaemonTemplate;
  try {
    parsed = (yaml.parse(yamlText) as ParsedDaemonTemplate) ?? {};
  } catch {
    return {
      form: { ...DEFAULT_FORM, id: existingId },
      formLossy: true,
      lossyReasons: ["YAML failed to parse — only YAML mode is available."],
    };
  }

  if (parsed.ship?.enabled) {
    reasons.push(
      "Template has `ship.enabled: true` — form mode can't edit ship config.",
    );
  }

  const phases: TemplatePhase[] = [];
  for (const p of parsed.phases ?? []) {
    const { phase, reasons: phaseReasons } = parsePhase(p);
    phases.push(phase);
    reasons.push(...phaseReasons);
  }

  let threshold: AgreementThreshold;
  let customThreshold: number | undefined;
  if (typeof parsed.agreementThreshold === "number") {
    threshold = thresholdFromNumber(parsed.agreementThreshold);
    // Form represents threshold as a 3-value enum (1, 0.66, 0.34).
    // Anything else round-trips through the closest preset, silently
    // rewriting the user's value. Capture it as `customThreshold` so the
    // emitter restores the exact number; surface a lossy reason so form
    // mode is hidden.
    const PRESETS = new Set([1, 0.66, 0.34]);
    if (!PRESETS.has(parsed.agreementThreshold)) {
      customThreshold = parsed.agreementThreshold;
      reasons.push(
        `Custom \`agreementThreshold: ${parsed.agreementThreshold}\` — form mode only offers 1.0 / 0.66 / 0.34 presets.`,
      );
    }
  } else {
    threshold =
      (parsed.agreementThreshold as AgreementThreshold) ?? "majority";
  }

  // Only the `review` literal is unrepresentable in the form (form has
  // auto-finalize/ask-user only). For merge/ask we let actionToDaemon do
  // the round-trip from form state on save — capturing the raw value here
  // would let it shadow a user's form edit. For "review" the form is
  // disabled (lossy) so capturing it is safe.
  const onThresholdMetRaw: "review" | undefined =
    parsed.onThresholdMet === "review" ? "review" : undefined;
  if (onThresholdMetRaw === "review") {
    reasons.push(
      "`onThresholdMet: review` — form mode only offers auto-finalize / ask-user.",
    );
  }

  return {
    form: {
      id: parsed.id ?? existingId,
      name: parsed.name ?? "",
      description: parsed.description ?? "",
      // author default 'chorus' must match TemplateSchema. Earlier `?? "you"`
      // corrupted builtin templates' author on round-trip: schema parses
      // missing author as 'chorus' → wizard sees absence → form shows 'you'
      // → save emits 'you' → builtin row promoted to user with wrong author.
      author: parsed.author ?? "chorus",
      category: deriveCategory(parsed.id ?? existingId),
      phases: phases.length > 0 ? phases : DEFAULT_FORM.phases,
      threshold,
      ...(customThreshold !== undefined ? { customThreshold } : {}),
      onThresholdMet: actionFromDaemon(parsed.onThresholdMet),
      ...(onThresholdMetRaw ? { onThresholdMetRaw } : {}),
      // Template-level maxRounds default = 3 (template-schema.ts:209).
      // The iterate.maxRounds inside each phase has its own default of 2.
      maxRounds: parsed.maxRounds ?? 3,
      yoloDefault: parsed.yoloDefault ?? false,
      // Flatten daemon shape (`models: [...]`) into one form row per
      // (lineage, model) so the cockpit's add/remove UI is row-oriented.
      fallbackDoer: flattenFallbackList(parsed.fallback?.doer),
      fallbackReviewer: flattenFallbackList(parsed.fallback?.reviewer),
    },
    formLossy: reasons.length > 0,
    lossyReasons: reasons,
  };
}
