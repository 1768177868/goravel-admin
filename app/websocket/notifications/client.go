package notifications

import (
	"time"

	"github.com/gorilla/websocket"
)

type notificationClient struct {
	hub     *NotificationHub
	conn    *websocket.Conn
	send    chan []byte
	adminID uint
}

func newNotificationClient(hub *NotificationHub, conn *websocket.Conn, adminID uint) *notificationClient {
	return &notificationClient{
		hub:     hub,
		conn:    conn,
		send:    make(chan []byte, 256),
		adminID: adminID,
	}
}

func (c *notificationClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *notificationClient) writePump() {
	ticker := time.NewTicker(25 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *NotificationHub) RegisterConnection(conn *websocket.Conn, adminID uint) {
	client := newNotificationClient(h, conn, adminID)
	h.register <- client
	go client.writePump()
	go client.readPump()
}

