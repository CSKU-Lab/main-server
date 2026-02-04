schema "public" {
  comment = "standard public schema"
}

enum "role" {
  schema = schema.public
  values = [ "student" , "instructor" , "admin" ]
}

enum "user_type" {
  schema = schema.public
  values = [ "oauth", "credential" ]
}

enum "course_visibility" {
  schema = schema.public
  values = [ "public" , "private" ]
}

table "users" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "email" {
    type = text
    null = true
  }
  column "type" {
    type = enum.user_type
  }
  column "username" {
    type = varchar(255)
  }
  column "display_name" {
    type = text
  }
  column "profile_image" {
    type = text
    null = true
  }
  column "roles" {
    type = sql("role[]")
  }
  column "group_id" {
    type = uuid
    null = true
  }
  column "created_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "is_deleted" {
    type = boolean
    default = false
  }
  column "deleted_at" {
    type = timestamp
    null = true
  }
  primary_key  {
    columns = [ column.id ]
  }
  foreign_key "fk_group_id" {
    columns = [ column.group_id ]
    ref_columns = [ table.user_groups.column.id ]
    on_delete = SET_NULL
  }
  index "unique_active_username" {
    columns = [ column.username ]
    where = "is_deleted = false"
    unique = true
  }
  index "unique_active_email" {
    columns = [ column.email ]
    where = "is_deleted = false AND email IS NOT NULL"
    unique = true
  }
}

table "user_refresh_tokens" {
  schema = schema.public
  column "user_id" {
    type = uuid
  }
  column "token" {
    type = text
  }
  primary_key {
    columns = [ column.user_id ]
  }
  foreign_key "fk_user_id" {
    columns = [ column.user_id ]
    ref_columns = [ table.users.column.id ]
    on_delete = CASCADE
  }
}

table "user_passwords" {
  schema = schema.public
  column "user_id" {
    type = uuid
  }
  column "password" {
    type = varchar(80)
  }
  primary_key {
    columns = [ column.user_id ]
  }
  foreign_key "fk_user_id" {
    columns = [ column.user_id ]
    ref_columns = [ table.users.column.id ]
    on_delete = CASCADE
  }
}

table "user_groups" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "name" {
    type = text
  }
  unique "name" {
    columns = [ column.name ]
  }
  primary_key {
    columns = [ column.id ]
  }
}

table "courses" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "name" {
    type = text
  }
  column "visibility" {
    type = enum.course_visibility
  }
  column "created_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "is_archived" {
    type = boolean
    default = false
  }
  column "is_deleted" {
    type = boolean
    default = false
  }
  column "deleted_at" {
    type = timestamp
    null = true
  }
  primary_key  {
    columns = [ column.id ]
  }
  index "unique_active_course" {
    columns = [ column.name ]
    where = "is_deleted = false AND is_archived = false"
    unique = true
  }
}

table "course_creators" {
  schema = schema.public
  column "course_id" {
    type = uuid
  }
  column "creator_id" {
    type = uuid
  }
  column "order" {
    type = integer
  }
  primary_key  {
    columns = [ column.course_id, column.creator_id , column.order ]
  }
  foreign_key "fk_course_id" {
    columns = [ column.course_id ]
    ref_columns = [ table.courses.column.id ]
    on_delete = CASCADE
  }
  foreign_key "fk_creator_id" {
    columns = [ column.creator_id ]
    ref_columns = [ table.users.column.id ]
    on_delete = CASCADE
  }
}

table "sections" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "name" {
    type = text
  }
  column "banner" {
    type = text
    null = true
  }
  column "course_id" {
    type = uuid
  }
  column "semester_id" {
    type = uuid
  }
  column "created_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "is_deleted" {
    type = boolean
    default = false
  }
  column "deleted_at" {
    type = timestamp
    null = true
  }
  primary_key  {
    columns = [ column.id ]
  }
  foreign_key "fk_course_id" {
    columns = [ column.course_id ]
    ref_columns = [ table.courses.column.id ]
    on_delete = CASCADE
  }
  foreign_key "fk_semester_id" {
    columns = [ column.semester_id  ]
    ref_columns = [ table.semesters.column.id ]
    on_delete = CASCADE
  }
  index "unique_active_section" {
    columns = [ column.name, column.course_id, column.semester_id ]
    where = "is_deleted = false"
    unique = true
  }
}

table "section_instructors" {
  schema = schema.public
  column "section_id" {
    type = uuid
  }
  column "instructor_id" {
    type = uuid
  }
  primary_key  {
    columns = [ column.section_id,  column.instructor_id ]
  }
  foreign_key "fk_section_id" {
    columns = [ column.section_id ]
    ref_columns = [ table.sections.column.id ]
    on_delete = CASCADE
  }
  foreign_key "fk_instructor_id" {
    columns = [ column.instructor_id ]
    ref_columns = [ table.users.column.id ]
    on_delete = CASCADE
  }
}

