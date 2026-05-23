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
	if req.ReadonlyAt != nil {
		t.Fatal("did not expect readonly_at to be set when opening without dates")
	}
}

func TestApplyLabSectionScheduleSetsReadonlyAtWhenClosing(t *testing.T) {
	service := &labSectionService{}
	status := "readonly"
	req := &requests.UpdateLabSectionStatus{Status: &status}

	err := service.applyLabSectionSchedule(req, &models.LabSection{Status: "open"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.ReadonlyAt == nil {
		t.Fatal("expected readonly_at to be set when closing without dates")
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
		ReadonlyAt: &closedAt,
	}

	err := service.applyLabSectionSchedule(req, &models.LabSection{Status: "open"})
	if err == nil {
		t.Fatal("expected error for opened_at after readonly_at")
	}
	var csErr *cserrors.Error
	if !errors.As(err, &csErr) {
		t.Fatalf("expected cserrors.Error, got %T", err)
	}
}

func TestApplyLabSectionForceOpenSetsOpenedAtNow(t *testing.T) {
	service := &labSectionService{}
	status := "open"
	force := true
	req := &requests.UpdateLabSectionStatus{Status: &status, Force: &force}

	for _, fromStatus := range []string{"hidden", "readonly", "disabled"} {
		req.OpenedAt = nil
		req.ReadonlyAt = nil
		err := service.applyLabSectionSchedule(req, &models.LabSection{Status: fromStatus})
		if err != nil {
			t.Fatalf("force open from %s: unexpected error: %v", fromStatus, err)
		}
		if req.OpenedAt == nil {
			t.Fatalf("force open from %s: expected opened_at to be set", fromStatus)
		}
		if req.ReadonlyAt != nil {
			t.Fatalf("force open from %s: expected readonly_at to be nil", fromStatus)
		}
	}
}

func TestApplyLabSectionForceNonOpenClearsTimes(t *testing.T) {
	service := &labSectionService{}
	force := true
	for _, toStatus := range []string{"hidden", "readonly", "disabled"} {
		status := toStatus
		req := &requests.UpdateLabSectionStatus{Status: &status, Force: &force}
		err := service.applyLabSectionSchedule(req, &models.LabSection{Status: "open"})
		if err != nil {
			t.Fatalf("force %s: unexpected error: %v", toStatus, err)
		}
		if req.OpenedAt != nil {
			t.Fatalf("force %s: expected opened_at to be nil", toStatus)
		}
		if req.ReadonlyAt != nil {
			t.Fatalf("force %s: expected readonly_at to be nil", toStatus)
		}
	}
}

func TestApplyLabSectionForceRejectsTimestamps(t *testing.T) {
	service := &labSectionService{}
	status := "open"
	force := true
	openedAt := time.Now()
	req := &requests.UpdateLabSectionStatus{Status: &status, Force: &force, OpenedAt: &openedAt}

	err := service.applyLabSectionSchedule(req, &models.LabSection{Status: "hidden"})
	if err == nil {
		t.Fatal("expected error when force=true with timestamps")
	}
	var csErr *cserrors.Error
	if !errors.As(err, &csErr) {
		t.Fatalf("expected cserrors.Error, got %T", err)
	}
}
