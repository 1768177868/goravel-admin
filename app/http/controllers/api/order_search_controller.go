package api

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	apirequests "goravel/app/http/requests/api"
	"goravel/app/http/response"
	"goravel/app/http/helpers"
	esorders "goravel/app/elasticsearch/orders"
)

type OrderSearchController struct{}

func NewOrderSearchController() *OrderSearchController {
	return &OrderSearchController{}
}

// DemoSearchMyOrders GET demo：当前登录用户在 ES 中按关键词检索自己的订单（需 ELASTICSEARCH_ENABLED）。
func (c *OrderSearchController) DemoSearchMyOrders(ctx http.Context) http.Response {
	if !facades.Config().GetBool("elasticsearch.enabled", false) {
		return response.Error(ctx, http.StatusServiceUnavailable, "elasticsearch_unavailable")
	}

	var req apirequests.OrderSearch
	errors, err := ctx.Request().ValidateRequest(&req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil || userID == 0 {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	searchCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	total, list, err := esorders.SearchMyOrders(searchCtx, userID, req.Q, page, pageSize)
	if err != nil {
		facades.Log().Errorf("demo order ES search: %v", err)
		return response.Error(ctx, http.StatusBadGateway, "elasticsearch_search_failed")
	}

	return response.Success(ctx, http.Json{
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"list":       list,
	})
}
