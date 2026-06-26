/**
 * Template adapter — rewrites a builtin template's YAML to use the
 * voices the user actually has installed and enabled.
 *
 * Why: shipped templates (Tri-Review, Red/Green, etc.) hardcode model
 * names like `claude-opus-4-7`, `gpt-5.5`, `kimi-k2.6`. If a user
 * doesn't have those CLIs (or has them under different model
 * versions), the template can't run as-is. The adapter:
 *   - For each slot whose lineage has matching enabled voices,
 *     replaces `models` with the user's actual model_ids for that
 *     lineage. The first model in the user's voices list becomes
 *     the primary; the rest become fallbacks (per chorus-083).
 *   - For each slot whose lineage has NO matching voices, substitutes
 *     a different lineage that the user does have, preferring a
 *     lineage not yet used in the same phase (diversity preservation).
 *   - If no substitute is available (user has no voices anywhere),
 *     leaves models empty and marks the template as `incomplete`.
 *     The cockpit gates "Use template" until the user fills the YAML.
 *
 * Ranking note: voices are picked in their natural order from the
 * voices table. The seed orders curated CLI catalogues with the
 * most-capable model first (e.g., Claude Opus before Sonnet before
 * Haiku), so first-match approximates "top model" without an
 * explicit priority column. Future improvement: vendor_family
 * priority list keyed by lineage.
 *
 * Vendor-family fallback: an opencode voice with vendor_family
 * matching the template's slot lineage (e.g., template wants
 * lineage=moonshot, user has `opencode-go/kimi-k2.5` with
 * vendor_family=moonshot) is preferred over a cross-lineage swap.
 * Same model family, different transport — the template's intent is
 * preserved.
 */

import yaml from 'yaml';

interface Voice {
  id: string;
  provider: string;
  model_id: string;
  lineage: string;
  vendor_family: string | null;
  enabled: boolean;
}

interface SlotSpec {
  lineage?: string;
  models?: string[];
  // Reviewer slots may carry a personaId; the adapter passes through
  // unchanged.
  candidatePersonas?: unknown;
}

interface PhaseSpec {
  doer?: SlotSpec | null;
  reviewer?: {
    candidates?: SlotSpec[];
    require?: number;
    crossLineage?: boolean;
  };
}

interface TemplateRoot {
  phases?: PhaseSpec[];
}

export interface AdaptResult {
  yaml: string;
  isComplete: boolean;
  changed: boolean;
}

/**
 * Pure function. Takes the canonical YAML and the user's voices,
 * returns adapted YAML + completeness flag. Same input → same output.
 */
