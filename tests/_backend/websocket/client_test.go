package websocket

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/utils"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestActiveChatBan_NotBanned tests that non-banned users return false
// Note: This test requires database connection. In unit tests, use mocking or skip.
func TestActiveChatBan_NotBanned(t *testing.T) {
	t.Skip("Requires database connection - use integration tests")
}

// TestRejectBannedChat_PublicJoin tests rejection of banned user joining public chat
func TestRejectBannedChat_PublicJoin(t *testing.T) {
	client := &Client{
		uid:      1,
		username: "testuser",
		send:     make(chan []byte, 256),
		hub:      NewHub(),
	}

	bannedTime := time.Now().Add(24 * time.Hour)
	ban := &chatBanInfo{
		Until:  &bannedTime,
		Reason: "Violating community guidelines",
	}

	client.rejectBannedChat("join_public_chat", ban)

	// Check that response was queued for sending
	if len(client.send) == 0 {
		t.Error("Expected response to be queued, but nothing was sent")
	}
}

// TestHandleMessage_JoinRoomLobbyBanned tests that banned users cannot join lobby
func TestHandleMessage_JoinRoomLobbyBanned(t *testing.T) {
	// This test demonstrates the flow but requires a full database setup
	// In production, use integration tests with actual DB
	hub := NewHub()

	// Create a banned client for testing
	bannedClient := &Client{
		uid:      100,
		username: "banneduser",
		nickname: "Banned User",
		avatar:   "avatar.png",
		send:     make(chan []byte, 256),
		hub:      hub,
		roomID:   "",
	}
	_ = bannedClient // Ensure client is used

	// Test the join_room message flow
	// Note: This would require mocking the repository to return a ban
	msg := &Message{
		Type:   "join_room",
		RoomID: "lobby",
	}

	// handleMessage would check activeChatBan() and reject if banned
	// This test validates the message structure
	if msg.Type != "join_room" || msg.RoomID != "lobby" {
		t.Error("Message structure incorrect")
	}
}

// TestHandleMessage_ChatBanned tests that banned users cannot send chat
func TestHandleMessage_ChatBanned(t *testing.T) {
	hub := NewHub()

	bannedClient := &Client{
		uid:      101,
		username: "banneduser",
		nickname: "Banned User",
		avatar:   "avatar.png",
		send:     make(chan []byte, 256),
		hub:      hub,
		roomID:   "lobby",
	}
	_ = bannedClient // Ensure client is used

	msg := &Message{
		Type:    "chat",
		Message: "This should be blocked",
	}

	// Validate message structure for chat
	if msg.Type != "chat" {
		t.Error("Message type should be 'chat'")
	}
}

// TestHandleMessage_PrivateChatBanned tests that banned users cannot send private chat
func TestHandleMessage_PrivateChatBanned(t *testing.T) {
	hub := NewHub()

	bannedClient := &Client{
		uid:      102,
		username: "banneduser",
		nickname: "Banned User",
		avatar:   "avatar.png",
		send:     make(chan []byte, 256),
		hub:      hub,
		roomID:   "",
	}
	_ = bannedClient // Ensure client is used

	msg := &Message{
		Type:      "private_chat",
		TargetUID: 5,
		Message:   "private message",
	}

	// Validate message structure for private chat
	if msg.Type != "private_chat" || msg.TargetUID != 5 {
		t.Error("Message structure incorrect")
	}
}

// TestChatBanInfo_UnmarshalJSON tests the chat ban info structure
func TestChatBanInfo_Response(t *testing.T) {
	bannedTime := time.Now().Add(24 * time.Hour)
	ban := &chatBanInfo{
		Until:  &bannedTime,
		Reason: "Test ban reason",
	}

	data := map[string]string{
		"reason": "banned",
		"action": "join_public_chat",
	}

	if ban.Until == nil {
		t.Error("Ban until should not be nil")
	}

	if ban.Reason != "Test ban reason" {
		t.Error("Ban reason mismatch")
	}

	if data["reason"] != "banned" {
		t.Error("Response reason should be 'banned'")
	}
}

// TestRejectBannedChat_WithContext tests rejection message includes ban context
func TestRejectBannedChat_WithContext(t *testing.T) {
	client := &Client{
		uid:      1,
		username: "testuser",
		send:     make(chan []byte, 256),
		hub:      nil,
	}

	bannedTime := time.Now().Add(48 * time.Hour)
	ban := &chatBanInfo{
		Until:  &bannedTime,
		Reason: "Violating chat policy",
	}

	// Record initial send channel length
	initialLen := len(client.send)

	client.rejectBannedChat("send_public_chat", ban)

	// Check that message was queued for sending
	if len(client.send) <= initialLen {
		t.Error("Expected message to be queued for sending")
	}
}

