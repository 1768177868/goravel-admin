package option_providers

import (
	"github.com/samber/lo"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

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
			"value": cast.ToString(role.ID),
		}
	})

	return map[string]any{
		"options": options,
	}, nil
}

