-- sql/migrations/rollback/DANGER_026_memory_lifecycle_split.sql
--
-- ┌─────────────────────────────────────────────────────────────────────┐
-- │ DANGER: ROLLBACK MIGRATION                                          │
-- │                                                                     │
-- │ This script DROPS the memory_lifecycle table + its triggers and    │
-- │ restores the pre-026 state on memories (stage columns re-added     │
-- │ and backfilled from memory_lifecycle before the table is dropped). │
-- │                                                                     │
-- │ Rollback is for development / recovery, not for routine operation. │
-- │ The Python MCP server v0.11.0+ expects memory_lifecycle to exist   │
-- │ and will fail after this rollback unless you also downgrade the    │
-- │ server to v0.10.x.                                                  │
-- │                                                                     │
-- │ To run this intentionally:                                          │
-- │   psql $DATABASE_URL <<'EOF'                                        │
-- │   SET ogham.confirm_rollback = 'I-KNOW-WHAT-I-AM-DOING';           │
-- │   \i sql/migrations/rollback/DANGER_026_memory_lifecycle_split.sql  │
-- │   EOF                                                               │
-- │                                                                     │
-- │ Piping this file naively (psql $URL < this_file) will FAIL by      │
-- │ design -- the session variable is checked before anything else.    │
-- └─────────────────────────────────────────────────────────────────────┘

-- Guard lives INSIDE the transaction so a missing session variable
-- aborts the whole rollback rather than just the DO block. Putting the
-- guard before BEGIN means a naive `psql $URL -f file.sql` (without
-- ON_ERROR_STOP=1) prints the ERROR and keeps running the destructive
-- ops below. Inside BEGIN, the abort is transactional.

BEGIN;

DO $$
BEGIN
    IF current_setting('ogham.confirm_rollback', true) IS DISTINCT FROM 'I-KNOW-WHAT-I-AM-DOING' THEN
        RAISE EXCEPTION USING
            MESSAGE = 'Rollback refused -- explicit opt-in required.',
            HINT = 'Run "SET ogham.confirm_rollback = ''I-KNOW-WHAT-I-AM-DOING'';" in the same session before this script. See file header for details.';
    END IF;
END$$;

-- Restore 025's schema on memories.
ALTER TABLE memories
    ADD COLUMN IF NOT EXISTS stage text NOT NULL DEFAULT 'fresh',
    ADD COLUMN IF NOT EXISTS stage_entered_at timestamptz NOT NULL DEFAULT now();

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'memories_stage_valid'
    ) THEN
        ALTER TABLE memories
            ADD CONSTRAINT memories_stage_valid
            CHECK (stage IN ('fresh', 'stable', 'editing'));
    END IF;
END$$;

-- Backfill from memory_lifecycle if it exists.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'memory_lifecycle') THEN
        UPDATE memories m
           SET stage = ml.stage, stage_entered_at = ml.stage_entered_at
          FROM memory_lifecycle ml
         WHERE ml.memory_id = m.id;
    END IF;
END$$;

CREATE INDEX IF NOT EXISTS memories_stage_idx
    ON memories (profile, stage, stage_entered_at);

-- Drop triggers + functions.
DROP TRIGGER IF EXISTS memories_init_lifecycle ON memories;
DROP TRIGGER IF EXISTS memories_sync_lifecycle_profile ON memories;
DROP FUNCTION IF EXISTS init_memory_lifecycle();
DROP FUNCTION IF EXISTS sync_memory_lifecycle_profile();

-- Drop memory_lifecycle.
DROP TABLE IF EXISTS memory_lifecycle;

COMMIT;
