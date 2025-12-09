package admin

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"strings"
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
	countData := r.getCountData()

	// 获取今日访问量（今日登录日志数）
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	todayVisits, _ := facades.Orm().Query().Model(&models.LoginLog{}).
		Where("created_at >= ?", todayStart).
		Where("created_at <= ?", todayEnd).
		Where("status", 1).
		Count()

	// 获取在线用户数
	onlineUserCount := r.getOnlineUserCount()

	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": "get_success",
		"data": map[string]any{
			"admin_count":  countData["admins"],
			"role_count":   countData["roles"],
			"menu_count":   countData["menus"],
			"today_visits": todayVisits,
			"online_users": onlineUserCount,
		},
	})
}

// GetUserAccessSource 获取用户访问来源数据（根据 UserAgent 判断设备类型）
func (r *DashboardController) GetUserAccessSource(ctx http.Context) http.Response {
	// 从登录日志中统计不同设备类型的访问量
	// 统计最近30天的数据
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	var loginLogs []models.LoginLog
	facades.Orm().Query().Model(&models.LoginLog{}).
		Where("created_at >= ?", thirtyDaysAgo).
		Where("status", 1).
		Get(&loginLogs)

	// 统计设备类型
	deviceStats := make(map[string]int64)
	for _, log := range loginLogs {
		deviceType := r.parseDeviceType(log.UserAgent)
		deviceStats[deviceType]++
	}

	// 转换为数组格式
	result := []map[string]any{
		{"name": "桌面端", "value": deviceStats["desktop"]},
		{"name": "移动端", "value": deviceStats["mobile"]},
		{"name": "平板端", "value": deviceStats["tablet"]},
		{"name": "其他", "value": deviceStats["other"]},
	}

	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": "get_success",
		"data":    result,
	})
}

// parseDeviceType 根据 UserAgent 解析设备类型
func (r *DashboardController) parseDeviceType(userAgent string) string {
	if userAgent == "" {
		return "other"
	}

	ua := strings.ToLower(userAgent)

	// 平板设备检测（需要在移动设备之前检测）
	if strings.Contains(ua, "ipad") || (strings.Contains(ua, "tablet") && !strings.Contains(ua, "mobile")) {
		return "tablet"
	}

	// 移动设备检测
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "mobile"
	}

	// 桌面设备
	return "desktop"
}

// GetWeeklyUserActivity 获取每周用户活跃量（从操作日志统计）
func (r *DashboardController) GetWeeklyUserActivity(ctx http.Context) http.Response {
	weeklyData := r.getWeeklyUserActivityData()
	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": "get_success",
		"data":    weeklyData,
	})
}

// GetMonthlySales 获取每月操作统计（替换销售额数据）
func (r *DashboardController) GetMonthlySales(ctx http.Context) http.Response {
	// 替换成操作日志月度统计
	monthlyData := r.getMonthlyOperationData()
	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": "get_success",
		"data":    monthlyData,
	})
}

// GetRecentActivities 获取最近活动
func (r *DashboardController) GetRecentActivities(ctx http.Context) http.Response {
	// 获取最近10条操作日志
	var logs []models.OperationLog
	facades.Orm().Query().Model(&models.OperationLog{}).
		With("Admin").
		Order("id desc").
		Limit(10).
		Get(&logs)

	activities := make([]map[string]any, 0, len(logs))
	for _, log := range logs {
		adminName := "未知用户"
		if log.Admin.ID > 0 {
			adminName = log.Admin.Nickname
			if adminName == "" {
				adminName = log.Admin.Username
			}
		}

		statusText := "成功"
		statusType := "success"
		if log.Status == 0 {
			statusText = "失败"
			statusType = "danger"
		}

		// 计算时间差
		var timeAgo string
		if log.CreatedAt != nil {
			// carbon.DateTime 转换为 time.Time（使用 ToDateTimeString 然后解析）
			timeStr := log.CreatedAt.ToDateTimeString()
			if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
				timeAgo = r.formatTimeAgo(t)
			} else {
				timeAgo = "未知"
			}
		} else {
			timeAgo = "未知"
		}

		activities = append(activities, map[string]any{
			"user":        adminName,
			"action":      log.Title,
			"time":        timeAgo,
			"status":      statusText,
			"type":        statusType,
			"avatarColor": r.getAvatarColor(adminName),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"code":    200,
		"message": "get_success",
		"data":    activities,
	})
}

// formatTimeAgo 格式化时间差
func (r *DashboardController) formatTimeAgo(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < time.Minute {
		return "刚刚"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d分钟前", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d小时前", hours)
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%d天前", days)
	}
}

// getAvatarColor 根据用户名生成头像颜色
func (r *DashboardController) getAvatarColor(name string) string {
	colors := []string{"#409EFF", "#67C23A", "#E6A23C", "#F56C6C", "#909399", "#606266"}
	if name == "" {
		return colors[0]
	}
	hash := 0
	for _, char := range name {
		hash = hash*31 + int(char)
	}
	return colors[hash%len(colors)]
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

// getWeeklyUserActivityData 获取每周用户活跃量（从操作日志统计）
func (r *DashboardController) getWeeklyUserActivityData() []map[string]any {
	now := time.Now()
	weeklyData := make([]map[string]any, 7)

	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		// 计算当天的开始和结束时间
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		// 统计当天的操作日志数（访问量）
		visitCount, _ := facades.Orm().Query().Model(&models.OperationLog{}).
			Where("created_at >= ?", startOfDay).
			Where("created_at <= ?", endOfDay).
			Where("status", 1).
			Count()

		// 统计当天活跃的管理员数（去重）
		var uniqueAdmins []uint
		facades.Orm().Query().Model(&models.OperationLog{}).
			Where("created_at >= ?", startOfDay).
			Where("created_at <= ?", endOfDay).
			Where("status", 1).
			Select("DISTINCT admin_id").
			Pluck("admin_id", &uniqueAdmins)

		userCount := int64(len(uniqueAdmins))

		weeklyData[6-i] = map[string]any{
			"date":   dateStr,
			"visits": visitCount,
			"users":  userCount,
		}
	}
	return weeklyData
}

// getMonthlySalesData 获取每月销售额（保留方法名以兼容 SSE）
func (r *DashboardController) getMonthlySalesData() []map[string]any {
	return r.getMonthlyOperationData()
}

// getMonthlyOperationData 获取每月操作统计（替换销售额）
func (r *DashboardController) getMonthlyOperationData() []map[string]any {
	now := time.Now()
	monthlyData := make([]map[string]any, 12)

	for i := 11; i >= 0; i-- {
		date := now.AddDate(0, -i, 0)
		monthStr := date.Format("2006-01")

		// 计算当月的开始和结束时间
		startOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		var endOfMonth time.Time
		if i == 0 {
			// 当前月，使用当前时间
			endOfMonth = now
		} else {
			// 历史月份，使用月末
			endOfMonth = startOfMonth.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}

		// 统计当月的操作日志数
		operationCount, _ := facades.Orm().Query().Model(&models.OperationLog{}).
			Where("created_at >= ?", startOfMonth).
			Where("created_at <= ?", endOfMonth).
			Where("status", 1).
			Count()

		monthlyData[11-i] = map[string]any{
			"month": monthStr,
			"count": operationCount,
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
