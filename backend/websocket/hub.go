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
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
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

	log.Printf("用户 %d 加入房间 %s", client.uid, roomID)
}

func (h *Hub) LeaveRoom(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if client.roomID != "" {
		if clients, ok := h.rooms[client.roomID]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.rooms, client.roomID)
			}
		}
		log.Printf("用户 %d 离开房间 %s", client.uid, client.roomID)
		client.roomID = ""
	}
}

func (h *Hub) BroadcastToRoom(roomID string, message interface{}) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("消息序列化失败: %v", err)
		return
	}

	if clients, ok := h.rooms[roomID]; ok {
		for client := range clients {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
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
				close(client.send)
				delete(h.clients, client)
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
			close(client.send)
			delete(h.clients, client)
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
