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
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	"github.com/CSKU-Lab/queue"

	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// mockConfigServiceClient is a mock for testing CMS config routes
type mockConfigServiceClient struct {
	getCompareFunc func(ctx context.Context, in *configPB.GetCompareRequest, opts ...grpc.CallOption) (*configPB.CompareResponse, error)
}

func (m *mockConfigServiceClient) GetRunner(ctx context.Context, in *configPB.GetRunnerRequest, opts ...grpc.CallOption) (*configPB.RunnerResponse, error) {
	return nil, nil
}

func (m *mockConfigServiceClient) CreateRunner(ctx context.Context, in *configPB.CreateRunnerRequest, opts ...grpc.CallOption) (*configPB.CreateRunnerResponse, error) {
	return nil, nil
}

func (m *mockConfigServiceClient) UpdateRunner(ctx context.Context, in *configPB.UpdateRunnerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *mockConfigServiceClient) DeleteRunner(ctx context.Context, in *configPB.DeleteRunnerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *mockConfigServiceClient) GetRunnersPagination(ctx context.Context, in *configPB.GetRunnersPaginationRequest, opts ...grpc.CallOption) (*configPB.GetRunnersPaginationResponse, error) {
	return nil, nil
}

func (m *mockConfigServiceClient) GetAllRunners(ctx context.Context, in *configPB.GetAllRunnersRequest, opts ...grpc.CallOption) (*configPB.GetAllRunnersResponse, error) {
	return nil, nil
}

func (m *mockConfigServiceClient) GetCompare(ctx context.Context, in *configPB.GetCompareRequest, opts ...grpc.CallOption) (*configPB.CompareResponse, error) {
	if m.getCompareFunc != nil {
		return m.getCompareFunc(ctx, in, opts...)
	}
	return &configPB.CompareResponse{
		Id:          in.Id,
		Name:        "Test Compare Script",
		Description: "Test description",
		BuildScript: "echo build",
		RunScript:   "echo run",
		RunName:     "test-run",
		Files:       []*configPB.File{},
	}, nil
}

func (m *mockConfigServiceClient) CreateCompare(ctx context.Context, in *configPB.CreateCompareRequest, opts ...grpc.CallOption) (*configPB.CreateCompareResponse, error) {
	return nil, nil
}

func (m *mockConfigServiceClient) UpdateCompare(ctx context.Context, in *configPB.UpdateCompareRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *mockConfigServiceClient) DeleteCompare(ctx context.Context, in *configPB.DeleteCompareRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *mockConfigServiceClient) GetComparesPagination(ctx context.Context, in *configPB.GetComparesPaginationRequest, opts ...grpc.CallOption) (*configPB.GetComparesPaginationResponse, error) {
	return nil, nil
}

func (m *mockConfigServiceClient) GetAllCompares(ctx context.Context, in *configPB.GetAllComparesRequest, opts ...grpc.CallOption) (*configPB.GetAllComparesResponse, error) {
	return nil, nil
}

// mockQueue is a mock queue for testing
type mockQueue struct{}

func (m *mockQueue) CreateQueue(ctx context.Context, name string, opts *queue.QueueOptions) (string, error) {
	return "", nil
}

func (m *mockQueue) DeleteQueue(ctx context.Context, name string) error {
	return nil
}

func (m *mockQueue) Publish(ctx context.Context, exchange, routingKey string, delivery *queue.Derivery) error {
	return nil
}

func (m *mockQueue) Consume(ctx context.Context, queueName string, prefetchCount int, requeue bool, handler func(*queue.Derivery, chan struct{}) error) error {
	return nil
}

func (m *mockQueue) Close() error {
	return nil
}