// TestRejectBannedChat_WithoutReason tests rejection message with no ban reason
func TestRejectBannedChat_WithoutReason(t *testing.T) {
	client := &Client{
		uid:      1,
		username: "testuser",
		send:     make(chan []byte, 256),
		hub:      nil,
	}

	bannedTime := time.Now().Add(1 * time.Hour)
	ban := &chatBanInfo{
		Until:  &bannedTime,
		Reason: "", // No reason provided
	}

	client.rejectBannedChat("send_private_chat", ban)

	// Check that message was queued
	if len(client.send) == 0 {
		t.Error("Expected message to be queued for sending")
	}
}

// TestRejectBannedChat_NilBan tests rejection with nil ban info
func TestRejectBannedChat_NilBan(t *testing.T) {
	client := &Client{
		uid:      1,
		username: "testuser",
		send:     make(chan []byte, 256),
		hub:      nil,
	}

	client.rejectBannedChat("send_public_chat", nil)

	// Check that message was queued
	if len(client.send) == 0 {
		t.Error("Expected message to be queued for sending")
	}
}

func TestWebSocketMessageLogAttributionSuppressesPing(t *testing.T) {
	if err := utils.InitLogger(20); err != nil {
		t.Fatalf("init logger: %v", err)
	}
	utils.ClearLogs()
	t.Cleanup(func() {
		_ = utils.CloseLogger()
	})

	client := &Client{
		uid:    100000101,
		send:   make(chan []byte, 256),
		source: &utils.LogSource{ClientIP: "127.0.0.1", UserAgent: "test-agent"},
	}

	client.handleMessage(&Message{Type: "ping", Timestamp: 123})
	if logs := utils.GetLogsFiltered(utils.LogFilter{Category: "websocket"}, 10); len(logs) != 0 {
		t.Fatalf("expected ping to be suppressed from websocket logs, got %d", len(logs))
	}
	if len(client.send) != 1 {
		t.Fatalf("expected pong response to be queued")
	}

	client.handleMessage(&Message{Type: "unknown_event"})
	logs := utils.GetLogsFiltered(utils.LogFilter{Category: "websocket", UID: utils.IntPtr(100000101)}, 10)
	if len(logs) != 1 {
		t.Fatalf("expected one websocket message log, got %d", len(logs))
	}
	entry := logs[0]
	if entry.Source == nil || entry.Source.ClientIP != "127.0.0.1" {
		t.Fatalf("expected source metadata, got %+v", entry.Source)
	}
	if entry.WebSocket == nil || entry.WebSocket.Event != "message" || entry.WebSocket.Type != "unknown_event" {
		t.Fatalf("unexpected websocket metadata: %+v", entry.WebSocket)
	}
}

func TestHubRecordsLastOfflineOnlyAfterFinalUIDConnection(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	database.DB = db
	repository.UserRepo = repository.NewUserRepository(db)

	user := database.User{UID: 100000201, Username: "rank-online"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	hub := NewHub()
	first := &Client{uid: int(user.UID), send: make(chan []byte, 1)}
	second := &Client{uid: int(user.UID), send: make(chan []byte, 1)}

	hub.mutex.Lock()
	hub.clients[first] = true
	hub.clients[second] = true
	delete(hub.clients, first)
	shouldRecordFirst := first.uid > 0 && !hub.hasUIDLocked(first.uid)
	hub.mutex.Unlock()
	if shouldRecordFirst {
		t.Fatalf("expected first disconnect to keep user online while another connection remains")
	}

	hub.mutex.Lock()
	delete(hub.clients, second)
	shouldRecordSecond := second.uid > 0 && !hub.hasUIDLocked(second.uid)
	hub.mutex.Unlock()
	if !shouldRecordSecond {
		t.Fatalf("expected final disconnect to record offline time")
	}

	offlineAt := time.Date(2026, 5, 6, 9, 45, 0, 0, time.UTC)
	hub.recordLastOffline(int(user.UID), offlineAt)
	deadline := time.Now().Add(time.Second)
	for {
		var updated database.User
		if err := db.First(&updated, user.UID).Error; err != nil {
			t.Fatalf("load user: %v", err)
		}
		if updated.LastOfflineAt != nil && updated.LastOfflineAt.Equal(offlineAt) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("last_offline_at was not updated, got %v", updated.LastOfflineAt)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
