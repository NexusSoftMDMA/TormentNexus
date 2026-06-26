-- sql/migrations/rollback/DANGER_025_memory_lifecycle.sql
--
-- ┌─────────────────────────────────────────────────────────────────────┐
-- │ DANGER: ROLLBACK MIGRATION                                          │
-- │                                                                     │
-- │ This script DROPS the lifecycle columns on memories and the two    │
-- │ decay-tuning columns on profile_settings. All data in those        │
-- │ columns is lost. Rollback is for development / recovery, not for   │
-- │ routine operation.                                                  │
-- │                                                                     │
-- │ To run this intentionally:                                          │
-- │   psql $DATABASE_URL <<'EOF'                                        │
-- │   SET ogham.confirm_rollback = 'I-KNOW-WHAT-I-AM-DOING';           │
-- │   \i sql/migrations/rollback/DANGER_025_memory_lifecycle.sql        │
-- │   EOF                                                               │
-- │                                                                     │
-- │ Piping this file naively (psql $URL < this_file) will FAIL by      │
-- │ design -- the session variable is checked before anything else.    │
-- └─────────────────────────────────────────────────────────────────────┘

-- Guard lives INSIDE the transaction so that a missing session variable
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

DROP INDEX IF EXISTS memories_stage_idx;

ALTER TABLE memories
    DROP CONSTRAINT IF EXISTS memories_stage_valid,
    DROP COLUMN IF EXISTS stage,
    DROP COLUMN IF EXISTS stage_entered_at;

ALTER TABLE profile_settings
    DROP COLUMN IF EXISTS decay_lambda,
    DROP COLUMN IF EXISTS decay_beta;

COMMIT;