table "section_tas" {
  schema = schema.public
  column "section_id" {
    type = uuid
  }
  column "ta_id" {
    type = uuid
  }
  primary_key  {
    columns = [ column.section_id,  column.ta_id ]
  }
  foreign_key "fk_section_id" {
    columns = [ column.section_id ]
    ref_columns = [ table.sections.column.id ]
    on_delete = CASCADE
  }
  foreign_key "fk_ta_id" {
    columns = [ column.ta_id ]
    ref_columns = [ table.users.column.id ]
    on_delete = CASCADE
  }
}

table "section_students" {
  schema = schema.public
  column "section_id" {
    type = uuid
  }
  column "student_id" {
    type = uuid
  }
  primary_key  {
    columns = [ column.section_id,  column.student_id ]
  }
  foreign_key "fk_section_id" {
    columns = [ column.section_id ]
    ref_columns = [ table.sections.column.id ]
    on_delete = CASCADE
  }
  foreign_key "fk_student_id" {
    columns = [ column.student_id ]
    ref_columns = [ table.users.column.id ]
    on_delete = CASCADE
  }
}

table "section_logs" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "user_id" {
    type = uuid
    null = true
  }
  column "section_id" {
    type = uuid
  }
  column "action" {
    type = text
  }
  column "timestamp" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "ip_address" {
    type = inet
  }
  primary_key  {
    columns = [ column.id ]
  }
  foreign_key "fk_user_id" {
    columns = [ column.user_id ]
    ref_columns = [ table.users.column.id ]
    on_delete = SET_NULL
  }
  foreign_key "fk_section_id" {
    columns = [ column.section_id ]
    ref_columns = [ table.sections.column.id ]
    on_delete = CASCADE
  }
}

enum "semester_type" {
  schema = schema.public
  values = [ "first" , "second" , "summer" ]
}

table "semesters" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "name" {
    type = varchar(255)
  }
  column "type" {
    type = enum.semester_type
  }
  column "started_date" {
    type = date
  }
  column "created_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "is_deleted" {
    type = boolean
    default = false
  }
  column "deleted_at" {
    type = timestamp
    null = true
  }
  primary_key  {
    columns = [ column.id ]
  }
  index "unique_active_semester" {
    columns = [ column.name, column.type ]
    where = "is_deleted = false"
    unique = true
  }
}

table "default_labs" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "course_id" {
    type = uuid
  }
  column "lab_id" {
    type = uuid
  }
  column "lab_name" {
    type = text
  }
  column "position" {
    type = int
  }
  column "created_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "is_deleted" {
    type = boolean
    default = false
  }
  column "deleted_at" {
    type = timestamp
    null = true
  }
  primary_key  {
    columns = [ column.id ]
  }
  index "unique_default_lab" {
    columns = [ column.lab_id, column.course_id ]
    where = "is_deleted = false"
    unique = true
  }
  foreign_key "fk_course_id" {
    columns = [ column.course_id ]
    ref_columns = [ table.courses.column.id ]
    on_delete = CASCADE
  }
  foreign_key "fk_lab_id" {
    columns = [ column.lab_id ]
    ref_columns = [ table.labs.column.id ]
    on_delete = CASCADE
  }
}

table "labs" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "display_name" {
    type = text
  }
  column "is_default" {
    type = boolean
    default = false
  }
  column "course_id" {
    type = uuid
  }
  column "created_by" {
    type = uuid
  }
  column "created_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "is_deleted" {
    type = boolean
    default = false
  }
  column "deleted_at" {
    type = timestamp
    null = true
  }
  primary_key  {
    columns = [ column.id ]
  }
  index "unique_display_name" {
    columns = [ column.display_name ]
    where = "is_deleted = false"
    unique = true
  }
  foreign_key "fk_created_by" {
    columns = [ column.created_by ]
    ref_columns = [ table.users.column.id ]
    on_delete = SET_NULL
  }
  foreign_key "fk_course_id" {
    columns = [ column.course_id ]
    ref_columns = [ table.courses.column.id ]
    on_delete = CASCADE
  }
}

table "lab_materials" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "lab_id" {
    type = uuid
  }
  column "material_id" {
    type = uuid
  }
  column "created_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "is_deleted" {
    type = boolean
    default = false
  }
  column "deleted_at" {
    type = timestamp
    null = true
  }
  primary_key  {
    columns = [ column.id ]
  }
  index "unique_lab_material" {
    columns = [ column.lab_id, column.material_id ]
    where = "is_deleted = false"
    unique = true
  }
  foreign_key "fk_lab_id" {
    columns = [ column.lab_id ]
    ref_columns = [ table.labs.column.id ]
    on_delete = CASCADE
  }
  foreign_key "fk_material_id" {
    columns = [ column.material_id ]
    ref_columns = [ table.materials.column.id ]
    on_delete = CASCADE
  }
}

table "lab_sections" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "lab_id" {
    type = uuid
  }
  column "section_id" {
    type = uuid
  }
  column "position" {
    type = int
  }
  column "created_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "is_deleted" {
    type = boolean
    default = false
  }
  column "deleted_at" {
    type = timestamp
    null = true
  }
  primary_key  {
    columns = [ column.id ]
  }
  index "unique_lab_section" {
    columns = [ column.lab_id, column.section_id ]
    where = "is_deleted = false"
    unique = true
  }
  foreign_key "fk_lab_id" {
    columns = [ column.lab_id ]
    ref_columns = [ table.labs.column.id ]
    on_delete = CASCADE
  }
  foreign_key "fk_section_id" {
    columns = [ column.section_id ]
    ref_columns = [ table.sections.column.id ]
    on_delete = CASCADE
  }
}

