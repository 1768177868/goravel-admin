package admin

import (
	"github.com/goravel/framework/contracts/http"
)

type DashboardController struct{}

func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

// GetCount 获取统计数据
func (r *DashboardController) GetCount(ctx http.Context) http.Response {
	// 暂时返回空数据
	emptyData := []interface{}{}
	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": "get_success",
		"data":    emptyData,
	})
}

// GetUserAccessSource 获取用户来源数据
func (r *DashboardController) GetUserAccessSource(ctx http.Context) http.Response {
	// 暂时返回空数据
	emptyData := []interface{}{}
	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": "get_success",
		"data":    emptyData,
	})
}

// GetWeeklyUserActivity 获取每周用户活跃量
func (r *DashboardController) GetWeeklyUserActivity(ctx http.Context) http.Response {
	// 暂时返回空数据
	emptyData := []interface{}{}
	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": "get_success",
		"data":    emptyData,
	})
}

// GetMonthlySales 获取每月销售额
func (r *DashboardController) GetMonthlySales(ctx http.Context) http.Response {
	// 暂时返回空数据
	emptyData := []interface{}{}
	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": "get_success",
		"data":    emptyData,
	})
}

