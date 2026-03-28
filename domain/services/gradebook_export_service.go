package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// GradebookExportService handles exporting gradebooks to different formats
type GradebookExportService interface {
	ExportCSV(ctx context.Context, sectionID string) ([]byte, error)
	ExportXLSX(ctx context.Context, sectionID string) ([]byte, error)
}

type gradebookExportService struct {
	submissionService SubmissionService
}

// NewGradebookExportService creates a new gradebook export service
func NewGradebookExportService(submissionService SubmissionService) GradebookExportService {
	return &gradebookExportService{
		submissionService: submissionService,
	}
}

// ExportCSV generates a CSV file from gradebook data
func (s *gradebookExportService) ExportCSV(ctx context.Context, sectionID string) ([]byte, error) {
	gradebook, err := s.submissionService.GetGradebookBySectionID(ctx, sectionID)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Build header row
	header := []string{"Username", "Display Name"}
	for _, lab := range gradebook.LabCol {
		header = append(header, fmt.Sprintf("%s (Auto Score)", lab.LabName))
		header = append(header, fmt.Sprintf("%s (Manual Score)", lab.LabName))
	}

	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write student rows
	for _, student := range gradebook.StudentRow {
		row := []string{student.Username, student.DisplayName}

		for _, lab := range gradebook.LabCol {
			labScore := student.LabScores[lab.LabID]
			row = append(row, strconv.Itoa(labScore.AutoScore))
			row = append(row, strconv.Itoa(labScore.ManualScore))
		}

		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportXLSX generates an XLSX file from gradebook data
func (s *gradebookExportService) ExportXLSX(ctx context.Context, sectionID string) ([]byte, error) {
	gradebook, err := s.submissionService.GetGradebookBySectionID(ctx, sectionID)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			// Log error but don't fail the export
		}
	}()

	sheetName := "Gradebook"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)

	// Create bold style for headers
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create header style: %w", err)
	}

	// Write header row
	col := 1
	f.SetCellValue(sheetName, cellName(1, col), "Username")
	f.SetCellStyle(sheetName, cellName(1, col), cellName(1, col), headerStyle)
	col++

	f.SetCellValue(sheetName, cellName(1, col), "Display Name")
	f.SetCellStyle(sheetName, cellName(1, col), cellName(1, col), headerStyle)
	col++

	for _, lab := range gradebook.LabCol {
		autoScoreHeader := fmt.Sprintf("%s (Auto Score)", lab.LabName)
		f.SetCellValue(sheetName, cellName(1, col), autoScoreHeader)
		f.SetCellStyle(sheetName, cellName(1, col), cellName(1, col), headerStyle)
		col++

		manualScoreHeader := fmt.Sprintf("%s (Manual Score)", lab.LabName)
		f.SetCellValue(sheetName, cellName(1, col), manualScoreHeader)
		f.SetCellStyle(sheetName, cellName(1, col), cellName(1, col), headerStyle)
		col++
	}

	// Write student rows
	row := 2
	for _, student := range gradebook.StudentRow {
		col := 1
		f.SetCellValue(sheetName, cellName(row, col), student.Username)
		col++

		f.SetCellValue(sheetName, cellName(row, col), student.DisplayName)
		col++

		for _, lab := range gradebook.LabCol {
			labScore := student.LabScores[lab.LabID]
			f.SetCellValue(sheetName, cellName(row, col), labScore.AutoScore)
			col++

			f.SetCellValue(sheetName, cellName(row, col), labScore.ManualScore)
			col++
		}

		row++
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 15) // Username
	f.SetColWidth(sheetName, "B", "B", 20) // Display Name
	// Set width for lab columns
	if len(gradebook.LabCol) > 0 {
		lastCol := columnName(2 + len(gradebook.LabCol)*2)
		f.SetColWidth(sheetName, "C", lastCol, 18)
	}

	// Delete default Sheet1 if it exists and is not our sheet
	if sheetName != "Sheet1" {
		f.DeleteSheet("Sheet1")
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write XLSX file: %w", err)
	}

	return buf.Bytes(), nil
}

// cellName converts row and column numbers to Excel cell reference (e.g., "A1", "B2")
func cellName(row, col int) string {
	return fmt.Sprintf("%s%d", columnName(col), row)
}

// columnName converts column number to Excel column letter (1=A, 2=B, ..., 27=AA)
func columnName(col int) string {
	name := ""
	for col > 0 {
		col--
		name = string(rune('A'+(col%26))) + name
		col /= 26
	}
	return name
}
