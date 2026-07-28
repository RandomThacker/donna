-- calendar_sources.sync_cursor briefly stored calendarList etags before events sync owned it.
-- Clear stale values so the next events sync performs a clean full sync + stores a real syncToken.

UPDATE calendar_sources
SET sync_cursor = NULL,
    updated_at = now()
WHERE sync_cursor IS NOT NULL;
