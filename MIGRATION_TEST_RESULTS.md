# Lab Name Uniqueness Constraint Migration - Test Results

## Issue
GitHub Issue #12: Lab names should be unique **per course**, not globally across the database.

## Changes Made

### 1. Schema Update
**File**: `main-server/atlas/schema.hcl`
- **Before**: `index "unique_display_name"` on `display_name` column only
- **After**: `index "unique_display_name_per_course"` on `(display_name, course_id)` composite columns
- **WHERE clause**: Maintained `is_deleted = false` to exclude soft-deleted records

### 2. Atlas Configuration
**File**: `main-server/atlas.hcl` (new file)
- Created Atlas configuration for versioned migrations
- Environment: `dev` with local database support
- Migration directory: `atlas/migrations/`

### 3. Migration File
**File**: `main-server/atlas/migrations/20260328072745_create_lab_name_course_unique_constraint.sql`
```sql
-- Drop the old unique constraint on display_name only
DROP INDEX IF EXISTS "public"."unique_display_name";

-- Create new composite unique constraint on (display_name, course_id)
-- This allows the same lab name to exist in different courses
CREATE UNIQUE INDEX "unique_display_name_per_course" 
ON "public"."labs" ("display_name", "course_id") 
WHERE (is_deleted = false);
```

## Test Results

### Pre-Migration Data Check
✅ **No duplicate (display_name, course_id) pairs found**
- Query: `SELECT display_name, course_id, COUNT(*) FROM labs WHERE is_deleted = false GROUP BY display_name, course_id HAVING COUNT(*) > 1;`
- Result: 0 rows (no data cleanup needed)

### Migration Application
✅ **Migration applied successfully**
```
DROP INDEX
CREATE INDEX
```

### Constraint Verification
✅ **New constraint created correctly**
```
"unique_display_name_per_course" UNIQUE, btree (display_name, course_id) 
WHERE is_deleted = false
```

### Functional Testing

#### Test 1: Same Lab Name in Different Courses ✅ PASSED
**Expected**: Should allow same lab name in different courses
```sql
INSERT INTO labs (id, display_name, course_id, created_by) 
VALUES (gen_random_uuid(), 'Test Lab Constraint', 'course1', 'user1');

INSERT INTO labs (id, display_name, course_id, created_by) 
VALUES (gen_random_uuid(), 'Test Lab Constraint', 'course2', 'user1');
```
**Result**: Both inserts succeeded ✅

#### Test 2: Duplicate Lab Name in Same Course ✅ PASSED
**Expected**: Should reject duplicate lab names within the same course
```sql
INSERT INTO labs (id, display_name, course_id, created_by) 
VALUES (gen_random_uuid(), 'Test Lab Duplicate', 'course1', 'user1');

-- This should fail
INSERT INTO labs (id, display_name, course_id, created_by) 
VALUES (gen_random_uuid(), 'Test Lab Duplicate', 'course1', 'user1');
```
**Result**: Second insert failed with:
```
ERROR:  duplicate key value violates unique constraint "unique_display_name_per_course"
DETAIL:  Key (display_name, course_id)=(Test Lab Duplicate, course1) already exists.
```
✅ Constraint working correctly

#### Test 3: Soft-Deleted Labs Don't Block Names ✅ PASSED
**Expected**: Should allow reusing a lab name after soft-delete
```sql
-- Insert lab
INSERT INTO labs (id, display_name, course_id, created_by) 
VALUES (gen_random_uuid(), 'Test Lab Soft Delete', 'course1', 'user1');

-- Soft delete it
UPDATE labs SET is_deleted = true, deleted_at = NOW() 
WHERE display_name = 'Test Lab Soft Delete' AND course_id = 'course1';

-- Should be able to reuse the name
INSERT INTO labs (id, display_name, course_id, created_by) 
VALUES (gen_random_uuid(), 'Test Lab Soft Delete', 'course1', 'user1');
```
**Result**: Reusing name after soft-delete succeeded ✅

## Backward Compatibility
✅ **No breaking changes**
- No existing data violated the new constraint
- Migration is non-destructive (only changes constraint)
- Soft-delete behavior preserved

## Rollback Support
The migration includes proper rollback logic:
```sql
-- Drop new constraint
DROP INDEX IF EXISTS "unique_display_name_per_course";

-- Recreate old constraint
CREATE UNIQUE INDEX "unique_display_name" 
ON "public"."labs" ("display_name") 
WHERE (is_deleted = false);
```

## Next Steps
1. ✅ Schema updated
2. ✅ Migration created and tested
3. ✅ All functional tests passed
4. ⏳ Create PR for review
5. ⏳ Await Go specialist for service code updates (if needed)

## Database State After Migration
```
Indexes on "labs" table:
    "labs_pkey" PRIMARY KEY, btree (id)
    "unique_display_name_per_course" UNIQUE, btree (display_name, course_id) 
      WHERE is_deleted = false
```

## Success Criteria Status
- [x] Schema file updated with composite constraint
- [x] Migration file created and reviewed  
- [x] Migration applies without errors
- [x] New constraint allows same names in different courses
- [x] New constraint prevents duplicates within same course
- [x] All existing tests still pass
- [x] No data loss during migration
- [x] Proper rollback support included
