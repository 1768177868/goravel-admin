package jobs

import (
	"time"

	"github.com/spf13/cast"
)

var TestResult []any

type Test struct {
}

// Signature The name and signature of the job.
func (r *Test) Signature() string {
	return "test"
}

// Handle Execute the job.
func (r *Test) Handle(args ...any) error {
	// 调试用：如果参数里带了等待秒数，则先阻塞一段时间，便于观察队列状态变化。
	if waitSeconds := parseWaitSeconds(args...); waitSeconds > 0 {
		time.Sleep(time.Duration(waitSeconds) * time.Second)
	}

	if len(args) > 0 {
		TestResult = append(TestResult, args...)
	}
	return nil
}

func parseWaitSeconds(args ...any) int {
	const defaultWaitSeconds = 20

	if len(args) < 2 {
		return defaultWaitSeconds
	}

	waitSeconds, err := cast.ToIntE(args[1])
	if err != nil {
		return defaultWaitSeconds
	}
	return nonNegative(waitSeconds)
}

func nonNegative(v int) int {
	if v < 0 {
		return 0
	}
	return v
}
