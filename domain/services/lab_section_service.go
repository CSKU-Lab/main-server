package services

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/google/uuid"
)

type LabSectionService interface {
	Create(ctx context.Context, req *requests.SetLabSection, userID string, sectionID string) error
	GetPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filterParams map[string]string) ([]models.LabSection, error)
	GetByLabAndSectionID(ctx context.Context, labID string, sectionID string) (*models.LabSection, error)
	Update(ctx context.Context, userID string, sectionID string, req *requests.UpdateLabSection) error
	UpdateStatus(ctx context.Context, userID string, sectionID string, labID string, req *requests.UpdateLabSectionStatus) error
	Delete(ctx context.Context, sectionID string, userID string, req *requests.DeleteLabSection) error
	Count(ctx context.Context, filterParams map[string]string) (int, error)
}

type labSectionService struct {
	labSectionRepo      repositories.LabSectionRepository
	sectionStudentRepo  repositories.SectionStudentRepository
	uowRepo             repositories.UoWRepository
	labRepo             repositories.LabRepository
	sectionRepo         repositories.SectionRepository
	allowedFilterFields map[string]bool
	allowedSortFields   map[string]bool
}

func NewLabSectionService(labSectionRepo repositories.LabSectionRepository, uowRepo repositories.UoWRepository, labRepo repositories.LabRepository, sectionRepo repositories.SectionRepository, sectionStudentRepo repositories.SectionStudentRepository) LabSectionService {
	return &labSectionService{
		labSectionRepo:     labSectionRepo,
		sectionStudentRepo: sectionStudentRepo,
		uowRepo:            uowRepo,
		labRepo:            labRepo,
		sectionRepo:        sectionRepo,
		allowedFilterFields: map[string]bool{
			"lab_id":        true,
			"section_id":    true,
			"status":        true,
			"l.display_name": true,
		},
		allowedSortFields: map[string]bool{
			"position": true,
		},
	}
}

func (ls *labSectionService) GetByLabAndSectionID(ctx context.Context, labID string, sectionID string) (*models.LabSection, error) {
	err := ls.rowExists(ctx, labID, sectionID)
	if err != nil {
		return nil, err
	}
	labSec, err := ls.labSectionRepo.GetByID(ctx, labID, sectionID)
	if err != nil {
		return nil, err
	}
	return labSec, nil
}

func (ls *labSectionService) rowExists(ctx context.Context, labID string, sectionID string) error {
	_, err := ls.labRepo.GetByID(ctx, labID)
	if err != nil {
		return err
	}
	_, err = ls.sectionRepo.GetByID(ctx, sectionID)
	if err != nil {
		return err
	}
	return nil
}

func (ls *labSectionService) mutationPermission(ctx context.Context, userID string, sectionID string) error {
	err := ls.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		user, err := u.User().GetByID(ctx, userID)
		if err != nil {
			return err
		}

		for _, role := range user.Roles {
			if role == string(models.ADMIN) {
				return nil
			}
		}

		sectionInstructors, err := u.SectionInstructor().Get(ctx, sectionID)
		if err != nil {
			return err
		}

		section, err := u.Section().GetRawByID(ctx, sectionID)
		if err != nil {
			return err
		}

		courseCreator, err := u.CourseCreator().GetCreators(ctx, section.CourseID)
		if err != nil {
			return err
		}

		for _, creator := range courseCreator {
			if creator.ID == userID {
				return nil
			}
		}

		for _, instructor := range sectionInstructors {
			if instructor.ID == userID {
				return nil
			}
		}

		return cserrors.New(&cserrors.Option{
			Message:    "No Permission",
			HttpStatus: http.StatusForbidden,
		})
	})
	if err != nil {
		return err
	}
	return nil
}

