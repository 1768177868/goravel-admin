package models

import "github.com/goravel/framework/database/orm"

// ApiEndpointMetric stores per-request API performance metrics for observability dashboards.
type ApiEndpointMetric struct {
	orm.Model
	TraceID       string  `gorm:"size:120;index;comment:链路ID" json:"trace_id"`
	Method        string  `gorm:"size:12;index;comment:请求方法" json:"method"`
	RouteTemplate string  `gorm:"size:255;index;comment:路由模板" json:"route_template"`
	StatusCode    int     `gorm:"index;comment:响应状态码" json:"status_code"`
	DurationMS    float64 `gorm:"type:decimal(12,3);index;comment:耗时毫秒" json:"duration_ms"`
	OccurredAt    string  `gorm:"size:32;index;comment:发生时间" json:"occurred_at"`
}
