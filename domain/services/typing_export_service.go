package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/xuri/excelize/v2"
)

type TypingExportService interface {
	ExportXLSX(ctx context.Context, sectionID string) ([]byte, error)
}

type typingExportService struct {
	submissionService  SubmissionService
	labSectionRepo     repositories.LabSectionRepository
	labMatRepo         repositories.LabMaterialRepository
	typingMaterialRepo repositories.TypingMaterialRepository
}

func NewTypingExportService(
	submissionService SubmissionService,
	labSectionRepo repositories.LabSectionRepository,
	labMatRepo repositories.LabMaterialRepository,
	typingMaterialRepo repositories.TypingMaterialRepository,
) TypingExportService {
	return &typingExportService{
		submissionService:  submissionService,
		labSectionRepo:     labSectionRepo,
		labMatRepo:         labMatRepo,
		typingMaterialRepo: typingMaterialRepo,
	}
}

type typingExportRow struct {
	labName      string
	materialName string
	studentName  string
	studentID    string
	wpm          any
	accuracy     any
	score        any
	mode         string
	submittedAt  string
}

func (s *typingExportService) ExportXLSX(ctx context.Context, sectionID string) ([]byte, error) {
	labs, err := s.labSectionRepo.GetBySectionID(ctx, sectionID)
	if err != nil {
		return nil, err
	}

	var rows []typingExportRow

	for _, lab := range labs {
		labMaterials, err := s.labMatRepo.GetByLabID(ctx, lab.ID)
		if err != nil {
			return nil, err
		}

		for _, lm := range labMaterials {
			if lm.MaterialData == nil || lm.MaterialData.Type != "typing" {
				continue
			}

			matID := lm.MaterialID
			materialName := lm.MaterialData.Name

			typingMat, err := s.typingMaterialRepo.GetByID(ctx, matID)
			if err != nil {
				continue
			}
			typingType := typingMat.TypingType
			if typingType == "" {
				typingType = "practice"
			}

			submissions, err := s.submissionService.GetSectionLabMaterialSubmissions(ctx, sectionID, lab.ID, matID)
			if err != nil {
				continue
			}

			for _, sub := range submissions {
				row := typingExportRow{
					labName:      lab.DisplayName,
					materialName: materialName,
					studentName:  sub.Student.DisplayName,
					studentID:    sub.Student.Username,
					mode:         typingType,
				}

				if sub.StudentSubmission == nil || sub.Status == models.NOT_SUBMITTED {
					row.wpm = "-"
					row.accuracy = "-"
					row.score = "-"
					row.submittedAt = "-"
				} else {
					var typingSub models.TypingSubmission
					if sub.Payload != nil {
						b, _ := json.Marshal(sub.Payload)
						json.Unmarshal(b, &typingSub)
					}
					row.wpm = math.Round(typingSub.AdjustedWPM)
					row.accuracy = math.Round((100-typingSub.ErrorRate)*10) / 10
					row.score = sub.AutoScore
					row.submittedAt = sub.CreatedAt.In(time.FixedZone("UTC+7", 7*60*60)).Format("2006-01-02 15:04:05")
				}

				rows = append(rows, row)
			}
		}
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Typing Submissions"
	index, err := f.NewSheet(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create header style: %w", err)
	}

	headers := []string{"Lab", "Material", "Student Name", "Student ID", "WPM", "Accuracy (%)", "Score", "Mode", "Submitted At"}
	for i, h := range headers {
		cell := cellName(1, i+1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	for i, row := range rows {
		r := i + 2
		f.SetCellValue(sheet, cellName(r, 1), row.labName)
		f.SetCellValue(sheet, cellName(r, 2), row.materialName)
		f.SetCellValue(sheet, cellName(r, 3), row.studentName)
		f.SetCellValue(sheet, cellName(r, 4), row.studentID)
		f.SetCellValue(sheet, cellName(r, 5), row.wpm)
		f.SetCellValue(sheet, cellName(r, 6), row.accuracy)
		f.SetCellValue(sheet, cellName(r, 7), row.score)
		f.SetCellValue(sheet, cellName(r, 8), row.mode)
		f.SetCellValue(sheet, cellName(r, 9), row.submittedAt)
	}

	f.SetColWidth(sheet, "A", "A", 20)
	f.SetColWidth(sheet, "B", "B", 25)
	f.SetColWidth(sheet, "C", "C", 25)
	f.SetColWidth(sheet, "D", "D", 15)
	f.SetColWidth(sheet, "E", "E", 10)
	f.SetColWidth(sheet, "F", "F", 14)
	f.SetColWidth(sheet, "G", "G", 10)
	f.SetColWidth(sheet, "H", "H", 12)
	f.SetColWidth(sheet, "I", "I", 22)

	if sheet != "Sheet1" {
		f.DeleteSheet("Sheet1")
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write XLSX: %w", err)
	}

	return buf.Bytes(), nil
}
