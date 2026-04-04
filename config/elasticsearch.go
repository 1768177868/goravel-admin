package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("elasticsearch", map[string]any{
		// 关闭时不注册容器绑定，避免未部署 ES 时启动失败
		"enabled": config.Env("ELASTICSEARCH_ENABLED", false),
		"default": config.Env("ELASTICSEARCH_CONNECTION", "default"),
		"connections": map[string]any{
			"default": map[string]any{
				"urls":     config.Env("ELASTICSEARCH_URLS", "http://127.0.0.1:9200"),
				"username": config.Env("ELASTICSEARCH_USERNAME", ""),
				"password": config.Env("ELASTICSEARCH_PASSWORD", ""),
				"api_key":  config.Env("ELASTICSEARCH_API_KEY", ""),
				"cloud_id": config.Env("ELASTICSEARCH_CLOUD_ID", ""),
				// 仅开发/自签证书：跳过 TLS 校验（生产务必 false）
				"insecure_skip_verify": config.Env("ELASTICSEARCH_INSECURE_SKIP_VERIFY", false),
			},
		},
		"index_prefix": config.Env("ELASTICSEARCH_INDEX_PREFIX", ""),
		// 示例索引短名（实际索引名 = index_prefix + demo_index）
		"demo_index": config.Env("ELASTICSEARCH_DEMO_INDEX", "goravel_demo"),
		// 订单写入 ES：需同时开启 enabled 与本项
		"sync_orders_enabled": config.Env("ELASTICSEARCH_SYNC_ORDERS", false),
		// 订单索引短名（完整 = index_prefix + orders_index）
		"orders_index": config.Env("ELASTICSEARCH_ORDERS_INDEX", "orders"),
		// 订单等 ES 同步任务使用的队列逻辑名（需 bootstrap 中注册对应 Worker）
		"sync_queue": config.Env("ELASTICSEARCH_QUEUE", "elasticsearch"),
		// 同步队列 Worker：auto=任一副开关开启则启动 | true | false
		"sync_worker": config.Env("ELASTICSEARCH_SYNC_WORKER", "auto"),
	})
}
