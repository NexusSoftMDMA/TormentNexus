-- 040_fix_lint_contradiction_filter.sql (#262)
--
-- Fix: wiki_lint_contradictions (migration 031) joined only the SOURCE endpoint
-- to memories and filtered m.profile = p_profile on that side alone. That has
-- two consequences:
--   * a cross-profile edge whose source happens to sit in p_profile was counted
--     even though the contradiction is not within-profile (over-count), and
--   * an edge whose source moved to another profile but whose target is still
--     in p_profile was invisible (under-count).
-- A within-profile contradiction means BOTH memories live in the profile, so
-- join and filter BOTH endpoints -- mirroring the both-endpoint scoping that
-- migration 039 (gap_contradictions_for_ids) already uses.
--
-- Behaviour change: lint_wiki contradiction counts now reflect only edges with
-- both endpoints in the profile. In normal operation the auto-linker only
-- creates within-profile edges (it searches the active profile and links to
-- same-profile memories), so realistic counts are unchanged; this hardens the
-- filter against cross-profile edges. RETURNS TABLE shape, ORDER, LIMIT, and
-- the SECURITY/search_path attributes are byte-identical to 031 so
-- wiki_lint.find_contradictions is otherwise untouched.
--
-- Note: not back-ported into 031 in place. 031 is already shipped, and
-- wiki_lint_contradictions is not inlined in any schema*.sql -- fresh installs
-- build it by running migrations in order (031 then 040), so they pick up this
-- corrected definition. Editing a shipped migration would violate migration
-- immutability for no benefit (the dual-tree hash gate is gone post-Phase-B).
CREATE OR REPLACE FUNCTION wiki_lint_contradictions(
    p_profile text,
    p_sample_size integer DEFAULT 10
)
RETURNS TABLE (
    source_id text,
    target_id text,
    strength float,
    created_at timestamptz,
    total_count bigint
)
LANGUAGE sql
SECURITY INVOKER
SET search_path = public, extensions, pg_catalog
AS $$
    WITH all_pairs AS (
        SELECT mr.source_id, mr.target_id, mr.strength, mr.created_at
          FROM memory_relationships mr
          JOIN memories ms ON ms.id = mr.source_id AND ms.profile = p_profile
          JOIN memories mt ON mt.id = mr.target_id AND mt.profile = p_profile
         WHERE mr.relationship = 'contradicts'
    )
    SELECT mr.source_id::text, mr.target_id::text, mr.strength, mr.created_at,
           (SELECT count(*) FROM all_pairs)
      FROM all_pairs mr
     ORDER BY mr.created_at DESC
     LIMIT p_sample_size;
$$;
