package websocket

import (
	"chemistryuno/backend/repository"
	"chemistryuno/backend/utils"
	"encoding/json"
	"log"
	"strconv"
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
	source   *utils.LogSource
}

type Message struct {
	Type      string      `json:"type"`
	RoomID    string      `json:"room_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	UID       int         `json:"uid,omitempty"`
	TargetUID int         `json:"target_uid,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"` // 用于ping/pong时间戳
}

type chatBanInfo struct {
	Until  *time.Time
	Reason string
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

func (c *Client) SetSource(source *utils.LogSource) {
	c.source = source
}

func (c *Client) logWebSocket(event string, msgType string, level string) {
	roomID := c.roomID
	utils.LogStructured(utils.LogEntry{
		Level:     level,
		Category:  "websocket",
		Message:   "websocket " + event + " uid=" + strconv.Itoa(c.uid),
		UID:       utils.IntPtr(c.uid),
		AuthState: "authenticated",
		Source:    c.source,
		WebSocket: &utils.LogWebSocket{
			Event:  event,
			Type:   msgType,
			RoomID: roomID,
		},
	})
}

func (c *Client) activeChatBan() (*chatBanInfo, bool) {
	bannedUntil, _, reason, err := repository.NewUserRepository().CheckBanStatus(uint(c.uid))
	if err != nil {
		log.Printf("failed to check ban status for websocket user %d: %v", c.uid, err)
		return nil, false
	}
	if bannedUntil == nil || !bannedUntil.After(time.Now()) {
		return nil, false
	}
	return &chatBanInfo{Until: bannedUntil, Reason: reason}, true
}

func (c *Client) rejectBannedChat(action string, ban *chatBanInfo) {
	data := map[string]string{
		"reason": "banned",
		"action": action,
	}
	if ban != nil {
		if ban.Until != nil {
			data["banned_until"] = ban.Until.Format(time.RFC3339)
		}
		if ban.Reason != "" {
			data["ban_reason"] = ban.Reason
		}
	}
	c.Send(Message{
		Type:    "chat_blocked",
		Message: "账号封禁期间无法使用聊天功能，请提交申诉或等待封禁结束。",
		Data:    data,
	})
}

func (c *Client) ReadPump() {
	defer func() {
		c.logWebSocket("disconnect", "", "INFO")
		// 防止在hub已关闭或roomID为空时仍尝试LeaveRoom
		if c.hub != nil {
			c.hub.LeaveRoom(c)
			c.hub.unregister <- c
		}
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
				log.Printf("❌ WebSocket 错误: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("❌ 消息解析失败: %v", err)
			continue
		}

		c.handleMessage(&msg)
	}
}

// Send 发送接口消息到客户端队列
func (c *Client) Send(msg interface{}) {
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("❌ 序列化失败: %v", err)
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
	if msg.Type != "ping" {
		log.Printf("📡 消息类型: %s (用户 %d)", msg.Type, c.uid)
		c.logWebSocket("message", msg.Type, "INFO")
	}

	switch msg.Type {
	case "ping":
		// 处理ping请求，立即返回pong响应
		c.Send(Message{
			Type:      "pong",
			Timestamp: msg.Timestamp, // 将客户端发送的时间戳原样返回
		})
		return

	case "join_room":
		if msg.RoomID == "lobby" {
			if ban, banned := c.activeChatBan(); banned {
				c.rejectBannedChat("join_public_chat", ban)
				return
			}
		}
		log.Printf("✅ 用户 %d 加入房间 %s", c.uid, msg.RoomID)
		c.hub.JoinRoom(c, msg.RoomID)
		displayName := c.nickname
		if displayName == "" {
			displayName = c.username
		}
		c.hub.BroadcastToRoom(msg.RoomID, Message{
			Type:    "player_joined",
			UID:     c.uid,
			Message: "安全门开启：研究员 " + displayName + " 已进入实验室。",
		})
		log.Printf("📢 广播 player_joined: 用户 %d 房间 %s", c.uid, msg.RoomID)

	case "leave_room":
		log.Printf("✅ 用户 %d 离开房间 %s", c.uid, c.roomID)
		roomID := c.roomID
		c.hub.LeaveRoom(c)
		displayName := c.nickname
		if displayName == "" {
			displayName = c.username
		}
		c.hub.BroadcastToRoom(roomID, Message{
			Type:    "player_left",
			UID:     c.uid,
			Message: "安全门关闭：研究员 " + displayName + " 已撤离实验室。",
		})
		log.Printf("📢 广播 player_left: 用户 %d 房间 %s", c.uid, roomID)

	case "chat":
		if ban, banned := c.activeChatBan(); banned {
			c.rejectBannedChat("send_public_chat", ban)
			return
		}
		targetRoom := c.roomID
		if targetRoom == "" {
			targetRoom = "lobby"
		}

		// 如果是大厅对话，保存到数据库
		if targetRoom == "lobby" {
			repository.ChatRepo.SaveChatMessage(uint(c.uid), c.username, c.nickname, c.avatar, msg.Message)
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
		if ban, banned := c.activeChatBan(); banned {
			c.rejectBannedChat("send_private_chat", ban)
			return
		}
		// 禁止游戏内（房间内）私聊
		if c.roomID != "" && c.roomID != "lobby" {
			c.hub.SendToUID(c.uid, Message{
				Type:    "error",
				Message: "量子纠缠协议异常：实验室内禁止开启加密私聊频道。",
			})
			return
		}
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

			// 检查是否是游戏邀请消息
			isGameInvite := false
			roomID := ""
			var inviteData map[string]interface{}
			if json.Unmarshal([]byte(msg.Message), &inviteData) == nil {
				if inviteType, ok := inviteData["type"].(string); ok && inviteType == "game_invite" {
					isGameInvite = true
					if rid, ok := inviteData["room_id"].(string); ok {
						roomID = rid
					}
				}
			}

			// 保存消息到数据库
			privateChatRepo := repository.NewPrivateChatRepository()
			err = privateChatRepo.SavePrivateMessage(uint(c.uid), uint(msg.TargetUID), msg.Message, isGameInvite, roomID)
			if err != nil {
				log.Printf("❌ 保存私聊消息失败: %v", err)
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
