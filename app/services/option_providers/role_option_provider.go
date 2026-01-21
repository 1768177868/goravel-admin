package option_providers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/samber/lo"

	"goravel/app/models"
)

type RoleOptionProvider struct{}

func NewRoleOptionProvider() *RoleOptionProvider {
	return &RoleOptionProvider{}
}

func (p *RoleOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	var roles []models.Role
	if err := facades.Orm().Query().Where("status", 1).Order("id asc").Get(&roles); err != nil {
		return nil, err
	}

	options := lo.Map(roles, func(role models.Role, _ int) map[string]any {
		return map[string]any{
			"label": role.Name,
			// "value": cast.ToString(role.ID),
			"value": role.ID,
		}
	})

	return map[string]any{
		"options": options,
	}, nil
}
