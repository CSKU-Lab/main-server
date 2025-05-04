schema "public" {
  comment = "standard public schema"
}

enum "role" {
  schema = schema.public
  values = [ "student" , "instructor" , "admin" ]
}

table "users" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "email" {
    type = text
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
  index "unique_active_username" {
    columns = [ column.username ]
    where = "is_deleted = false"
    unique = true
  }
  index "unique_active_email" {
    columns = [ column.email ]
    where = "is_deleted = false"
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

table "courses" {
  schema = schema.public
  column "id" {
    type = uuid
  }
  column "code" {
    type = varchar(8)
  }
  column "name" {
    type = text
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
  }
  index "unique_active_course" {
    columns = [ column.name ]
    where = "is_deleted = false"
    unique = true
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
  column "started_at" {
    type = timestamp
  }
  column "ended_at" {
    type = timestamp
  }
  column "icon" {
    type = text
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
  }
  foreign_key "fk_instructor_id" {
    columns = [ column.instructor_id ]
    ref_columns = [ table.users.column.id ]
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
    type = timestamp
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
