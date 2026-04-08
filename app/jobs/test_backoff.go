package jobs

import (
	"fmt"
	"sync"
	"time"
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

	// 前 3 次故意失败，触发 ShouldRetry 的 5/10/20 秒退避。
	if attempt <= 3 {
		return fmt.Errorf("backoff test failure: marker=%s attempt=%d", marker, attempt)
	}

	return nil
}

func (r *TestBackoff) ShouldRetry(err error, attempt int) (bool, time.Duration) {
	delays := []time.Duration{
		5 * time.Second,
		10 * time.Second,
		15 * time.Second,
	}

	if attempt <= 0 || attempt > len(delays) {
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

	return true, delay
}

func ResetTestBackoff() {
	testBackoffMu.Lock()
	defer testBackoffMu.Unlock()

	testBackoffAttempt = map[string]int{}
	TestBackoffResult = nil
}