func (ls *labSectionService) rearrangeUpdatedIndex(ctx context.Context, sectionID string, newPos *int, labID string, oldPos int) error {
	if newPos == nil || *newPos == oldPos {
		return nil
	}

	err := ls.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		maxPos, err := ls.labSectionRepo.GetMaxPosition(ctx, sectionID, "")
		if err != nil {
			return err
		}

		if *newPos >= maxPos {
			*newPos = maxPos - 1
		}

		if *newPos < 1 {
			*newPos = 1
		}

		if *newPos < oldPos {
			err = ls.labSectionRepo.ShiftRangeDown(ctx, sectionID, *newPos, oldPos-1)
		} else {
			err = ls.labSectionRepo.ShiftRangeUp(ctx, sectionID, oldPos+1, *newPos)
		}
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (ls *labSectionService) rearrangeDeletedIndex(ctx context.Context, sectionID string, position int) error {
	err := ls.labSectionRepo.ShiftUpPositions(ctx, sectionID, position)
	if err != nil {
		return err
	}
	return nil
}

func (ls *labSectionService) Create(ctx context.Context, req *requests.SetLabSection, userID string, sectionID string) error {
	err := ls.mutationPermission(ctx, userID, sectionID)
	if err != nil {
		return err
	}

	for _, labID := range req.LabIDs {
		err := ls.rowExists(ctx, labID, sectionID)
		if err != nil {
			return err
		}
		labSection, err := ls.labSectionRepo.GetByID(ctx, labID, sectionID)
		if err != nil && labSection != nil {
			return err
		}
		if labSection != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusConflict,
				Message:    "Item already exists",
			})
		}

		// Get the max position and append at the end
		maxPos, err := ls.labSectionRepo.GetMaxPosition(ctx, sectionID, "")
		if err != nil {
			return err
		}

		position := maxPos

		ID, err := uuid.NewV7()
		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Cannot generate uuid",
			})
		}

		err = ls.labSectionRepo.Create(ctx, repositories.CreateLabSectionParams{
			LabID:     labID,
			SectionID: sectionID,
			Position:  position,
			ID:        ID.String(),
			Status:    "hidden",
			OpenedAt:  nil,
			ReadonlyAt:  nil,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (ls *labSectionService) GetPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filterParams map[string]string) ([]models.LabSection, error) {
	sanitizedSortBy, err := sanitize.SortBy(sortBy, ls.allowedSortFields)
	if err != nil {
		return nil, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid sort by field",
			})
	}

	sanitizedSortOrder, err := sanitize.SortOrder(sortOrder)
	if err != nil {
		return nil, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid sort order",
			})
	}

	sanitizedFilters, err := sanitize.Filters(filterParams, ls.allowedFilterFields)
	if err != nil {
		return nil, err
	}

	labSections, err := ls.labSectionRepo.GetPagination(ctx, page, limit, sanitizedSortBy, sanitizedSortOrder, sanitizedFilters)
	if err != nil {
		return nil, err
	}

	return labSections, nil
}

func (ls *labSectionService) Count(ctx context.Context, filterParams map[string]string) (int, error) {
	filters, err := sanitize.Filters(filterParams, ls.allowedFilterFields)
	if err != nil {
		return 0, err
	}

	return ls.labSectionRepo.Count(ctx, filters)
}

func (ls *labSectionService) Update(ctx context.Context, userID string, sectionID string, req *requests.UpdateLabSection) error {
	err := ls.rowExists(ctx, req.LabID, sectionID)
	if err != nil {
		return err
	}

	err = ls.mutationPermission(ctx, userID, sectionID)
	if err != nil {
		return err
	}

	labSection, err := ls.labSectionRepo.GetByID(ctx, req.LabID, sectionID)
	if err != nil {
		return err
	}

	err = ls.rearrangeUpdatedIndex(ctx, sectionID, &req.Position, req.LabID, labSection.Position)
	if err != nil {
		return err
	}

	err = ls.labSectionRepo.UpdateByID(ctx, req.LabID, sectionID, labSection.ID, req)
	if err != nil {
		return err
	}
	return nil
}

func (ls *labSectionService) UpdateStatus(ctx context.Context, userID string, sectionID string, labID string, req *requests.UpdateLabSectionStatus) error {
	err := ls.rowExists(ctx, labID, sectionID)
	if err != nil {
		return err
	}

	err = ls.mutationPermission(ctx, userID, sectionID)
	if err != nil {
		return err
	}

	labSection, err := ls.labSectionRepo.GetByID(ctx, labID, sectionID)
	if err != nil {
		return err
	}

	updateReq := &requests.UpdateLabSectionStatus{
		Status:     req.Status,
		OpenedAt:   req.OpenedAt,
		ReadonlyAt: req.ReadonlyAt,
		Force:      req.Force,
	}

	err = ls.applyLabSectionSchedule(updateReq, labSection)
	if err != nil {
		return err
	}

	return ls.labSectionRepo.UpdateStatusByID(ctx, labID, sectionID, labSection.ID, updateReq)
}

