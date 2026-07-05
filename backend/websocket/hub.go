package websocket

import (
	"chemistryuno/backend/metrics"
	"chemistryuno/backend/repository"
	"encoding/json"
	"log"
	"sync"
	"time"
)

type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool // roomID -> clients
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
	OnRegister func(*Client)
}

func (h *Hub) SendToUser(uid int, message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ 消息序列化失败: %v", err)
		return
	}

	for client := range h.clients {
		if client.uid == uid {
			select {
			case client.send <- jsonData:
			default:
				h.unregister <- client
			}
		}
	}
}

var GlobalHub *Hub

func NewHub() *Hub {
	GlobalHub = &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	return GlobalHub
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Println("👤 用户连接")
			log.Printf("   ✅ 用户 %d 已连接到系统", client.uid)
			h.BroadcastOnlineCount()

			if h.OnRegister != nil {
				h.OnRegister(client)
			}

		case client := <-h.unregister:
			shouldRecordOffline := false
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				shouldRecordOffline = client.uid > 0 && !h.hasUIDLocked(client.uid)

				// 从所有房间中移除
				for roomID, clients := range h.rooms {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.rooms, roomID)
					}
				}
			}
			h.mutex.Unlock()
			log.Println("👤 用户断开")
			log.Printf("   ❌ 用户 %d 已断开连接", client.uid)
			if shouldRecordOffline {
				h.recordLastOffline(client.uid, time.Now())
			}
			h.BroadcastOnlineCount()

		case message := <-h.broadcast:
			h.mutex.RLock()
			clients := make([]*Client, 0, len(h.clients))
			for client := range h.clients {
				clients = append(clients, client)
			}
			h.mutex.RUnlock()

			for _, client := range clients {
				select {
				case client.send <- message:
				default:
					h.unregister <- client
				}
			}
		}
	}
}

// IsUIDOnline 检查指定 UID 是否在线
func (h *Hub) IsUIDOnline(uid int) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	for client := range h.clients {
		if client.uid == uid {
			return true
		}
	}
	return false
}

// Register 注册一个新客户端
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销一个客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) hasUIDLocked(uid int) bool {
	for client := range h.clients {
		if client.uid == uid {
			return true
		}
	}
	return false
}

func (h *Hub) recordLastOffline(uid int, offlineAt time.Time) {
	go func() {
		if err := repository.NewUserRepository().UpdateLastOfflineAt(uint(uid), offlineAt); err != nil {
			log.Printf("failed to update last_offline_at for uid %d: %v", uid, err)
		}
	}()
}

func (h *Hub) JoinRoom(client *Client, roomID string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
	client.roomID = roomID

	log.Println("🎮 房间操作 - 加入")
	log.Printf("   ✅ 用户 %d 加入房间 %s", client.uid, roomID)
	log.Printf("   📊 房间现有人数: %d", len(h.rooms[roomID]))

	// 打印房间内所有用户
	var uids []int
	for c := range h.rooms[roomID] {
		uids = append(uids, c.uid)
	}
	log.Printf("   👥 成员列表: %v", uids)
}

func (h *Hub) LeaveRoom(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.leaveRoomInternal(client)
}

// LeaveRoomByUID 让指定 UID 的用户离开其当前所在的房间
func (h *Hub) LeaveRoomByUID(uid int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for client := range h.clients {
		if client.uid == uid {
			h.leaveRoomInternal(client)
		}
	}
}

func (h *Hub) leaveRoomInternal(client *Client) {
	if client.roomID != "" {
		roomID := client.roomID
		if clients, ok := h.rooms[client.roomID]; ok {
			delete(clients, client)
			log.Println("🎮 房间操作 - 离开")
			log.Printf("   ✅ 用户 %d 离开房间 %s", client.uid, roomID)
			log.Printf("   📊 房间剩余人数: %d", len(clients))
			if len(clients) == 0 {
				delete(h.rooms, client.roomID)
				log.Printf("   🗑️  房间 %s 已清空，自动删除", roomID)
			}
		}
		client.roomID = ""
	}
}

