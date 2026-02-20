# AGENTS.md — main-server

## Project Overview

Go 1.25 REST API server following **Clean Architecture** (domain → adapters → infrastructure).

- **HTTP:** GoFiber v3
- **Database:** PostgreSQL via `sqlx` (raw SQL, no ORM)
- **Migrations:** Atlas (`atlas/schema.hcl`)
- **Message Queue:** RabbitMQ (`CSKU-Lab/queue` wrapper)
- **PubSub:** Redis + PostgreSQL `LISTEN/NOTIFY`
- **Object Storage:** MinIO
- **Auth:** JWT (`golang-jwt`) + Google OAuth2
- **gRPC:** Client-side only (grader, config, task services)
- **Module path:** `github.com/CSKU-Lab/main-server`

---

## Build & Run Commands

```sh
# Build binary
go build -o ./tmp/main ./cmd/app

# Run directly
go run ./cmd/app

# Dev server with hot reload (uses .air.toml)
air

# Seed the database
go run ./cmd/seed

# Start dev infrastructure (Postgres, Redis, RabbitMQ, MinIO)
./scripts/compose-infra.sh

# Start full dev stack
./scripts/compose.sh up

# Run database migrations (Atlas)
./scripts/migrate.sh

# Apply PostgreSQL triggers
./scripts/postgres_triggers.sh
```

---

## Test Commands

```sh
# Run all tests
go test ./...

# Run all tests with verbose output
go test -v ./...

# Run all tests with race detector
go test -race ./...

# Run tests in a single package
go test ./internal/transaction/
go test ./domain/services/

# Run a single test case by name (regex matched against test function name)
go test -run TestOneStepSuccess ./internal/transaction/
go test -run TestApplyLabSectionScheduleDerivesStatusFromDates ./domain/services/

# Run a single test with verbose output
go test -v -run TestCommitWithRetry ./internal/transaction/
```

**Test framework:** standard `testing` package + `github.com/stretchr/testify/assert`.

White-box tests (same package, access unexported symbols) use `package <name>` declaration.
Black-box tests use `package <name>_test` declaration.

---

## Code Style

### Import Ordering

Imports must be grouped into **three blocks** separated by blank lines:

1. Standard library
2. Internal project packages (`github.com/CSKU-Lab/main-server/...`)
3. Third-party external packages

```go
import (
    "context"
    "errors"
    "net/http"

    "github.com/CSKU-Lab/main-server/domain/cserrors"
    "github.com/CSKU-Lab/main-server/domain/models"
    "github.com/CSKU-Lab/main-server/domain/repositories"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
)
```

Use import aliases for disambiguation (protobuf packages, adapter packages):

```go
import (
    configPB   "github.com/CSKU-Lab/main-server/genproto/config/v1"
    graderPB   "github.com/CSKU-Lab/main-server/genproto/grader/v1"
    sqlxAdapter "github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
)
```

### Naming Conventions

| Element | Convention | Examples |
|---|---|---|
| Packages | lowercase single word | `services`, `middlewares`, `pubsub` |
| Exported types/interfaces | PascalCase | `UserService`, `CMSRouter`, `UoWInstance` |
| Unexported types | camelCase | `userService`, `labSectionService`, `uowImpl` |
| Constructors (exported) | `New<Type>()` | `NewUserService()`, `NewAuthRouter()` |
| Methods/functions (unexported) | camelCase | `mutationPermission()`, `rowExists()` |
| Struct fields (exported) | PascalCase | `HttpStatus`, `DisplayName`, `ProfileImage` |
| `db:` tag fields | PascalCase field, snake_case tag | `GroupID *string \`db:"group_id"\`` |
| `json:` tags | snake_case | `\`json:"display_name"\`` |
| Constants | SCREAMING_SNAKE_CASE | `CODE_EXECUTION_QUEUED`, `ADMIN`, `STUDENT` |
| File names | snake_case | `user_service.go`, `cms_course_routes.go` |
| Test files | `<name>_test.go` | `transaction_test.go` |

### Formatting

Use standard `gofmt` / `goimports`. No custom formatter config is present.

---

## Architecture Patterns

The project is divided into three layers. **Dependencies only flow inward** — adapters depend on domain, never the reverse.

