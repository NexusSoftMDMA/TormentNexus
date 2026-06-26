/**
 * Re-export the brand gradient map from the single-source-of-truth so the
 * old import path keeps working. New code should pull from
 * `@/lib/lineage-maps` directly via `UI_LINEAGE_BRAND` or
 * `uiLineageGradient(lineage)`.
 */
import { UI_LINEAGE_BRAND, type UILineage } from "@/lib/lineage-maps";

export const LINEAGE_GRADIENT: Record<string, string> = Object.fromEntries(
  Object.entries(UI_LINEAGE_BRAND).map(([k, v]) => [k, v.gradient]),
) as Record<UILineage, string>;
