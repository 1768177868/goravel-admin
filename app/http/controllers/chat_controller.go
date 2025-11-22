package controllers

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/gorilla/websocket"
)

// 客户端身份结构体
type Client struct {
	Conn     *websocket.Conn // WebSocket连接
	Username string          // 客户端用户名（唯一标识）
	Groups   []string        // 客户端所属群ID
	lastPong time.Time       // 最后收到Pong的时间（用于检测超时）
}

// 消息结构体：新增system类型和Groups字段支持退群结果同步
type Message struct {
	Type    string   `json:"type"`    // 消息类型：all/group/private/leave_group/system
	From    string   `json:"from"`    // 发送方用户名
	To      string   `json:"to"`      // 接收方：群ID/用户名
	Content string   `json:"content"` // 消息内容
	Groups  []string `json:"groups"`  // 客户端最新群列表（系统消息用）
}

// 聊天室管理器
type ChatManager struct {
	Clients    map[*Client]bool // 所有在线客户端
	Register   chan *Client     // 注册客户端
	Unregister chan *Client     // 注销客户端
	Broadcast  chan []byte      // 消息广播通道
	mu         sync.Mutex       // 并发安全锁
	// 心跳配置
	PingInterval time.Duration // 发送Ping的间隔
	PongTimeout  time.Duration // Pong响应超时时间
}

// 初始化管理器：设置心跳间隔30秒，超时10秒
var chatManager = &ChatManager{
	Clients:      make(map[*Client]bool),
	Register:     make(chan *Client),
	Unregister:   make(chan *Client),
	Broadcast:    make(chan []byte),
	PingInterval: 30 * time.Second, // 每30秒发一次Ping
	PongTimeout:  10 * time.Second, // 10秒内没收到Pong则断开
}

func init() {
	go chatManager.Run()      // 启动核心管理器
	go chatManager.PingLoop() // 启动心跳检测循环
}

// 心跳检测循环，定时给所有客户端发Ping
func (m *ChatManager) PingLoop() {
	ticker := time.NewTicker(m.PingInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		for client := range m.Clients {
			// 检测Pong超时：最后收到Pong的时间超过（Ping间隔+超时时间）
			if time.Since(client.lastPong) > m.PingInterval+m.PongTimeout {
				m.Unregister <- client // 超时则注销
				continue
			}
			// 发送Ping帧（使用WebSocket原生Ping，客户端会自动回复Pong）
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				m.Unregister <- client // 发送失败则注销
			}
		}
		m.mu.Unlock()
	}
}

// 核心消息循环：处理注册、注销、广播、退群
func (m *ChatManager) Run() {
	for {
		select {
		case client := <-m.Register:
			m.mu.Lock()
			client.lastPong = time.Now() // 初始化最后Pong时间
			m.Clients[client] = true
			m.mu.Unlock()
			fmt.Printf("用户[%s]上线，当前在线：%d\n", client.Username, len(m.Clients))

		case client := <-m.Unregister:
			m.mu.Lock()
			if _, ok := m.Clients[client]; ok {
				delete(m.Clients, client)
				client.Conn.Close()
				fmt.Printf("用户[%s]下线，当前在线：%d\n", client.Username, len(m.Clients))
			}
			m.mu.Unlock()

		case msgBytes := <-m.Broadcast:
			// 解析消息
			var msg Message
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				fmt.Println("消息解析失败：", err)
				continue
			}

			m.mu.Lock()
			switch msg.Type {
			case "all": // 全员发送
				for client := range m.Clients {
					m.sendMsg(client.Conn, msg)
				}
			case "group": // 群发送
				for client := range m.Clients {
					for _, group := range client.Groups {
						if group == msg.To {
							m.sendMsg(client.Conn, msg)
							break
						}
					}
				}
			case "private": // 私人发送
				for client := range m.Clients {
					if client.Username == msg.To {
						m.sendMsg(client.Conn, msg)
						break
					}
				}
			case "leave_group": // 退群处理
				// 查找发送方客户端
				var targetClient *Client
				for client := range m.Clients {
					if client.Username == msg.From {
						targetClient = client
						break
					}
				}
				if targetClient == nil {
					fmt.Printf("退群失败：未找到用户[%s]\n", msg.From)
					m.mu.Unlock()
					return
				}
				// 校验并删除目标群
				targetGroup := msg.To
				groupIndex := -1
				for i, g := range targetClient.Groups {
					if g == targetGroup {
						groupIndex = i
						break
					}
				}
				if groupIndex == -1 {
					m.sendSystemMsg(targetClient.Conn, fmt.Sprintf("退群失败：未加入群[%s]", targetGroup), targetClient.Groups)
					m.mu.Unlock()
					return
				}
				// 安全删除切片元素
				targetClient.Groups = append(targetClient.Groups[:groupIndex], targetClient.Groups[groupIndex+1:]...)
				fmt.Printf("用户[%s]退出群[%s]，当前群列表：%v\n", msg.From, targetGroup, targetClient.Groups)
				// 返回退群成功消息
				m.sendSystemMsg(targetClient.Conn, fmt.Sprintf("退群成功：已退出群[%s]", targetGroup), targetClient.Groups)
			}
			m.mu.Unlock()
		}
	}
}

