-- DANGER_040: rollback of migration 040_fix_lint_contradiction_filter.sql
--
-- Restores the pre-040 wiki_lint_contradictions definition (migration 031's
-- source-side-only profile filter). This re-introduces the known under/over-
-- count bug, so it is guarded: the confirm_rollback check sits AFTER BEGIN; so
-- a missing session variable aborts the whole transaction before the function
-- is replaced. Piping this file naively (psql -f without ON_ERROR_STOP) FAILS
-- by design.
--
-- Manual usage:
--     SET ogham.confirm_rollback = 'I-KNOW-WHAT-I-AM-DOING';
--     \i sql/migrations/rollback/DANGER_040_fix_lint_contradiction_filter.sql

BEGIN;

DO $$
BEGIN
    IF current_setting('ogham.confirm_rollback', true) IS DISTINCT FROM 'I-KNOW-WHAT-I-AM-DOING' THEN
        RAISE EXCEPTION 'Refusing to run DANGER_040 rollback. Set ogham.confirm_rollback = ''I-KNOW-WHAT-I-AM-DOING'' first.';
    END IF;
END
$$;

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
          JOIN memories m ON m.id = mr.source_id
         WHERE mr.relationship = 'contradicts'
           AND m.profile = p_profile
    )
    SELECT mr.source_id::text, mr.target_id::text, mr.strength, mr.created_at,
           (SELECT count(*) FROM all_pairs)
      FROM all_pairs mr
     ORDER BY mr.created_at DESC
     LIMIT p_sample_size;
$$;

COMMIT;
