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

enum "course_type" {
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
  column "type" {
    type = enum.course_type
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
  }
  foreign_key "fk_semester_id" {
    columns = [ column.semester_id  ]
    ref_columns = [ table.semesters.column.id ]
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
  }
  foreign_key "fk_ta_id" {
    columns = [ column.ta_id ]
    ref_columns = [ table.users.column.id ]
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
  }
  foreign_key "fk_student_id" {
    columns = [ column.student_id ]
    ref_columns = [ table.users.column.id ]
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
