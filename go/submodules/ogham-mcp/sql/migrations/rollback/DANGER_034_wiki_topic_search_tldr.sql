-- sql/migrations/rollback/DANGER_034_wiki_topic_search_tldr.sql
--
-- ROLLBACK for migration 034. Restores the v0.12-era wiki_topic_search
-- signature (12 columns, no tldr_one_line / tldr_short). Use only when
-- rolling back a v0.13.x deployment to v0.12.x — leaving the function
-- with the 034 shape on a v0.12 codebase doesn't break anything (the
-- old code just ignores the new columns), but explicit rollback keeps
-- the schema strictly v0.12-shaped.
--
-- Guard: requires session variable `ogham.confirm_rollback` to be set
-- to 'I-KNOW-WHAT-I-AM-DOING' before any DDL runs. Test harness sets
-- this via `_Harness.apply_rollback`.

BEGIN;

-- Guard must live INSIDE the transaction so a missing session variable
-- aborts the whole statement and rolls back any DDL that follows.
DO $$
BEGIN
    IF current_setting('ogham.confirm_rollback', true) IS DISTINCT FROM 'I-KNOW-WHAT-I-AM-DOING' THEN
        RAISE EXCEPTION
            'rollback gate: set ogham.confirm_rollback = ''I-KNOW-WHAT-I-AM-DOING'' first'
            USING HINT = 'this drops + recreates wiki_topic_search to the v0.12 shape';
    END IF;
END;
$$;

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
    WITH top_k AS (
        SELECT id, topic_key, profile_id, content, source_count,
               source_cursor, source_hash, model_used, version, status,
               updated_at,
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
