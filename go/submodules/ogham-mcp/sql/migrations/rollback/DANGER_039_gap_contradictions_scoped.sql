-- DANGER_039: rollback of migration 039_gap_contradictions_scoped.sql
--
-- Drops the gap-analysis result-ID-scoped contradiction function
-- (gap_contradictions_for_ids). Read-only helper, so rollback is low-risk,
-- but it follows the standard DANGER guard pattern: the confirm_rollback
-- check sits AFTER BEGIN; so a missing session variable aborts the
-- transaction before the DROP runs.
--
-- Manual usage:
--     SET ogham.confirm_rollback = 'I-KNOW-WHAT-I-AM-DOING';
--     \i sql/migrations/rollback/DANGER_039_gap_contradictions_scoped.sql

BEGIN;

DO $$
BEGIN
    IF current_setting('ogham.confirm_rollback', true) IS DISTINCT FROM 'I-KNOW-WHAT-I-AM-DOING' THEN
        RAISE EXCEPTION 'Refusing to run DANGER_039 rollback. Set ogham.confirm_rollback = ''I-KNOW-WHAT-I-AM-DOING'' first.';
    END IF;
END
$$;

DROP FUNCTION IF EXISTS gap_contradictions_for_ids(text, uuid[], integer);

COMMIT;
