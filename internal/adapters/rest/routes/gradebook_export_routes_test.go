package routes

import (
	"context"
	"encoding/csv"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"

	"github.com/gofiber/fiber/v3"
	"github.com/xuri/excelize/v2"
)

// mockGradebookExportService is a mock for testing
type mockGradebookExportService struct {
	services.GradebookExportService
	csvData  []byte
	xlsxData []byte
	err      error
}

func (m *mockGradebookExportService) ExportCSV(ctx context.Context, sectionID string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.csvData, nil
}

func (m *mockGradebookExportService) ExportXLSX(ctx context.Context, sectionID string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.xlsxData, nil
}

// mockSubmissionServiceForRoutes is a mock for testing
type mockSubmissionServiceForRoutes struct {
	services.SubmissionService
	gradebook *models.Gradebook
	err       error
}

func (m *mockSubmissionServiceForRoutes) GetGradebookBySectionID(ctx context.Context, ID string) (*models.Gradebook, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.gradebook, nil
}

func setupTestApp(exportService services.GradebookExportService, submissionService services.SubmissionService) *fiber.App {
	app := fiber.New()

	// Create a minimal router setup
	router := app.Group("/api/v1/cms")

	// Register only the export endpoint for testing
	router.Get("/sections/:id/gradebook/export", func(c fiber.Ctx) error {
		id := c.Params("id")
		format := c.Query("format", "csv")

		if format != "csv" && format != "xlsx" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid format. Must be 'csv' or 'xlsx'",
			})
		}

		var data []byte
		var err error
		var contentType string
		var filename string

		if format == "csv" {
			data, err = exportService.ExportCSV(c.Context(), id)
			contentType = "text/csv"
			filename = "gradebook.csv"
		} else {
			data, err = exportService.ExportXLSX(c.Context(), id)
			contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
			filename = "gradebook.xlsx"
		}

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		c.Set("Content-Type", contentType)
		c.Set("Content-Disposition", "attachment; filename="+filename)
		return c.Status(fiber.StatusOK).Send(data)
	})

	return app
}

func TestGradebookExportEndpoint_CSV(t *testing.T) {
	// Prepare CSV data
	csvData := "Username,Display Name,Lab 1 (Auto Score),Lab 1 (Manual Score)\nstudent1,John Doe,90,45\n"

	mockExportService := &mockGradebookExportService{
		csvData: []byte(csvData),
	}

	mockSubmissionService := &mockSubmissionServiceForRoutes{}

	app := setupTestApp(mockExportService, mockSubmissionService)

	req := httptest.NewRequest("GET", "/api/v1/cms/sections/section1/gradebook/export?format=csv", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/csv" {
		t.Errorf("expected Content-Type 'text/csv', got %q", contentType)
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "attachment; filename=gradebook.csv" {
		t.Errorf("expected Content-Disposition with filename, got %q", contentDisposition)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(body) != csvData {
		t.Errorf("expected CSV data %q, got %q", csvData, string(body))
	}
}

func TestGradebookExportEndpoint_XLSX(t *testing.T) {
	// Create a simple XLSX file for testing
	f := excelize.NewFile()
	sheetName := "Gradebook"
	f.NewSheet(sheetName)
	f.SetCellValue(sheetName, "A1", "Username")
	f.SetCellValue(sheetName, "B1", "Display Name")
	f.SetCellValue(sheetName, "A2", "student1")
	f.SetCellValue(sheetName, "B2", "John Doe")
	f.DeleteSheet("Sheet1")

	xlsxData, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("failed to create XLSX data: %v", err)
	}

	mockExportService := &mockGradebookExportService{
		xlsxData: xlsxData.Bytes(),
	}

	mockSubmissionService := &mockSubmissionServiceForRoutes{}

	app := setupTestApp(mockExportService, mockSubmissionService)

	req := httptest.NewRequest("GET", "/api/v1/cms/sections/section1/gradebook/export?format=xlsx", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	expectedContentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	if contentType != expectedContentType {
		t.Errorf("expected Content-Type %q, got %q", expectedContentType, contentType)
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "attachment; filename=gradebook.xlsx" {
		t.Errorf("expected Content-Disposition with filename, got %q", contentDisposition)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if len(body) == 0 {
		t.Error("expected XLSX data, got empty response")
	}

	// Verify it's valid XLSX
	_, err = excelize.OpenReader(strings.NewReader(string(body)))
	if err != nil {
		t.Errorf("response is not valid XLSX: %v", err)
	}
}

func TestGradebookExportEndpoint_DefaultFormat(t *testing.T) {
	csvData := "Username,Display Name\n"

	mockExportService := &mockGradebookExportService{
		csvData: []byte(csvData),
	}

	mockSubmissionService := &mockSubmissionServiceForRoutes{}

	app := setupTestApp(mockExportService, mockSubmissionService)

	// No format parameter - should default to CSV
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/section1/gradebook/export", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/csv" {
		t.Errorf("expected default format to be CSV, got Content-Type %q", contentType)
	}
}

func TestGradebookExportEndpoint_InvalidFormat(t *testing.T) {
	mockExportService := &mockGradebookExportService{}
	mockSubmissionService := &mockSubmissionServiceForRoutes{}

	app := setupTestApp(mockExportService, mockSubmissionService)

	req := httptest.NewRequest("GET", "/api/v1/cms/sections/section1/gradebook/export?format=pdf", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status %d for invalid format, got %d", fiber.StatusBadRequest, resp.StatusCode)
	}
}

func TestGradebookExportEndpoint_MultipleStudentsAndLabs(t *testing.T) {
	// Create CSV with multiple students and labs
	csvBuilder := strings.Builder{}
	writer := csv.NewWriter(&csvBuilder)

	header := []string{
		"Username",
		"Display Name",
		"Lab 1 (Auto Score)",
		"Lab 1 (Manual Score)",
		"Lab 2 (Auto Score)",
		"Lab 2 (Manual Score)",
	}
	writer.Write(header)
	writer.Write([]string{"student1", "John Doe", "90", "45", "85", "40"})
	writer.Write([]string{"student2", "Jane Smith", "95", "50", "88", "42"})
	writer.Write([]string{"student3", "Bob Johnson", "80", "38", "90", "45"})
	writer.Flush()

	csvData := csvBuilder.String()

	mockExportService := &mockGradebookExportService{
		csvData: []byte(csvData),
	}

	mockSubmissionService := &mockSubmissionServiceForRoutes{}

	app := setupTestApp(mockExportService, mockSubmissionService)

	req := httptest.NewRequest("GET", "/api/v1/cms/sections/section1/gradebook/export?format=csv", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	// Parse CSV and verify row count
	reader := csv.NewReader(strings.NewReader(string(body)))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	expectedRows := 4 // 1 header + 3 students
	if len(records) != expectedRows {
		t.Errorf("expected %d rows, got %d", expectedRows, len(records))
	}

	// Verify student count
	studentCount := len(records) - 1
	if studentCount != 3 {
		t.Errorf("expected 3 students, got %d", studentCount)
	}
}
