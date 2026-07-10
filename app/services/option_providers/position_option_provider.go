package option_providers

import (
	"context"
	appfacades "goravel/app/facades"

	"github.com/goravel/framework/contracts/http"
	"github.com/samber/lo"

	"goravel/app/models"
)

type PositionOptionProvider struct {
	ctx context.Context
}

func NewPositionOptionProvider(ctx context.Context) *PositionOptionProvider {
	return &PositionOptionProvider{
		ctx: ctx}
}

func (p *PositionOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	var positions []models.Position
	if err := appfacades.OrmQuery(ctx).Where("status", 1).Order("sort asc, id asc").Get(&positions); err != nil {
		return nil, err
	}
	options := lo.Map(positions, func(pos models.Position, _ int) map[string]any {
		return map[string]any{
			"label": pos.Name,
			"value": pos.ID,
		}
	})
	return map[string]any{
		"options": options,
	}, nil
}
