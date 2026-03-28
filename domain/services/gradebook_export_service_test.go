package services

import (
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"testing"

	"github.com/CSKU-Lab/main-server/domain/models"

	"github.com/xuri/excelize/v2"
)

// mockSubmissionService is a mock implementation for testing
type mockSubmissionService struct {
	SubmissionService
	gradebook *models.Gradebook
	err       error
}

func (m *mockSubmissionService) GetGradebookBySectionID(ctx context.Context, ID string) (*models.Gradebook, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.gradebook, nil
}

func createTestGradebook() *models.Gradebook {
	return &models.Gradebook{
		LabCol: []models.LabCol{
			{
				LabID:          "lab1",
				LabName:        "Lab 1",
				MaxAutoScore:   100,
				MaxManualScore: 50,
			},
			{
				LabID:          "lab2",
				LabName:        "Lab 2",
				MaxAutoScore:   80,
				MaxManualScore: 20,
			},
		},
		StudentRow: []models.StudentRow{
			{
				Username:    "student1",
				DisplayName: "John Doe",
				LabScores: map[string]models.LabScore{
					"lab1": {AutoScore: 90, ManualScore: 45},
					"lab2": {AutoScore: 70, ManualScore: 15},
				},
			},
			{
				Username:    "student2",
				DisplayName: "Jane Smith",
				LabScores: map[string]models.LabScore{
					"lab1": {AutoScore: 85, ManualScore: 50},
					"lab2": {AutoScore: 75, ManualScore: 18},
				},
			},
		},
	}
}

func TestExportCSV_Success(t *testing.T) {
	mockService := &mockSubmissionService{
		gradebook: createTestGradebook(),
	}

	exportService := NewGradebookExportService(mockService)
	data, err := exportService.ExportCSV(context.Background(), "section1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected CSV data, got empty bytes")
	}

	// Parse CSV to verify structure
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	// Verify header
	if len(records) < 1 {
		t.Fatal("expected at least header row")
	}

	expectedHeader := []string{
		"Username",
		"Display Name",
		"Lab 1 (Auto Score)",
		"Lab 1 (Manual Score)",
		"Lab 2 (Auto Score)",
		"Lab 2 (Manual Score)",
	}

	header := records[0]
	if len(header) != len(expectedHeader) {
		t.Fatalf("expected %d columns, got %d", len(expectedHeader), len(header))
	}

	for i, expected := range expectedHeader {
		if header[i] != expected {
			t.Errorf("header[%d]: expected %q, got %q", i, expected, header[i])
		}
	}

	// Verify student rows
	if len(records) != 3 { // header + 2 students
		t.Fatalf("expected 3 rows (1 header + 2 students), got %d", len(records))
	}

	// Verify first student
	student1 := records[1]
	if student1[0] != "student1" {
		t.Errorf("expected username 'student1', got %q", student1[0])
	}
	if student1[1] != "John Doe" {
		t.Errorf("expected display name 'John Doe', got %q", student1[1])
	}
	if student1[2] != "90" {
		t.Errorf("expected auto score '90', got %q", student1[2])
	}
	if student1[3] != "45" {
		t.Errorf("expected manual score '45', got %q", student1[3])
	}

	// Verify second student
	student2 := records[2]
	if student2[0] != "student2" {
		t.Errorf("expected username 'student2', got %q", student2[0])
	}
	if student2[1] != "Jane Smith" {
		t.Errorf("expected display name 'Jane Smith', got %q", student2[1])
	}
}

