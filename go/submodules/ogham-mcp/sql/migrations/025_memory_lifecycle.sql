-- src/ogham/sql/migrations/025_memory_lifecycle.sql
--
-- Migration 025: memory lifecycle (FRESH / STABLE / EDITING).
--
-- Adds two columns to memories (stage, stage_entered_at) and two
-- columns to profile_settings (decay_lambda, decay_beta) for per-
-- profile decay-curve tuning.
--
-- ADDITIVE ONLY -- no columns dropped, no columns renamed, no data
-- touched except for the one-time backfill to set existing rows to
-- stage='fresh' with stage_entered_at = created_at.
--
-- Idempotent -- `IF NOT EXISTS` guards mean running twice is a no-op.
-- Safe on mixed-version deployments: a pre-migration Python or Go
-- client reading rows still works (extra columns ignored by
-- `SELECT known_columns FROM memories`).

BEGIN;

-- memories: add lifecycle state
ALTER TABLE memories
    ADD COLUMN IF NOT EXISTS stage text NOT NULL DEFAULT 'fresh',
    ADD COLUMN IF NOT EXISTS stage_entered_at timestamptz NOT NULL DEFAULT now();

-- Constrain stage to known values. Drop-then-add pattern is safe on
-- Postgres 15+; we only ship against 15+.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'memories_stage_valid'
    ) THEN
        ALTER TABLE memories
            ADD CONSTRAINT memories_stage_valid
            CHECK (stage IN ('fresh', 'stable', 'editing'));
    END IF;
END$$;

-- Backfill: existing rows get stage='fresh' with stage_entered_at =
-- created_at (so the first sweep will advance them naturally on time
-- gate if they're old enough + pass the surprise/importance gate).
-- Guard: only touch rows where we introduced the default (they all
-- currently have stage='fresh' from the ADD COLUMN DEFAULT).
UPDATE memories
   SET stage_entered_at = created_at
 WHERE stage_entered_at > created_at;

-- Index on (profile, stage, stage_entered_at) for fast lookups during
-- sweep + dashboard pipeline counts.
CREATE INDEX IF NOT EXISTS memories_stage_idx
    ON memories (profile, stage, stage_entered_at);

-- profile_settings: per-profile decay curve params.
-- Defaults matching Shodh's recommended values (Wixted 2004).
ALTER TABLE profile_settings
    ADD COLUMN IF NOT EXISTS decay_lambda double precision NOT NULL DEFAULT 0.1,
    ADD COLUMN IF NOT EXISTS decay_beta double precision NOT NULL DEFAULT 0.4;

COMMIT;
