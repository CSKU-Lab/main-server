# E2E (End-to-End) Test Suite for main-server

This directory contains comprehensive end-to-end regression tests for all API routes in the main-server service. These tests establish a baseline of working behavior before applying permission checks (Phase 1 of Permission Service integration).

## Overview

The E2E test suite covers all public API routes and tests both happy paths and error cases. Tests use a real PostgreSQL database to ensure realistic behavior.

## Test Structure

```
tests/e2e/
├── setup_test.go           # Test infrastructure and base suite
├── fixtures.go             # Test data creators
├── helpers.go              # Request helpers and assertions
├── routes/
│   ├── auth_test.go        # Authentication E2E tests
│   ├── users_test.go       # User management E2E tests
│   ├── courses_test.go     # Course management E2E tests
│   ├── sections_test.go    # Section management E2E tests
│   ├── labs_test.go        # Lab management E2E tests
│   ├── submissions_test.go # Submission E2E tests
│   └── grading_test.go     # Grading E2E tests
└── README.md               # This file
```

## Running the Tests

### Prerequisites

1. PostgreSQL database running (via docker-compose or local installation)
2. Environment variables configured:
   - `PGHOST` - Database host (default: localhost)
   - `PGPORT` - Database port (default: 5432)
   - `PGUSER` - Database user (default: cs_pg_user)
   - `PGPASSWORD` - Database password (default: cs_pg_password)
   - `PGDATABASE` - Database name (default: main-server)
   - `JWT_SECRET` - JWT secret for token generation
   - `JWT_REFRESH_SECRET` - JWT refresh secret

### Run All E2E Tests

```bash
# Set environment variable to enable E2E tests
export RUN_E2E_TESTS=true

# Run all E2E tests
go test -tags=e2e ./tests/e2e/...

# Run with verbose output
go test -tags=e2e -v ./tests/e2e/...
```

### Run Specific Test Suites

```bash
# Run only auth tests
go test -tags=e2e -v ./tests/e2e/routes -run TestAuthTestSuite

# Run only user tests
go test -tags=e2e -v ./tests/e2e/routes -run TestUsersTestSuite

# Run only course tests
go test -tags=e2e -v ./tests/e2e/routes -run TestCoursesTestSuite
```

### Run Individual Tests

```bash
# Run a specific test
go test -tags=e2e -v ./tests/e2e/routes -run TestAuthTestSuite/TestCredentialLogin_Success
```

## Test Coverage

### Authentication Routes (`auth_test.go`)

- POST /auth/sign-in/credential
  - ✓ Valid credentials
  - ✓ Wrong password
  - ✓ User not found
  - ✓ Missing credentials
- POST /auth/refresh-token
  - ✓ Valid refresh token
  - ✓ Invalid token
  - ✓ No cookies
  - ✓ Expired access token with valid refresh token
- Protected routes
  - ✓ Without auth (401)
  - ✓ With invalid token (401)
  - ✓ With valid token (success)

### User Routes (`users_test.go`)

- GET /admin/users
  - ✓ Admin listing (success)
  - ✓ With pagination
  - ✓ With search
  - ✓ Non-admin forbidden (403)
  - ✓ Unauthorized (401)
- POST /admin/users
  - ✓ Admin creating user (201)
  - ✓ Invalid email (400)
  - ✓ Short password (400)
  - ✓ Missing required fields (400)
  - ✓ Non-admin forbidden (403)
- GET /users/:userID
  - ✓ Admin getting any user (success)
  - ✓ User getting own profile (success)
  - ✓ Not found (500)
  - ✓ Unauthorized (401)
- PATCH /admin/users/:userID
  - ✓ Admin updating user (202)
  - ✓ Invalid data (400)
  - ✓ Non-admin forbidden (403)
- DELETE /admin/users/:userID
  - ✓ Admin deleting user (204)
  - ✓ Non-admin forbidden (403)
- POST /admin/users/deleteMany
  - ✓ Admin deleting multiple users (204)
  - ✓ Invalid IDs (400)
  - ✓ Empty list (400)
- POST /admin/users/import
  - ✓ Admin importing users (201)

### Course Routes (`courses_test.go`)

- GET /cms/courses
  - ✓ Admin listing (success)
  - ✓ Instructor listing (success)
  - ✓ Student forbidden (403)
  - ✓ Unauthorized (401)
  - ✓ With filters
- POST /cms/courses
  - ✓ Admin creating (201)
  - ✓ Instructor creating (201)
  - ✓ Student forbidden (403)
  - ✓ Invalid data (400)
  - ✓ Missing fields (400)
- GET /cms/courses/:courseID
  - ✓ Admin getting course (success)
  - ✓ Not found (500)
- PATCH /cms/courses/:courseID
  - ✓ Admin updating (204)
  - ✓ Instructor updating own course (204)
  - ✓ Invalid data (400)
