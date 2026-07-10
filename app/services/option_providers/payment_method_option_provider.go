package option_providers

import (
	"context"
	appfacades "goravel/app/facades"

	"github.com/goravel/framework/contracts/http"
	"github.com/samber/lo"

	"goravel/app/models"
)

type PaymentMethodOptionProvider struct {
	ctx context.Context
}

func NewPaymentMethodOptionProvider(ctx context.Context) *PaymentMethodOptionProvider {
	return &PaymentMethodOptionProvider{
		ctx: ctx}
}

func (p *PaymentMethodOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	var paymentMethods []models.PaymentMethod

	// 只查询已启用的支付方式
	query := appfacades.OrmQuery(ctx).Model(&models.PaymentMethod{}).Where("is_active", true).Order("sort asc").Order("id asc")
	if err := query.Get(&paymentMethods); err != nil {
		return nil, err
	}

	options := lo.Map(paymentMethods, func(pm models.PaymentMethod, _ int) map[string]any {
		return map[string]any{
			"label": pm.Name,
			"value": pm.ID,
			// "value": cast.ToString(pm.ID),
		}
	})

	return map[string]any{
		"options": options,
	}, nil
}