// BroadcastToRoom 将消息广播到指定房间
func (h *Hub) BroadcastToRoom(roomID string, message interface{}) {
	start := time.Now()

	// Serialize message BEFORE acquiring lock (I/O operation)
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ 消息序列化失败: %v", err)
		return
	}

	// Acquire lock only to copy client list
	h.mutex.RLock()
	clients, ok := h.rooms[roomID]
	if !ok {
		h.mutex.RUnlock()
		log.Printf("   ⚠️  房间 %s 不存在，无法广播消息", roomID)
		return
	}

	// Copy client slice to release lock quickly
	clientList := make([]*Client, 0, len(clients))
	for client := range clients {
		clientList = append(clientList, client)
	}
	clientCount := len(clientList)
	h.mutex.RUnlock()

	// Send messages WITHOUT holding lock
	log.Println("📡 广播消息")
	log.Printf("   🎯 目标房间: %s", roomID)
	log.Printf("   👥 连接客户端数: %d", clientCount)

	sentCount := 0
	for _, client := range clientList {
		select {
		case client.send <- data:
			sentCount++
		default:
			log.Printf("   ⚠️  无法发送给用户 %d，准备断开连接", client.uid)
			h.unregister <- client
		}
	}

	duration := time.Since(start)
	log.Printf("   ✅ 消息发送完成: %d/%d 个客户端收到 (耗时: %v)", sentCount, clientCount, duration)

	// Record metrics
	metrics.WebSocketMessagesTotal.WithLabelValues(roomID).Inc()
	metrics.WebSocketBroadcastDuration.WithLabelValues(roomID).Observe(duration.Seconds())
}

// SendToUID 发送消息给指定用户的所有连接
func (h *Hub) SendToUID(uid int, message interface{}) {
	// Serialize message BEFORE acquiring lock
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ 消息序列化失败: %v", err)
		return
	}

	// Acquire lock only to find matching clients
	h.mutex.RLock()
	matchingClients := make([]*Client, 0)
	for client := range h.clients {
		if client.uid == uid {
			matchingClients = append(matchingClients, client)
		}
	}
	h.mutex.RUnlock()

	// Send messages WITHOUT holding lock
	for _, client := range matchingClients {
		select {
		case client.send <- data:
		default:
			h.unregister <- client
		}
	}
}

// IsUIDInRoom 检查指定用户是否在指定房间的连接列表中
func (h *Hub) IsUIDInRoom(roomID string, uid int) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clients, ok := h.rooms[roomID]; ok {
		for client := range clients {
			if client.uid == uid {
				return true
			}
		}
	}
	return false
}

func (h *Hub) BroadcastToAll(message interface{}) {
	// Serialize message BEFORE acquiring lock
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ 消息序列化失败: %v", err)
		return
	}

	// Acquire lock only to copy client list
	h.mutex.RLock()
	clientList := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clientList = append(clientList, client)
	}
	h.mutex.RUnlock()

	// Send messages WITHOUT holding lock
	for _, client := range clientList {
		select {
		case client.send <- data:
		default:
			h.unregister <- client
		}
	}
}

// BroadcastOnlineCount 广播当前在线人数
func (h *Hub) BroadcastOnlineCount() {
	h.mutex.RLock()
	count := len(h.clients)
	h.mutex.RUnlock()

	h.BroadcastToAll(Message{
		Type: "online_count",
		Data: count,
	})
}

// Stop 优雅停止Hub
func (h *Hub) Stop() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 关闭所有客户端连接
	for client := range h.clients {
		close(client.send)
		client.conn.Close()
	}

	// 清空数据
	h.clients = make(map[*Client]bool)
	h.rooms = make(map[string]map[*Client]bool)
}
