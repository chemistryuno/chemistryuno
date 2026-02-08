package websocket

import (
	"chemistryuno/repository"
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
	nickname string
	avatar   string
	roomID   string
}

type Message struct {
	Type      string      `json:"type"`
	RoomID    string      `json:"room_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	UID       int         `json:"uid,omitempty"`
	TargetUID int         `json:"target_uid,omitempty"`
	Message   string      `json:"message,omitempty"`
}

func NewClient(hub *Hub, conn *websocket.Conn, uid int, username string, nickname string, avatar string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		uid:      uid,
		username: username,
		nickname: nickname,
		avatar:   avatar,
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

// Send 发送接口消息到客户端队列
func (c *Client) Send(msg interface{}) {
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Send error: %v", err)
		return
	}
	c.send <- payload
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
	log.Printf("[WebSocket] User %d handling message type: %s", c.uid, msg.Type)

	switch msg.Type {
	case "join_room":
		log.Printf("[WebSocket] User %d joining room %s", c.uid, msg.RoomID)
		c.hub.JoinRoom(c, msg.RoomID)
		c.hub.BroadcastToRoom(msg.RoomID, Message{
			Type:    "player_joined",
			UID:     c.uid,
			Message: c.username + " 加入了房间",
		})
		log.Printf("[WebSocket] Broadcasted player_joined for user %d to room %s", c.uid, msg.RoomID)

	case "leave_room":
		log.Printf("[WebSocket] User %d leaving room %s", c.uid, c.roomID)
		roomID := c.roomID
		c.hub.LeaveRoom(c)
		c.hub.BroadcastToRoom(roomID, Message{
			Type:    "player_left",
			UID:     c.uid,
			Message: c.username + " 离开了房间",
		})
		log.Printf("[WebSocket] Broadcasted player_left for user %d from room %s", c.uid, roomID)

	case "chat":
		targetRoom := c.roomID
		if targetRoom == "" {
			targetRoom = "lobby"
		}

		// 如果是大厅对话，保存到数据库
		if targetRoom == "lobby" {
			repository.ChatRepo.SaveChatMessage(uint(c.uid), c.username, c.avatar, msg.Message)
		}

		c.hub.BroadcastToRoom(targetRoom, Message{
			Type:    "chat",
			UID:     c.uid,
			Message: msg.Message,
			Data: map[string]string{
				"username": c.username,
				"nickname": c.nickname,
				"avatar":   c.avatar,
			},
		})

	case "private_chat":
		if msg.TargetUID != 0 {
			// 检查是否是好友
			isFriend, err := repository.FriendshipRepo.IsFriend(uint(c.uid), uint(msg.TargetUID))
			if err != nil || !isFriend {
				// 发送系统通知告知不能私聊
				c.hub.SendToUID(c.uid, Message{
					Type:    "error",
					Message: "只有互为好友的研究员才能进行单向加密通信",
				})
				return
			}

			// 发送给目标用户
			payload := Message{
				Type:      "private_chat",
				UID:       c.uid,
				TargetUID: msg.TargetUID,
				Message:   msg.Message,
				Data: map[string]string{
					"username": c.username, "nickname": c.nickname, "avatar": c.avatar,
				},
			}
			c.hub.SendToUID(msg.TargetUID, payload)
			// 同时发送给自己，以便在发送者的 UI 上显示
			c.hub.SendToUID(c.uid, payload)
		}
	}
}
