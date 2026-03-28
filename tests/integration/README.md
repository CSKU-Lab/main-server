# Integration Tests

This directory contains integration tests for the main-server that test against real infrastructure (PostgreSQL database).

## Running Integration Tests

### Prerequisites

1. **Docker and Docker Compose** must be installed
2. **PostgreSQL database** must be running (via docker-compose.dev.yaml)

### Start the Infrastructure

```bash
# From the project root
docker-compose -f docker-compose.dev.yaml up -d db
```

### Run Integration Tests

```bash
# From main-server directory
cd main-server

# Run all integration tests
go test -v ./tests/integration/... -tags=integration

# Run specific integration test
go test -v ./tests/integration/... -tags=integration -run TestLabUniquenessIntegrationTestSuite

# Run with race detector
go test -race -v ./tests/integration/... -tags=integration

# Run in short mode (skips integration tests)
go test -v ./tests/integration/... -short
```

### Environment Variables

The integration tests use standard PostgreSQL environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PGHOST` | `localhost` | PostgreSQL host |
| `PGPORT` | `5432` | PostgreSQL port |
| `PGUSER` | `cs_pg_user` | PostgreSQL username |
| `PGPASSWORD` | `cs_pg_password` | PostgreSQL password |
| `PGDATABASE` | `main-server` | PostgreSQL database name |

### Test Structure

- `lab_uniqueness_integration_test.go` - Tests for lab name uniqueness constraint (GitHub issue #13)
  - Verifies same lab name can exist in different courses
  - Verifies duplicate lab names fail in same course
  - Verifies soft-deleted labs allow name reuse
  - Verifies composite unique index exists
  - Verifies old global index does not exist

### Database State

Integration tests:
1. Connect to the real PostgreSQL database
2. Run migrations to ensure correct schema
3. Create test data (users, courses, labs)
4. Run assertions against real database
5. Clean up test data after each test

Tests use unique identifiers to avoid conflicts with existing data.

### Skipping Integration Tests

Integration tests are skipped if:
- The `-short` flag is provided
- The PostgreSQL database is not available
- The `integration` build tag is not specified

### Troubleshooting

**Tests fail with "database not available"**
- Ensure PostgreSQL container is running: `docker-compose -f docker-compose.dev.yaml ps`
- Check database logs: `docker-compose -f docker-compose.dev.yaml logs db`
- Verify environment variables match docker-compose configuration

**Migration failures**
- Check that atlas migrations are applied: `./scripts/migrate.sh`
- Verify database schema: `psql -h localhost -U cs_pg_user -d main-server -c "\d labs"`

**Port conflicts**
- Ensure port 5432 is not in use by another PostgreSQL instance
- Check with: `lsof -i :5432`
