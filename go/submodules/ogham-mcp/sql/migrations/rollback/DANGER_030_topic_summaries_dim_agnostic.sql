-- DANGER: rollback for migration 030 — restore the hardcoded
-- vector(512) embedding column on topic_summaries.
--
-- Drops every existing summary embedding (no embedding fits multiple
-- dims) and rebuilds the column at the legacy 512 dim. Self-hosters
-- running at any non-512 dim will start emitting "expected 512
-- dimensions, not N" errors on the next recompute, so this rollback
-- is only useful if you also roll memories.embedding back to 512
-- (which is even more destructive).
--
-- Refuses to run unless `ogham.confirm_rollback = 'I-KNOW-WHAT-I-AM-DOING'`
-- is set in the session, matching the rollback ritual from the
-- migration 025/026/028 series.

-- Guard lives INSIDE the transaction so a missing session variable
-- aborts the whole rollback rather than just the DO block. Putting the
-- guard before BEGIN means a naive `psql $URL -f file.sql` (without
-- ON_ERROR_STOP=1) prints the ERROR and keeps running the destructive
-- ops below. Inside BEGIN, the abort is transactional.

BEGIN;

DO $$
BEGIN
    IF current_setting('ogham.confirm_rollback', true) IS DISTINCT FROM 'I-KNOW-WHAT-I-AM-DOING' THEN
        RAISE EXCEPTION
            'Rollback refused. Set ogham.confirm_rollback = ''I-KNOW-WHAT-I-AM-DOING'' first. '
            'See sql/migrations/rollback/README.md.';
    END IF;
END $$;

DROP INDEX IF EXISTS topic_summaries_embedding_hnsw_idx;
ALTER TABLE topic_summaries DROP COLUMN embedding;
ALTER TABLE topic_summaries ADD COLUMN embedding vector(512);
CREATE INDEX IF NOT EXISTS topic_summaries_embedding_hnsw_idx
    ON topic_summaries USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64)
    WHERE status = 'fresh';

COMMIT;
