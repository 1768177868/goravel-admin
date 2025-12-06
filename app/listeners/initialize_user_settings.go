package listeners

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
)

// InitializeUserSettings 初始化用户设置监听器（同步执行）
type InitializeUserSettings struct {
}

func (receiver *InitializeUserSettings) Signature() string {
	return "initialize_user_settings"
}

// Queue 禁用队列，同步执行（需要立即生效）
func (receiver *InitializeUserSettings) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false, // 同步执行
		Connection: "",
		Queue:      "",
	}
}

func (receiver *InitializeUserSettings) Handle(args ...any) error {
	if len(args) > 0 {
		userID := args[0]

		// 同步执行：立即初始化用户设置（需要立即生效）
		facades.Log().Infof("⚙️ [同步] 初始化用户设置，用户 ID: %v", userID)
		// 实际场景中这里会立即初始化用户默认设置
	}
	return nil
}
