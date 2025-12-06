package listeners

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
)

// CreateUserProfile 创建用户档案监听器（启用队列）
type CreateUserProfile struct {
}

func (receiver *CreateUserProfile) Signature() string {
	return "create_user_profile"
}

// Queue 启用队列，异步创建用户档案
func (receiver *CreateUserProfile) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     true,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *CreateUserProfile) Handle(args ...any) error {
	if len(args) > 0 {
		userID := args[0]

		// 模拟创建用户档案（耗时操作）
		facades.Log().Infof("📝 [队列] 创建用户档案，用户 ID: %v", userID)
		// 实际场景中这里会创建用户档案、初始化设置等
	}
	return nil
}
