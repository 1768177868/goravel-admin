package listeners

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/errors"
)

// OrderShippedArgs 订单发货事件的参数结构体
type OrderShippedArgs struct {
	OrderID uint
	Remark  string
}

type SendShipmentNotification struct {
}

func NewSendShipmentNotification() *SendShipmentNotification {
	return &SendShipmentNotification{}
}

func (receiver *SendShipmentNotification) Signature() string {
	return "send_shipment_notification"
}

func (receiver *SendShipmentNotification) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     true,
		Connection: "",
		Queue:      "",
	}
}

// Handle 处理发送发货通知
// 
// 参数:
//   - args[0]: OrderShippedArgs 结构体或 orderID/remark (int/uint/string)
//
// 返回:
//   - error: 错误信息
func (receiver *SendShipmentNotification) Handle(args ...any) error {
	if len(args) < 1 {
		return errors.ErrInvalidArgument.WithMessage("missing shipment notification arguments")
	}

	var orderID uint
	var remark string

	if osa, ok := args[0].(OrderShippedArgs); ok {
		orderID = osa.OrderID
		remark = osa.Remark
	} else {
		// 兼容旧版本：按位置解析
		orderID = cast.ToUint(args[0])
		if len(args) > 1 {
			remark = cast.ToString(args[1])
		}
	}

	if orderID == 0 {
		return errors.ErrInvalidArgument.WithMessage("invalid order ID")
	}

	facades.Log().Infof("📦 [队列] 发送发货通知，订单 ID: %d, 备注: %s", orderID, remark)
	// 实际场景中这里会发送发货通知
	return nil
}
