package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/helpers"
	apirequests "goravel/app/http/requests/api"
	"goravel/app/http/response"
	"goravel/app/http/trans"
	"goravel/app/services"
	"goravel/app/utils"
)

type OrderController struct {
	orderService services.OrderService
}

func NewOrderController() *OrderController {
	return &OrderController{
		orderService: services.NewOrderService(),
	}
}

// SearchMyOrders GET：当前登录用户检索自己的订单。开启 ELASTICSEARCH_ENABLED 时走 ES（可多字段含商品名）；否则走分表数据库（关键词仅订单号、备注；无时间参数时默认近 3 个月）。
func (c *OrderController) SearchMyOrders(ctx http.Context) http.Response {
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

	timeRange, errField, errMsgKey := helpers.ParseOrderSearchCreatedRange(req.CreatedFrom, req.CreatedTo)
	if errField != "" {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", map[string]map[string]string{
			errField: {"time": trans.Get(ctx, errMsgKey)},
		})
	}

	searchCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	list, total, err := c.orderService.SearchMyOrdersForUser(searchCtx, userID, req.Q, page, pageSize, timeRange)
	if err != nil {
		if timeRangeErr, ok := err.(*utils.TimeRangeError); ok {
			message := trans.Get(ctx, timeRangeErr.Key)
			if timeRangeErr.Params != nil {
				for key, value := range timeRangeErr.Params {
					placeholder := fmt.Sprintf("{%s}", key)
					message = strings.ReplaceAll(message, placeholder, fmt.Sprintf("%v", value))
				}
			}
			return response.Error(ctx, http.StatusBadRequest, message)
		}
		if facades.Config().GetBool("elasticsearch.enabled", false) {
			facades.Log().Errorf("order ES search: %v", err)
			return response.Error(ctx, http.StatusBadGateway, "elasticsearch_search_failed")
		}
		facades.Log().Errorf("order DB search: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Paginate(ctx, "success", list, total, page, pageSize)
}
