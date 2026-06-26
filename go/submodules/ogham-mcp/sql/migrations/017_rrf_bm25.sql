-- Migration 017: True Reciprocal Rank Fusion + BM25-approximating keyword scoring.
--
-- Background
-- ----------
-- Migration 013 (halfvec_compression) rewrote hybrid_search_memories with a
-- raw-score linear-combination fusion formula:
--
--     semantic_weight * similarity + full_text_weight * keyword_rank
--
-- That is not Reciprocal Rank Fusion. It produces unstable ranking because the
-- two scores live on incompatible scales: cosine similarity is bounded to
-- [0, 1], but ts_rank_cd is unbounded and typically << 1. In practice the
-- semantic term dominates and keyword matches are effectively ignored.
--
-- The unnumbered file sql/migrations/update_search_function.sql (removed in
-- v0.9.2) sorted after 017 alphabetically in sql/upgrade.sh and reinforced the
-- same broken formula on every upgrade, so no user running upgrade.sh could
-- end up with correct RRF via migrations alone.
--
-- This migration
-- ---------------
-- Restores position-based Reciprocal Rank Fusion using row_number() and a
-- scale-invariant 1 / (rrf_k + rank) weighting. Cormack, Clarke, Büttcher
-- (SIGIR 2009) "Reciprocal Rank Fusion outperforms Condorcet and individual
-- rank learning methods". It also adds a 10th optional parameter
-- `filter_profiles text[]` so this migration replaces the 9-param overload
-- left behind by 013, preventing Postgres function-overload ambiguity.
--
-- Halfvec-specific. Safe no-op on non-halfvec deployments (Supabase Cloud,
-- 768-dim): the DO block detects the column type and exits early. Those
-- deployments should re-run sql/schema.sql to apply RRF.
--
-- Idempotent. No data migration.

do $mig$
declare
    col_type text;
begin
    select format_type(a.atttypid, a.atttypmod) into col_type
    from pg_attribute a
    join pg_class c on c.oid = a.attrelid
    where c.relname = 'memories'
      and a.attname = 'embedding'
      and a.attnum > 0;

    if col_type is null then
        raise notice 'Migration 017: memories.embedding not found; run your schema file first. Skipping.';
        return;
    end if;

    if col_type not like '%halfvec%' then
        raise notice 'Migration 017: memories.embedding is % (not halfvec); skipping. Re-run sql/schema.sql or sql/schema_selfhost_supabase.sql to apply RRF for this backend.', col_type;
        return;
    end if;

    -- Drop the 9-param overload introduced by migration 013 so the new 10-param
    -- version below is unambiguous for function resolution.
    execute 'drop function if exists hybrid_search_memories(text, vector, integer, text, text[], text, float, float, integer)';

    execute $fn$
        create or replace function hybrid_search_memories(
            query_text text,
            query_embedding vector,
            match_count integer default 10,
            filter_profile text default 'default',
            filter_tags text[] default null,
            filter_source text default null,
            full_text_weight float default 0.3,
            semantic_weight float default 0.7,
            rrf_k integer default 10,
            filter_profiles text[] default null
        )
        returns table(
            id uuid, content text, metadata jsonb, source text, profile text, tags text[],
            similarity float, keyword_rank float, relevance float,
            access_count integer, last_accessed_at timestamptz, confidence float,
            created_at timestamptz, updated_at timestamptz
        )
        language sql
        set search_path = public, extensions
        as $func$
        with semantic as (
            select
                m.id,
                (1 - (m.embedding::halfvec(512) <=> query_embedding::halfvec(512)))::float as similarity,
                row_number() over (order by m.embedding::halfvec(512) <=> query_embedding::halfvec(512)) as rank_ix
            from memories m
            where (filter_profiles is not null and m.profile = any(filter_profiles)
                   or filter_profiles is null and m.profile = filter_profile)
              and (filter_tags is null or m.tags && filter_tags)
              and (filter_source is null or m.source = filter_source)
              and (m.expires_at is null or m.expires_at > now())
            order by m.embedding::halfvec(512) <=> query_embedding::halfvec(512)
            limit match_count * 3
        ),
        keyword as (
            select
                m.id,
                ts_rank_cd(m.fts, websearch_to_tsquery(query_text), 34)::float as keyword_rank,
                row_number() over (order by ts_rank_cd(m.fts, websearch_to_tsquery(query_text), 34) desc) as rank_ix
            from memories m
            where (filter_profiles is not null and m.profile = any(filter_profiles)
                   or filter_profiles is null and m.profile = filter_profile)
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
                -- Reciprocal Rank Fusion: position-based, score-agnostic.
                (
                    semantic_weight * (1.0 / (rrf_k + coalesce(s.rank_ix, match_count * 3)))
                    + full_text_weight * (1.0 / (rrf_k + coalesce(k.rank_ix, match_count * 3)))
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
        limit match_count
        $func$
    $fn$;

    raise notice 'Migration 017: hybrid_search_memories updated to true Reciprocal Rank Fusion.';
end
$mig$;
