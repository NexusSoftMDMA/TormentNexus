-- sql/migrations/034_wiki_topic_search_tldr.sql
--
-- Migration 034: extend wiki_topic_search to return the two TLDR columns
-- added in 033. Without this, hybrid_search wiki_preamble silently ignores
-- the v0.13 `level=` parameter and always serves body content via the
-- back-compat fallback path in service._wiki_injection_results.
--
-- Bug discovered post-033 deploy when comparing wiki_preamble token cost
-- across levels: all three returned identical token counts because the
-- RPC didn't surface the new columns to Python, so row.get("tldr_short")
-- was always None.
--
-- DROP + CREATE because CREATE OR REPLACE FUNCTION cannot alter RETURNS
-- TABLE — Postgres errors with "cannot change return type". The signature
-- (parameter list) is unchanged, only the return shape grows two columns.
--
-- Depends on: migration 031 (wiki_topic_search baseline) + 033 (tldr cols).
-- Idempotent: DROP IF EXISTS, then CREATE.

BEGIN;

DROP FUNCTION IF EXISTS wiki_topic_search(text, vector, integer, float);

CREATE OR REPLACE FUNCTION wiki_topic_search(
    p_profile text,
    p_query_embedding vector,
    p_top_k integer DEFAULT 3,
    p_min_similarity float DEFAULT 0.0
)
RETURNS TABLE (
    id uuid,
    topic_key text,
    profile_id text,
    content text,
    tldr_one_line text,
    tldr_short text,
    source_count integer,
    source_cursor uuid,
    source_hash bytea,
    model_used text,
    version integer,
    status text,
    updated_at timestamptz,
    similarity float
)
LANGUAGE sql
SECURITY INVOKER
SET search_path = public, extensions, pg_catalog
AS $$
    -- HNSW + threshold trap: combining `WHERE similarity >= threshold`
    -- with `ORDER BY <=> ... LIMIT k` defeats the index when the
    -- threshold filters out top-k. Postgres falls back to scanning the
    -- HNSW tail row-by-row. Wrap the index-driven top-k in a CTE,
    -- apply the threshold AFTER. The index path then runs unfiltered
    -- and the threshold trims the (already small) output.
    WITH top_k AS (
        SELECT id, topic_key, profile_id, content,
               tldr_one_line, tldr_short,
               source_count, source_cursor, source_hash,
               model_used, version, status, updated_at,
               1 - (embedding <=> p_query_embedding) AS similarity
          FROM topic_summaries
         WHERE profile_id = p_profile
           AND status = 'fresh'
           AND embedding IS NOT NULL
         ORDER BY embedding <=> p_query_embedding
         LIMIT p_top_k
    )
    SELECT * FROM top_k WHERE similarity >= p_min_similarity;
$$;

COMMIT;
