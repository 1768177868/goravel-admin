package option_providers

import (
	"context"
	appfacades "goravel/app/facades"

	"github.com/goravel/framework/contracts/http"
	"github.com/samber/lo"

	"goravel/app/models"
)

type RoleOptionProvider struct {
	ctx context.Context
}

func NewRoleOptionProvider(ctx context.Context) *RoleOptionProvider {
	return &RoleOptionProvider{
		ctx: ctx}
}

func (p *RoleOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	var roles []models.Role
	if err := appfacades.OrmQuery(ctx).Where("status", 1).Order("id asc").Get(&roles); err != nil {
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
