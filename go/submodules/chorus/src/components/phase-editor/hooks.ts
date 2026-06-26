import { useEffect, useState } from "react";
import { listVoices } from "@/lib/api/voices";
import { listPersonas, type Persona } from "@/lib/api/personas";
import type { ReviewerLineage } from "@/lib/cockpit-types";
import { DAEMON_TO_COCKPIT_LINEAGE } from "./constants";

export interface ConnectedVoiceMap {
  /** Per-cockpit-lineage list of enabled model_ids. */
  byLineage: Partial<Record<ReviewerLineage, string[]>>;
  /** Set of cockpit-lineages with at least one enabled voice. */
  connectedLineages: Set<ReviewerLineage>;
  /** True once the initial fetch settled (success or error). */
  loaded: boolean;
}

/**
 * Loads every enabled voice once and groups by cockpit-lineage so the
 * doer + reviewer dropdowns can show real options. Tolerates fetch
 * failure — falls back to empty maps; freeform fallback still lets the
 * user type a model id by hand.
 *
 * OpenCode-provider voices are intentionally double-bucketed: under
 * their model-family lineage AND under cockpit "opencode". Without
 * this, opencode-go/kimi only appeared under "Kimi" and never under
 * "OpenCode" even though the user picked OpenCode CLI as the path.
 *
 * OpenRouter voices are bucketed under their own cockpit "openrouter"
 * lineage and the picker value is the prefixed voice id
 * (`openrouter:<model_id>`) so the runner's pickShimForVoice can detect
 * the prefix and route through the OpenRouter HTTP shim. The voice's
 * underlying daemon lineage (anthropic/openai/...) stays on the row
 * for diversity scoring but isn't used for grouping.
 */
export function useConnectedVoices(): ConnectedVoiceMap {
  const [state, setState] = useState<ConnectedVoiceMap>({
    byLineage: {},
    connectedLineages: new Set(),
    loaded: false,
  });
  useEffect(() => {
    listVoices({ enabled: true })
      .then((voices) => {
        const byLineage: Partial<Record<ReviewerLineage, string[]>> = {};
        const connectedLineages = new Set<ReviewerLineage>();
        for (const v of voices) {
          if (v.provider === "openrouter") {
            connectedLineages.add("openrouter");
            (byLineage["openrouter"] ??= []).push(v.id);
            continue;
          }
          const cockpitLineage = DAEMON_TO_COCKPIT_LINEAGE[v.lineage];
          if (cockpitLineage) {
            connectedLineages.add(cockpitLineage);
            (byLineage[cockpitLineage] ??= []).push(v.model_id);
          }
          if (
            v.provider.startsWith("opencode") &&
            cockpitLineage !== "opencode"
          ) {
            connectedLineages.add("opencode");
            (byLineage["opencode"] ??= []).push(v.model_id);
          }
        }
        for (const k of Object.keys(byLineage) as ReviewerLineage[]) {
          byLineage[k] = Array.from(new Set(byLineage[k]!));
        }
        setState({ byLineage, connectedLineages, loaded: true });
      })
      .catch(() =>
        setState({ byLineage: {}, connectedLineages: new Set(), loaded: true }),
      );
  }, []);
  return state;
}

/**
 * Fetches the persona catalog once. Returns `[]` while loading and on
 * fetch failure so the picker just renders an empty list rather than
 * crashing the dialog. Persona is optional on every slot.
 */
export function usePersonas(): { personas: Persona[]; loaded: boolean } {
  const [state, setState] = useState<{ personas: Persona[]; loaded: boolean }>({
    personas: [],
    loaded: false,
  });
  useEffect(() => {
    listPersonas()
      .then((personas) => setState({ personas, loaded: true }))
      .catch(() => setState({ personas: [], loaded: true }));
  }, []);
  return state;
}