```
domain/          ← pure business logic, no framework imports
  models/        ← domain structs
  repositories/  ← repository interfaces + DTOs
  services/      ← service interfaces + implementations
  cserrors/      ← custom error types

internal/
  adapters/
    rest/        ← Fiber route handlers (HTTP adapter)
    sqlx/        ← repository implementations (DB adapter)
    middlewares/ ← Fiber middlewares
    pubsub/      ← Redis + Postgres pub/sub adapters
  requests/      ← request DTOs with ozzo-validation
  transaction/   ← Unit of Work / saga utilities

infrastructure/
  auth/          ← JWT + Google OAuth2 utilities

cmd/app/         ← entrypoint: wires DI, starts HTTP server + worker
```

### Constructor Pattern

Constructors **always return the interface type**, never the concrete struct:

```go
func NewUserService(repo repositories.User, ...) UserService {
    return &userService{repo: repo, ...}
}

func NewUserRepository(db instance) repositories.User {
    return &userRepository{db: db}
}
```

Manual constructor injection is used throughout — no DI framework.

---

## Error Handling

Use the custom `cserrors` package for all domain and HTTP errors. Never return bare `errors.New` for user-facing errors.

```go
// Create a structured error with HTTP status
return nil, cserrors.New(&cserrors.Option{
    HttpStatus: http.StatusBadRequest,
    Message:    "Invalid sort by field",
})

// Create a redirect error (OAuth flows)
return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
```

Standard early-return pattern — always check `err != nil` immediately:

```go
result, err := someOperation(ctx)
if err != nil {
    return nil, err
}
```

Discriminate error types with `errors.As`:

```go
var csErr *cserrors.Error
if errors.As(err, &csErr) {
    return err
}
return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "unexpected error"})
```

The global Fiber error handler (`internal/adapters/middlewares/error_handler_middleware.go`) translates
`*cserrors.Error` → JSON response, `RedirectError` → HTTP redirect, `*fiber.Error` → status code.

**No `panic`/`recover` in business logic.** Use `log.Fatal` / `logger.Fatalln` only at startup for
unrecoverable failures (missing config, failed DB connection).

---

## Database Patterns

All queries are **raw SQL** with `sqlx`. There is no ORM.

```go
// Named query with sqlx
query, args, err := sqlx.Named(`SELECT * FROM users WHERE id = :id`, map[string]any{"id": id})
if err != nil { return nil, err }
query = db.Rebind(query)
err = db.GetContext(ctx, &result, query, args...)

// Bulk IN query
query, args, err := sqlx.In(`SELECT * FROM users WHERE id IN (?)`, ids)
query = db.Rebind(query)
```

**Struct tags** — use `db:"column_name"` for sqlx mapping; pointer types for nullable columns:

```go
type UserRow struct {
    ID          string     `db:"id"`
    DisplayName string     `db:"display_name"`
    DeletedAt   *time.Time `db:"deleted_at"`
}
```

**Soft deletes:** use `is_deleted = false` / `deleted_at IS NULL` in WHERE clauses.

**Transactions:** use the `UoWInstance` interface from `internal/transaction/`. Never raw `db.BeginTxx` in service code.

**Migrations:** edit `atlas/schema.hcl` and run `./scripts/migrate.sh`.

---

## HTTP Layer (GoFiber v3)

Handler signature:

```go
router.Get("/:id", func(c fiber.Ctx) error {
    id := c.Params("id")
    page := c.Query("page", "1")

    var req SomeRequest
    if err := c.Bind().Body(&req); err != nil {
        return err
    }

    result, err := service.DoSomething(c.Context(), id, &req)
    if err != nil {
        return err   // forwarded to global error handler
    }

    return c.JSON(result)
})
```

- Pass `c.Context()` (a `context.Context`) to all service/repository calls.
- Store request-scoped data in locals: `c.Locals("user", &user)`.
- Return errors directly — the global error handler middleware converts them to HTTP responses.
- Use `c.Status(http.StatusCreated).JSON(result)` for non-200 success responses.
- Use `c.SendStatus(http.StatusNoContent)` for empty responses.

---

## Context & Concurrency

Every service and repository method must accept `context.Context` as its **first parameter**:

```go
func (s *userService) GetByID(ctx context.Context, id string) (*models.User, error) { ... }
```

Use goroutines + channels for background work:

```go
payloadChan := make(chan []byte)
go func() {
    select {
    case <-ctx.Done():
        close(payloadChan)
        return
    case msg := <-sub.Channel():
        payloadChan <- []byte(msg.Payload)
    }
}()
```

Always propagate context cancellation — do not use `context.Background()` inside handlers or services
unless spawning a truly independent background goroutine (e.g., the submission worker).
