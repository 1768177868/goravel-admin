package esworker

import (
	"strings"

	"github.com/goravel/framework/facades"
)

// ShouldRunQueueWorker 是否启动 ES 同步专用队列 Worker。
// elasticsearch.sync_worker: auto（默认）= 任一同步子模块开启则运行；true/false 强制开/关。
func ShouldRunQueueWorker() bool {
	if !facades.Config().GetBool("elasticsearch.enabled", false) {
		return false
	}
	sw := strings.ToLower(strings.TrimSpace(facades.Config().GetString("elasticsearch.sync_worker", "auto")))
	switch sw {
	case "false", "0", "no", "off":
		return false
	case "true", "1", "yes", "on":
		return true
	default:
		return anySyncModuleEnabled()
	}
}

func anySyncModuleEnabled() bool {
	if facades.Config().GetBool("elasticsearch.sync_orders_enabled", false) {
		return true
	}
	// 后续例如: sync_articles_enabled
	return false
}
