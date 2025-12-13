package option_providers

import (
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

	var options []map[string]any
	for _, role := range roles {
		options = append(options, map[string]any{
			"label": role.Name,
			"value": cast.ToString(role.ID),
		})
	}

	return map[string]any{
		"options": options,
	}, nil
}

