package listeners

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
)

// UpdateInventory 更新库存监听器（同步执行）
type UpdateInventory struct {
}

func (receiver *UpdateInventory) Signature() string {
	return "update_inventory"
}

// Queue 禁用队列，同步执行（库存需要立即更新）
func (receiver *UpdateInventory) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false, // 同步执行，库存需要立即更新
		Connection: "",
		Queue:      "",
	}
}

func (receiver *UpdateInventory) Handle(args ...any) error {
	if len(args) > 0 {
		orderID := args[0]

		// 同步执行：立即更新库存（需要立即生效，避免超卖）
		facades.Log().Infof("📦 [同步] 更新库存，订单 ID: %v", orderID)
		// 实际场景中这里会立即更新商品库存
	}
	return nil
}
