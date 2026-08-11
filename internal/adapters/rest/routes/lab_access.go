package routes

import (
	"context"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
)

// requireLabAccess enforces the student-facing lab schedule after section
// membership has been checked.  Status must be checked server-side because UI
// controls alone can be bypassed with direct API requests.
func requireLabAccess(ctx context.Context, labSectionService services.LabSectionService, labID, sectionID string, submission bool) error {
	labSection, err := labSectionService.GetByLabAndSectionID(ctx, labID, sectionID)
	if err != nil {
		return err
	}

	switch labSection.EffectiveStatus() {
	case "disabled":
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Code:       cserrors.Forbidden,
			Message:    "This lab is disabled",
		})
	case "readonly":
		if submission {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusForbidden,
				Code:       cserrors.Forbidden,
				Message:    "This lab is readonly and no longer accepts submissions",
			})
		}
	}

	return nil
}

// requireMaterialView conceals materials belonging to disabled labs. This
// intentionally returns 404 so a direct student URL renders the normal
// not-found state instead of an authorization error.
func requireMaterialView(labSection *models.LabSection) error {
	if labSection.EffectiveStatus() != "disabled" {
		return nil
	}

	return cserrors.New(&cserrors.Option{
		HttpStatus: http.StatusNotFound,
		Message:    "Material not found",
	})
}
