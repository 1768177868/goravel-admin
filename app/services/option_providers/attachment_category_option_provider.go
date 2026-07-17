package option_providers

import (
	"context"

	"github.com/goravel/framework/contracts/http"
	"github.com/samber/lo"

	appfacades "goravel/app/facades"
	"goravel/app/models"
)

type AttachmentCategoryOptionProvider struct {
	ctx context.Context
}

func NewAttachmentCategoryOptionProvider(ctx context.Context) *AttachmentCategoryOptionProvider {
	return &AttachmentCategoryOptionProvider{ctx: ctx}
}

func (p *AttachmentCategoryOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	var categories []models.AttachmentCategory
	if err := appfacades.OrmQuery(ctx).Where("status", 1).Order("is_system desc, sort asc, id asc").Get(&categories); err != nil {
		return nil, err
	}
	options := lo.Map(categories, func(item models.AttachmentCategory, _ int) map[string]any {
		return map[string]any{
			"label":     item.Name,
			"value":     item.ID,
			"is_system": item.IsSystem,
		}
	})
	return map[string]any{"options": options}, nil
}
