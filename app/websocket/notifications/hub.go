package notifications

import (
	"encoding/json"
	"sync"
	"time"

	"goravel/app/models"
)

type NotificationHub struct {
	clients    map[uint]map[*notificationClient]bool
	register   chan *notificationClient
	unregister chan *notificationClient
	broadcast  chan *models.Notification
	stop       chan struct{} // 停止信号
	mu         sync.RWMutex
}

var hubInstance = newNotificationHub()

func init() {
	go hubInstance.run()
}

func Hub() *NotificationHub {
	return hubInstance
}

func newNotificationHub() *NotificationHub {
	return &NotificationHub{
		clients:    make(map[uint]map[*notificationClient]bool),
		register:   make(chan *notificationClient),
		unregister: make(chan *notificationClient),
		broadcast:  make(chan *models.Notification, 100),
		stop:       make(chan struct{}),
	}
}

func (h *NotificationHub) run() {
	for {
		select {
		case <-h.stop:
			// 收到停止信号，关闭所有客户端连接并退出
			h.mu.Lock()
			for _, adminClients := range h.clients {
				for client := range adminClients {
					close(client.send)
				}
			}
			h.clients = make(map[uint]map[*notificationClient]bool)
			h.mu.Unlock()
			return
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		case notification := <-h.broadcast:
			h.dispatch(notification)
		}
	}
}

func (h *NotificationHub) addClient(client *notificationClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.adminID]; !ok {
		h.clients[client.adminID] = make(map[*notificationClient]bool)
	}
	h.clients[client.adminID][client] = true
}

func (h *NotificationHub) removeClient(client *notificationClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if adminClients, ok := h.clients[client.adminID]; ok {
		if _, exists := adminClients[client]; exists {
			delete(adminClients, client)
			close(client.send)
		}
		if len(adminClients) == 0 {
			delete(h.clients, client.adminID)
		}
	}
}

func (h *NotificationHub) dispatch(notification *models.Notification) {
	payload := h.payload(notification)
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	if notification.ReceiverID == nil {
		h.mu.RLock()
		defer h.mu.RUnlock()
		for _, adminClients := range h.clients {
			for client := range adminClients {
				client.send <- data
			}
		}
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	if adminClients, ok := h.clients[*notification.ReceiverID]; ok {
		for client := range adminClients {
			client.send <- data
		}
	}
}

func (h *NotificationHub) payload(notification *models.Notification) map[string]any {
	var readAt *string
	if notification.ReadAt != nil {
		formatted := notification.ReadAt.Format(time.RFC3339)
		readAt = &formatted
	}

	var receiver uint
	if notification.ReceiverID != nil {
		receiver = *notification.ReceiverID
	}

	var sender uint
	if notification.SenderID != nil {
		sender = *notification.SenderID
	}

	return map[string]any{
		"id":          notification.ID,
		"title":       notification.Title,
		"content":     notification.Content,
		"type":        notification.Type,
		"sender_id":   sender,
		"receiver_id": receiver,
		"is_read":     notification.IsRead,
		"read_at":     readAt,
		"created_at":  notification.CreatedAt.Format(time.RFC3339),
	}
}

func (h *NotificationHub) Broadcast(notification *models.Notification) {
	select {
	case h.broadcast <- notification:
		// 成功发送
	case <-h.stop:
		// Hub 已停止，忽略广播
	}
}

// Stop 停止 NotificationHub，关闭所有连接并退出 goroutine
func (h *NotificationHub) Stop() {
	close(h.stop)
}

func (h *NotificationHub) Stats() (int, int) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	admins := len(h.clients)
	connections := 0
	for _, adminClients := range h.clients {
		connections += len(adminClients)
	}

	return admins, connections
}
