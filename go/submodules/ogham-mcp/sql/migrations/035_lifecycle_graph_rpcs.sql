-- Migration 035: Supabase RPC parity for lifecycle, graph, density, and
-- suggest_connections operations.
--
-- Why this migration exists
-- =========================
-- v0.13 health audit caught six call sites outside `backends/` calling
-- `backend._execute(...)` directly. `PostgresBackend` has `_execute`;
-- `SupabaseBackend` does NOT (PostgREST has no SQL passthrough). Result:
-- the six operations have been silently failing on Supabase since v0.11.
--
-- The cleanest fix is to publish the operations as RPC functions, accessible
-- to both backends through the standard PostgREST surface. Postgres can keep
-- its fast `_execute` path internally; Supabase calls the RPC over HTTP.
--
-- The seven functions defined below:
--   * `lifecycle_advance_fresh_to_stable`     -- fresh -> stable transition
--   * `lifecycle_close_editing_windows`       -- close stale editing windows
--   * `lifecycle_open_editing_window`         -- flip stable -> editing
--   * `lifecycle_pipeline_counts`             -- dashboard stage counts
--   * `hebbian_strengthen_edges`              -- co-retrieval edge UPSERT
--   * `entity_graph_density`                  -- entities / edges per profile
--   * `suggest_unlinked_by_shared_entities`   -- "hidden links" tool body
--
-- All seven follow the project conventions:
--   * `SECURITY DEFINER` so callers don't need direct table grants
--   * `SET search_path = public, extensions, pg_catalog` so pgvector
--     operators resolve on Supabase (the gotcha behind migration 031)
--   * `LANGUAGE plpgsql` (or `sql` where the body is a single statement)
--   * `CREATE OR REPLACE FUNCTION` -- idempotent, safe to re-apply
--
-- Rollback: see sql/migrations/rollback/DANGER_035_lifecycle_graph_rpcs.sql

BEGIN;

-- ── 1. lifecycle_advance_fresh_to_stable ─────────────────────────────────
CREATE OR REPLACE FUNCTION public.lifecycle_advance_fresh_to_stable(
    p_profile text,
    p_cutoff  timestamptz,
    p_s_gate  float,
    p_i_gate  float
)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public, extensions, pg_catalog
AS $$
DECLARE
    v_count integer;
BEGIN
    WITH advanced AS (
        UPDATE memory_lifecycle AS ml
           SET stage            = 'stable',
               stage_entered_at = now(),
               updated_at       = now()
          FROM memories AS m
         WHERE ml.memory_id        = m.id
           AND ml.profile          = p_profile
           AND ml.stage            = 'fresh'
           AND ml.stage_entered_at <= p_cutoff
           AND (m.surprise >= p_s_gate OR m.importance >= p_i_gate)
        RETURNING ml.memory_id
    )
    SELECT count(*)::integer INTO v_count FROM advanced;
    RETURN v_count;
END;
$$;

-- ── 2. lifecycle_close_editing_windows ───────────────────────────────────
CREATE OR REPLACE FUNCTION public.lifecycle_close_editing_windows(
    p_profile text,
    p_cutoff  timestamptz
)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public, extensions, pg_catalog
AS $$
DECLARE
    v_count integer;
BEGIN
    WITH closed AS (
        UPDATE memory_lifecycle
           SET stage            = 'stable',
               stage_entered_at = now(),
               updated_at       = now()
         WHERE profile           = p_profile
           AND stage             = 'editing'
           AND stage_entered_at <= p_cutoff
        RETURNING memory_id
    )
    SELECT count(*)::integer INTO v_count FROM closed;
    RETURN v_count;
END;
$$;

-- ── 3. lifecycle_open_editing_window ─────────────────────────────────────
CREATE OR REPLACE FUNCTION public.lifecycle_open_editing_window(
    p_ids uuid[]
)
RETURNS void
LANGUAGE sql
SECURITY DEFINER
SET search_path = public, extensions, pg_catalog
AS $$
    UPDATE memory_lifecycle
       SET stage            = 'editing',
           stage_entered_at = now(),
           updated_at       = now()
     WHERE memory_id = ANY(p_ids)
       AND stage = 'stable';
$$;

-- ── 4. lifecycle_pipeline_counts ─────────────────────────────────────────
CREATE OR REPLACE FUNCTION public.lifecycle_pipeline_counts(
    p_profile text
)
RETURNS TABLE (stage text, n bigint)
LANGUAGE sql
SECURITY DEFINER
SET search_path = public, extensions, pg_catalog
AS $$
    SELECT stage, count(*)::bigint AS n
      FROM memory_lifecycle
     WHERE profile = p_profile
     GROUP BY stage;
$$;

-- ── 5. hebbian_strengthen_edges ──────────────────────────────────────────
CREATE OR REPLACE FUNCTION public.hebbian_strengthen_edges(
    p_sources    text[],
    p_targets    text[],
    p_bootstrap  real,
    p_rate       real
)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public, extensions, pg_catalog
AS $$
DECLARE
    v_count integer;