func TestExportCSV_EmptyGradebook(t *testing.T) {
	mockService := &mockSubmissionService{
		gradebook: &models.Gradebook{
			LabCol:     []models.LabCol{},
			StudentRow: []models.StudentRow{},
		},
	}

	exportService := NewGradebookExportService(mockService)
	data, err := exportService.ExportCSV(context.Background(), "section1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	// Should have only header row
	if len(records) != 1 {
		t.Fatalf("expected 1 row (header only), got %d", len(records))
	}

	expectedHeader := []string{"Username", "Display Name"}
	if len(records[0]) != len(expectedHeader) {
		t.Fatalf("expected %d columns, got %d", len(expectedHeader), len(records[0]))
	}
}

func TestExportCSV_ServiceError(t *testing.T) {
	expectedErr := errors.New("database error")
	mockService := &mockSubmissionService{
		err: expectedErr,
	}

	exportService := NewGradebookExportService(mockService)
	_, err := exportService.ExportCSV(context.Background(), "section1")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestExportXLSX_Success(t *testing.T) {
	mockService := &mockSubmissionService{
		gradebook: createTestGradebook(),
	}

	exportService := NewGradebookExportService(mockService)
	data, err := exportService.ExportXLSX(context.Background(), "section1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected XLSX data, got empty bytes")
	}

	// Open XLSX to verify structure
	f, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("failed to open XLSX: %v", err)
	}
	defer f.Close()

	// Verify sheet exists
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		t.Fatal("expected at least one sheet")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		t.Fatalf("failed to get rows: %v", err)
	}

	// Verify header
	if len(rows) < 1 {
		t.Fatal("expected at least header row")
	}

	header := rows[0]
	expectedHeader := []string{
		"Username",
		"Display Name",
		"Lab 1 (Auto Score)",
		"Lab 1 (Manual Score)",
		"Lab 2 (Auto Score)",
		"Lab 2 (Manual Score)",
	}

	if len(header) != len(expectedHeader) {
		t.Fatalf("expected %d columns, got %d", len(expectedHeader), len(header))
	}

	for i, expected := range expectedHeader {
		if header[i] != expected {
			t.Errorf("header[%d]: expected %q, got %q", i, expected, header[i])
		}
	}

	// Verify student rows
	if len(rows) != 3 { // header + 2 students
		t.Fatalf("expected 3 rows (1 header + 2 students), got %d", len(rows))
	}

	// Verify first student
	student1 := rows[1]
	if student1[0] != "student1" {
		t.Errorf("expected username 'student1', got %q", student1[0])
	}
	if student1[1] != "John Doe" {
		t.Errorf("expected display name 'John Doe', got %q", student1[1])
	}
	if student1[2] != "90" {
		t.Errorf("expected auto score '90', got %q", student1[2])
	}
	if student1[3] != "45" {
		t.Errorf("expected manual score '45', got %q", student1[3])
	}
}

func TestExportXLSX_EmptyGradebook(t *testing.T) {
	mockService := &mockSubmissionService{
		gradebook: &models.Gradebook{
			LabCol:     []models.LabCol{},
			StudentRow: []models.StudentRow{},
		},
	}

	exportService := NewGradebookExportService(mockService)
	data, err := exportService.ExportXLSX(context.Background(), "section1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Open XLSX
	f, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("failed to open XLSX: %v", err)
	}
	defer f.Close()

	sheetName := f.GetSheetList()[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		t.Fatalf("failed to get rows: %v", err)
	}

	// Should have only header row
	if len(rows) != 1 {
		t.Fatalf("expected 1 row (header only), got %d", len(rows))
	}
}

func TestExportXLSX_ServiceError(t *testing.T) {
	expectedErr := errors.New("database error")
	mockService := &mockSubmissionService{
		err: expectedErr,
	}

	exportService := NewGradebookExportService(mockService)
	_, err := exportService.ExportXLSX(context.Background(), "section1")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestColumnName(t *testing.T) {
	tests := []struct {
		col      int
		expected string
	}{
		{1, "A"},
		{2, "B"},
		{26, "Z"},
		{27, "AA"},
		{28, "AB"},
		{52, "AZ"},
		{53, "BA"},
	}

	for _, test := range tests {
		result := columnName(test.col)
		if result != test.expected {
			t.Errorf("columnName(%d): expected %q, got %q", test.col, test.expected, result)
		}
	}
}

func TestCellName(t *testing.T) {
	tests := []struct {
		row      int
		col      int
		expected string
	}{
		{1, 1, "A1"},
		{1, 2, "B1"},
		{2, 1, "A2"},
		{10, 26, "Z10"},
		{5, 27, "AA5"},
	}

	for _, test := range tests {
		result := cellName(test.row, test.col)
		if result != test.expected {
			t.Errorf("cellName(%d, %d): expected %q, got %q", test.row, test.col, test.expected, result)
		}
	}
}
