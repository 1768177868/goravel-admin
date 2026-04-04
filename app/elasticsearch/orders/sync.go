package esorders

import "github.com/goravel/framework/facades"

// SyncEnabled 为 true 时表示 ES 已启用且允许订单写入 ES（elasticsearch.sync_orders_enabled）。
func SyncEnabled() bool {
	return facades.Config().GetBool("elasticsearch.enabled", false) &&
		facades.Config().GetBool("elasticsearch.sync_orders_enabled", false)
}
