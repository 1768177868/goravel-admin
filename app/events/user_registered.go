package events

import "github.com/goravel/framework/contracts/event"

// UserRegistered 用户注册事件
type UserRegistered struct {
}

func (receiver *UserRegistered) Handle(args []event.Arg) ([]event.Arg, error) {
	// 可以在这里对数据进行加工
	// 例如：添加注册时间、IP地址等信息
	return args, nil
}
