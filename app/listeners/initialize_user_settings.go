package listeners

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/errors"
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

// Handle 处理初始化用户设置
// 
// 参数:
//   - args[0]: UserRegisteredArgs 结构体或 userID (int/uint)
//
// 返回:
//   - error: 错误信息
func (receiver *InitializeUserSettings) Handle(args ...any) error {
	if len(args) < 1 {
		return errors.ErrInvalidArgument.WithMessage("missing user ID")
	}

	var userID uint
	if ura, ok := args[0].(UserRegisteredArgs); ok {
		userID = ura.UserID
	} else {
		// 兼容旧版本：按位置解析
		userID = cast.ToUint(args[0])
	}

	if userID == 0 {
		return errors.ErrInvalidArgument.WithMessage("invalid user ID")
	}

	// 同步执行：立即初始化用户设置（需要立即生效）
	facades.Log().Infof("⚙️ [同步] 初始化用户设置，用户 ID: %d", userID)
	// 实际场景中这里会立即初始化用户默认设置
	return nil
}