// 发送普通消息给客户端
func (m *ChatManager) sendMsg(conn *websocket.Conn, msg Message) {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("消息序列化失败：", err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		fmt.Println("消息发送失败：", err)
		conn.Close()
	}
}

// 发送系统消息（用于退群结果、状态通知）
func (m *ChatManager) sendSystemMsg(conn *websocket.Conn, content string, groups []string) {
	systemMsg := Message{
		Type:    "system",
		Content: content,
		Groups:  groups,
	}
	msgBytes, err := json.Marshal(systemMsg)
	if err != nil {
		fmt.Println("系统消息序列化失败：", err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		fmt.Println("系统消息发送失败：", err)
	}
}

// ChatController：处理WebSocket连接和HTTP消息发送
type ChatController struct{}

func NewChatController() *ChatController {
	return &ChatController{}
}

// Server 处理WebSocket连接
func (c *ChatController) Server(ctx http.Context) http.Response {
	upGrader := websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *nethttp.Request) bool {
			return true
		},
	}

	ws, err := upGrader.Upgrade(ctx.Response().Writer(), ctx.Request().Origin(), nil)
	if err != nil {
		return ctx.Response().String(http.StatusInternalServerError, err.Error())
	}
	defer ws.Close()

	// 身份注册
	_, registerBytes, err := ws.ReadMessage()
	if err != nil {
		return ctx.Response().String(http.StatusBadRequest, "身份注册失败")
	}
	type RegisterMsg struct {
		Username string   `json:"username"`
		Groups   []string `json:"groups"`
	}
	var register RegisterMsg
	if err := json.Unmarshal(registerBytes, &register); err != nil || register.Username == "" {
		return ctx.Response().String(http.StatusBadRequest, "身份格式错误")
	}

	// 初始化客户端
	client := &Client{
		Conn:     ws,
		Username: register.Username,
		Groups:   register.Groups,
		lastPong: time.Now(), // 初始化Pong时间
	}
	chatManager.Register <- client

	// 重写Pong处理器：绑定当前客户端的lastPong更新
	ws.SetPongHandler(func(s string) error {
		client.lastPong = time.Now()
		return nil
	})

	// 循环读取消息
	for {
		_, msgBytes, err := ws.ReadMessage()
		if err != nil {
			chatManager.Unregister <- client
			break
		}
		chatManager.Broadcast <- msgBytes
	}

	return nil
}

// HTTP消息发送参数结构体
type HttpChatMsg struct {
	Type    string `json:"type" form:"type"`
	From    string `json:"from" form:"from"`
	To      string `json:"to" form:"to"`
	Content string `json:"content" form:"content"`
}

// SendMsgByHttp 处理HTTP请求发送聊天室消息
// 全员发送消息（GET 请求）: http://localhost:3008/api/chat/send?type=all&from=系统通知&content=大家好，这是一条系统全员消息！
// 群发送消息（POST 请求，JSON 体）: http://localhost:3008/api/chat/send
//
//	{
//	    "type": "group",
//	    "from": "系统通知",
//	    "to": "group1",
//	    "content": "group1的各位成员，这是一条群专属通知！"
//	}

// 私人发送消息（POST 请求，JSON 体）: http://localhost:3008/api/chat/send
//
//	{
//	    "type": "private",
//	    "from": "系统通知",
//	    "to": "user1",
//	    "content": "user1你好，这是一条私人系统消息！"
//	}
func (c *ChatController) SendMsgByHttp(ctx http.Context) http.Response {
	var msgParam HttpChatMsg
	if err := ctx.Request().Bind(&msgParam); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, map[string]any{
			"code": 400,
			"msg":  "参数绑定失败：" + err.Error(),
			"data": nil,
		})
	}

	// 参数校验
	if msgParam.Type == "" || msgParam.From == "" || msgParam.Content == "" {
		return ctx.Response().Json(http.StatusBadRequest, map[string]any{
			"code": 400,
			"msg":  "类型/发送方/内容为必填参数",
			"data": nil,
		})
	}
	if (msgParam.Type == "group" || msgParam.Type == "private") && msgParam.To == "" {
		return ctx.Response().Json(http.StatusBadRequest, map[string]any{
			"code": 400,
			"msg":  "群/私人消息需指定接收方（To）",
			"data": nil,
		})
	}
	if msgParam.Type != "all" && msgParam.Type != "group" && msgParam.Type != "private" {
		return ctx.Response().Json(http.StatusBadRequest, map[string]any{
			"code": 400,
			"msg":  "类型仅支持all/group/private",
			"data": nil,
		})
	}

	// 构造并发送消息
	chatMsg := Message{
		Type:    msgParam.Type,
		From:    msgParam.From,
		To:      msgParam.To,
		Content: msgParam.Content,
	}
	msgBytes, err := json.Marshal(chatMsg)
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, map[string]any{
			"code": 500,
			"msg":  "消息序列化失败",
			"data": nil,
		})
	}
	chatManager.Broadcast <- msgBytes

	return ctx.Response().Json(http.StatusOK, map[string]any{
		"code": 200,
		"msg":  "消息发送成功",
		"data": chatMsg,
	})
}
