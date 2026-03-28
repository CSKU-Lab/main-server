package middlewares

import (
	"errors"
	"net/http"
	"testing"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

// mockCondition is a test condition that returns a configurable value
type mockCondition struct {
	shouldSatisfy bool
}

func (m mockCondition) IsSatisfied(userID string) bool {
	return m.shouldSatisfy
}

func TestRequirePermission_AllowsAccessWhenConditionSatisfied(t *testing.T) {
	app := fiber.New()

	// Create a mock condition that always allows
	condition := mockCondition{shouldSatisfy: true}

	app.Get("/test", func(c fiber.Ctx) error {
		c.Locals("user", &models.User{ID: "user-123"})
		return c.Next()
	}, RequirePermission(condition), func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequirePermission_DeniesAccessWhenConditionNotSatisfied(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			// Return the error status code directly
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return c.Status(csErr.HttpStatus).JSON(fiber.Map{"error": csErr.Message})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Create a mock condition that always denies
	condition := mockCondition{shouldSatisfy: false}

	app.Get("/test", func(c fiber.Ctx) error {
		c.Locals("user", &models.User{ID: "user-123"})
		return c.Next()
	}, RequirePermission(condition), func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRequirePermission_WithIsAdmin(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return c.Status(csErr.HttpStatus).JSON(fiber.Map{"error": csErr.Message})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// IsAdmin mock returns false, so this should deny
	app.Get("/admin", func(c fiber.Ctx) error {
		c.Locals("user", &models.User{ID: "user-123"})
		return c.Next()
	}, RequirePermission(permission.IsAdmin), func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "admin access granted"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRequirePermission_WithIsAuthenticated(t *testing.T) {
	app := fiber.New()

	// IsAuthenticated always returns true
	app.Get("/me", func(c fiber.Ctx) error {
		c.Locals("user", &models.User{ID: "user-123"})
		return c.Next()
	}, RequirePermission(permission.IsAuthenticated), func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "authenticated access granted"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/me", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequirePermission_ReturnsCsError(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			// Check that the error is a cserrors.Error with correct status
			var csErr *cserrors.Error
			if assert.ErrorAs(t, err, &csErr) {
				assert.Equal(t, http.StatusForbidden, csErr.HttpStatus)
				assert.Equal(t, "Forbidden: insufficient permissions", csErr.Message)
			}
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		},
	})

	condition := mockCondition{shouldSatisfy: false}

	app.Get("/protected", func(c fiber.Ctx) error {
		c.Locals("user", &models.User{ID: "user-123"})
		return c.Next()
	}, RequirePermission(condition))

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
