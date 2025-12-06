package listeners

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/errors"
)

// OrderCreatedArgs 订单创建事件的参数结构体
type OrderCreatedArgs struct {
	OrderID uint
	Remark  string
}

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

// Handle 处理更新库存
// 
// 参数:
//   - args[0]: OrderCreatedArgs 结构体或 orderID (int/uint)
//
// 返回:
//   - error: 错误信息
func (receiver *UpdateInventory) Handle(args ...any) error {
	if len(args) < 1 {
		return errors.ErrInvalidArgument.WithMessage("missing order ID")
	}

	var orderID uint
	if oca, ok := args[0].(OrderCreatedArgs); ok {
		orderID = oca.OrderID
	} else {
		// 兼容旧版本：按位置解析
		orderID = cast.ToUint(args[0])
	}

	if orderID == 0 {
		return errors.ErrInvalidArgument.WithMessage("invalid order ID")
	}

	// 同步执行：立即更新库存（需要立即生效，避免超卖）
	facades.Log().Infof("📦 [同步] 更新库存，订单 ID: %d", orderID)
	// 实际场景中这里会立即更新商品库存
	return nil
}
