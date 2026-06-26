-- Add sparse_embedding column for BGE-M3 neural sparse vectors.
-- Uses pgvector's sparsevec type. No index needed initially —
-- sequential scan is fine for benchmarking with small profile sizes.

ALTER TABLE memories ADD COLUMN IF NOT EXISTS sparse_embedding sparsevec;
