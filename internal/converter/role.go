package converter

import "github.com/CSKU-Lab/main-server/domain/models"

func ToRoleSlice(roles []string) ([]models.Role, error) {
	result := make([]models.Role, len(roles))
	for i, role := range roles {
		role := models.Role(role)
		if err := role.Validate(); err != nil {
			return nil, err
		}
		result[i] = role
	}
	return result, nil
}
