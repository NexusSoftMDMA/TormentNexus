import type {
  AgreementThreshold,
  ReviewerLineage,
  Template,
  TemplatePhase,
  ThresholdAction,
} from "@/lib/cockpit-types";

export interface FallbackVoice {
  lineage: ReviewerLineage;
  model: string;
  persona?: string;
}

export interface FormState {
  /** Stable id once known — set on edit, derived on save for new templates. */
  id: string;
  name: string;
  description: string;
  /**
   * Original `author` from the YAML, preserved verbatim so editing a
   * community template doesn't silently rewrite ownership.
   */
  author: string;
  category: Template["category"];
  phases: TemplatePhase[];
  threshold: AgreementThreshold;
  /**
   * Custom (non-preset) numeric threshold from the YAML. When set, the
   * `threshold` enum is the closest preset and the form tab is disabled
   * (lossy). Emitter prefers `customThreshold` when present so the YAML
   * round-trips unchanged.
   */
  customThreshold?: number;
  onThresholdMet: ThresholdAction;
  /**
   * Captures the literal "review" value from the daemon when present.
   * Form's 2-value enum can't represent `review`, so for that case the
   * form is disabled (lossy) and the emitter restores the literal from
   * here. Merge/ask are derivable from `onThresholdMet`, so we
   * deliberately do NOT capture them here (a stale raw value would
   * shadow a user's form edit).
   */
  onThresholdMetRaw?: "review";
  maxRounds: number;
  yoloDefault: boolean;
  /**
   * Template-level fallback chains, split by role. Tried in order when a
   * slot's per-slot model chain exhausts. Strict (lineage, model) dedup
   * against active slots of the same role in the same phase.
   */
  fallbackDoer: FallbackVoice[];
  fallbackReviewer: FallbackVoice[];
}

export interface DaemonPhaseYaml {
  id: string;
  kind: string;
  title: string;
  description?: string;
  doer?: { lineage: string; models?: string[]; persona?: string };
  reviewer?: {
    require: number;
    crossLineage: boolean;
    candidates: { lineage: string; models?: string[]; persona?: string }[];
  };
  inputs: { include: string[]; exclude: string[] };
  /** Standard phases only — review_only is single-pass and has no iterate. */
  iterate?: {
    maxRounds: number;
    onDisagreement: "continue" | "escalate" | "accept-doer";
    shareSessionAcrossRounds: boolean;
    shareSessionAcrossPhases: boolean;
  };
  /** review_only phases only — runtime artifact spec. */
  artifact?: {
    label: string;
    hint: string;
    maxBytes: number;
  };
}

export interface DaemonTemplateYaml {
  id: string;
  name: string;
  description: string;
  author?: string;
  agreementThreshold: number;
  onThresholdMet: "merge" | "ask" | "review";
  maxRounds: number;
  yoloDefault: boolean;
  phases: DaemonPhaseYaml[];
  fallback?: {
    doer?: Array<{ lineage: string; models: string[]; persona?: string }>;
    reviewer?: Array<{ lineage: string; models: string[]; persona?: string }>;
  };
}

export interface ParsedDaemonTemplate {
  id?: string;
  name?: string;
  description?: string;
  author?: string;
  agreementThreshold?: number | string;
  onThresholdMet?: string;
  maxRounds?: number;
  yoloDefault?: boolean;
  ship?: { enabled?: boolean };
  phases?: Array<{
    id?: string;
    kind?: string;
    title?: string;
    name?: string;
    description?: string;
    doer?: { lineage?: string; models?: string[]; persona?: string };
    reviewer?: {
      require?: number;
      crossLineage?: boolean;
      candidates?: Array<{
        lineage?: string;
        models?: string[];
        persona?: string;
      }>;
    };
    inputs?: { include?: string[]; exclude?: string[] };
    iterate?: { maxRounds?: number; onDisagreement?: string };
    artifact?: { label?: string; hint?: string; maxBytes?: number };
    timeoutMs?: number;
  }>;
  fallback?: {
    doer?: Array<{ lineage?: string; models?: string[]; persona?: string }>;
    reviewer?: Array<{ lineage?: string; models?: string[]; persona?: string }>;
  };
}

export interface ParseResult {
  form: FormState;
  /** True if YAML contains fields the form can't represent (form tab → readonly). */
  formLossy: boolean;
  lossyReasons: string[];
}
