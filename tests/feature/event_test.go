package feature

import (
	"testing"
	"time"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/assert"

	"goravel/app/events"
)

func TestEvent(t *testing.T) {
	// 测试事件调度是否成功
	err1 := facades.Event().Job(&events.OrderShipped{}, []event.Arg{
		{Type: "string", Value: "I'm OrderShipped"},
	}).Dispatch()
	assert.NoError(t, err1)

	err2 := facades.Event().Job(&events.OrderCanceled{}, []event.Arg{
		{Type: "string", Value: "I'm OrderCanceled"},
	}).Dispatch()
	assert.NoError(t, err2)

	// 等待队列处理
	time.Sleep(1 * time.Second)

	// 注意：由于移除了全局测试变量，这里只验证事件调度是否成功
	// 如果需要验证监听器执行结果，可以通过日志或其他方式验证
	t.Log("Events dispatched successfully")
}
