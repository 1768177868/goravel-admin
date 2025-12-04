package admin

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type DashboardController struct{}

func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

// GetCount 获取统计数据
func (r *DashboardController) GetCount(ctx http.Context) http.Response {
	// 暂时返回空数据
	emptyData := []any{}
	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": "get_success",
		"data":    emptyData,
	})
}

// GetUserAccessSource 获取用户来源数据
func (r *DashboardController) GetUserAccessSource(ctx http.Context) http.Response {
	// 暂时返回空数据
	emptyData := []any{}
	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": "get_success",
		"data":    emptyData,
	})
}

// GetWeeklyUserActivity 获取每周用户活跃量
func (r *DashboardController) GetWeeklyUserActivity(ctx http.Context) http.Response {
	// 暂时返回空数据
	emptyData := []any{}
	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": "get_success",
		"data":    emptyData,
	})
}

// GetMonthlySales 获取每月销售额
func (r *DashboardController) GetMonthlySales(ctx http.Context) http.Response {
	// 暂时返回空数据
	emptyData := []any{}
	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": "get_success",
		"data":    emptyData,
	})
}

// StreamDashboardData SSE 实时推送 Dashboard 数据
// 定期推送所有 Dashboard 统计数据，包括计数、用户来源、用户活跃度、销售额等
func (r *DashboardController) StreamDashboardData(ctx http.Context) http.Response {
	// 获取推送间隔（秒），默认 5 秒
	interval := 5
	if intervalStr := ctx.Request().Query("interval", ""); intervalStr != "" {
		if parsed, err := time.ParseDuration(intervalStr + "s"); err == nil {
			interval = int(parsed.Seconds())
			if interval < 2 {
				interval = 2
			}
			if interval > 60 {
				interval = 60
			}
		}
	}

	// 设置 SSE 响应头
	writer := ctx.Response().Writer()
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲

	// 发送初始连接消息
	initMsg := map[string]any{
		"type":     "connected",
		"message":  "SSE连接已建立，开始推送 Dashboard 数据",
		"interval": interval,
	}
	initData, _ := json.Marshal(initMsg)
	fmt.Fprintf(writer, "data: %s\n\n", string(initData))
	if flusher, ok := writer.(nethttp.Flusher); ok {
		flusher.Flush()
	}

	// 创建 ticker，定期推送数据
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// 检测客户端断开连接
	clientGone := ctx.Request().Origin().Context().Done()

	for {
		select {
		case <-clientGone:
			// 客户端断开连接
			return nil
		case <-ticker.C:
			// 收集所有 Dashboard 数据
			dashboardData := r.collectDashboardData(ctx)

			// 构造 SSE 消息
			message := map[string]any{
				"type":      "dashboard_data",
				"data":      dashboardData,
				"timestamp": time.Now().Format(time.RFC3339),
			}

			messageData, err := json.Marshal(message)
			if err != nil {
				// 记录错误但继续推送
				facades.Log().Errorf("Dashboard SSE: failed to marshal data: %v", err)
				continue
			}

			// 发送 SSE 消息
			fmt.Fprintf(writer, "data: %s\n\n", string(messageData))

			// 刷新缓冲区
			if flusher, ok := writer.(nethttp.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

// collectDashboardData 收集 Dashboard 数据
func (r *DashboardController) collectDashboardData(ctx http.Context) map[string]any {
	data := make(map[string]any)

	// 1. 获取统计数据（管理员、角色、权限等）
	countData := r.getCountData()
	data["count"] = countData

	// 2. 获取用户访问来源数据
	accessSourceData := r.getUserAccessSourceData()
	data["user_access_source"] = accessSourceData

	// 3. 获取每周用户活跃量
	weeklyActivityData := r.getWeeklyUserActivityData()
	data["weekly_user_activity"] = weeklyActivityData

	// 4. 获取每月销售额
	monthlySalesData := r.getMonthlySalesData()
	data["monthly_sales"] = monthlySalesData

	// 5. 获取在线用户数
	onlineUserCount := r.getOnlineUserCount()
	data["online_user_count"] = onlineUserCount

	return data
}

// getCountData 获取统计数据
func (r *DashboardController) getCountData() map[string]any {
	// 统计各种数据
	adminCount, _ := facades.Orm().Query().Model(&models.Admin{}).Count()
	roleCount, _ := facades.Orm().Query().Model(&models.Role{}).Count()
	permissionCount, _ := facades.Orm().Query().Model(&models.Permission{}).Count()
	menuCount, _ := facades.Orm().Query().Model(&models.Menu{}).Count()
	departmentCount, _ := facades.Orm().Query().Model(&models.Department{}).Count()
	dictionaryCount, _ := facades.Orm().Query().Model(&models.Dictionary{}).Count()
	configCount, _ := facades.Orm().Query().Model(&models.Config{}).Count()

	return map[string]any{
		"admins":       adminCount,
		"roles":        roleCount,
		"permissions":  permissionCount,
		"menus":        menuCount,
		"departments":  departmentCount,
		"dictionaries": dictionaryCount,
		"configs":      configCount,
	}
}

// getUserAccessSourceData 获取用户访问来源数据
func (r *DashboardController) getUserAccessSourceData() []map[string]any {
	// 这里可以根据实际业务逻辑查询用户访问来源
	// 例如：根据登录日志统计不同来源的用户数
	// 暂时返回示例数据
	return []map[string]any{
		{"source": "web", "count": 0},
		{"source": "mobile", "count": 0},
		{"source": "api", "count": 0},
	}
}

// getWeeklyUserActivityData 获取每周用户活跃量
func (r *DashboardController) getWeeklyUserActivityData() []map[string]any {
	// 这里可以根据实际业务逻辑查询每周用户活跃量
	// 例如：统计最近7天每天的用户活跃数
	// 暂时返回示例数据
	now := time.Now()
	weeklyData := make([]map[string]any, 7)
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		weeklyData[6-i] = map[string]any{
			"date":  date.Format("2006-01-02"),
			"count": 0,
		}
	}
	return weeklyData
}

// getMonthlySalesData 获取每月销售额
func (r *DashboardController) getMonthlySalesData() []map[string]any {
	// 这里可以根据实际业务逻辑查询每月销售额
	// 例如：统计最近12个月每月的销售额
	// 暂时返回示例数据
	now := time.Now()
	monthlyData := make([]map[string]any, 12)
	for i := 11; i >= 0; i-- {
		date := now.AddDate(0, -i, 0)
		monthlyData[11-i] = map[string]any{
			"month": date.Format("2006-01"),
			"sales": 0,
		}
	}
	return monthlyData
}

// getOnlineUserCount 获取在线用户数
func (r *DashboardController) getOnlineUserCount() int64 {
	// 统计最近15分钟内有活动的用户（在线用户）
	onlineThreshold := time.Now().Add(-15 * time.Minute)
	count, _ := facades.Orm().Query().Model(&models.PersonalAccessToken{}).
		Where("tokenable_type", "admin").
		Where("last_used_at IS NOT NULL").
		Where("last_used_at >= ?", onlineThreshold).
		Count()
	return count
}