enum "material_type" {
  schema = schema.public
  values = [ "document" , "code" , "type" ]
}

enum "visibility" {
  schema = schema.public
  values = [ "public" , "private" ]
}

table "materials" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "name" {
    type = text
  }
  column "type" {
    type = enum.material_type
  }
  column "visibility" {
    type = enum.visibility
  }
  column "created_by" {
    type = uuid
  }
  column "created_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  column "is_deleted" {
    type = boolean
    default = false
  }
  column "deleted_at" {
    type = timestamp
    null = true
  }
  primary_key  {
    columns = [ column.id ]
  }
  foreign_key "fk_created_by" {
    columns = [ column.created_by ]
    ref_columns = [ table.users.column.id ]
    on_delete = SET_NULL
  }
}

table "code_materials" {
  schema = schema.public
  column "material_id" {
    type = uuid
  }
  column "description" {
    type = text
    null = true
    default = null
  }
  column "task_id" {
    type = uuid
  }
  column "hide_test_cases" {
    type = boolean
    default = true
  }
  primary_key  {
    columns = [ column.material_id ]
  }
  foreign_key "fk_material_id" {
    columns = [ column.material_id ]
    ref_columns = [ table.materials.column.id ]
    on_delete = CASCADE
  }
}

table "material_tags" {
  schema = schema.public
  column "material_id" {
    type = uuid
  }
  column "tag_id" {
    type = uuid
  }
  primary_key  {
    columns = [ column.material_id, column.tag_id ]
  }
  foreign_key "fk_material_id" {
    columns = [ column.material_id ]
    ref_columns = [ table.materials.column.id ]
    on_delete = CASCADE
  }
  foreign_key "fk_tag_id" {
    columns = [ column.tag_id ]
    ref_columns = [ table.tags.column.id ]
    on_delete = CASCADE
  }
}

table "tags" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "name" {
    type = text
  }
  unique "unique_tag_name" {
    columns = [ column.name ]
  }
  primary_key {
    columns = [ column.id ]
  }
}

enum "action" {
  schema = schema.public
  values = [ "sign-in" , "sign-out" , "sign-in-failed" ]
}

table "auth_logs" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "user_id" {
    type = uuid
  }
  column "action" {
    type = enum.action
  }
  column "created_at" {
    type = timestamp
  }
  primary_key  {
    columns = [ column.id ]
  }
  foreign_key "fk_user_id" {
    columns = [ column.user_id ]
    ref_columns = [ table.users.column.id ]
  }
}

enum "submission_status" {
  schema = schema.public
  values = [ "queued", "running", "passed", "failed"]
}

table "submissions" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "user_id" {
    type = uuid
  }
  column "material_id" {
    type = uuid
  }
  column "section_id" {
    type = uuid
    null = true
  }
  column "course_id" {
    type = uuid
    null = true
  }
  column "created_at" {
    type = timestamp
  }
  column "updated_at" {
    type = timestamp
  }
  column "status"{
    type = enum.submission_status
  }
  primary_key {
    columns = [ column.id ]
  }
  foreign_key "fk_user_id" {
    columns = [ column.user_id ]
    ref_columns = [ table.users.column.id ]
  }
  foreign_key "fk_material_id" {
    columns = [ column.material_id ]
    ref_columns = [ table.materials.column.id ]
  }
  foreign_key "fk_section_id" {
    columns = [ column.section_id ]
    ref_columns = [ table.sections.column.id ]
  }
  foreign_key "fk_course_id" {
    columns = [ column.course_id ]
    ref_columns = [ table.courses.column.id ]
  }
}

table "code_submissions" {
  schema = schema.public
  column "submission_id" {
    type = uuid
  }
  column "files" {
    type = jsonb
  }
  column "status"{
    type = text
    null = true
  }
  column "avg_wall_time" {
    type = float
    null = true
  }
  column "avg_memory" {
    type = int
    null = true
  }
  column "test_case_groups" {
    type = jsonb
    null = true
  }
  column "score" {
    type = int
    null = true
  }
  primary_key {
    columns = [ column.submission_id ]
  }
  foreign_key "fk_submission_id" {
    columns = [ column.submission_id ]
    ref_columns = [ table.submissions.column.id ]
    on_delete = CASCADE
  }
}

table "code_submissions_outbox" {
  schema = schema.public
  column "id"{
    type = uuid
  }
  column "submission_id" {
    type = uuid
  }
  column "is_sent" {
    type = boolean
    default = false
  }
  column "payload" {
    type = text
  }
  column "created_at" {
    type = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  primary_key {
    columns = [ column.id ]
  }
  foreign_key "fk_aggregate_id" {
    columns = [ column.submission_id ]
    ref_columns = [ table.submissions.column.id ]
    on_delete = CASCADE
  }
}
