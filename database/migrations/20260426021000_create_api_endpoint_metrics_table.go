package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260426021000CreateApiEndpointMetricsTable struct{}

func (r *M20260426021000CreateApiEndpointMetricsTable) Signature() string {
	return "20260426021000_create_api_endpoint_metrics_table"
}

func (r *M20260426021000CreateApiEndpointMetricsTable) Up() error {
	if facades.Schema().HasTable("api_endpoint_metrics") {
		return nil
	}

	return facades.Schema().Create("api_endpoint_metrics", func(table schema.Blueprint) {
		table.BigIncrements("id")
		table.String("trace_id", 120).Nullable().Comment("链路ID")
		table.String("method", 12).Comment("请求方法")
		table.String("route_template", 255).Comment("路由模板")
		table.Integer("status_code").Default(200).Comment("响应状态码")
		table.Decimal("duration_ms").Total(12).Places(3).Default(0).Comment("耗时毫秒")
		table.String("occurred_at", 32).Nullable().Comment("发生时间")
		table.Timestamps()

		table.Index("trace_id")
		table.Index("status_code")
		table.Index("duration_ms")
		table.Index("occurred_at")
		table.Index("created_at")
		table.Index("occurred_at", "method", "route_template")
		table.Index("occurred_at", "status_code")
		table.Comment("接口性能指标表")
	})
}

func (r *M20260426021000CreateApiEndpointMetricsTable) Down() error {
	return facades.Schema().DropIfExists("api_endpoint_metrics")
}
