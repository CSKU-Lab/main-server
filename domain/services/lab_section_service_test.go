package services

import (
	"errors"
	"testing"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

func TestApplyLabSectionScheduleRejectsDatesForReadonly(t *testing.T) {
	service := &labSectionService{}
	status := "readonly"
	openedAt := time.Now().Add(1 * time.Hour)
	req := &requests.UpdateLabSectionStatus{
		Status:   &status,
		OpenedAt: &openedAt,
	}

	err := service.applyLabSectionSchedule(req, &models.LabSection{Status: "readonly"})
	if err == nil {
		t.Fatal("expected error for readonly status with dates")
	}
	var csErr *cserrors.Error
	if !errors.As(err, &csErr) {
		t.Fatalf("expected cserrors.Error, got %T", err)
	}
}

func TestApplyLabSectionScheduleSetsOpenedAtWhenOpening(t *testing.T) {
	service := &labSectionService{}
	status := "open"
	req := &requests.UpdateLabSectionStatus{Status: &status}

	err := service.applyLabSectionSchedule(req, &models.LabSection{Status: "hidden"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.OpenedAt == nil {
		t.Fatal("expected opened_at to be set when opening without dates")
	}
	if req.ClosedAt != nil {
		t.Fatal("did not expect closed_at to be set when opening without dates")
	}
}

func TestApplyLabSectionScheduleSetsClosedAtWhenClosing(t *testing.T) {
	service := &labSectionService{}
	status := "readonly"
	req := &requests.UpdateLabSectionStatus{Status: &status}

	err := service.applyLabSectionSchedule(req, &models.LabSection{Status: "open"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.ClosedAt == nil {
		t.Fatal("expected closed_at to be set when closing without dates")
	}
	if req.OpenedAt != nil {
		t.Fatal("did not expect opened_at to be set when closing without dates")
	}
}

func TestApplyLabSectionScheduleDerivesStatusFromDates(t *testing.T) {
	service := &labSectionService{}
	status := "open"
	openedAt := time.Now().Add(1 * time.Hour)
	req := &requests.UpdateLabSectionStatus{
		Status:   &status,
		OpenedAt: &openedAt,
	}

	err := service.applyLabSectionSchedule(req, &models.LabSection{Status: "open"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Status == nil || *req.Status != "hidden" {
		t.Fatalf("expected derived status hidden, got %v", req.Status)
	}
}

func TestApplyLabSectionScheduleRejectsInvalidRange(t *testing.T) {
	service := &labSectionService{}
	status := "open"
	openedAt := time.Now().Add(2 * time.Hour)
	closedAt := time.Now().Add(1 * time.Hour)
	req := &requests.UpdateLabSectionStatus{
		Status:   &status,
		OpenedAt: &openedAt,
		ClosedAt: &closedAt,
	}

	err := service.applyLabSectionSchedule(req, &models.LabSection{Status: "open"})
	if err == nil {
		t.Fatal("expected error for opened_at after closed_at")
	}
	var csErr *cserrors.Error
	if !errors.As(err, &csErr) {
		t.Fatalf("expected cserrors.Error, got %T", err)
	}
}
