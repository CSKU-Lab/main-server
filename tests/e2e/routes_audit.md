# API Routes Audit - main-server

This document contains a complete audit of all REST API routes in the main-server application.
Generated from: `main-server/cmd/app/api.go` and all route files in `internal/adapters/rest/routes/`.

## Route Structure

All routes are prefixed with `/api/v1`.

---

## 1. Public Routes (No Authentication Required)

### POST /api/v1/playground/execute
- **Handler**: Anonymous function in api.go
- **Auth Required**: No
- **Description**: Execute code in playground mode with Server-Sent Events (SSE) response
- **Request Body**:
  ```json
  {
    "files": [{"name": "string", "content": "string"}],
    "input": "string",
    "runner_id": "string"
  }
  ```
- **Response**: SSE stream with execution results

---

## 2. Authentication Routes (/api/v1/auth/*)

All auth routes are **PUBLIC** (no JWT required).

### GET /api/v1/auth/sign-in/google
- **Handler**: Anonymous function in auth_router.go
- **Auth Required**: No
- **Description**: Initiate Google OAuth2 sign-in flow
- **Response**: Redirects to Google OAuth URL

### GET /api/v1/auth/sign-in/google/callback
- **Handler**: Anonymous function in auth_router.go
- **Auth Required**: No
- **Description**: Google OAuth2 callback handler
- **Query Params**: `state`, `code`
- **Response**: 
  - Dev mode: JSON with access_token and refresh_token
  - Production: Redirect to frontend with cookies set

### POST /api/v1/auth/sign-in/credential
- **Handler**: Anonymous function in auth_router.go
- **Auth Required**: No
- **Validation**: requests.Credential
- **Request Body**:
  ```json
  {
    "username": "string",
    "password": "string"
  }
  ```
- **Success Response (200)**:
  ```json
  {
    "message": "success"
  }
  ```
  - Sets cookies: access_token, refresh_token
- **Error Response (401)**: Unauthorized for invalid credentials

### POST /api/v1/auth/refresh-token
- **Handler**: Anonymous function in auth_router.go
- **Auth Required**: No (uses cookies)
- **Cookies Required**: access_token, refresh_token
- **Response**:
  - 200: New tokens issued
  - 401: Unauthorized if refresh token invalid

---

## 3. Admin Routes (/api/v1/admin/*)

**Required Role**: ADMIN only
Middleware: `RBACMiddleware([ADMIN])`

### 3.1 User Management (/api/v1/admin/users)

#### GET /api/v1/admin/users
- **Handler**: Anonymous function in admin_user_routes.go
- **Auth Required**: Yes (Admin)
- **Query Params**:
  - `page` (default: 1)
  - `page_size` (default: 10)
  - `search` (optional)
  - `sort_by` (default: created_at)
  - `sort_order` (default: desc)
  - Filter params with `__` suffix (e.g., `role__is`)
- **Response (200)**:
  ```json
  {
    "pagination": {
      "page": 1,
      "total_page": 10,
      "total_rows": 100
    },
    "data": [/* User objects */]
  }
  ```

#### POST /api/v1/admin/users
- **Handler**: Anonymous function in admin_user_routes.go
- **Auth Required**: Yes (Admin)
- **Validation**: requests.CreateMultiTypeUser
- **Request Body**: User creation data
- **Success Response**: 201 Created

#### POST /api/v1/admin/users/import
- **Handler**: Anonymous function in admin_user_routes.go
- **Auth Required**: Yes (Admin)
- **Validation**: requests.CreateManyUsers
- **Request Body**: Bulk user import data
- **Success Response**: 201 Created

#### GET /api/v1/admin/users/:userID
- **Handler**: Anonymous function in admin_user_routes.go
- **Auth Required**: Yes (Admin)
- **Response (200)**: User object

#### PATCH /api/v1/admin/users/:userID
- **Handler**: Anonymous function in admin_user_routes.go
- **Auth Required**: Yes (Admin)
- **Validation**: requests.UpdateUser
- **Success Response**: 202 Accepted

#### DELETE /api/v1/admin/users/:userID
- **Handler**: Anonymous function in admin_user_routes.go
- **Auth Required**: Yes (Admin)
- **Success Response**: 204 No Content

