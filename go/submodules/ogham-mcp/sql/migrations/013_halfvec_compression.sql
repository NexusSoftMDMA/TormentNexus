-- Migration 013: Switch HNSW index to halfvec (float16) for ~50% index size reduction
--
-- The embedding column stays vector(512) (full float32 precision at rest).
-- Only the HNSW index and search operations use halfvec(512) (float16).
-- This cuts index size roughly in half with negligible quality impact.
--
-- Requires pgvector >= 0.7.0 for halfvec support.
-- Neon: pgvector 0.7+ is available on all regions.
-- Supabase: pgvector 0.7+ available since late 2024.

-- Step 1: Replace the HNSW index with halfvec expression index
DROP INDEX IF EXISTS memories_embedding_idx;

CREATE INDEX memories_embedding_idx
    ON memories USING hnsw ((embedding::halfvec(512)) halfvec_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- Step 2: Update RPC functions to cast embeddings to halfvec in <=> operations
-- The cast is required for Postgres to use the halfvec HNSW index.

-- 2a. auto_link_memory
CREATE OR REPLACE FUNCTION auto_link_memory(
    new_memory_id uuid,
    new_embedding vector(512),
    link_threshold float DEFAULT 0.85,
    max_links int DEFAULT 5,
    filter_profile text DEFAULT 'default'
)
RETURNS integer
LANGUAGE sql
SECURITY INVOKER
SET search_path = public, extensions
AS $$
    WITH candidates AS (
        SELECT m.id, (1 - (m.embedding::halfvec(512) <=> new_embedding::halfvec(512)))::float AS similarity
        FROM memories m
        WHERE m.id != new_memory_id
          AND m.profile = filter_profile
          AND (m.expires_at IS NULL OR m.expires_at > now())
          AND 1 - (m.embedding::halfvec(512) <=> new_embedding::halfvec(512)) > link_threshold
        ORDER BY m.embedding::halfvec(512) <=> new_embedding::halfvec(512)
        LIMIT max_links
    ),
    inserted AS (
        INSERT INTO memory_relationships (source_id, target_id, relationship, strength, created_by)
        SELECT new_memory_id, c.id, 'similar', c.similarity, 'auto'
        FROM candidates c
        ON CONFLICT (source_id, target_id, relationship) DO NOTHING
        RETURNING 1
    )
    SELECT count(*)::integer FROM inserted;
$$;

-- 2b. match_memories
CREATE OR REPLACE FUNCTION match_memories(
    query_embedding vector(512),
    match_threshold float DEFAULT 0.7,
    match_count int DEFAULT 10,
    filter_tags text[] DEFAULT NULL,
    filter_source text DEFAULT NULL,
    filter_profile text DEFAULT 'default'
)
RETURNS TABLE (
    id uuid,
    content text,
    metadata jsonb,
    source text,
    profile text,
    tags text[],
    similarity float,
    relevance float,
    access_count integer,
    last_accessed_at timestamptz,
    confidence float,
    created_at timestamptz,
    updated_at timestamptz
)
LANGUAGE plpgsql
SECURITY INVOKER
SET search_path = public, extensions
AS $$
BEGIN
    RETURN QUERY
    SELECT
        m.id, m.content, m.metadata, m.source, m.profile, m.tags,
        (1 - (m.embedding::halfvec(512) <=> query_embedding::halfvec(512)))::float AS similarity,
        (
            (1 - (m.embedding::halfvec(512) <=> query_embedding::halfvec(512))) *
            ln(1.0 + exp(
                ln(m.access_count + 1.0) -
                0.5 * ln(
                    greatest(
                        extract(epoch from now() - coalesce(m.last_accessed_at, m.created_at)) / 86400.0,
                        0.01
                    ) / (m.access_count + 1.0)
                )
            ))
            * m.confidence
            * (1.0 + g.graph_boost * 0.2)
        )::float AS relevance,
        m.access_count, m.last_accessed_at, m.confidence, m.created_at, m.updated_at
    FROM public.memories m
    LEFT JOIN LATERAL (
        SELECT coalesce(sum(r.strength), 0.0) AS graph_boost
        FROM memory_relationships r
        WHERE r.target_id = m.id OR r.source_id = m.id
    ) g ON true
    WHERE
        1 - (m.embedding::halfvec(512) <=> query_embedding::halfvec(512)) > match_threshold
        AND (filter_tags IS NULL OR m.tags && filter_tags)
        AND (filter_source IS NULL OR m.source = filter_source)
        AND m.profile = filter_profile
        AND (m.expires_at IS NULL OR m.expires_at > now())
    ORDER BY relevance DESC
    LIMIT match_count;
END;
$$;

-- 2c. hybrid_search_memories (CCF version)
CREATE OR REPLACE FUNCTION hybrid_search_memories(
    query_text text,
    query_embedding vector,
    match_count integer DEFAULT 10,
    filter_profile text DEFAULT 'default',
    filter_tags text[] DEFAULT NULL,
    filter_source text DEFAULT NULL,
    full_text_weight float DEFAULT 0.3,
    semantic_weight float DEFAULT 0.7,
    rrf_k integer DEFAULT 10
)
RETURNS TABLE(
    id uuid, content text, metadata jsonb, source text, profile text, tags text[],
    similarity float, keyword_rank float, relevance float,
    access_count integer, last_accessed_at timestamptz, confidence float,
    created_at timestamptz, updated_at timestamptz
)
LANGUAGE sql
SET search_path = public, extensions
AS $function$
WITH semantic AS (
    SELECT
        m.id,
        (1 - (m.embedding::halfvec(512) <=> query_embedding::halfvec(512)))::float AS similarity
    FROM memories m
    WHERE m.profile = filter_profile
      AND (filter_tags IS NULL OR m.tags && filter_tags)
      AND (filter_source IS NULL OR m.source = filter_source)
      AND (m.expires_at IS NULL OR m.expires_at > now())
    ORDER BY m.embedding::halfvec(512) <=> query_embedding::halfvec(512)
    LIMIT match_count * 3
),
keyword AS (
    SELECT
        m.id,
        ts_rank_cd(m.fts, websearch_to_tsquery(query_text))::float AS keyword_rank
    FROM memories m
    WHERE m.profile = filter_profile
      AND m.fts @@ websearch_to_tsquery(query_text)
      AND (filter_tags IS NULL OR m.tags && filter_tags)
      AND (filter_source IS NULL OR m.source = filter_source)
      AND (m.expires_at IS NULL OR m.expires_at > now())
    ORDER BY keyword_rank DESC
    LIMIT match_count * 3
),
fused AS (
    SELECT
        coalesce(s.id, k.id) AS id,
        coalesce(s.similarity, 0.0) AS similarity,
        coalesce(k.keyword_rank, 0.0) AS keyword_rank,
        (
            semantic_weight * coalesce(s.similarity, 0.0)
            + full_text_weight * coalesce(k.keyword_rank, 0.0)
        ) AS score
    FROM semantic s
    FULL OUTER JOIN keyword k ON s.id = k.id
)
SELECT
    m.id, m.content, m.metadata, m.source, m.profile, m.tags,
    f.similarity, f.keyword_rank,
    (
        f.score
        * (1.0 + ln(m.access_count + 1.0) * 0.1)
        * m.confidence
        * (1.0 + g.graph_boost * 0.2)
    )::float AS relevance,
    m.access_count, m.last_accessed_at, m.confidence, m.created_at, m.updated_at
FROM fused f
JOIN memories m ON m.id = f.id
LEFT JOIN LATERAL (
    SELECT coalesce(sum(r.strength), 0.0) AS graph_boost
    FROM memory_relationships r
    WHERE r.target_id = m.id OR r.source_id = m.id
) g ON true
ORDER BY relevance DESC
LIMIT match_count;
$function$;

-- 2d. batch_check_duplicates
CREATE OR REPLACE FUNCTION batch_check_duplicates(
    query_embeddings vector(512)[],
    match_threshold float DEFAULT 0.8,
    filter_profile text DEFAULT 'default'
)
RETURNS boolean[]
LANGUAGE plpgsql
SECURITY INVOKER
SET search_path = public, extensions
AS $$
DECLARE
    results boolean[];
    i integer;
    found boolean;
BEGIN
    PERFORM set_config('hnsw.ef_search', '40', true);
    results := array[]::boolean[];
    FOR i IN 1..array_length(query_embeddings, 1) LOOP
        SELECT exists(
            SELECT 1 FROM memories m
            WHERE m.profile = filter_profile
              AND (m.expires_at IS NULL OR m.expires_at > now())
              AND 1 - (m.embedding::halfvec(512) <=> query_embeddings[i]::halfvec(512)) > match_threshold
            LIMIT 1
        ) INTO found;
        results := array_append(results, found);
    END LOOP;
    RETURN results;
END;
$$;