- DELETE /cms/courses/:courseID
  - ✓ Admin deleting (204)
  - ✓ Instructor forbidden (403)
- GET /cms/courses/:courseID/sections
  - ✓ Admin listing sections (success)
- GET /cms/courses/:courseID/labs
  - ✓ Admin listing labs (success)
- GET /cms/courses/:courseID/default-labs
  - ✓ Admin listing default labs (success)
- POST /cms/courses/:courseID/default-labs
  - ✓ Admin setting default lab (201)

### Section Routes (`sections_test.go`)

- GET /cms/sections
  - ✓ Admin listing (success)
  - ✓ Instructor listing (success)
  - ✓ Student forbidden (403)
- POST /cms/sections
  - ✓ Admin creating (201)
  - ✓ Instructor creating (201)
  - ✓ Missing fields (400)
- GET /cms/sections/:id
  - ✓ Admin getting section (success)
  - ✓ Not found (500)
- PATCH /cms/sections/:id
  - ✓ Admin updating (202)
- DELETE /cms/sections/:id
  - ✓ Admin deleting (204)
- GET /cms/sections/:id/students
  - ✓ Admin listing students (success)
- POST /cms/sections/:id/students
  - ✓ Admin adding students (success)
- POST /cms/sections/:id/students/remove
  - ✓ Admin removing students (success)
- GET /cms/sections/:sectionID/labs
  - ✓ Admin listing labs (success)
- POST /cms/sections/:sectionID/labs
  - ✓ Admin adding lab (201)
- PATCH /cms/sections/:sectionID/labs/:labID
  - ✓ Admin updating lab status (202)
- GET /cms/sections/:sectionID/logs
  - ✓ Admin getting logs (success)
- GET /cms/sections/:id/gradebook
  - ✓ Admin getting gradebook (success)
- GET /cms/sections/:id/gradebook/export
  - ✓ Admin exporting CSV (success)
  - ✓ Admin exporting XLSX (success)
  - ✓ Invalid format (400)

### Lab Routes (`labs_test.go`)

- GET /cms/labs
  - ✓ Admin listing (success)
  - ✓ Instructor listing (success)
  - ✓ Student forbidden (403)
  - ✓ With filters
- POST /cms/labs
  - ✓ Admin creating (201)
  - ✓ Instructor creating (201)
  - ✓ Student forbidden (403)
  - ✓ Invalid data (400)
  - ✓ Missing fields (400)
- GET /cms/labs/:labID
  - ✓ Admin getting lab (success)
  - ✓ Not found (500)
- PATCH /cms/labs/:labID
  - ✓ Admin updating (success)
  - ✓ Instructor updating own lab (success)
  - ✓ Invalid data (400)
- DELETE /cms/labs/:labID
  - ✓ Admin deleting (success)
  - ✓ Instructor deleting own lab (success)
- GET /cms/labs/:labID/sections
  - ✓ Admin listing sections (success)
- GET /cms/labs/:labID/materials
  - ✓ Admin listing materials (success)
- GET /cms/labs/:labID/materials/all
  - ✓ Admin getting all materials (success)
- POST /cms/labs/:labID/materials
  - ✓ Admin adding material (201)
- POST /cms/labs/:labID/materials/delete
  - ✓ Admin removing material (success)

### Submission Routes (`submissions_test.go`)

- POST /submissions
  - ✓ Student creating (success)
  - ✓ Unauthorized (401)
  - ✓ Invalid data (400)
  - ✓ Missing payload (400)
- GET /submissions/:id
  - ✓ Owner getting own (success)
  - ✓ Instructor getting student's (success)
  - ✓ Not found (500)
- GET /submissions
  - ✓ Student listing own (success)
  - ✓ With filters (success)
- PATCH /cms/submissions/:id/manual-score
  - ✓ Instructor updating (204)
  - ✓ Admin updating (204)
  - ✓ Student forbidden (403)
  - ✓ Invalid score (400)

### Grading Routes (`grading_test.go`)

- GET /cms/sections/:sectionID/labs/:labID/materials/:materialID/submissions
  - ✓ Instructor getting submissions (success)
  - ✓ Admin getting submissions (success)
  - ✓ Student forbidden (403)
  - ✓ With student_id filter (success)
- GET /cms/sections/:sectionID/labs/:labID/student-status
  - ✓ Instructor getting status (success)
  - ✓ Admin getting status (success)
  - ✓ Student forbidden (403)
- GET /cms/sections/:sectionID/labs/:labID
  - ✓ Instructor getting lab with stats (success)
- PATCH /cms/submissions/:id/manual-score
  - ✓ Instructor submitting grade (204)
  - ✓ Invalid score (400)
  - ✓ Non-instructor forbidden (403)
- GET /cms/sections/:id/gradebook
  - ✓ Admin getting gradebook (success)
  - ✓ Instructor getting gradebook (success)
  - ✓ Student forbidden (403)
