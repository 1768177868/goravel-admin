package events

import "github.com/goravel/framework/contracts/event"

// OrderCreated 订单创建事件
type OrderCreated struct {
}

func (receiver *OrderCreated) Handle(args []event.Arg) ([]event.Arg, error) {
	// 可以在这里对订单数据进行加工
	return args, nil
}
