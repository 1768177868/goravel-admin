package api

import (
	"strconv"
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/response"
	"goravel/app/jobs"
)

type QueueTestController struct{}

func NewQueueTestController() *QueueTestController {
	return &QueueTestController{}
}

func (c *QueueTestController) Dispatch(ctx http.Context) http.Response {
	args := []contractsqueue.Arg{
		{Type: "string", Value: "queue-test-dispatch"},
		{Type: "string", Value: time.Now().Format(time.RFC3339)},
	}
	if err := facades.Queue().Job(&jobs.Test{}, args).Dispatch(); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, "success", http.Json{
		"queued":     true,
		"type":       "dispatch",
		"connection": facades.Config().GetString("queue.default", "sync"),
		"queue":      "default",
	})
}

func (c *QueueTestController) Delay(ctx http.Context) http.Response {
	delaySeconds, _ := strconv.Atoi(ctx.Request().Query("seconds", "5"))
	if delaySeconds <= 0 {
		delaySeconds = 5
	}

	args := []contractsqueue.Arg{
		{Type: "string", Value: "queue-test-delay"},
		{Type: "int", Value: delaySeconds},
		{Type: "string", Value: time.Now().Format(time.RFC3339)},
	}
	if err := facades.Queue().Job(&jobs.Test{}, args).Delay(time.Now().Add(time.Duration(delaySeconds) * time.Second)).Dispatch(); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, "success", http.Json{
		"queued":      true,
		"type":        "delay",
		"delay_second": delaySeconds,
		"connection":  facades.Config().GetString("queue.default", "sync"),
		"queue":       "default",
	})
}

func (c *QueueTestController) LongRunning(ctx http.Context) http.Response {
	args := []contractsqueue.Arg{
		{Type: "string", Value: "queue-test-long-running"},
		{Type: "string", Value: time.Now().Format(time.RFC3339)},
	}
	if err := facades.Queue().Job(&jobs.Test{}, args).OnQueue("long-running").Dispatch(); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, "success", http.Json{
		"queued":     true,
		"type":       "long-running",
		"connection": facades.Config().GetString("queue.default", "sync"),
		"queue":      "long-running",
	})
}

func (c *QueueTestController) Fail(ctx http.Context) http.Response {
	args := []contractsqueue.Arg{
		{Type: "string", Value: "queue-test-fail"},
		{Type: "string", Value: time.Now().Format(time.RFC3339)},
	}
	if err := facades.Queue().Job(&jobs.TestErr{}, args).Dispatch(); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, "success", http.Json{
		"queued":     true,
		"type":       "fail",
		"connection": facades.Config().GetString("queue.default", "sync"),
		"queue":      "default",
	})
}

func (c *QueueTestController) AllInOne(ctx http.Context) http.Response {
	delaySeconds, _ := strconv.Atoi(ctx.Request().Query("seconds", "5"))
	if delaySeconds <= 0 {
		delaySeconds = 5
	}

	now := time.Now().Format(time.RFC3339)
	connection := facades.Config().GetString("queue.default", "sync")

	if err := facades.Queue().Job(&jobs.Test{}, []contractsqueue.Arg{
		{Type: "string", Value: "queue-test-all-dispatch"},
		{Type: "string", Value: now},
	}).Dispatch(); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	if err := facades.Queue().Job(&jobs.Test{}, []contractsqueue.Arg{
		{Type: "string", Value: "queue-test-all-delay"},
		{Type: "int", Value: delaySeconds},
		{Type: "string", Value: now},
	}).Delay(time.Now().Add(time.Duration(delaySeconds) * time.Second)).Dispatch(); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	if err := facades.Queue().Job(&jobs.Test{}, []contractsqueue.Arg{
		{Type: "string", Value: "queue-test-all-long-running"},
		{Type: "string", Value: now},
	}).OnQueue("long-running").Dispatch(); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	if err := facades.Queue().Job(&jobs.TestErr{}, []contractsqueue.Arg{
		{Type: "string", Value: "queue-test-all-fail"},
		{Type: "string", Value: now},
	}).Dispatch(); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, "success", http.Json{
		"queued":       true,
		"type":         "all-in-one",
		"connection":   connection,
		"delay_second": delaySeconds,
		"items": []string{
			"default:dispatch",
			"default:delay",
			"long-running:dispatch",
			"default:fail",
		},
	})
}

