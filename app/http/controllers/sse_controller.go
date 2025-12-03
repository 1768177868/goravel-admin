package controllers

import (
	"fmt"
	nethttp "net/http"
	"time"

	"github.com/goravel/framework/contracts/http"
)

/*********************************
SSE (Server-Sent Events) 示例

1. SSE 是一种服务器向客户端推送数据的技术
2. 比 WebSocket 更简单，但只支持单向通信（服务器到客户端）
3. 使用标准的 HTTP 协议，不需要升级连接

使用方法：
1. 在浏览器中打开：http://localhost:3000/sse
2. 或者使用 EventSource API：
   const eventSource = new EventSource('/sse');
   eventSource.onmessage = function(event) {
     console.log('收到消息:', event.data);
   };

4. 或者使用 curl 测试：
   curl -N http://localhost:3000/sse

特点：
- 自动重连：如果连接断开，浏览器会自动重连
- 简单易用：不需要额外的库，使用标准 HTTP
- 单向通信：只能服务器向客户端发送数据
- 文本格式：只能发送文本数据（JSON 字符串）
 ********************************/

type SseController struct {
	// Dependent services
}

func NewSseController() *SseController {
	return &SseController{
		// Inject services
	}
}

// Server SSE 服务器端实现
// 向客户端推送实时数据
func (r *SseController) Server(ctx http.Context) http.Response {
	// 设置 SSE 响应头
	writer := ctx.Response().Writer()
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("Access-Control-Allow-Origin", "*") // 允许跨域

	// 发送初始连接消息
	fmt.Fprintf(writer, "data: %s\n\n", `{"type":"connected","message":"SSE连接已建立"}`)
	if flusher, ok := writer.(nethttp.Flusher); ok {
		flusher.Flush()
	}

	// 创建一个 ticker，每秒发送一次数据
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// 创建一个通道用于检测客户端断开连接
	clientGone := ctx.Request().Origin().Context().Done()

	// 计数器
	counter := 0

	for {
		select {
		case <-clientGone:
			// 客户端断开连接
			fmt.Println("客户端断开连接")
			return nil
		case <-ticker.C:
			counter++

			// 构造消息数据（JSON 格式）
			message := fmt.Sprintf(`{"type":"message","counter":%d,"timestamp":"%s","data":"这是第 %d 条消息"}`,
				counter,
				time.Now().Format("2006-01-02 15:04:05"),
				counter)

			// 发送 SSE 消息
			// SSE 格式：data: {内容}\n\n
			fmt.Fprintf(writer, "data: %s\n\n", message)

			// 刷新缓冲区，确保数据立即发送
			if flusher, ok := writer.(nethttp.Flusher); ok {
				flusher.Flush()
			}

			// 示例：发送10条消息后发送完成消息
			if counter >= 10 {
				fmt.Fprintf(writer, "data: %s\n\n", `{"type":"completed","message":"消息发送完成"}`)
				if flusher, ok := writer.(nethttp.Flusher); ok {
					flusher.Flush()
				}
				// 可以选择继续发送或关闭连接
				// return nil // 取消注释此行可以发送10条后关闭连接
			}
		}
	}
}

// StreamData 示例：推送自定义数据流
// 可以根据业务需求推送不同类型的数据
func (r *SseController) StreamData(ctx http.Context) http.Response {
	writer := ctx.Response().Writer()
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("Access-Control-Allow-Origin", "*")

	// 获取查询参数（可选）
	topic := ctx.Request().Query("topic", "default")

	// 发送初始消息
	fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", "init", fmt.Sprintf(`{"topic":"%s","message":"开始推送数据"}`, topic))
	if flusher, ok := writer.(nethttp.Flusher); ok {
		flusher.Flush()
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	clientGone := ctx.Request().Origin().Context().Done()
	counter := 0

	for {
		select {
		case <-clientGone:
			return nil
		case <-ticker.C:
			counter++

			// 使用 event 字段指定事件类型
			eventType := "update"
			if counter%5 == 0 {
				eventType = "notification"
			}

			data := fmt.Sprintf(`{"topic":"%s","counter":%d,"event":"%s","time":"%s"}`,
				topic, counter, eventType, time.Now().Format(time.RFC3339))

			// SSE 格式：event: {事件类型}\ndata: {数据}\n\n
			fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", eventType, data)

			if flusher, ok := writer.(nethttp.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}