#### POST /api/v1/admin/users/deleteMany
- **Handler**: Anonymous function in admin_user_routes.go
- **Auth Required**: Yes (Admin)
- **Validation**: requests.DeleteManyUser
- **Success Response**: 204 No Content

### 3.2 User Group Management (/api/v1/admin/user-groups)

#### POST /api/v1/admin/user-groups
- **Handler**: Anonymous function in admin_user_group_routes.go
- **Auth Required**: Yes (Admin)
- **Request Body**:
  ```json
  {
    "name": "string"
  }
  ```

#### GET /api/v1/admin/user-groups
- **Handler**: Anonymous function in admin_user_group_routes.go
- **Auth Required**: Yes (Admin)
- **Query Params**: page, page_size, search, sort_by, sort_order

#### PATCH /api/v1/admin/user-groups/:id
- **Handler**: Anonymous function in admin_user_group_routes.go
- **Auth Required**: Yes (Admin)

#### DELETE /api/v1/admin/user-groups/:id
- **Handler**: Anonymous function in admin_user_group_routes.go
- **Auth Required**: Yes (Admin)

---

## 4. CMS Routes (/api/v1/cms/*)

**Required Roles**: ADMIN or INSTRUCTOR
Middleware: `RBACMiddleware([ADMIN, INSTRUCTOR])`

### 4.1 Semester Management (/api/v1/cms/semesters)

#### POST /api/v1/cms/semesters
- **Handler**: Anonymous function in admin_semester_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.CreateSemester
- **Success Response**: 201 Created

#### GET /api/v1/cms/semesters
- **Handler**: Anonymous function in admin_semester_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: page, page_size, search, sort_by, sort_order, filter params

#### GET /api/v1/cms/semesters/:semID
- **Handler**: Anonymous function in admin_semester_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: Semester object

#### PATCH /api/v1/cms/semesters/:semID
- **Handler**: Anonymous function in admin_semester_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.UpdateSemester
- **Success Response**: 202 Accepted

#### DELETE /api/v1/cms/semesters/:semID
- **Handler**: Anonymous function in admin_semester_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Success Response**: 204 No Content

#### GET /api/v1/cms/semesters/:semID/affected-sections
- **Handler**: Anonymous function in admin_semester_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: Course with sections data

### 4.2 Course Management (/api/v1/cms/courses)

#### GET /api/v1/cms/courses/:courseID/sections
- **Handler**: Anonymous function in cms_course_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Description**: Get sections grouped by semester
- **Query Params**: page, page_size, search, sort_by, sort_order

#### POST /api/v1/cms/courses
- **Handler**: Anonymous function in cms_course_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.CreateCourse
- **Success Response**: 201 Created with course object

#### GET /api/v1/cms/courses
- **Handler**: Anonymous function in cms_course_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: page, page_size, search, sort_by, sort_order, show (active/archived)

#### GET /api/v1/cms/courses/:courseID
- **Handler**: Anonymous function in cms_course_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: Course object

#### PATCH /api/v1/cms/courses/:courseID
- **Handler**: Anonymous function in cms_course_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.UpdateCourse
- **Success Response**: 204 No Content

#### DELETE /api/v1/cms/courses/:courseID
- **Handler**: Anonymous function in cms_course_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Success Response**: 204 No Content

#### POST /api/v1/cms/courses/:courseID/default-labs
- **Handler**: Anonymous function in cms_course_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.SetDefaultLab
- **Success Response**: 201 Created

#### GET /api/v1/cms/courses/:courseID/default-labs
- **Handler**: Anonymous function in cms_course_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: page, page_size, search, sort_by, sort_order

#### POST /api/v1/cms/courses/:courseID/default-labs/delete
- **Handler**: Anonymous function in cms_course_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.DeleteDefaultLab

#### PATCH /api/v1/cms/courses/:courseID/default-labs
- **Handler**: Anonymous function in cms_course_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.UpdateDefaultLab

#### GET /api/v1/cms/courses/:courseID/labs
- **Handler**: Anonymous function in cms_course_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: page, page_size, search, sort_by, sort_order

### 4.3 Section Management (/api/v1/cms/sections)

#### POST /api/v1/cms/sections
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Content-Type**: multipart/form-data
- **Form Fields**: name, semester_id, course_id, instructors[], students[], banner (file)
- **Success Response**: 201 Created with `{ "id": "section-id" }`

