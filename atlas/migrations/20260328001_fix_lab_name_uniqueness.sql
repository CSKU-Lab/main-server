-- Migration: Change lab name uniqueness from global to course-level
-- Issue: #13
-- Date: 2026-03-28
-- Author: Database Specialist

-- This migration changes the uniqueness constraint on labs.display_name
-- from global (across all courses) to course-level (unique within a course)

BEGIN;

-- Drop the old global unique index
DROP INDEX IF EXISTS unique_display_name;

-- Create new composite unique index on (course_id, display_name)
-- This allows the same lab name to exist in different courses
-- but prevents duplicate names within the same course
CREATE UNIQUE INDEX unique_display_name_per_course 
ON labs (course_id, display_name) 
WHERE is_deleted = false;

COMMIT;
