package admin

import (
	apphttp "github.com/goravel/framework/contracts/http"

	"goravel/app/http/response"
	"goravel/app/services"
)

// ElasticsearchController HTTP 层使用示例（需开启 ELASTICSEARCH_ENABLED）。
type ElasticsearchController struct {
	svc *services.ElasticsearchService
}

func NewElasticsearchController() *ElasticsearchController {
	return &ElasticsearchController{
		svc: services.NewElasticsearchService(),
	}
}

// Ping GET api/admin/elasticsearch/ping — 需登录，用于健康检查。
func (r *ElasticsearchController) Ping(ctx apphttp.Context) apphttp.Response {
	if err := r.svc.Ping(ctx.Context()); err != nil {
		return response.ErrorWithLog(ctx, "elasticsearch", err, map[string]any{
			"action": "ping",
		})
	}
	return response.Success(ctx, map[string]any{
		"elasticsearch": "ok",
	})
}
