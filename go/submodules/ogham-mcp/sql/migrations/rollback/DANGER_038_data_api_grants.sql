-- DANGER_038: rollback of migration 038_data_api_grants.sql
--
-- Revokes the explicit Data API GRANTs that migration 038 added.
-- **Only run this if you have a specific reason to roll back to
-- Supabase's pre-2026-04-28 default-grant behaviour** (which is being
-- phased out entirely by 2026-10-30 -- see
-- https://github.com/orgs/supabase/discussions/45329).
--
-- Running this rollback on a Supabase project that has already passed
-- the 2026-10-30 enforcement date will leave Ogham unable to talk to
-- Supabase through PostgREST. Only useful on still-grandfathered
-- projects where Supabase's old default behaviour is still active.
--
-- Same vanilla-Postgres guard as the forward migration: no-ops with a
-- NOTICE if `service_role` doesn't exist.
--
-- Manual usage:
--     SET ogham.confirm_rollback = 'I-KNOW-WHAT-I-AM-DOING';
--     \i sql/migrations/rollback/DANGER_038_data_api_grants.sql

BEGIN;

DO $$
BEGIN
    IF current_setting('ogham.confirm_rollback', true) IS DISTINCT FROM 'I-KNOW-WHAT-I-AM-DOING' THEN
        RAISE EXCEPTION 'Refusing to run DANGER_038 rollback. Set ogham.confirm_rollback = ''I-KNOW-WHAT-I-AM-DOING'' first.';
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'service_role') THEN
        RAISE NOTICE
            'service_role not found -- skipping Data API revokes '
            '(non-Supabase install).';
        RETURN;
    END IF;

    EXECUTE 'REVOKE SELECT, INSERT, UPDATE, DELETE ON public.memories             FROM service_role';
    EXECUTE 'REVOKE SELECT, INSERT, UPDATE, DELETE ON public.profile_settings     FROM service_role';
    EXECUTE 'REVOKE SELECT, INSERT, UPDATE, DELETE ON public.memory_lifecycle     FROM service_role';
    EXECUTE 'REVOKE SELECT, INSERT, UPDATE, DELETE ON public.memory_relationships FROM service_role';
    EXECUTE 'REVOKE SELECT, INSERT, UPDATE, DELETE ON public.audit_log            FROM service_role';
    EXECUTE 'REVOKE SELECT, INSERT, UPDATE, DELETE ON public.entities             FROM service_role';
    EXECUTE 'REVOKE SELECT, INSERT, UPDATE, DELETE ON public.memory_entities      FROM service_role';
    EXECUTE 'REVOKE SELECT, INSERT, UPDATE, DELETE ON public.topic_summaries       FROM service_role';
    EXECUTE 'REVOKE SELECT, INSERT, UPDATE, DELETE ON public.topic_summary_sources FROM service_role';

    EXECUTE 'REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM service_role';
END
$$;

COMMIT;
