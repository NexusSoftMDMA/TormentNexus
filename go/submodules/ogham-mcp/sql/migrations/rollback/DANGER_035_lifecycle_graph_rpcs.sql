-- DANGER_035: rollback of migration 035_lifecycle_graph_rpcs.sql
--
-- Drops the seven RPC functions defined by migration 035. Pure DDL drops
-- on functions that were added (not modified), so the rollback is fully
-- reversible by re-running 035. No data loss -- functions only.
--
-- This file follows the v0.12 DANGER_* convention:
--   * Filename prefix DANGER_ flags it as destructive
--   * Session-variable guard inside the BEGIN transaction (per the v0.13
--     test_danger_rollback_guards_live_inside_transaction test pattern)
--   * The harness apply_rollback() helper sets ogham.confirm_rollback
--     before executing
--
-- Manual usage:
--     SET ogham.confirm_rollback = 'I-KNOW-WHAT-I-AM-DOING';
--     \i sql/migrations/rollback/DANGER_035_lifecycle_graph_rpcs.sql

BEGIN;

DO $$
BEGIN
    IF current_setting('ogham.confirm_rollback', true) IS DISTINCT FROM 'I-KNOW-WHAT-I-AM-DOING' THEN
        RAISE EXCEPTION 'Refusing to run DANGER_035 rollback. Set ogham.confirm_rollback = ''I-KNOW-WHAT-I-AM-DOING'' first.';
    END IF;
END
$$;

DROP FUNCTION IF EXISTS public.suggest_unlinked_by_shared_entities(uuid, text, integer, integer);
DROP FUNCTION IF EXISTS public.entity_graph_density(text);
DROP FUNCTION IF EXISTS public.hebbian_strengthen_edges(text[], text[], real, real);
DROP FUNCTION IF EXISTS public.lifecycle_pipeline_counts(text);
DROP FUNCTION IF EXISTS public.lifecycle_open_editing_window(uuid[]);
DROP FUNCTION IF EXISTS public.lifecycle_close_editing_windows(text, timestamptz);
DROP FUNCTION IF EXISTS public.lifecycle_advance_fresh_to_stable(text, timestamptz, float, float);

COMMIT;
