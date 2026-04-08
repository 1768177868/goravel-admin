package jobs

import (
	"fmt"
	"strconv"
	"time"
)

var TestClaimResult []any

type TestClaim struct{}

func (r *TestClaim) Signature() string {
	return "test_claim"
}

func (r *TestClaim) Handle(args ...any) error {
	if len(args) < 2 {
		return fmt.Errorf("test_claim requires marker and sleep seconds")
	}

	marker := fmt.Sprint(args[0])
	sleepSeconds, err := parseSleepSeconds(args[1])
	if err != nil {
		return err
	}
	if sleepSeconds < 1 {
		sleepSeconds = 1
	}

	TestClaimResult = append(TestClaimResult, map[string]any{
		"marker": marker,
		"stage":  "start",
		"at":     time.Now().Format(time.RFC3339),
		"sleep":  sleepSeconds,
	})

	time.Sleep(time.Duration(sleepSeconds) * time.Second)

	TestClaimResult = append(TestClaimResult, map[string]any{
		"marker": marker,
		"stage":  "done",
		"at":     time.Now().Format(time.RFC3339),
		"sleep":  sleepSeconds,
	})

	return nil
}

func parseSleepSeconds(raw any) (int, error) {
	switch v := raw.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("invalid sleep seconds type: %T", raw)
	}
}
