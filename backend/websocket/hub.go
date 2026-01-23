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
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("客户端 %d 已连接", client.userID)

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
			log.Printf("客户端 %d 已断开", client.userID)

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

func (h *Hub) JoinRoom(client *Client, roomID string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
	client.roomID = roomID
	
	log.Printf("用户 %d 加入房间 %s", client.userID, roomID)
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
		log.Printf("用户 %d 离开房间 %s", client.userID, client.roomID)
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
