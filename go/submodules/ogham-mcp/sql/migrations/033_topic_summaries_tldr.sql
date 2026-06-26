-- sql/migrations/033_topic_summaries_tldr.sql
--
-- Migration 033: multi-resolution topic summaries (v0.13 TLDR spike).
--
-- Adds two new TEXT columns to topic_summaries:
--   * tldr_one_line  -- single sentence, ~30-50 tokens
--   * tldr_short     -- one paragraph, ~150-300 tokens
--
-- The existing `content` column keeps the full body (~1000 words). Callers
-- pick the form that fits their context budget via the `level` param on
-- query_topic_summary; hybrid_search defaults to injecting tldr_short
-- in the wiki preamble (vs body in v0.12, which was heavy).
--
-- Both columns are NULLABLE. Existing rows keep working unchanged; the
-- next compile_wiki run for a topic populates them. No backfill task --
-- lazy regeneration is consistent with how stale rows are handled.
--
-- Generation strategy: a single LLM call produces all three forms via
-- JSON-structured output. Same source-memory load, marginal output cost.
-- The Python compile path passes all three to wiki_topic_upsert in one
-- transaction so the row is never partially populated.
--
-- Depends on: migration 028 (topic_summaries table) + 031 (wiki RPC
-- functions, of which wiki_topic_upsert needs a signature update for the
-- two new params).

BEGIN;

-- ── Table columns ────────────────────────────────────────────────────

ALTER TABLE topic_summaries
    ADD COLUMN IF NOT EXISTS tldr_one_line text,
    ADD COLUMN IF NOT EXISTS tldr_short text;

-- ── wiki_topic_upsert: new signature with tldr params ────────────────
--
-- CREATE OR REPLACE FUNCTION cannot change parameter list (only body),
-- so we DROP and CREATE. Old callers passing the original 10 positional
-- args still work because the two new params have DEFAULT NULL at the
-- end -- but Python callers using named params (which is everything in
-- src/ogham/backends/) get the new fields naturally as they're added.

DROP FUNCTION IF EXISTS wiki_topic_upsert(
    text, text, text, vector, uuid[], text, uuid, bytea, integer, float
);

CREATE OR REPLACE FUNCTION wiki_topic_upsert(
    p_profile text,
    p_topic_key text,
    p_content text,
    p_embedding vector,
    p_source_memory_ids uuid[],
    p_model_used text,
    p_source_cursor uuid,
    p_source_hash bytea,
    p_token_count integer DEFAULT NULL,
    p_importance float DEFAULT 0.5,
    p_tldr_one_line text DEFAULT NULL,
    p_tldr_short text DEFAULT NULL
)
RETURNS topic_summaries
LANGUAGE plpgsql
SECURITY INVOKER
SET search_path = public, extensions, pg_catalog
AS $$
DECLARE
    upserted topic_summaries;
BEGIN
    INSERT INTO topic_summaries (
        topic_key, profile_id, content, embedding,
        source_count, source_cursor, source_hash,
        token_count, importance, model_used,
        tldr_one_line, tldr_short
    )
    VALUES (
        p_topic_key, p_profile, p_content, p_embedding,
        cardinality(p_source_memory_ids), p_source_cursor, p_source_hash,
        p_token_count, p_importance, p_model_used,
        p_tldr_one_line, p_tldr_short
    )
    ON CONFLICT (profile_id, topic_key) DO UPDATE SET
        content = EXCLUDED.content,
        embedding = EXCLUDED.embedding,
        source_count = EXCLUDED.source_count,
        source_cursor = EXCLUDED.source_cursor,
        source_hash = EXCLUDED.source_hash,
        token_count = EXCLUDED.token_count,
        importance = EXCLUDED.importance,
        model_used = EXCLUDED.model_used,
        tldr_one_line = EXCLUDED.tldr_one_line,
        tldr_short = EXCLUDED.tldr_short,
        version = topic_summaries.version + 1,
        status = 'fresh',
        stale_reason = NULL
    RETURNING * INTO upserted;

    -- Concurrent-delete safety: if another transaction deleted the topic
    -- between our row-lock release and the RETURNING, INSERT...DO UPDATE
    -- can yield zero rows. Bail rather than crash on the FK insert.
    IF upserted.id IS NULL THEN
        RETURN NULL;
    END IF;

    DELETE FROM topic_summary_sources WHERE summary_id = upserted.id;

    -- JOIN against memories so concurrently-deleted memory ids drop
    -- silently instead of throwing a FK violation. Wiki content is a
    -- best-effort snapshot; missing one source is preferable to
    -- failing the whole upsert.
    INSERT INTO topic_summary_sources (summary_id, memory_id)
    SELECT upserted.id, m.id
      FROM unnest(p_source_memory_ids) AS t(id)
      JOIN memories m ON m.id = t.id
    ON CONFLICT DO NOTHING;

    RETURN upserted;
END;
$$;

COMMIT;
