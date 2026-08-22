package api

import (
	appfacades "goravel/app/facades"
	"goravel/app/http/response"
	"goravel/app/models"

	"github.com/goravel/framework/contracts/http"
)

type PublicConfigController struct{}

func NewPublicConfigController() *PublicConfigController {
	return &PublicConfigController{}
}

// CustomerService 公开返回客服配置（仅 customer_service 分组，供 C 端展示链接）
func (r *PublicConfigController) CustomerService(ctx http.Context) http.Response {
	var configs []models.Config
	_ = appfacades.OrmQuery(ctx).Where("group", "customer_service").Order("sort asc, id asc").Get(&configs)

	data := make(map[string]string, len(configs))
	for _, item := range configs {
		data[item.Key] = item.Value
	}

	if _, ok := data["cs_enabled"]; !ok {
		data["cs_enabled"] = "0"
	}

	return response.Success(ctx, http.Json{
		"configs": data,
	})
}
