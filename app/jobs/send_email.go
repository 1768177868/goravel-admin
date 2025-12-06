package jobs

import (
	"time"

	"github.com/goravel/framework/facades"
)

// SendEmail 发送邮件任务（支持递增延迟重试）
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

		// 实际场景中这里会调用邮件服务发送邮件
		err := sendEmail(to, subject, content)
		if err != nil {
			facades.Log().Errorf("📧 [Job] 发送邮件失败 - 收件人: %s, 主题: %s, 错误: %v", to, subject, err)
			return err // 返回错误，触发重试
		}

		facades.Log().Infof("📧 [Job] 发送邮件成功 - 收件人: %s, 主题: %s", to, subject)
	}
	return nil
}

// ShouldRetry 自定义重试逻辑：递增延迟重试
// attempt: 当前重试次数（从1开始，第1次重试时attempt=1，第2次重试时attempt=2）
// 返回: (是否重试, 延迟时间)
func (r *SendEmail) ShouldRetry(err error, attempt int) (retryable bool, delay time.Duration) {
	// 最多重试 3 次
	maxRetries := 3
	if attempt > maxRetries {
		facades.Log().Errorf("📧 [Job] 已达到最大重试次数 %d，不再重试", maxRetries)
		return false, 0 // 不再重试
	}

	// 递增延迟：第1次重试延迟3秒，第2次延迟10秒，第3次延迟20秒
	delays := []time.Duration{
		3 * time.Second,  // 第1次重试：3秒
		10 * time.Second, // 第2次重试：10秒
		20 * time.Second, // 第3次重试：20秒
	}

	// 获取当前重试的延迟时间（attempt从1开始，所以减1作为索引）
	delayIndex := attempt - 1
	if delayIndex < len(delays) {
		delay = delays[delayIndex]
	} else {
		// 如果超过配置的延迟数组，使用最后一个延迟时间
		delay = delays[len(delays)-1]
	}

	facades.Log().Infof("📧 [Job] 第 %d 次重试，将在 %v 后执行", attempt, delay)
	return true, delay
}

// sendEmail 模拟发送邮件函数
func sendEmail(to, subject, content string) error {
	// 实际场景中这里会调用邮件服务发送邮件
	// 模拟：随机失败（用于测试重试）
	// if rand.Intn(3) == 0 {
	//     return errors.New("邮件服务暂时不可用")
	// }
	return nil
}
