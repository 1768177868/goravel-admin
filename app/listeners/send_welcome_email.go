package listeners

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/errors"
)

// UserRegisteredArgs 用户注册事件的参数结构体
type UserRegisteredArgs struct {
	UserID uint
	Email  string
}

// SendWelcomeEmail 发送欢迎邮件监听器（启用队列）
type SendWelcomeEmail struct {
}

func (receiver *SendWelcomeEmail) Signature() string {
	return "send_welcome_email"
}

// Queue 启用队列，异步发送邮件
func (receiver *SendWelcomeEmail) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     true,
		Connection: "", // 使用默认连接
		Queue:      "", // 使用默认队列
	}
}

// Handle 处理发送欢迎邮件
// 
// 参数:
//   - args[0]: UserRegisteredArgs 结构体或 userID (int/uint)
//   - args[1]: 如果 args[0] 不是结构体，则 args[1] 是 email (string)
//
// 返回:
//   - error: 错误信息
func (receiver *SendWelcomeEmail) Handle(args ...any) error {
	if len(args) < 1 {
		return errors.ErrInvalidArgument.WithMessage("missing user ID")
	}

	var userID uint
	var email string

	// 尝试解析为结构体
	if len(args) == 1 {
		if ura, ok := args[0].(UserRegisteredArgs); ok {
			userID = ura.UserID
			email = ura.Email
		} else {
			// 兼容旧版本：按位置解析
			userID = cast.ToUint(args[0])
		}
	} else {
		// 兼容旧版本：按位置解析
		userID = cast.ToUint(args[0])
		if len(args) > 1 {
			if e, ok := args[1].(string); ok {
				email = e
			} else {
				email = cast.ToString(args[1])
			}
		}
	}

	if userID == 0 {
		return errors.ErrInvalidArgument.WithMessage("invalid user ID")
	}

	// 模拟发送邮件（耗时操作）
	facades.Log().Infof("📧 [队列] 发送欢迎邮件给用户 ID: %d, Email: %s", userID, email)
	// 实际场景中这里会调用邮件服务发送邮件
	return nil
}