func (ls *labSectionService) applyLabSectionSchedule(req *requests.UpdateLabSectionStatus, current *models.LabSection) error {
	status := ""
	if req.Status != nil {
		status = *req.Status
	}
	if status == "" {
		status = current.Status
	}

	if req.Force != nil && *req.Force {
		if req.OpenedAt != nil || req.ReadonlyAt != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "opened_at and readonly_at must be empty when force is true",
			})
		}
		now := time.Now()
		switch status {
		case "open":
			req.OpenedAt = &now
			req.ReadonlyAt = nil
			req.ClearTimes = true
		default:
			req.OpenedAt = nil
			req.ReadonlyAt = nil
			req.ClearTimes = true
		}
		return nil
	}

	if status == "disabled" {
		if req.OpenedAt != nil || req.ReadonlyAt != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "opened_at and readonly_at are not allowed for disabled status",
			})
		}
		req.ClearTimes = true
		return nil
	}
	if status != "" && status != "hidden" && status != "open" && status != "readonly" {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "status must be hidden, open or readonly when using opened_at/readonly_at",
		})
	}
	if status == "readonly" && req.OpenedAt != nil {
		req.OpenedAt = nil
	}
	if status == "hidden" && (req.OpenedAt != nil || req.ReadonlyAt != nil) {
		req.OpenedAt = nil
		req.ReadonlyAt = nil
	}
	if req.OpenedAt == nil && req.ReadonlyAt == nil {
		switch status {
		case "open":
			if current.OpenedAt == nil {
				now := time.Now()
				req.OpenedAt = &now
			}
			req.ClearTimes = true
		case "readonly":
			if current.ReadonlyAt == nil {
				now := time.Now()
				req.ReadonlyAt = &now
			}
			req.ClearTimes = true
		case "hidden":
			req.ClearTimes = true
		}
		return nil
	}
	if status == "" {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "status is required when opened_at or readonly_at is provided",
		})
	}

	effectiveOpenedAt := req.OpenedAt
	if effectiveOpenedAt == nil {
		effectiveOpenedAt = current.OpenedAt
	}
	effectiveReadonlyAt := req.ReadonlyAt
	if effectiveReadonlyAt == nil {
		effectiveReadonlyAt = current.ReadonlyAt
	}

	if effectiveOpenedAt != nil && effectiveReadonlyAt != nil && effectiveOpenedAt.After(*effectiveReadonlyAt) {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "opened_at must be before readonly_at",
		})
	}

	now := time.Now()
	derivedStatus := deriveScheduledStatus(now, effectiveOpenedAt, effectiveReadonlyAt)
	req.Status = &derivedStatus
	return nil
}

func deriveScheduledStatus(now time.Time, openedAt *time.Time, closedAt *time.Time) string {
	if openedAt != nil && now.Before(*openedAt) {
		return "hidden"
	}
	if closedAt != nil && now.After(*closedAt) {
		return "readonly"
	}
	return "open"
}

func (ls *labSectionService) Delete(ctx context.Context, sectionID string, userID string, req *requests.DeleteLabSection) error {
	err := ls.mutationPermission(ctx, userID, sectionID)
	if err != nil {
		return err
	}

	// Collect labSections with their positions
	type labToDelete struct {
		labID    string
		position int
		id       string
	}
	var labsToDelete []labToDelete

	for _, labID := range req.LabIDs {
		err := ls.rowExists(ctx, labID, sectionID)
		if err != nil {
			return err
		}

		labSection, err := ls.labSectionRepo.GetByID(ctx, labID, sectionID)
		if err != nil {
			return err
		}

		labsToDelete = append(labsToDelete, labToDelete{
			labID:    labID,
			position: labSection.Position,
			id:       labSection.ID,
		})
	}

	// Sort by position descending to delete from highest first
	sort.Slice(labsToDelete, func(i, j int) bool {
		return labsToDelete[i].position > labsToDelete[j].position
	})

	for _, lab := range labsToDelete {
		err = ls.rearrangeDeletedIndex(ctx, sectionID, lab.position)
		if err != nil {
			return err
		}

		err = ls.labSectionRepo.DeleteByID(ctx, lab.id)
		if err != nil {
			return err
		}
	}

	return nil
}
