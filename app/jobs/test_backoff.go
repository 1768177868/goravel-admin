package jobs

import (
	"fmt"
	"sync"
	"time"

	"github.com/goravel/framework/facades"
)

var (
	testBackoffMu      sync.Mutex
	testBackoffAttempt = map[string]int{}
	TestBackoffResult  []any
)

type TestBackoff struct{}

func (r *TestBackoff) Signature() string {
	return "test_backoff"
}

func (r *TestBackoff) Handle(args ...any) error {
	marker := "default"
	if len(args) > 0 {
		marker = fmt.Sprint(args[0])
	}

	testBackoffMu.Lock()
	testBackoffAttempt[marker]++
	attempt := testBackoffAttempt[marker]
	TestBackoffResult = append(TestBackoffResult, map[string]any{
		"marker":  marker,
		"attempt": attempt,
		"stage":   "handle",
		"at":      time.Now().Format(time.RFC3339),
	})
	testBackoffMu.Unlock()

	facades.Log().Infof("[queue-test/backoff] handle marker=%s attempt=%d", marker, attempt)

	// 前 3 次故意失败，触发 ShouldRetry 的 5/10/20 秒退避。
	if attempt <= 3 {
		facades.Log().Warningf("[queue-test/backoff] simulated failure marker=%s attempt=%d", marker, attempt)
		return fmt.Errorf("backoff test failure: marker=%s attempt=%d", marker, attempt)
	}

	facades.Log().Infof("[queue-test/backoff] success marker=%s attempt=%d", marker, attempt)
	return nil
}

func (r *TestBackoff) ShouldRetry(err error, attempt int) (bool, time.Duration) {
	delays := []time.Duration{
		5 * time.Second,
		10 * time.Second,
		20 * time.Second,
	}

	if attempt <= 0 || attempt > len(delays) {
		facades.Log().Warningf("[queue-test/backoff] should_retry=false attempt=%d reason=out_of_range", attempt)
		return false, 0
	}

	delay := delays[attempt-1]
	testBackoffMu.Lock()
	TestBackoffResult = append(TestBackoffResult, map[string]any{
		"stage":         "should_retry",
		"attempt":       attempt,
		"retryable":     true,
		"delay_seconds": int(delay.Seconds()),
		"at":            time.Now().Format(time.RFC3339),
	})
	testBackoffMu.Unlock()

	facades.Log().Infof("[queue-test/backoff] should_retry=true attempt=%d delay=%s err=%v", attempt, delay, err)

	return true, delay
}

func ResetTestBackoff() {
	testBackoffMu.Lock()
	defer testBackoffMu.Unlock()

	testBackoffAttempt = map[string]int{}
	TestBackoffResult = nil
}
