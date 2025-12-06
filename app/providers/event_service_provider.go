package providers

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/facades"

	"goravel/app/events"
	"goravel/app/listeners"
)

type EventServiceProvider struct {
}

func (receiver *EventServiceProvider) Register(app foundation.Application) {
	facades.Event().Register(receiver.listen())
}

func (receiver *EventServiceProvider) Boot(app foundation.Application) {

}

func (receiver *EventServiceProvider) listen() map[event.Event][]event.Listener {
	return map[event.Event][]event.Listener{
		// 订单发货事件
		events.NewOrderShipped(): {
			listeners.NewSendShipmentNotification(),
		},
		// 订单取消事件
		events.NewOrderCanceled(): {
			listeners.NewSendShipmentNotification(),
		},
		// 用户注册事件 - 演示事件+队列结合使用
		&events.UserRegistered{}: {
			&listeners.SendWelcomeEmail{},       // 启用队列：异步发送欢迎邮件
			&listeners.CreateUserProfile{},      // 启用队列：异步创建用户档案
			&listeners.InitializeUserSettings{}, // 同步执行：立即初始化用户设置
		},
		// 订单创建事件 - 演示混合使用（同步+异步）
		&events.OrderCreated{}: {
			&listeners.UpdateInventory{},       // 同步执行：立即更新库存
			&listeners.SendOrderNotification{}, // 启用队列：异步发送通知
		},
	}
}
