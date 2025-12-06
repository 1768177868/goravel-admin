package controllers

import (
	"time"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"

	"goravel/app/events"
	"goravel/app/http/response"
	"goravel/app/jobs"
)

// TestController 测试控制器 - 演示事件、监听器、Job的使用
type TestController struct {
}

func NewTestController() *TestController {
	return &TestController{}
}

// TestEvent 测试纯事件调用（监听器可能启用队列）
// GET /test/event
func (r *TestController) TestEvent(ctx http.Context) http.Response {
	// 触发订单发货事件
	// 这个事件会触发 SendShipmentNotification 监听器（已启用队列）
	err := facades.Event().Job(&events.OrderShipped{}, []event.Arg{
		{Type: "string", Value: "订单发货测试"},
		{Type: "int", Value: 12345},
	}).Dispatch()

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "事件调度失败: "+err.Error())
	}

	return ctx.Response().Success().Json(http.Json{
		"message": "事件已触发",
		"note":    "监听器 SendShipmentNotification 已启用队列，会在后台异步执行",
	})
}

// TestEventWithMultipleListeners 测试事件触发多个监听器（混合：同步+异步）
// GET /test/event-multiple
func (r *TestController) TestEventWithMultipleListeners(ctx http.Context) http.Response {
	userID := ctx.Request().QueryInt("user_id", 1001)
	email := ctx.Request().Query("email", "test@example.com")

	// 触发用户注册事件
	// 这个事件会触发多个监听器：
	// 1. SendWelcomeEmail - 启用队列（异步）
	// 2. CreateUserProfile - 启用队列（异步）
	// 3. InitializeUserSettings - 同步执行（立即生效）
	err := facades.Event().Job(&events.UserRegistered{}, []event.Arg{
		{Type: "int", Value: userID},
		{Type: "string", Value: email},
	}).Dispatch()

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "事件调度失败: "+err.Error())
	}

	return ctx.Response().Success().Json(http.Json{
		"message": "用户注册事件已触发",
		"user_id": userID,
		"email":   email,
		"listeners": []map[string]any{
			{
				"name":   "SendWelcomeEmail",
				"type":   "异步（队列）",
				"status": "已加入队列，后台执行",
			},
			{
				"name":   "CreateUserProfile",
				"type":   "异步（队列）",
				"status": "已加入队列，后台执行",
			},
			{
				"name":   "InitializeUserSettings",
				"type":   "同步",
				"status": "已立即执行",
			},
		},
	})
}

// TestEventOrderCreated 测试订单创建事件（混合场景）
// GET /test/event-order
func (r *TestController) TestEventOrderCreated(ctx http.Context) http.Response {
	orderID := ctx.Request().QueryInt("order_id", 2001)

	// 触发订单创建事件
	// 这个事件会触发多个监听器：
	// 1. UpdateInventory - 同步执行（库存需要立即更新）
	// 2. SendOrderNotification - 启用队列（异步发送通知）
	err := facades.Event().Job(&events.OrderCreated{}, []event.Arg{
		{Type: "int", Value: orderID},
		{Type: "string", Value: "订单创建测试"},
	}).Dispatch()

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "事件调度失败: "+err.Error())
	}

	return ctx.Response().Success().Json(http.Json{
		"message":  "订单创建事件已触发",
		"order_id": orderID,
		"listeners": []map[string]any{
			{
				"name":   "UpdateInventory",
				"type":   "同步",
				"status": "已立即执行，库存已更新",
			},
			{
				"name":   "SendOrderNotification",
				"type":   "异步（队列）",
				"status": "已加入队列，后台发送通知",
			},
		},
	})
}

// TestJob 测试纯Job调用
// GET /test/job
func (r *TestController) TestJob(ctx http.Context) http.Response {
	// 直接调度Job任务
	err := facades.Queue().Job(&jobs.SendEmail{}, []queue.Arg{
		{Type: "string", Value: "user@example.com"},
		{Type: "string", Value: "测试邮件"},
		{Type: "string", Value: "这是一封测试邮件"},
	}).Dispatch()

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "Job调度失败: "+err.Error())
	}

	return ctx.Response().Success().Json(http.Json{
		"message": "Job已加入队列",
		"job":     "SendEmail",
		"note":    "邮件将在后台异步发送",
	})
}

// TestJobWithDelay 测试延迟Job
// GET /test/job-delay
func (r *TestController) TestJobWithDelay(ctx http.Context) http.Response {
	// 延迟10秒后执行
	err := facades.Queue().Job(&jobs.SendEmail{}, []queue.Arg{
		{Type: "string", Value: "user@example.com"},
		{Type: "string", Value: "延迟邮件"},
		{Type: "string", Value: "这封邮件将在10秒后发送"},
	}).Delay(time.Now().Add(10 * time.Second)).Dispatch()

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "Job调度失败: "+err.Error())
	}

	return ctx.Response().Success().Json(http.Json{
		"message":    "延迟Job已加入队列",
		"job":        "SendEmail",
		"delay":      "10秒",
		"execute_at": time.Now().Add(10 * time.Second).Format("2006-01-02 15:04:05"),
	})
}

