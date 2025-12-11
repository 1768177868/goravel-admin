package providers

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/facades"
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
		// events.NewOrderShipped(): {
		// 	listeners.NewSendShipmentNotification(),
		// },
		// 订单取消事件
		// events.NewOrderCanceled(): {
		// 	listeners.NewSendShipmentNotification(),
		// },
		// 订单创建事件 - 演示混合使用（同步+异步）
		// &events.OrderCreated{}: {
		// 	&listeners.UpdateInventory{},       // 同步执行：立即更新库存
		// 	&listeners.SendOrderNotification{}, // 启用队列：异步发送通知
		// },
	}
}
