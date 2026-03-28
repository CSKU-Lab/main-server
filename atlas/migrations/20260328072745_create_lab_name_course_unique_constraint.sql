-- Drop the old unique constraint on display_name only
DROP INDEX IF EXISTS "public"."unique_display_name";

-- Create new composite unique constraint on (display_name, course_id)
-- This allows the same lab name to exist in different courses
CREATE UNIQUE INDEX "unique_display_name_per_course" 
ON "public"."labs" ("display_name", "course_id") 
WHERE (is_deleted = false);