// TestJobOnQueue 测试指定队列的Job
// GET /test/job-queue
func (r *TestController) TestJobOnQueue(ctx http.Context) http.Response {
	queueName := ctx.Request().Query("queue", "emails")

	// 指定队列执行
	err := facades.Queue().Job(&jobs.SendEmail{}, []queue.Arg{
		{Type: "string", Value: "user@example.com"},
		{Type: "string", Value: "指定队列邮件"},
		{Type: "string", Value: "这封邮件将在指定队列中处理"},
	}).OnQueue(queueName).Dispatch()

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "Job调度失败: "+err.Error())
	}

	return ctx.Response().Success().Json(http.Json{
		"message": "Job已加入指定队列",
		"job":     "SendEmail",
		"queue":   queueName,
	})
}

// TestJobProcessImage 测试图片处理Job
// GET /test/job-image
func (r *TestController) TestJobProcessImage(ctx http.Context) http.Response {
	imagePath := ctx.Request().Query("path", "/uploads/image.jpg")

	err := facades.Queue().Job(&jobs.ProcessImage{}, []queue.Arg{
		{Type: "string", Value: imagePath},
	}).Dispatch()

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "Job调度失败: "+err.Error())
	}

	return ctx.Response().Success().Json(http.Json{
		"message":    "图片处理Job已加入队列",
		"job":        "ProcessImage",
		"image_path": imagePath,
		"note":       "图片将在后台异步处理（压缩、裁剪等）",
	})
}

// TestJobGenerateReport 测试生成报表Job
// GET /test/job-report
func (r *TestController) TestJobGenerateReport(ctx http.Context) http.Response {
	startDate := ctx.Request().Query("start_date", "2024-01-01")
	endDate := ctx.Request().Query("end_date", "2024-01-31")

	err := facades.Queue().Job(&jobs.GenerateReport{}, []queue.Arg{
		{Type: "string", Value: startDate},
		{Type: "string", Value: endDate},
	}).Dispatch()

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "Job调度失败: "+err.Error())
	}

	return ctx.Response().Success().Json(http.Json{
		"message":    "报表生成Job已加入队列",
		"job":        "GenerateReport",
		"start_date": startDate,
		"end_date":   endDate,
		"note":       "报表将在后台异步生成",
	})
}

// TestCombined 测试事件和Job结合使用的完整场景
// GET /test/combined
func (r *TestController) TestCombined(ctx http.Context) http.Response {
	userID := ctx.Request().QueryInt("user_id", 1001)
	email := ctx.Request().Query("email", "user@example.com")

	// 1. 触发用户注册事件（会触发多个监听器，部分启用队列）
	err := facades.Event().Job(&events.UserRegistered{}, []event.Arg{
		{Type: "int", Value: userID},
		{Type: "string", Value: email},
	}).Dispatch()

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "事件调度失败: "+err.Error())
	}

	// 2. 同时调度一个独立的Job任务
	err = facades.Queue().Job(&jobs.SendEmail{}, []queue.Arg{
		{Type: "string", Value: email},
		{Type: "string", Value: "欢迎注册"},
		{Type: "string", Value: "感谢您注册我们的服务！"},
	}).Dispatch()

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "Job调度失败: "+err.Error())
	}

	return ctx.Response().Success().Json(http.Json{
		"message":  "事件和Job结合使用测试",
		"scenario": "用户注册完整流程",
		"steps": []map[string]any{
			{
				"step":   1,
				"action": "触发 UserRegistered 事件",
				"result": "触发多个监听器",
			},
			{
				"step":   2,
				"action": "监听器 SendWelcomeEmail（队列）",
				"result": "已加入队列，异步发送欢迎邮件",
			},
			{
				"step":   3,
				"action": "监听器 CreateUserProfile（队列）",
				"result": "已加入队列，异步创建用户档案",
			},
			{
				"step":   4,
				"action": "监听器 InitializeUserSettings（同步）",
				"result": "已立即执行，用户设置已初始化",
			},
			{
				"step":   5,
				"action": "独立Job SendEmail",
				"result": "已加入队列，异步发送邮件",
			},
		},
		"note": "查看日志可以看到同步和异步执行的区别",
	})
}

// TestAll 测试所有功能
// GET /test/all
func (r *TestController) TestAll(ctx http.Context) http.Response {
	results := make(map[string]any)

	// 1. 测试事件
	err1 := facades.Event().Job(&events.OrderShipped{}, []event.Arg{
		{Type: "string", Value: "测试订单"},
		{Type: "int", Value: 9999},
	}).Dispatch()
	results["event"] = map[string]any{
		"status": func() string {
			if err1 != nil {
				return "失败: " + err1.Error()
			}
			return "成功"
		}(),
	}

	// 2. 测试Job
	err2 := facades.Queue().Job(&jobs.SendEmail{}, []queue.Arg{
		{Type: "string", Value: "test@example.com"},
		{Type: "string", Value: "测试主题"},
		{Type: "string", Value: "测试内容"},
	}).Dispatch()
	results["job"] = map[string]any{
		"status": func() string {
			if err2 != nil {
				return "失败: " + err2.Error()
			}
			return "成功"
		}(),
	}

	// 3. 测试事件+队列结合
	err3 := facades.Event().Job(&events.UserRegistered{}, []event.Arg{
		{Type: "int", Value: 8888},
		{Type: "string", Value: "test@example.com"},
	}).Dispatch()
	results["event_with_queue"] = map[string]any{
		"status": func() string {
			if err3 != nil {
				return "失败: " + err3.Error()
			}
			return "成功"
		}(),
	}

	return ctx.Response().Success().Json(http.Json{
		"message": "所有测试已完成",
		"results": results,
		"note":    "请查看日志和队列状态确认执行结果",
	})
}
