package option_providers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/samber/lo"

	"goravel/app/models"
)

type PaymentMethodOptionProvider struct{}

func NewPaymentMethodOptionProvider() *PaymentMethodOptionProvider {
	return &PaymentMethodOptionProvider{}
}

func (p *PaymentMethodOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	var paymentMethods []models.PaymentMethod

	// 只查询已启用的支付方式
	query := facades.Orm().Query().Model(&models.PaymentMethod{}).Where("is_active", true).Order("sort asc").Order("id asc")
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
