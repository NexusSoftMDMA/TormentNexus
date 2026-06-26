-- Result-ID-scoped out-of-result contradiction lookup for gap-analysis (#262).
-- Surfaces 'contradicts' edges where ONE endpoint is in the supplied result-set
-- IDs and the OTHER is in the same profile but OUTSIDE the result set. Profile is
-- filtered on BOTH endpoints (memories.profile).
CREATE OR REPLACE FUNCTION gap_contradictions_for_ids(
    p_profile text,
    p_memory_ids uuid[],
    p_sample_size integer DEFAULT 10
)
RETURNS TABLE (
    in_result_id text,
    other_id text,
    strength float,
    total_count bigint
)
LANGUAGE sql
STABLE
AS $$
    WITH edges AS (
        SELECT
            CASE WHEN mr.source_id = ANY(p_memory_ids) THEN mr.source_id ELSE mr.target_id END AS in_id,
            CASE WHEN mr.source_id = ANY(p_memory_ids) THEN mr.target_id ELSE mr.source_id END AS other_id,
            mr.strength
        FROM memory_relationships mr
        JOIN memories ms ON ms.id = mr.source_id AND ms.profile = p_profile
        JOIN memories mt ON mt.id = mr.target_id AND mt.profile = p_profile
        WHERE mr.relationship = 'contradicts'
          AND (mr.source_id = ANY(p_memory_ids) OR mr.target_id = ANY(p_memory_ids))
          AND NOT (mr.source_id = ANY(p_memory_ids) AND mr.target_id = ANY(p_memory_ids))
    )
    SELECT in_id::text, other_id::text, strength, count(*) OVER () AS total_count
    FROM edges
    LIMIT p_sample_size;
$$;
