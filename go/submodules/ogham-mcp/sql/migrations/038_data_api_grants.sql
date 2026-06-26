-- Migration 038: explicit Data API grants for service_role
--
-- WHY THIS MIGRATION EXISTS
-- On 2026-04-28, Supabase announced a breaking change to platform
-- defaults. Tables created in the `public` schema will no longer be
-- automatically exposed to the Data API (PostgREST + GraphQL +
-- supabase-js) via implicit grants to anon / authenticated /
-- service_role. The new behaviour becomes the default for all new
-- projects on 2026-05-30 and is applied to all existing projects on
-- 2026-10-30.
-- Source: https://github.com/orgs/supabase/discussions/45329
--
-- Ogham's Python backend talks to Supabase through PostgREST using the
-- service_role / sb_secret_ key (see CLAUDE.md gotchas: anon is
-- explicitly denied by our RLS policies). Without explicit
-- table-level GRANTs, PostgREST will start returning
--   {"code":"42501","message":"permission denied for table ..."}
-- once Supabase's platform-level default grant is revoked on existing
-- projects.
--
-- This migration adds explicit GRANTs on every Ogham table in `public`
-- to `service_role` only. We deliberately do NOT grant to anon or
-- authenticated:
--   * anon is blocked at the RLS layer via "Deny anon access" policies
--     on every table, and locked out of RPC EXECUTE in migration 037.
--   * authenticated is currently unused by Ogham's access patterns.
--
-- RLS policies and the migration 037 lockdown both remain in force.
-- This migration is additive, idempotent, and reversible (see
-- DANGER_038_data_api_grants.sql).
--
-- VANILLA-POSTGRES GUARD
-- The Supabase-specific `service_role` does not exist on vanilla
-- Postgres installs (Neon, self-hosted PG, etc.). Following the
-- v0.14.1 pattern set by migrations 032/036/037, we wrap the GRANTs
-- in a DO block that checks `pg_roles` and no-ops with a NOTICE on
-- non-Supabase installs. Vanilla Postgres users connect via psycopg
-- with a role that already has full access to its own tables.
--
-- COVERAGE (9 tables)
--   Defined in schema.sql:
--     * memories
--     * profile_settings
--     * memory_lifecycle
--     * memory_relationships
--     * audit_log
--     * entities
--     * memory_entities
--   Defined in migration 028:
--     * topic_summaries
--     * topic_summary_sources
--
-- Sequences in public are also granted for completeness.
--
-- VERIFICATION (Supabase only)
-- After running, this query should return 4 rows per table (one row
-- per SELECT, INSERT, UPDATE, DELETE) for each of the 9 tables = 36
-- rows total:
--
--   SELECT table_name, privilege_type
--     FROM information_schema.table_privileges
--    WHERE table_schema = 'public'
--      AND grantee = 'service_role'
--      AND table_name IN (
--          'memories', 'profile_settings', 'memory_lifecycle',
--          'memory_relationships', 'audit_log', 'entities',
--          'memory_entities', 'topic_summaries',
--          'topic_summary_sources')
--    ORDER BY table_name, privilege_type;

BEGIN;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'service_role') THEN
        RAISE NOTICE
            'service_role not found -- skipping Data API grants '
            '(non-Supabase install). On vanilla Postgres, the psycopg '
            'connection role already has full access to its own tables.';
        RETURN;
    END IF;

    EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON public.memories             TO service_role';
    EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON public.profile_settings     TO service_role';
    EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON public.memory_lifecycle     TO service_role';
    EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON public.memory_relationships TO service_role';
    EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON public.audit_log            TO service_role';
    EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON public.entities             TO service_role';
    EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON public.memory_entities      TO service_role';
    EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON public.topic_summaries       TO service_role';
    EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON public.topic_summary_sources TO service_role';

    EXECUTE 'GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO service_role';
END
$$;

COMMIT;