func (c *QueueTestController) Reclaim(ctx http.Context) http.Response {
	sleepSeconds, _ := strconv.Atoi(ctx.Request().Query("sleep", "30"))
	if sleepSeconds <= 0 {
		sleepSeconds = 30
	}
	targetQueue := ctx.Request().Query("queue", "default")
	marker := "claim-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	args := []contractsqueue.Arg{
		{Type: "string", Value: marker},
		{Type: "int", Value: sleepSeconds},
	}
	if err := facades.Queue().Job(&jobs.TestClaim{}, args).OnQueue(targetQueue).Dispatch(); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, "success", http.Json{
		"queued":       true,
		"type":         "reclaim",
		"connection":   facades.Config().GetString("queue.default", "sync"),
		"queue":        targetQueue,
		"marker":       marker,
		"sleep_second": sleepSeconds,
		"hint":         "start worker A, then kill before ack, wait retry_after, start worker B",
	})
}

func (c *QueueTestController) AllSpecial(ctx http.Context) http.Response {
	delaySeconds, _ := strconv.Atoi(ctx.Request().Query("delay_seconds", "10"))
	if delaySeconds <= 0 {
		delaySeconds = 10
	}
	reclaimSleep, _ := strconv.Atoi(ctx.Request().Query("reclaim_sleep", "30"))
	if reclaimSleep <= 0 {
		reclaimSleep = 30
	}
	reclaimQueue := ctx.Request().Query("reclaim_queue", "default")
	now := time.Now().Format(time.RFC3339)
	reclaimMarker := "all-special-claim-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	if err := facades.Queue().Job(&jobs.Test{}, []contractsqueue.Arg{
		{Type: "string", Value: "queue-test-all-special-delay"},
		{Type: "int", Value: delaySeconds},
		{Type: "string", Value: now},
	}).Delay(time.Now().Add(time.Duration(delaySeconds) * time.Second)).Dispatch(); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	if err := facades.Queue().Job(&jobs.TestErr{}, []contractsqueue.Arg{
		{Type: "string", Value: "queue-test-all-special-fail"},
		{Type: "string", Value: now},
	}).Dispatch(); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	if err := facades.Queue().Job(&jobs.TestClaim{}, []contractsqueue.Arg{
		{Type: "string", Value: reclaimMarker},
		{Type: "int", Value: reclaimSleep},
	}).OnQueue(reclaimQueue).Dispatch(); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, "success", http.Json{
		"queued":          true,
		"type":            "all-special",
		"connection":      facades.Config().GetString("queue.default", "sync"),
		"delay_second":    delaySeconds,
		"reclaim_sleep":   reclaimSleep,
		"reclaim_queue":   reclaimQueue,
		"reclaim_marker":  reclaimMarker,
		"contains":        []string{"delay", "fail", "reclaim"},
		"reclaim_test_tip": "kill worker before ack, wait retry_after, then start another worker",
	})
}

func (c *QueueTestController) Result(ctx http.Context) http.Response {
	return response.Success(ctx, "success", http.Json{
		"test_result":       jobs.TestResult,
		"test_err_result":   jobs.TestErrResult,
		"test_claim_result": jobs.TestClaimResult,
	})
}

func (c *QueueTestController) Reset(ctx http.Context) http.Response {
	jobs.TestResult = nil
	jobs.TestErrResult = nil
	jobs.TestClaimResult = nil

	return response.Success(ctx, "success", http.Json{
		"reset": true,
	})
}