export function adaptTemplate(
  canonicalYaml: string,
  voices: Voice[],
): AdaptResult {
  const enabled = voices.filter((v) => v.enabled);
  // Group enabled voices by lineage and by vendor_family for the
  // two-step fallback (lineage match → vendor_family match → swap).
  const byLineage = new Map<string, Voice[]>();
  const byFamily = new Map<string, Voice[]>();
  for (const v of enabled) {
    const ll = byLineage.get(v.lineage) ?? [];
    ll.push(v);
    byLineage.set(v.lineage, ll);
    if (v.vendor_family) {
      const fl = byFamily.get(v.vendor_family) ?? [];
      fl.push(v);
      byFamily.set(v.vendor_family, fl);
    }
  }

  // Parse with the YAML library that round-trips comments/formatting
  // reasonably; we re-serialise at the end. The parsed shape is loose
  // because templates evolve — we only touch fields we know about.
  let parsed: TemplateRoot;
  try {
    parsed = yaml.parse(canonicalYaml) as TemplateRoot;
  } catch {
    // Malformed canonical YAML — caller's bug. Return as-is and let
    // the validation layer surface it.
    return { yaml: canonicalYaml, isComplete: false, changed: false };
  }
  if (!parsed || typeof parsed !== 'object' || !Array.isArray(parsed.phases)) {
    return { yaml: canonicalYaml, isComplete: false, changed: false };
  }

  let isComplete = true;
  let changed = false;

  for (const phase of parsed.phases) {
    if (!phase) continue;
    // Sequential slot assignment in template order. Each slot's
    // assignment honors its preferred lineage when possible, then
    // falls through to a diversity-preserving substitute (a lineage
    // not yet used in this phase), then last-ditch any-available.
    //
    // Two diversity dimensions tracked per phase:
    //   usedLineages — which lineages are filled (prefer unused next)
    //   usedTuples   — which (lineage, model) pairs are filled
    //                  (rotate across multiple voices within a lineage
    //                   so 3 anthropic slots get opus + sonnet + haiku,
    //                   not opus three times)
    const usedLineages = new Set<string>();
    const usedTuples = new Set<string>();

    if (phase.doer) {
      const assigned = assignSlot(
        phase.doer,
        byLineage,
        byFamily,
        usedLineages,
        usedTuples,
      );
      if (!assigned) {
        phase.doer.models = [];
        isComplete = false;
      }
    }

    if (phase.reviewer?.candidates) {
      for (const slot of phase.reviewer.candidates) {
        const assigned = assignSlot(
          slot,
          byLineage,
          byFamily,
          usedLineages,
          usedTuples,
        );
        if (!assigned) {
          slot.models = [];
          isComplete = false;
        }
      }
    }
  }
  // changed flag — compare round-tripped JSON. Cheap, deterministic,
  // catches lineage swaps + model reductions in one shot.
  changed =
    JSON.stringify(parsed) !== JSON.stringify(yaml.parse(canonicalYaml));

  // Re-serialise. yaml lib preserves keys/order and quoting style well
  // enough for builtin templates that are themselves machine-authored.
  // Hand-edited templates aren't passed to this function (caller gates
  // on source='builtin').
  const output = yaml.stringify(parsed, { lineWidth: 0 });
  return { yaml: output, isComplete, changed };
}

/**
 * Assign a single slot. Returns true iff the slot was actually
 * written into. Sequential resolution:
 *   1. Exact lineage match.
 *   2. Vendor-family match (template wants moonshot, user has
 *      opencode-go/kimi with vendor_family=moonshot).
 *   3. Diversity-preserving substitute: a lineage not yet used in
 *      this phase. Maximizes cross-lineage reviewer diversity.
 *   4. Last-ditch: any available lineage (accepts duplication).
 *   5. No voices anywhere → returns false; caller marks incomplete.
 */
function assignSlot(
  slot: SlotSpec,
  byLineage: Map<string, Voice[]>,
  byFamily: Map<string, Voice[]>,
  usedInPhase: Set<string>,
  usedTuples: Set<string>,
): boolean {
  const preferredLineage = slot.lineage;
  if (!preferredLineage) return false;

  const direct = byLineage.get(preferredLineage);
  if (direct && direct.length > 0) {
    slot.lineage = preferredLineage;
    slot.models = voicesToModelKeys(direct, usedTuples);
    usedInPhase.add(preferredLineage);
    return true;
  }

  const familyMatch = byFamily.get(preferredLineage);
  if (familyMatch && familyMatch.length > 0) {
    slot.lineage = familyMatch[0].lineage;
    slot.models = voicesToModelKeys(familyMatch, usedTuples);
    usedInPhase.add(familyMatch[0].lineage);
    return true;
  }

  for (const [lineage, voices] of byLineage) {
    if (!usedInPhase.has(lineage)) {
      slot.lineage = lineage;
      slot.models = voicesToModelKeys(voices, usedTuples);
      usedInPhase.add(lineage);
      return true;
    }
  }

  for (const [lineage, voices] of byLineage) {
    slot.lineage = lineage;
    slot.models = voicesToModelKeys(voices, usedTuples);
    usedInPhase.add(lineage);
    return true;
  }

  return false;
}

