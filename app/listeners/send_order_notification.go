package listeners

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
)

// SendOrderNotification 发送订单通知监听器（启用队列）
type SendOrderNotification struct {
}

func (receiver *SendOrderNotification) Signature() string {
	return "send_order_notification"
}

// Queue 启用队列，异步发送通知
func (receiver *SendOrderNotification) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     true,
		Connection: "",
		Queue:      "notifications", // 使用专门的通知队列
	}
}

func (receiver *SendOrderNotification) Handle(args ...any) error {
	if len(args) > 0 {
		orderID := args[0]

		// 模拟发送通知（耗时操作）
		facades.Log().Infof("🔔 [队列] 发送订单通知，订单 ID: %v", orderID)
		// 实际场景中这里会发送短信、推送通知等
	}
	return nil
}
