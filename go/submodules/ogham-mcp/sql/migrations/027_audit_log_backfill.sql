-- src/ogham/sql/migrations/027_audit_log_backfill.sql
--
-- Migration 027: backfill the audit_log table for Supabase (or Postgres)
-- deployments created before audit_log was added to the baseline schema.
--
-- Idempotent. Safe to run against any DB -- if audit_log already exists,
-- this is a no-op.
--
-- Matches the definition in sql/schema_postgres.sql (lines 827-847).

BEGIN;

CREATE TABLE IF NOT EXISTS audit_log (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_time timestamptz NOT NULL DEFAULT now(),
    profile text NOT NULL,
    operation text NOT NULL,              -- store | search | delete | update | re_embed
    resource_id uuid,                     -- memories.id if applicable
    outcome text NOT NULL DEFAULT 'success',
    source text,                          -- claude-code | cursor | gateway | cli
    embedding_model text,
    tokens_used integer,
    cost_usd numeric(10,6),
    result_ids uuid[],                    -- memory IDs returned by search
    result_count integer,
    query_hash text,                      -- sha256 of query text (not the query itself)
    metadata jsonb DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_audit_log_profile_time
    ON audit_log (profile, event_time DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_resource
    ON audit_log (resource_id) WHERE resource_id IS NOT NULL;

-- Mirror the "Deny anon access" RLS pattern used on memories + memory_lifecycle.
-- Only applies on Supabase (or any DB where the 'anon' role exists).
-- On plain self-hosted Postgres there is no 'anon' role, so RLS + policy are
-- skipped entirely -- service_role bypass is Supabase-specific and doesn't
-- apply there either.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
        RAISE NOTICE 'anon role not found -- skipping RLS setup for audit_log (non-Supabase install)';
        RETURN;
    END IF;

    IF NOT (SELECT rowsecurity FROM pg_tables
              WHERE tablename = 'audit_log' AND schemaname = 'public') THEN
        EXECUTE 'ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_policy p
          JOIN pg_class c ON c.oid = p.polrelid
         WHERE c.relname = 'audit_log' AND p.polname = 'Deny anon access'
    ) THEN
        EXECUTE $policy$
            CREATE POLICY "Deny anon access" ON audit_log
                FOR ALL TO anon
                USING (false) WITH CHECK (false)
        $policy$;
    END IF;
END$$;

COMMIT;
