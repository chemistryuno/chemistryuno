package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/team"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"

	"github.com/gin-gonic/gin"
)

var teamChatUpgrader = ws.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// teamChatRoom tracks active WS connections per team.
type teamChatRoom struct {
	mu      sync.RWMutex
	conns   map[uint]*ws.Conn // uid -> conn
}

var (
	teamChatRooms   = make(map[uint]*teamChatRoom) // teamID -> room
	teamChatRoomsMu sync.RWMutex
)

func getOrCreateTeamChatRoom(teamID uint) *teamChatRoom {
	teamChatRoomsMu.Lock()
	defer teamChatRoomsMu.Unlock()
	r, ok := teamChatRooms[teamID]
	if !ok {
		r = &teamChatRoom{conns: make(map[uint]*ws.Conn)}
		teamChatRooms[teamID] = r
	}
	return r
}

func broadcastToTeamChat(teamID uint, data []byte) {
	teamChatRoomsMu.RLock()
	r, ok := teamChatRooms[teamID]
	teamChatRoomsMu.RUnlock()
	if !ok {
		return
	}
	r.mu.RLock()
	conns := make(map[uint]*ws.Conn, len(r.conns))
	for uid, c := range r.conns {
		conns[uid] = c
	}
	r.mu.RUnlock()
	for uid, c := range conns {
		if err := c.WriteMessage(ws.TextMessage, data); err != nil {
			log.Printf("[TeamChat] write error uid=%d: %v", uid, err)
		}
	}
}

// TeamChat upgrades the connection to a WebSocket for team chat.
func TeamChat(c *gin.Context) {
	uid := uint(c.GetInt("uid"))
	state := team.GlobalManager.GetTeamByUID(uid)
	if state == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "不在任何队伍中"})
		return
	}
	teamID := state.Team.ID

	conn, err := teamChatUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[TeamChat] upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Register connection.
	room := getOrCreateTeamChatRoom(teamID)
	room.mu.Lock()
	room.conns[uid] = conn
	room.mu.Unlock()
	defer func() {
		room.mu.Lock()
		delete(room.conns, uid)
		room.mu.Unlock()
	}()

	// Send history first.
	var history []database.TeamChatMessage
	database.DB.Where("team_id = ?", teamID).Order("created_at asc").Limit(50).Find(&history)
	if histJSON, err := json.Marshal(map[string]interface{}{
		"type":     "team_chat_history",
		"messages": history,
	}); err == nil {
		conn.WriteMessage(ws.TextMessage, histJSON)
	}

	// Read loop.
	for {
		conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// Verify sender is still in the team.
		if team.GlobalManager.GetTeamByUID(uid) == nil {
			break
		}
		var incoming struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(msgBytes, &incoming); err != nil || incoming.Content == "" {
			continue
		}

		// Persist message.
		msg := database.TeamChatMessage{
			TeamID:  teamID,
			UID:     uid,
			Content: incoming.Content,
		}
		database.DB.Create(&msg)

		// Broadcast to all team members.
		payload, _ := json.Marshal(map[string]interface{}{
			"type":       "team_chat_message",
			"id":         msg.ID,
			"team_id":    teamID,
			"uid":        uid,
			"content":    incoming.Content,
			"created_at": msg.CreatedAt,
		})
		broadcastToTeamChat(teamID, payload)
	}
}
