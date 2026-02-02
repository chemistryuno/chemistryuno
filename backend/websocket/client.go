package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 65536 // 增加最大消息限制为 64KB，防止加载游戏数据时断连
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	uid      int
	username string
	roomID   string
}

type Message struct {
	Type    string      `json:"type"`
	RoomID  string      `json:"room_id,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	UID     int         `json:"uid,omitempty"`
	Message string      `json:"message,omitempty"`
}

func NewClient(hub *Hub, conn *websocket.Conn, uid int, username string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		uid:      uid,
		username: username,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.LeaveRoom(c)
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("消息解析失败: %v", err)
			continue
		}

		c.handleMessage(&msg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case "join_room":
		c.hub.JoinRoom(c, msg.RoomID)
		c.hub.BroadcastToRoom(msg.RoomID, Message{
			Type:    "player_joined",
			UID:     c.uid,
			Message: c.username + " 加入了房间",
		})

	case "leave_room":
		roomID := c.roomID
		c.hub.LeaveRoom(c)
		c.hub.BroadcastToRoom(roomID, Message{
			Type:    "player_left",
			UID:     c.uid,
			Message: c.username + " 离开了房间",
		})

	case "chat":
		if c.roomID != "" {
			c.hub.BroadcastToRoom(c.roomID, Message{
				Type:    "chat",
				UID:     c.uid,
				Message: msg.Message,
				Data: map[string]string{
					"username": c.username,
				},
			})
		}
	}
}
