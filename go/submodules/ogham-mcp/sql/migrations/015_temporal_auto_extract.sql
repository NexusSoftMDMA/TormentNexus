-- Migration 015: Auto-extract occurrence_period from content [Date: YYYY-MM-DD] prefix
--
-- Fires on INSERT/UPDATE. If occurrence_period is NULL and content contains
-- a [Date: YYYY-MM-DD] prefix, sets occurrence_period to that day's range.
-- Production memories should use the Python extraction pipeline for richer
-- temporal data; this trigger catches raw imports and API-ingested content.

CREATE OR REPLACE FUNCTION extract_occurrence_from_content()
RETURNS trigger
LANGUAGE plpgsql
SECURITY INVOKER
SET search_path = public
AS $$
DECLARE
    date_str text;
    parsed date;
BEGIN
    -- Only fire if occurrence_period is not already set
    IF NEW.occurrence_period IS NOT NULL THEN
        RETURN NEW;
    END IF;

    -- Extract [Date: YYYY-MM-DD] prefix
    date_str := substring(NEW.content FROM '\[Date:\s*(\d{4}-\d{2}-\d{2})\]');
    IF date_str IS NOT NULL THEN
        BEGIN
            parsed := date_str::date;
            NEW.occurrence_period := tstzrange(
                parsed::timestamptz,
                (parsed + interval '1 day')::timestamptz
            );
        EXCEPTION WHEN OTHERS THEN
            -- Invalid date string, skip
            NULL;
        END;
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS memories_extract_occurrence ON memories;
CREATE TRIGGER memories_extract_occurrence
    BEFORE INSERT OR UPDATE ON memories
    FOR EACH ROW
    EXECUTE FUNCTION extract_occurrence_from_content();
