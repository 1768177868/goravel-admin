package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/queue"
)

// queueProvider 包装框架的 queue.ServiceProvider，禁用默认队列 runner。
// 队列由 bootstrap/app.go 的 WithRunners 中注册的自定义 runner 启动。
type queueProvider struct {
	queue.ServiceProvider
}

func (p *queueProvider) Runners(app foundation.Application) []foundation.Runner {
	return nil
}