#### PATCH /api/v1/cms/sections/:id
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Content-Type**: multipart/form-data
- **Form Fields**: name, semester_id, instructors[], students[], banner (file)
- **Success Response**: 202 Accepted

#### GET /api/v1/cms/sections
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Description**: List sections grouped by semester
- **Query Params**: page, page_size, search, sort_by, sort_order

#### GET /api/v1/cms/sections/:id
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: Section object

#### DELETE /api/v1/cms/sections/:id
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Success Response**: 204 No Content

#### GET /api/v1/cms/sections/:id/students
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: `{ "data": [/* Students */] }`

#### POST /api/v1/cms/sections/:id/students
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.SectionStudents
- **Request Body**:
  ```json
  {
    "student_usernames": ["string"]
  }
  ```

#### POST /api/v1/cms/sections/:id/students/remove
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.RemoveStudent
- **Request Body**:
  ```json
  {
    "student_ids": ["string"]
  }
  ```

#### GET /api/v1/cms/sections/:sectionID/labs
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: page, page_size, sort_by, sort_order

#### GET /api/v1/cms/sections/:sectionID/labs/:labID
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: Lab detail with completion stats

#### GET /api/v1/cms/sections/:sectionID/labs/:labID/materials/:materialID/submissions
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: student_id (optional), page, page_size, sort_by, sort_order

#### PATCH /api/v1/cms/sections/:sectionID/labs/:labID
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.UpdateLabSectionStatus
- **Success Response**: 202 Accepted

#### GET /api/v1/cms/sections/:sectionID/labs/:labID/student-status
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: Student status data

#### POST /api/v1/cms/sections/:sectionID/labs
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.SetLabSection
- **Success Response**: 201 Created

#### PATCH /api/v1/cms/sections/:sectionID/labs
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.UpdateLabSection

#### POST /api/v1/cms/sections/:sectionID/labs/delete
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.DeleteLabSection

#### GET /api/v1/cms/sections/:sectionID/logs
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: page, page_size, search, sort_by, sort_order

#### GET /api/v1/cms/sections/:id/gradebook
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: Gradebook data

#### GET /api/v1/cms/sections/:id/gradebook/export
- **Handler**: Anonymous function in cms_section_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: format (csv or xlsx)
- **Response**: File download

### 4.4 Lab Management (/api/v1/cms/labs)

#### GET /api/v1/cms/labs/:labID
- **Handler**: Anonymous function in cms_lab_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: Lab object

#### POST /api/v1/cms/labs
- **Handler**: Anonymous function in cms_lab_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.CreateLab
- **Success Response**: 201 Created with `{ "id": "lab-id" }`

#### GET /api/v1/cms/labs
- **Handler**: Anonymous function in cms_lab_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: page, page_size, search, sort_by, sort_order

#### PATCH /api/v1/cms/labs/:labID
- **Handler**: Anonymous function in cms_lab_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.BaseUpdateLab

#### DELETE /api/v1/cms/labs/:labID
- **Handler**: Anonymous function in cms_lab_routes.go
- **Auth Required**: Yes (Admin/Instructor)

#### GET /api/v1/cms/labs/:labID/sections
- **Handler**: Anonymous function in cms_lab_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: page, page_size, sort_by, sort_order

#### POST /api/v1/cms/labs/:labID/materials
- **Handler**: Anonymous function in cms_lab_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.SetLabMaterial
- **Success Response**: 201 Created

#### GET /api/v1/cms/labs/:labID/materials/all
- **Handler**: Anonymous function in cms_lab_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: All lab materials

#### GET /api/v1/cms/labs/:labID/materials
- **Handler**: Anonymous function in cms_lab_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: page, page_size, sort_by, sort_order

#### POST /api/v1/cms/labs/:labID/materials/delete
- **Handler**: Anonymous function in cms_lab_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.DeleteLabMaterial

### 4.5 Material Management (/api/v1/cms/materials)

#### POST /api/v1/cms/materials
- **Handler**: Anonymous function in cms_material_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.CreateMaterial
- **Success Response**: 201 Created with `{ "id": "material-id" }`

#### GET /api/v1/cms/materials
- **Handler**: Anonymous function in cms_material_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: page, page_size, search, sort_by, sort_order

#### GET /api/v1/cms/materials/:id
- **Handler**: Anonymous function in cms_material_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: Material object