/**
 * Voices reference one of two id forms in the wire/template space:
 *   - openrouter voices use the prefixed `id` (`openrouter:<model>`)
 *     because the runner's pickShimForVoice detects the prefix to
 *     dispatch via the HTTP shim (see chorus-086).
 *   - All other voices (CLI-backed) use the bare `model_id` because
 *     the CLI shim is keyed by lineage + model_id.
 *
 * Returns AT MOST 1 model — the highest-ranked voice for the lineage.
 * Pre-fix the adapter dumped every available voice into the slot's
 * fallback chain (per chorus-083), which produced 5+ form rows per
 * slot and confused users.
 *
 * `usedTuples` (lineage:model_id) tracks what's already been assigned
 * in this phase so multiple slots wanting the same lineage rotate
 * through available models instead of all picking the same top one.
 * Without rotation, review-only's 3 opencode reviewer slots all got
 * the same kimi model — wasted API calls and no internal diversity.
 */
export function voicesToModelKeys(
  voices: Voice[],
  usedTuples?: Set<string>,
): string[] {
  if (voices.length === 0) return [];
  const ranked = [...voices].sort(
    (a, b) => capabilityScore(b) - capabilityScore(a),
  );
  const tupleKey = (v: Voice): string => `${v.lineage}:${v.model_id}`;
  // Prefer the highest-ranked voice whose (lineage, model_id) tuple
  // hasn't been used elsewhere in this phase. If all are used, fall
  // through to the top — duplication is acceptable as last resort.
  let pick = ranked[0];
  if (usedTuples) {
    for (const v of ranked) {
      if (!usedTuples.has(tupleKey(v))) {
        pick = v;
        break;
      }
    }
    usedTuples.add(tupleKey(pick));
  }
  return [pick.provider === 'openrouter' ? pick.id : pick.model_id];
}

/**
 * Capability heuristic for ordering voices within a lineage.
 *
 * Picks the top model when the user has multiple voices for the slot's
 * lineage. Higher score = more capable. The function is per-family —
 * Anthropic Opus beats Sonnet beats Haiku, OpenAI gpt-5.5 beats 5.4,
 * Gemini 3.x beats 2.x, Pro beats Flash, and so on.
 *
 * Heuristic by design: model naming changes over time, so this scoring
 * will need touch-ups when new models land. Keep the logic in one
 * function to make those updates easy. Falls back to 0 (alphabetical
 * tiebreaker via the sort's stability) for unknown models — better
 * than picking arbitrarily.
 */
export function capabilityScore(v: Voice): number {
  const id = v.model_id.toLowerCase();
  let score = 0;

  // Anthropic family tier.
  if (id.includes('opus')) score += 1000;
  else if (id.includes('sonnet')) score += 700;
  else if (id.includes('haiku')) score += 400;

  // OpenAI / Codex (gpt-X.Y).
  const gptMatch = id.match(/gpt-(\d+)\.(\d+)/);
  if (gptMatch) {
    score += Number.parseInt(gptMatch[1], 10) * 100;
    score += Number.parseInt(gptMatch[2], 10) * 10;
    // -codex variant counts as production (vs preview).
    if (id.includes('-codex')) score += 5;
  }

  // Gemini family — major.minor with pro/flash modifier.
  const geminiMatch = id.match(/gemini-(\d+)\.(\d+)/);
  if (geminiMatch) {
    score += Number.parseInt(geminiMatch[1], 10) * 100;
    score += Number.parseInt(geminiMatch[2], 10) * 10;
    if (id.includes('pro')) score += 50;
    else if (id.includes('flash')) score += 20;
    if (id.includes('preview')) score += 5;
  }

  // Kimi (k2.x). Thinking variant scores higher than base.
  const kimiMatch = id.match(/k(\d+)\.(\d+)/);
  if (kimiMatch) {
    score += Number.parseInt(kimiMatch[1], 10) * 100;
    score += Number.parseInt(kimiMatch[2], 10) * 10;
    if (id.includes('thinking')) score += 50;
    if (id.includes('turbo')) score += 30;
  }

  // Anthropic version suffix (-4-7 > -4-6 > -4-5).
  const versionMatch = id.match(/-(\d+)-(\d+)$/);
  if (versionMatch) {
    score += Number.parseInt(versionMatch[1], 10) * 5;
    score += Number.parseInt(versionMatch[2], 10);
  }

  return score;
}

