-- Migration 014: Switch TOAST compression from pglz to lz4
--
-- lz4 is faster to compress and decompress than the default pglz,
-- with similar compression ratios. Benefits content-heavy reads
-- (search result retrieval, export, graph traversal).
--
-- Only affects new/updated rows. Existing rows keep their current
-- compression until rewritten (VACUUM FULL or pg_repack).
--
-- Requires PostgreSQL 14+ (lz4 TOAST support added in PG14).

ALTER TABLE memories ALTER COLUMN content SET COMPRESSION lz4;
ALTER TABLE memories ALTER COLUMN original_content SET COMPRESSION lz4;
ALTER TABLE memories ALTER COLUMN metadata SET COMPRESSION lz4;