#### PATCH /api/v1/cms/materials/:id
- **Handler**: Anonymous function in cms_material_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.BaseUpdateMaterial

#### DELETE /api/v1/cms/materials/:id
- **Handler**: Anonymous function in cms_material_routes.go
- **Auth Required**: Yes (Admin/Instructor)

#### POST /api/v1/cms/materials/:id/assets
- **Handler**: Anonymous function in cms_material_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Content-Type**: multipart/form-data
- **Form Field**: file
- **Response**: `{ "url": "file-url" }`

### 4.6 Submission Management (/api/v1/cms/submissions)

#### PATCH /api/v1/cms/submissions/:id/manual-score
- **Handler**: Anonymous function in cms_submission_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.UpdateSubmissionManualScore
- **Success Response**: 204 No Content

### 4.7 Affected Entities (/api/v1/cms/affected-entities)

#### POST /api/v1/cms/affected-entities
- **Handler**: Anonymous function in cms_affected_entities_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.GetAffectedEntities

### 4.8 User Existance (/api/v1/cms/user-existances)

#### POST /api/v1/cms/user-existances
- **Handler**: Anonymous function in cms_user_existances.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.GetInvalidUsers
- **Response**:
  - Valid: `{ "code": "OK", "message": "All users are valid" }`
  - Invalid: `{ "code": "INVALID_USERS", "error": "...", "users": [...] }`

### 4.9 CMS Users (/api/v1/cms/users)

#### GET /api/v1/cms/users/:userID
- **Handler**: Anonymous function in cms_user_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: CMSUser object (limited fields)

### 4.10 Config Management (/api/v1/cms/configs)

**Note**: These routes require gRPC connection to config-server

#### GET /api/v1/cms/configs/runners/:id
- **Handler**: Anonymous function in cms_config_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: RunnerConfigDetail

#### POST /api/v1/cms/configs/runners
- **Handler**: Anonymous function in cms_config_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.CreateRunnerRequest

#### PATCH /api/v1/cms/configs/runners/:id
- **Handler**: Anonymous function in cms_config_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.UpdateRunnerRequest

#### DELETE /api/v1/cms/configs/runners/:id
- **Handler**: Anonymous function in cms_config_routes.go
- **Auth Required**: Yes (Admin/Instructor)

#### POST /api/v1/cms/configs/runners/:id/test
- **Handler**: Anonymous function in cms_config_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Response**: SSE stream with test results

#### GET /api/v1/cms/configs/runners
- **Handler**: Anonymous function in cms_config_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: include_scripts, page, page_size, sort_order, search

#### GET /api/v1/cms/configs/compare-scripts
- **Handler**: Anonymous function in cms_config_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Query Params**: page, page_size, sort_order, search

#### POST /api/v1/cms/configs/compare-scripts
- **Handler**: Anonymous function in cms_config_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.CreateCompareRequest

#### PATCH /api/v1/cms/configs/compare-scripts/:id
- **Handler**: Anonymous function in cms_config_routes.go
- **Auth Required**: Yes (Admin/Instructor)
- **Validation**: requests.UpdateCompareRequest

#### DELETE /api/v1/cms/configs/compare-scripts/:id
- **Handler**: Anonymous function in cms_config_routes.go
- **Auth Required**: Yes (Admin/Instructor)

---

## 5. Core Routes (/api/v1/*)

**Required Roles**: ADMIN, INSTRUCTOR, or STUDENT
Middleware: `RBACMiddleware([ADMIN, INSTRUCTOR, STUDENT])`

### 5.1 Student Sections (/api/v1/sections)

#### GET /api/v1/sections
- **Handler**: Anonymous function in core_section_route.go
- **Auth Required**: Yes (Any authenticated user)
- **Description**: List sections for current student
- **Query Params**: page, page_size, sort_by, sort_order

#### GET /api/v1/sections/:sectionID
- **Handler**: Anonymous function in core_section_route.go
- **Auth Required**: Yes (Any authenticated user)
- **Response**: `{ "section": {...}, "course": {...} }`

#### GET /api/v1/sections/:sectionID/labs/:labID
- **Handler**: Anonymous function in core_section_route.go
- **Auth Required**: Yes (Any authenticated user)
- **Response**: LabSection object

