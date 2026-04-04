package option_providers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/samber/lo"

	"goravel/app/models"
)

type PositionOptionProvider struct{}

func NewPositionOptionProvider() *PositionOptionProvider {
	return &PositionOptionProvider{}
}

func (p *PositionOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	var positions []models.Position
	if err := facades.Orm().Query().Where("status", 1).Order("sort asc, id asc").Get(&positions); err != nil {
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
