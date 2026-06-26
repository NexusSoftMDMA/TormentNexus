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
with semantic as (
    select
        m.id,
        (1 - (m.embedding <=> query_embedding))::float as similarity
    from memories m
    where m.profile = filter_profile
      and (filter_tags is null or m.tags && filter_tags)
      and (filter_source is null or m.source = filter_source)
      and (m.expires_at is null or m.expires_at > now())
    order by m.embedding <=> query_embedding
    limit match_count * 3
),
keyword as (
    select
        m.id,
        ts_rank_cd(m.fts, websearch_to_tsquery(query_text))::float as keyword_rank
    from memories m
    where m.profile = filter_profile
      and m.fts @@ websearch_to_tsquery(query_text)
      and (filter_tags is null or m.tags && filter_tags)
      and (filter_source is null or m.source = filter_source)
      and (m.expires_at is null or m.expires_at > now())
    order by keyword_rank desc
    limit match_count * 3
),
fused as (
    select
        coalesce(s.id, k.id) as id,
        coalesce(s.similarity, 0.0) as similarity,
        coalesce(k.keyword_rank, 0.0) as keyword_rank,
        (
            semantic_weight * coalesce(s.similarity, 0.0)
            + full_text_weight * coalesce(k.keyword_rank, 0.0)
        ) as score
    from semantic s
    full outer join keyword k on s.id = k.id
)
select
    m.id, m.content, m.metadata, m.source, m.profile, m.tags,
    f.similarity, f.keyword_rank,
    (
        f.score
        * (1.0 + ln(m.access_count + 1.0) * 0.1)
        * m.confidence
        * (1.0 + g.graph_boost * 0.2)
    )::float as relevance,
    m.access_count, m.last_accessed_at, m.confidence, m.created_at, m.updated_at
from fused f
join memories m on m.id = f.id
left join lateral (
    select coalesce(sum(r.strength), 0.0) as graph_boost
    from memory_relationships r
    where r.target_id = m.id or r.source_id = m.id
) g on true
order by relevance desc
limit match_count;
$function$;