#### GET /api/v1/sections/:sectionID/labs
- **Handler**: Anonymous function in core_section_route.go
- **Auth Required**: Yes (Any authenticated user)
- **Query Params**: page, page_size, sort_by, sort_order

#### DELETE /api/v1/sections/:sectionID/unenroll
- **Handler**: Anonymous function in core_section_route.go
- **Auth Required**: Yes (Student only)
- **Response**: `{ "message": "Students removed successfully" }`

### 5.2 Student Labs (/api/v1/labs)

#### POST /api/v1/labs/:labID
- **Handler**: Anonymous function in core_lab_routes.go
- **Auth Required**: Yes (Any authenticated user)
- **Validation**: requests.GetSection
- **Request Body**:
  ```json
  {
    "section_id": "string"
  }
  ```
- **Response**: Lab object

#### GET /api/v1/labs/:labID/materials
- **Handler**: Anonymous function in core_lab_routes.go
- **Auth Required**: Yes (Any authenticated user)
- **Query Params**: section_id (required), page, page_size, sort_by, sort_order

### 5.3 Sidebar (/api/v1/sidebar)

#### GET /api/v1/sidebar
- **Handler**: Anonymous function in core_sidebar_routes.go
- **Auth Required**: Yes (Any authenticated user)
- **Response**: Sidebar data

### 5.4 Submissions (/api/v1/submissions)

#### POST /api/v1/submissions
- **Handler**: Anonymous function in core_submission_routes.go
- **Auth Required**: Yes (Any authenticated user)
- **Validation**: requests.Submission
- **Success Response**: `{ "id": "submission-id" }`

#### GET /api/v1/submissions/:id
- **Handler**: Anonymous function in core_submission_routes.go
- **Auth Required**: Yes (Any authenticated user)
- **Response**: Submission object

#### GET /api/v1/submissions/:id/listen
- **Handler**: Anonymous function in core_submission_routes.go
- **Auth Required**: Yes (Any authenticated user)
- **Description**: SSE endpoint for real-time submission updates
- **Response**: SSE stream

#### GET /api/v1/submissions
- **Handler**: Anonymous function in core_submission_routes.go
- **Auth Required**: Yes (Any authenticated user)
- **Description**: List user's submissions
- **Query Params**: material_id, section_id, lab_id, page, page_size, sort_order

### 5.5 Materials (/api/v1/materials)

#### GET /api/v1/materials/:materialID
- **Handler**: Anonymous function in core_material_routes.go
- **Auth Required**: Yes (Any authenticated user)
- **Query Params**: section_id, lab_id
- **Response**: Material with latest submission status

#### GET /api/v1/materials/:materialID/submissions
- **Handler**: Anonymous function in core_material_routes.go
- **Auth Required**: Yes (Any authenticated user)
- **Query Params**: section_id, lab_id, page, page_size, sort_order
- **Response**: User's submissions for this material

---

## Summary Statistics

| Category | Count |
|----------|-------|
| Public Routes | 1 |
| Auth Routes | 4 |
| Admin Routes | 10 |
| CMS Routes | 47 |
| Core Routes | 12 |
| **Total** | **74** |

---

## Authentication Methods

1. **JWT Cookie**: `access_token` cookie (HTTPOnly)
2. **JWT Header**: `Authorization: Bearer <token>` (also supported)
3. **Refresh Token**: `refresh_token` cookie for token refresh

## Common Response Patterns

### Success Responses
- **200 OK**: Standard success with data
- **201 Created**: Resource created successfully
- **202 Accepted**: Update accepted
- **204 No Content**: Delete successful, no body

### Error Responses
All errors use the `cserrors.Error` format:
```json
{
  "message": "Error description",
  "code": "ERROR_CODE" // optional
}
```

Common HTTP status codes:
- **400**: Bad Request (validation error)
- **401**: Unauthorized (missing/invalid token)
- **403**: Forbidden (insufficient permissions)
- **404**: Not Found
- **500**: Internal Server Error

## Pagination Pattern

Most list endpoints support pagination:
```json
{
  "pagination": {
    "page": 1,
    "total_page": 10,
    "total_rows": 100
  },
  "data": [...]
}
```

Query parameters:
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 10 or 20)
- `search`: Search term
- `sort_by`: Field to sort by
- `sort_order`: asc or desc