- GET /cms/sections/:id/gradebook/export
  - ✓ Admin exporting CSV (success)
  - ✓ Admin exporting XLSX (success)
  - ✓ Invalid format (400)

## Test Data Strategy

### Test Users

Each test suite creates test users with different roles:

- **Admin**: Full access to all endpoints
- **Instructor**: Access to instructor endpoints and own resources
- **Student**: Access to student endpoints and own resources
- **Student2**: Additional student for cross-user permission tests

### Test Data Lifecycle

1. **SetupSuite**: Creates base test users
2. **SetupTest**: Individual tests create specific test data
3. **TearDownTest**: Cleans up test-specific data
4. **TearDownSuite**: Cleans up all test data

### Data Isolation

- Each test uses unique identifiers (prefixed with `e2e_test_`)
- Database transactions ensure isolation
- Cleanup functions remove all test data after tests

## Writing New E2E Tests

### Basic Test Structure

```go
//go:build e2e
// +build e2e

package e2e

import (
    "testing"
    "github.com/stretchr/testify/suite"
)

type MyFeatureTestSuite struct {
    TestSuite
}

func TestMyFeatureTestSuite(t *testing.T) {
    CheckE2EEnabled(t)
    suite.Run(t, new(MyFeatureTestSuite))
}

func (s *MyFeatureTestSuite) TestMyFeature_Success() {
    // Create test data
    resourceID := s.CreateTestResource()
    defer s.CleanupTestResource(resourceID)
    
    // Get auth token
    token := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)
    
    // Make request
    resp := s.RequestWithAuth("GET", BuildURL("/my-endpoint/"+resourceID), nil, token)
    
    // Assert response
    s.AssertSuccess(resp)
}
```

### Available Helpers

**Request Helpers:**
- `RequestWithAuth(method, path, body, token)` - Make authenticated request
- `RequestWithoutAuth(method, path, body)` - Make unauthenticated request
- `RequestWithFormData(method, path, fields, files, token)` - Make multipart request

**Assertion Helpers:**
- `AssertSuccess(resp)` - Assert 200 OK
- `AssertCreated(resp)` - Assert 201 Created
- `AssertNoContent(resp)` - Assert 204 No Content
- `AssertBadRequest(resp)` - Assert 400 Bad Request
- `AssertUnauthorized(resp)` - Assert 401 Unauthorized
- `AssertForbidden(resp)` - Assert 403 Forbidden
- `AssertNotFound(resp)` - Assert 404 Not Found
- `AssertErrorResponse(resp, message)` - Assert error message

**Fixture Helpers:**
- `CreateTestUser(role, roles)` - Create test user
- `CreateTestCourse(creatorID)` - Create test course
- `CreateTestSection(courseID, semesterID, instructors, students)` - Create test section
- `CreateTestLab(courseID, createdBy)` - Create test lab
- `CreateTestMaterial(labID, materialType)` - Create test material
- `CreateTestSubmission(userID, materialID, labID, sectionID)` - Create test submission
- `GenerateTestJWT(userID, username, roles)` - Generate JWT token

## CI/CD Integration

To run E2E tests in CI/CD:

```yaml
# Example GitHub Actions step
- name: Run E2E Tests
  env:
    RUN_E2E_TESTS: true
    PGHOST: localhost
    PGPORT: 5432
    PGUSER: test_user
    PGPASSWORD: test_password
    PGDATABASE: test_db
    JWT_SECRET: test-secret
    JWT_REFRESH_SECRET: test-refresh-secret
  run: |
    go test -tags=e2e -v ./tests/e2e/...
```

## Troubleshooting

### Database Connection Issues

```bash
# Test database connection
psql $DATABASE_URL -c "SELECT 1"

# Check if migrations are applied
./scripts/migrate.sh status
```

### Test Failures

1. Check that all environment variables are set
2. Ensure database is running and accessible
3. Verify migrations are up to date
4. Check test logs for specific error messages

### Skipped Tests

Tests are skipped if `RUN_E2E_TESTS` is not set to `true`. This prevents accidental running of E2E tests during unit test execution.

## Maintenance

### Adding New Routes

When adding new API routes:

1. Add corresponding E2E tests in the appropriate `*_test.go` file
2. Test happy path and error cases
3. Test with different user roles
4. Update this README with new test coverage

### Updating Tests

When API behavior changes:

1. Update affected E2E tests
2. Ensure all tests pass before merging
3. Document breaking changes in commit messages

## Success Criteria

✅ All route files created with tests
✅ E2E test infrastructure (setup, fixtures, helpers)
✅ >80% API route coverage
✅ All tests passing consistently (no flaky tests)
✅ Tests include happy path + error cases
✅ Test documentation with descriptions
✅ No external dependencies mocked - use real DB

## References

- [Go Test Documentation](https://golang.org/pkg/testing/)
- [Testify Suite Documentation](https://github.com/stretchr/testify#suite-package)
- [main-server API Documentation](../../README.md)
