-- Migration 012: Add temporal columns for calendar/timeline views
-- Adds occurrence_period (tstzrange) and recurrence_days (int[]) to memories table.
-- These columns are already in the main schema files for new installs.

-- Temporal columns
ALTER TABLE memories ADD COLUMN IF NOT EXISTS occurrence_period tstzrange;
ALTER TABLE memories ADD COLUMN IF NOT EXISTS recurrence_days int[];

-- GIST index for fast range overlap queries (calendar/timeline)
CREATE INDEX IF NOT EXISTS idx_memories_occurrence ON memories USING GIST (occurrence_period)
    WHERE occurrence_period IS NOT NULL;

-- GIN index for recurrence day lookups
CREATE INDEX IF NOT EXISTS idx_memories_recurrence ON memories USING GIN (recurrence_days)
    WHERE recurrence_days IS NOT NULL;

-- Backfill function: extract recurrence from memory text via regex
-- Handles: "every Monday", "weekly Wednesday", "every Tue and Thu", comma lists
CREATE OR REPLACE FUNCTION backfill_recurrence()
RETURNS TABLE(updated_count bigint) AS $$
DECLARE
    cnt bigint := 0;
    day_patterns text[][] := ARRAY[
        ARRAY['sunday', '0'],
        ARRAY['monday', '1'],
        ARRAY['tuesday', '2'],
        ARRAY['wednesday', '3'],
        ARRAY['thursday', '4'],
        ARRAY['friday', '5'],
        ARRAY['saturday', '6']
    ];
    pat text[];
BEGIN
    FOREACH pat SLICE 1 IN ARRAY day_patterns LOOP
        -- "every <day>"
        WITH updated AS (
            UPDATE memories
            SET recurrence_days = CASE
                WHEN recurrence_days IS NULL THEN ARRAY[pat[2]::int]
                WHEN NOT (pat[2]::int = ANY(recurrence_days)) THEN recurrence_days || pat[2]::int
                ELSE recurrence_days
            END
            WHERE content ~* ('\yevery\s+' || pat[1] || '\y')
              AND (recurrence_days IS NULL OR NOT (pat[2]::int = ANY(recurrence_days)))
            RETURNING 1
        )
        SELECT cnt + count(*) INTO cnt FROM updated;

        -- "weekly <day>" or "weekly on <day>"
        WITH updated AS (
            UPDATE memories
            SET recurrence_days = CASE
                WHEN recurrence_days IS NULL THEN ARRAY[pat[2]::int]
                WHEN NOT (pat[2]::int = ANY(recurrence_days)) THEN recurrence_days || pat[2]::int
                ELSE recurrence_days
            END
            WHERE content ~* ('\yweekly\s+(on\s+)?' || pat[1] || '\y')
              AND (recurrence_days IS NULL OR NOT (pat[2]::int = ANY(recurrence_days)))
            RETURNING 1
        )
        SELECT cnt + count(*) INTO cnt FROM updated;

        -- "every <other_day> and <day>" (conjunction partner)
        WITH updated AS (
            UPDATE memories
            SET recurrence_days = CASE
                WHEN recurrence_days IS NULL THEN ARRAY[pat[2]::int]
                WHEN NOT (pat[2]::int = ANY(recurrence_days)) THEN recurrence_days || pat[2]::int
                ELSE recurrence_days
            END
            WHERE content ~* ('\yevery\y.*\y' || pat[1] || '\y')
              AND content ~* '\yevery\s+(monday|tuesday|wednesday|thursday|friday|saturday|sunday)\y'
              AND (recurrence_days IS NULL OR NOT (pat[2]::int = ANY(recurrence_days)))
            RETURNING 1
        )
        SELECT cnt + count(*) INTO cnt FROM updated;
    END LOOP;

    updated_count := cnt;
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

-- Run the backfill (idempotent -- safe to run multiple times)
SELECT * FROM backfill_recurrence();
