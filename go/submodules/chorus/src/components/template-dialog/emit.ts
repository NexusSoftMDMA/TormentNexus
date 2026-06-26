import yaml from "yaml";
import type {
  AgreementThreshold,
  ThresholdAction,
} from "@/lib/cockpit-types";
import {
  COCKPIT_TO_DAEMON,
  DAEMON_DEFAULT_MODEL,
} from "./constants";
import type { DaemonPhaseYaml, DaemonTemplateYaml, FormState } from "./types";

function thresholdToNumber(t: AgreementThreshold): number {
  switch (t) {
    case "unanimous":
      return 1;
    case "majority":
      return 0.66;
    case "any":
      return 0.34;
  }
}

export function thresholdFromNumber(n: number): AgreementThreshold {
  if (n >= 0.99) return "unanimous";
  if (n >= 0.5) return "majority";
  return "any";
}

function actionToDaemon(a: ThresholdAction): "merge" | "ask" | "review" {
  return a === "auto-finalize" ? "merge" : "ask";
}

export function actionFromDaemon(s: string | undefined): ThresholdAction {
  return s === "merge" ? "auto-finalize" : "ask-user";
}

export function slugify(s: string): string {
  return (
    s
      .trim()
      .toLowerCase()
      .replace(/[^\w\s-]/g, "")
      .replace(/\s+/g, "-")
      .replace(/-+/g, "-") || "untitled"
  );
}

function formToDaemonShape(f: FormState): DaemonTemplateYaml {
  const id = f.id || slugify(f.name);
  return {
    id,
    name: f.name || "Untitled template",
    description: f.description || "Describe what this template is for.",
    author: f.author || "chorus",
    agreementThreshold:
      f.customThreshold !== undefined
        ? f.customThreshold
        : thresholdToNumber(f.threshold),
    onThresholdMet: f.onThresholdMetRaw ?? actionToDaemon(f.onThresholdMet),
    maxRounds: f.maxRounds,
    yoloDefault: f.yoloDefault,
    phases: f.phases.map((p): DaemonPhaseYaml => {
      const isReviewOnly = p.kind === "review_only";
      const reviewerBlock =
        p.reviewer.candidates.length > 0
          ? {
              require: p.reviewer.require,
              crossLineage: p.reviewer.crossLineage,
              // flatMap so each lineage with multiple models in
              // candidateModels emits one candidate-entry per model. Lets
              // the user pick e.g. codex gpt-5.5 + codex gpt-5.5-pro as
              // two reviewers.
              candidates: p.reviewer.candidates.flatMap((l) => {
                const userPicked = (p.reviewer.candidateModels?.[l] ?? [])
                  .map((m) => m.trim())
                  .filter((m) => m.length > 0);
                const personasByModel = p.reviewer.candidatePersonas?.[l];
                const daemonLineage = COCKPIT_TO_DAEMON[l] ?? "anthropic";
                if (userPicked.length === 0) {
                  const fallback = DAEMON_DEFAULT_MODEL[l] ?? "claude-opus-4-7";
                  // Lineage-only slots store their persona under the "" key.
                  const persona = personasByModel?.[""];
                  return [
                    {
                      lineage: daemonLineage,
                      models: [fallback],
                      ...(persona ? { persona } : {}),
                    },
                  ];
                }
                return userPicked.map((m) => {
                  const persona = personasByModel?.[m];
                  return {
                    lineage: daemonLineage,
                    models: [m],
                    ...(persona ? { persona } : {}),
                  };
                });
              }),
            }
          : undefined;

      // review_only: drop doer + iterate, add artifact block. The schema
      // rejects a doer field on review_only and also rejects iterate, so
      // the keys must literally not be present.
      if (isReviewOnly) {
        return {
          id: p.id,
          kind: p.kind,
          title: p.name,
          description: p.description,
          reviewer: reviewerBlock,
          inputs: { include: p.inputs.include, exclude: p.inputs.exclude },
          artifact: {
            label: p.artifact?.label ?? "Artifact to review",
            hint:
              p.artifact?.hint ??
              "Paste a unified diff, a markdown draft, code, or any text blob.",
            maxBytes: p.artifact?.maxBytes ?? 1024 * 1024,
          },
        };
      }

      return {
        id: p.id,
        kind: p.kind,
        title: p.name,
        description: p.description,
        doer: {
          lineage: COCKPIT_TO_DAEMON[p.doer.lineage] ?? "anthropic",
          models:
            p.doer.models.length > 0
              ? p.doer.models
              : [DAEMON_DEFAULT_MODEL[p.doer.lineage] ?? "claude-opus-4-7"],
          ...(p.doer.persona ? { persona: p.doer.persona } : {}),
        },
        reviewer: reviewerBlock,
        inputs: { include: p.inputs.include, exclude: p.inputs.exclude },
        iterate: {
          maxRounds: p.iterate.max,
          // Form's onMax → daemon's onDisagreement enum mapping:
          //   loopback  → 'continue'    (keep iterating with revisions)
          //   ask-user  → 'escalate'    (surface to user, halt loop)
          //   fail      → 'accept-doer' (drop reviewer veto, accept doer)
          // Pre-fix this emitted 'ask-user' which the schema rejects with
          // "Expected 'continue' | 'escalate' | 'accept-doer'", causing
          // every wizard-saved template to fail validation.
          onDisagreement:
            p.iterate.onMax === "loopback"
              ? "continue"
              : p.iterate.onMax === "fail"
                ? "accept-doer"
                : "escalate",
          shareSessionAcrossRounds: true,
          shareSessionAcrossPhases: false,
        },
      };
    }),
    // Template-level fallback chains, split by role. Emitted only when at
    // least one role has rows so older templates round-trip unchanged.
    ...(f.fallbackDoer.length > 0 || f.fallbackReviewer.length > 0
      ? {
          fallback: {
            ...(f.fallbackDoer.length > 0
              ? {
                  doer: f.fallbackDoer.map((fb) => ({
                    lineage: COCKPIT_TO_DAEMON[fb.lineage] ?? "anthropic",
                    models: [fb.model],
                    ...(fb.persona ? { persona: fb.persona } : {}),
                  })),
                }
              : {}),
            ...(f.fallbackReviewer.length > 0
              ? {
                  reviewer: f.fallbackReviewer.map((fb) => ({
                    lineage: COCKPIT_TO_DAEMON[fb.lineage] ?? "anthropic",
                    models: [fb.model],
                    ...(fb.persona ? { persona: fb.persona } : {}),
                  })),
                }
              : {}),
          },
        }
      : {}),
  };
}

export function buildYamlFromForm(f: FormState): string {
  return yaml.stringify(formToDaemonShape(f), { lineWidth: 0 });
}
