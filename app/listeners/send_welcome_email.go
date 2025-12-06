package listeners

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
)

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

func (receiver *SendWelcomeEmail) Handle(args ...any) error {
	if len(args) > 0 {
		userID := args[0]
		email := ""
		if len(args) > 1 {
			email = args[1].(string)
		}

		// 模拟发送邮件（耗时操作）
		facades.Log().Infof("📧 [队列] 发送欢迎邮件给用户 ID: %v, Email: %s", userID, email)
		// 实际场景中这里会调用邮件服务发送邮件
	}
	return nil
}
