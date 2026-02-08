package websocket

import (
	"encoding/json"
	"log"
	"sync"
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
		log.Printf("JSON 序列化失败: %v", err)
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
			log.Printf("客户端 %d 已连接", client.uid)
			h.BroadcastOnlineCount()

			if h.OnRegister != nil {
				h.OnRegister(client)
			}

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// 从所有房间中移除
				for roomID, clients := range h.rooms {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.rooms, roomID)
					}
				}
			}
			h.mutex.Unlock()
			log.Printf("客户端 %d 已断开", client.uid)
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

func (h *Hub) JoinRoom(client *Client, roomID string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
	client.roomID = roomID

	log.Printf("用户 %d 加入房间 %s (房间内共 %d 人)", client.uid, roomID, len(h.rooms[roomID]))

	// 打印房间内所有用户
	var uids []int
	for c := range h.rooms[roomID] {
		uids = append(uids, c.uid)
	}
	log.Printf("房间 %s 当前成员 UIDs: %v", roomID, uids)
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
			log.Printf("用户 %d 离开房间 %s (房间剩余 %d 人)", client.uid, roomID, len(clients))
			if len(clients) == 0 {
				delete(h.rooms, client.roomID)
				log.Printf("房间 %s 已清空并从 Hub 中删除", roomID)
			}
		}
		client.roomID = ""
	}
}

// BroadcastToRoom 将消息广播到指定房间
func (h *Hub) BroadcastToRoom(roomID string, message interface{}) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("消息序列化失败: %v", err)
		return
	}

	if clients, ok := h.rooms[roomID]; ok {
		log.Printf("广播消息到房间 %s (共 %d 个客户端)", roomID, len(clients))
		sentCount := 0
		for client := range clients {
			select {
			case client.send <- data:
				sentCount++
			default:
				log.Printf("发送失败，注销客户端 %d", client.uid)
				h.unregister <- client
			}
		}
		log.Printf("成功发送消息给 %d/%d 个客户端", sentCount, len(clients))
	} else {
		log.Printf("警告: 房间 %s 不存在，无法广播消息", roomID)
	}
}

// SendToUID 发送消息给指定用户的所有连接
func (h *Hub) SendToUID(uid int, message interface{}) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("消息序列化失败: %v", err)
		return
	}

	for client := range h.clients {
		if client.uid == uid {
			select {
			case client.send <- data:
			default:
				h.unregister <- client
			}
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
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("消息序列化失败: %v", err)
		return
	}

	for client := range h.clients {
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