BEGIN
    -- Caller is responsible for canonicalising pairs (sorted) -- see
    -- src/ogham/graph.py docstring for deadlock + idempotency rationale.
    WITH touched AS (
        INSERT INTO memory_relationships
            (source_id, target_id, relationship, strength, created_by)
        SELECT s::uuid, t::uuid, 'related', p_bootstrap, 'hebbian'
          FROM unnest(p_sources, p_targets) AS p(s, t)
        ON CONFLICT (source_id, target_id, relationship) DO UPDATE
            SET strength = LEAST(1.0,
                                 memory_relationships.strength * (1 + p_rate))
        RETURNING source_id
    )
    SELECT count(*)::integer INTO v_count FROM touched;
    RETURN v_count;
END;
$$;

-- ── 6. entity_graph_density ──────────────────────────────────────────────
--
-- Both 6 and 7 depend on the v0.10 `memory_entities` + `entities` tables
-- (added to bootstrap schemas but never retrofitted via a standalone
-- migration -- see #161). Older deployments may not have them.
--
-- We use LANGUAGE plpgsql + a `to_regclass()` runtime guard so the
-- CREATE statement parses cleanly even on databases without those
-- tables, and the function returns empty/zero results at call time
-- instead of raising. Once `memory_entities` is later added (or the
-- profile populates entries), the function behaves correctly without
-- a re-create.
CREATE OR REPLACE FUNCTION public.entity_graph_density(
    p_profile text
)
RETURNS TABLE (entities double precision, edges double precision)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public, extensions, pg_catalog
AS $$
BEGIN
    IF to_regclass('public.memory_entities') IS NULL THEN
        RETURN QUERY SELECT 0.0::double precision, 0.0::double precision;
        RETURN;
    END IF;
    RETURN QUERY EXECUTE
        'SELECT
            count(DISTINCT entity_id)::double precision AS entities,
            count(*)::double precision                  AS edges
           FROM memory_entities
          WHERE profile = $1'
        USING p_profile;
END;
$$;

-- ── 7. suggest_unlinked_by_shared_entities ───────────────────────────────
CREATE OR REPLACE FUNCTION public.suggest_unlinked_by_shared_entities(
    p_memory_id  uuid,
    p_profile    text,
    p_min_shared integer,
    p_limit      integer
)
RETURNS TABLE (
    id              text,
    shared_count    bigint,
    shared_entities text[],
    content         text,
    created_at      timestamptz,
    tags            text[]
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public, extensions, pg_catalog
AS $$
BEGIN
    -- Same to_regclass guard as entity_graph_density. Both `entities`
    -- and `memory_entities` are needed; check the more-derived one.
    IF to_regclass('public.memory_entities') IS NULL
       OR to_regclass('public.entities') IS NULL THEN
        RETURN;
    END IF;
    RETURN QUERY EXECUTE
        'WITH target_entities AS (
            SELECT entity_id FROM memory_entities
             WHERE memory_id = $1
        ),
        shared AS (
            SELECT
                me.memory_id,
                count(*)::bigint AS shared_count,
                array_agg(e.entity_type || '':'' || e.canonical_name) AS shared_entities
              FROM memory_entities me
              JOIN target_entities te ON te.entity_id = me.entity_id
              JOIN entities e         ON e.id        = me.entity_id
             WHERE me.memory_id != $1
               AND me.profile    = $2
             GROUP BY me.memory_id
            HAVING count(*) >= $3
        ),
        unlinked AS (
            SELECT s.* FROM shared s
             WHERE NOT EXISTS (
                 SELECT 1 FROM memory_relationships mr
                  WHERE (mr.source_id = $1 AND mr.target_id = s.memory_id)
                     OR (mr.target_id = $1 AND mr.source_id = s.memory_id)
             )
        )
        SELECT
            u.memory_id::text AS id,
            u.shared_count,
            u.shared_entities,
            m.content,
            m.created_at,
            m.tags
          FROM unlinked u
          JOIN memories m ON m.id = u.memory_id
         WHERE m.expires_at IS NULL OR m.expires_at > now()
         ORDER BY u.shared_count DESC, m.created_at DESC
         LIMIT $4'
        USING p_memory_id, p_profile, p_min_shared, p_limit;
END;
$$;

-- ── Grants ───────────────────────────────────────────────────────────────
-- The seven functions are SECURITY DEFINER, so granting EXECUTE to
-- authenticated + service_role is sufficient. anon stays denied (matches the
-- v0.12 RLS pattern from migration 032).
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated') THEN
        EXECUTE 'GRANT EXECUTE ON FUNCTION public.lifecycle_advance_fresh_to_stable(text, timestamptz, float, float) TO authenticated, service_role';
        EXECUTE 'GRANT EXECUTE ON FUNCTION public.lifecycle_close_editing_windows(text, timestamptz) TO authenticated, service_role';
        EXECUTE 'GRANT EXECUTE ON FUNCTION public.lifecycle_open_editing_window(uuid[]) TO authenticated, service_role';
        EXECUTE 'GRANT EXECUTE ON FUNCTION public.lifecycle_pipeline_counts(text) TO authenticated, service_role';
        EXECUTE 'GRANT EXECUTE ON FUNCTION public.hebbian_strengthen_edges(text[], text[], real, real) TO authenticated, service_role';
        EXECUTE 'GRANT EXECUTE ON FUNCTION public.entity_graph_density(text) TO authenticated, service_role';
        EXECUTE 'GRANT EXECUTE ON FUNCTION public.suggest_unlinked_by_shared_entities(uuid, text, integer, integer) TO authenticated, service_role';
    END IF;
END
$$;

COMMIT;
