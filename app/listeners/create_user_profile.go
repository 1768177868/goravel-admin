package listeners

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/errors"
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

// Handle 处理创建用户档案
// 
// 参数:
//   - args[0]: UserRegisteredArgs 结构体或 userID (int/uint)
//
// 返回:
//   - error: 错误信息
func (receiver *CreateUserProfile) Handle(args ...any) error {
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

	// 模拟创建用户档案（耗时操作）
	facades.Log().Infof("📝 [队列] 创建用户档案，用户 ID: %d", userID)
	// 实际场景中这里会创建用户档案、初始化设置等
	return nil
}
