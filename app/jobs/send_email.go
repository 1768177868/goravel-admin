package jobs

import (
	"github.com/goravel/framework/facades"
)

// SendEmail 发送邮件任务
type SendEmail struct {
}

func (r *SendEmail) Signature() string {
	return "send_email"
}

func (r *SendEmail) Handle(args ...any) error {
	if len(args) >= 2 {
		to := args[0].(string)
		subject := args[1].(string)
		content := ""
		if len(args) >= 3 {
			content = args[2].(string)
		}

		facades.Log().Infof("📧 [Job] 发送邮件 - 收件人: %s, 主题: %s, 内容: %s", to, subject, content)
		// 实际场景中这里会调用邮件服务发送邮件
	}
	return nil
}
