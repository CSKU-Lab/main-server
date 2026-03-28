package routes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"

	"github.com/gofiber/fiber/v3"
)

// mockUserServiceForAdminRoutes is a mock for testing admin user routes
type mockUserServiceForAdminRoutes struct{}

func (m *mockUserServiceForAdminRoutes) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserServiceForAdminRoutes) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserServiceForAdminRoutes) GetByID(ctx context.Context, ID string) (*models.User, error) {
	return &models.User{ID: ID}, nil
}

func (m *mockUserServiceForAdminRoutes) GetPasswordByID(ctx context.Context, ID string) (string, error) {
	return "", nil
}

func (m *mockUserServiceForAdminRoutes) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.User, error) {
	return []models.User{}, nil
}

func (m *mockUserServiceForAdminRoutes) GetInvalidUsers(ctx context.Context, req *requests.GetInvalidUsers) ([]string, error) {
	return []string{}, nil
}

func (m *mockUserServiceForAdminRoutes) Count(ctx context.Context, search string, filterParams map[string]string) (int, error) {
	return 0, nil
}

func (m *mockUserServiceForAdminRoutes) Create(ctx context.Context, user *requests.CreateMultiTypeUser) error {
	return nil
}

func (m *mockUserServiceForAdminRoutes) CreateMany(ctx context.Context, users *requests.CreateManyUsers) error {
	return nil
}

func (m *mockUserServiceForAdminRoutes) SetPassword(ctx context.Context, ID string, password string) error {
	return nil
}

func (m *mockUserServiceForAdminRoutes) Update(ctx context.Context, ID string, user *requests.UpdateUser) error {
	return nil
}

func (m *mockUserServiceForAdminRoutes) Delete(ctx context.Context, ID string) error {
	return nil
}

func (m *mockUserServiceForAdminRoutes) DeleteMany(ctx context.Context, IDs []string) error {
	return nil
}

// setupAdminUserTestApp creates a test app with admin user routes (with non-admin user)
func setupAdminUserTestApp() (*fiber.App, services.UserService) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			// Simple error handler for tests
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return c.Status(csErr.HttpStatus).JSON(fiber.Map{
					"code":  csErr.Code,
					"error": csErr.Message,
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		},
	})
	mockService := &mockUserServiceForAdminRoutes{}

	// Middleware to set a non-admin user in context
	app.Use(func(c fiber.Ctx) error {
		user := &models.User{
			ID:    "test-user-id",
			Roles: []models.Role{models.STUDENT}, // Non-admin user
		}
		c.Locals("user", user)
		return c.Next()
	})

	// Create admin router with RequireAdmin middleware
	adminRouter := app.Group("/api/v1/admin", middlewares.RequireAdmin())

	// Register routes
	NewAdminUserRoutes(adminRouter, mockService)

	return app, mockService
}

// setupAdminUserTestAppWithUser creates a test app with a user set in context
func setupAdminUserTestAppWithUser(roles []models.Role) (*fiber.App, services.UserService) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			// Simple error handler for tests
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return c.Status(csErr.HttpStatus).JSON(fiber.Map{
					"code":  csErr.Code,
					"error": csErr.Message,
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		},
	})
	mockService := &mockUserServiceForAdminRoutes{}

	// Middleware to set user in context (simulating auth middleware)
	app.Use(func(c fiber.Ctx) error {
		user := &models.User{
			ID:    "test-user-id",
			Roles: roles,
		}
		c.Locals("user", user)
		return c.Next()
	})

	// Create admin router with RequireAdmin middleware
	adminRouter := app.Group("/api/v1/admin", middlewares.RequireAdmin())

	// Register routes
	NewAdminUserRoutes(adminRouter, mockService)

	return app, mockService
}

func TestAdminUserRoutes_AdminAccess(t *testing.T) {
	app, _ := setupAdminUserTestAppWithUser([]models.Role{models.ADMIN})

	// Test GET /api/v1/admin/users
	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	// Should be OK (200) or have some other success/error related to the actual handler
	// but NOT 403 Forbidden
	if resp.StatusCode == http.StatusForbidden {
		t.Errorf("admin user should not get 403 Forbidden, got status %d", resp.StatusCode)
	}
}

func TestAdminUserRoutes_NonAdminAccess(t *testing.T) {
	app, _ := setupAdminUserTestAppWithUser([]models.Role{models.STUDENT})

	// Test GET /api/v1/admin/users
	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status %d for non-admin user, got %d", http.StatusForbidden, resp.StatusCode)
	}

	// Verify error message
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Permission denied: admin access required") {
		t.Errorf("expected error message 'Permission denied: admin access required', got %q", string(body))
	}
}

func TestAdminUserRoutes_InstructorOnlyAccess(t *testing.T) {
	app, _ := setupAdminUserTestAppWithUser([]models.Role{models.INSTRUCTOR})

	// Test GET /api/v1/admin/users
	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status %d for instructor user, got %d", http.StatusForbidden, resp.StatusCode)
	}
}

func TestAdminUserRoutes_AllEndpointsProtected(t *testing.T) {
	app, _ := setupAdminUserTestAppWithUser([]models.Role{models.STUDENT})

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/admin/users"},
		{"POST", "/api/v1/admin/users"},
		{"POST", "/api/v1/admin/users/import"},
		{"GET", "/api/v1/admin/users/user123"},
		{"PATCH", "/api/v1/admin/users/user123"},
		{"DELETE", "/api/v1/admin/users/user123"},
		{"POST", "/api/v1/admin/users/deleteMany"},
	}

	for _, endpoint := range endpoints {
		req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to send %s request to %s: %v", endpoint.method, endpoint.path, err)
		}

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("endpoint %s %s: expected status %d for non-admin, got %d", endpoint.method, endpoint.path, http.StatusForbidden, resp.StatusCode)
		}
	}
}