// setupCMSConfigTestApp creates a test app with CMS config routes
func setupCMSConfigTestApp(roles []models.Role, mockClient *mockConfigServiceClient) (*fiber.App, *mockConfigServiceClient) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return c.Status(csErr.HttpStatus).JSON(fiber.Map{
					"code":  csErr.Code,
					"error": csErr.Message,
				})
			}
			// Check if it's a type assertion error (user not in context)
			if strings.Contains(err.Error(), "interface conversion") || strings.Contains(err.Error(), "nil pointer") {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Authentication required",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		},
	})

	if mockClient == nil {
		mockClient = &mockConfigServiceClient{}
	}
	mockQ := &mockQueue{}

	// Middleware to set user in context (must be first)
	app.Use(func(c fiber.Ctx) error {
		user := &models.User{
			ID:    "test-user-id",
			Roles: roles,
		}
		c.Locals("user", user)
		return c.Next()
	})

	// Create API router - NewCMSConfigRoutes will apply RBAC middleware internally
	apiRouter := app.Group("/api/v1")

	// Register routes - RBAC middleware is applied inside NewCMSConfigRoutes
	NewCMSConfigRoutes(apiRouter, mockClient, mockQ)

	return app, mockClient
}

func TestGetCompareScript_Success(t *testing.T) {
	mockClient := &mockConfigServiceClient{
		getCompareFunc: func(ctx context.Context, in *configPB.GetCompareRequest, opts ...grpc.CallOption) (*configPB.CompareResponse, error) {
			return &configPB.CompareResponse{
				Id:          "compare-123",
				Name:        "My Compare Script",
				Description: "A test compare script",
				BuildScript: "make build",
				RunScript:   "make run",
				RunName:     "custom-run",
				Files: []*configPB.File{
					{Name: "test.txt", Content: "test content"},
				},
			}, nil
		},
	}

	app, _ := setupCMSConfigTestApp([]models.Role{models.ADMIN}, mockClient)

	// Test GET /api/v1/configs/compare-scripts/compare-123
	req := httptest.NewRequest("GET", "/api/v1/configs/compare-scripts/compare-123", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify response body contains expected fields
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "compare-123") {
		t.Errorf("expected response to contain compare ID, got %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "My Compare Script") {
		t.Errorf("expected response to contain compare name, got %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "make build") {
		t.Errorf("expected response to contain build_script, got %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "custom-run") {
		t.Errorf("expected response to contain run_name, got %s", bodyStr)
	}
}

func TestGetCompareScript_NotFound(t *testing.T) {
	mockClient := &mockConfigServiceClient{
		getCompareFunc: func(ctx context.Context, in *configPB.GetCompareRequest, opts ...grpc.CallOption) (*configPB.CompareResponse, error) {
			return nil, cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusNotFound,
				Message:    "Compare script not found",
			})
		},
	}

	app, _ := setupCMSConfigTestApp([]models.Role{models.ADMIN}, mockClient)

	// Test GET /api/v1/configs/compare-scripts/non-existent
	req := httptest.NewRequest("GET", "/api/v1/configs/compare-scripts/non-existent", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d for non-existent compare, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestGetCompareScript_Forbidden_Student(t *testing.T) {
	app, _ := setupCMSConfigTestApp([]models.Role{models.STUDENT}, nil)

	// Test GET /api/v1/configs/compare-scripts/123 as student
	req := httptest.NewRequest("GET", "/api/v1/configs/compare-scripts/123", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status %d for student user, got %d", http.StatusForbidden, resp.StatusCode)
	}
}

func TestGetCompareScript_AdminAccess(t *testing.T) {
	app, _ := setupCMSConfigTestApp([]models.Role{models.ADMIN}, nil)

	// Test GET /api/v1/configs/compare-scripts/123 as admin
	req := httptest.NewRequest("GET", "/api/v1/configs/compare-scripts/123", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d for admin user, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestGetCompareScript_InstructorAccess(t *testing.T) {
	app, _ := setupCMSConfigTestApp([]models.Role{models.INSTRUCTOR}, nil)

	// Test GET /api/v1/configs/compare-scripts/123 as instructor
	req := httptest.NewRequest("GET", "/api/v1/configs/compare-scripts/123", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d for instructor user, got %d", http.StatusOK, resp.StatusCode)
	}
}
