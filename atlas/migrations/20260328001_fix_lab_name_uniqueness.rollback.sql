-- Rollback: Revert lab name uniqueness back to global
-- This reverses the course-level uniqueness change

BEGIN;

-- Drop the composite unique index
DROP INDEX IF EXISTS unique_display_name_per_course;

-- Recreate the original global unique index
CREATE UNIQUE INDEX unique_display_name 
ON labs (display_name) 
WHERE is_deleted = false;

COMMIT;
