package api

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	esorders "goravel/app/elasticsearch/orders"
	"goravel/app/http/helpers"
	apirequests "goravel/app/http/requests/api"
	"goravel/app/http/response"
	"goravel/app/http/trans"
)

type OrderSearchController struct{}

func NewOrderSearchController() *OrderSearchController {
	return &OrderSearchController{}
}

// SearchMyOrders GET ：当前登录用户在 ES 中按关键词检索自己的订单（需 ELASTICSEARCH_ENABLED）。
func (c *OrderSearchController) SearchMyOrders(ctx http.Context) http.Response {
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

	page, pageSize := helpers.ValidatePaginationEx(req.Page, req.PageSize, helpers.PaginationLimits{})

	gte, lte, errField, errMsgKey := helpers.ParseOrderCreatedAtRangeForES(req.CreatedFrom, req.CreatedTo)
	if errField != "" {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", map[string]map[string]string{
			errField: {"time": trans.Get(ctx, errMsgKey)},
		})
	}

	searchCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	total, list, err := esorders.SearchMyOrders(searchCtx, userID, req.Q, page, pageSize, gte, lte)
	if err != nil {
		facades.Log().Errorf("order ES search: %v", err)
		return response.Error(ctx, http.StatusBadGateway, "elasticsearch_search_failed")
	}

	// 与后台 Paginate 一致：data.list / total / page / page_size / total_pages
	return response.Paginate(ctx, "success", list, total, page, pageSize)
}
