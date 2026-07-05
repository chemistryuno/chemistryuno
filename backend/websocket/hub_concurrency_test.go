package websocket

import (
	"sync"
	"testing"
	"time"
)

func TestHubBroadcastConcurrency(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create test clients
	numClients := 50
	clients := make([]*Client, numClients)
	for i := 0; i < numClients; i++ {
		client := &Client{
			hub:  hub,
			conn: nil, // Mock connection not needed for this test
			send: make(chan []byte, 256),
			uid:  i + 1,
		}
		clients[i] = client
		hub.Register(client)
	}

	// Wait for registration to complete
	time.Sleep(100 * time.Millisecond)

	// Concurrent broadcast test
	var wg sync.WaitGroup
	numGoroutines := 20
	broadcastsPerGoroutine := 50

	// Broadcast to all clients concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < broadcastsPerGoroutine; j++ {
				message := Message{
					Type: "test",
					Data: map[string]interface{}{
						"goroutine": id,
						"iteration": j,
					},
				}
				hub.BroadcastToAll(message)
			}
		}(i)
	}

	// Concurrent client registration/unregistration
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				newClient := &Client{
					hub:  hub,
					conn: nil,
					send: make(chan []byte, 256),
					uid:  1000 + id*10 + j,
				}
				hub.Register(newClient)
				time.Sleep(5 * time.Millisecond)
				hub.Unregister(newClient)
			}
		}(i)
	}

	wg.Wait()

	// If we got here without panic, concurrency test passed
	t.Log("Concurrency test passed - no data races detected")
}

func TestHubRoomBroadcastConcurrency(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create clients and join rooms
	numClients := 30
	numRooms := 5
	clients := make([]*Client, numClients)

	for i := 0; i < numClients; i++ {
		client := &Client{
			hub:  hub,
			conn: nil,
			send: make(chan []byte, 256),
			uid:  i + 1,
		}
		clients[i] = client
		hub.Register(client)

		// Distribute clients across rooms
		roomID := "room-" + string(rune('A'+i%numRooms))
		hub.JoinRoom(client, roomID)
	}

	time.Sleep(100 * time.Millisecond)

	// Concurrent room broadcasts
	var wg sync.WaitGroup
	numGoroutines := 10
	broadcastsPerGoroutine := 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < broadcastsPerGoroutine; j++ {
				roomID := "room-" + string(rune('A'+j%numRooms))
				message := Message{
					Type: "room_test",
					Data: map[string]interface{}{
						"room":      roomID,
						"goroutine": id,
						"iteration": j,
					},
				}
				hub.BroadcastToRoom(roomID, message)
			}
		}(i)
	}

	// Concurrent room join/leave operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				clientIdx := (id*10 + j) % numClients
				client := clients[clientIdx]

				// Leave current room
				hub.LeaveRoom(client)

				// Join a different room
				newRoomID := "room-" + string(rune('A'+(id+j)%numRooms))
				hub.JoinRoom(client, newRoomID)

				time.Sleep(5 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	t.Log("Room broadcast concurrency test passed")
}

func TestHubSendToUIDConcurrency(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create clients with duplicate UIDs
	numClients := 40
	clients := make([]*Client, numClients)
	for i := 0; i < numClients; i++ {
		client := &Client{
			hub:  hub,
			conn: nil,
			send: make(chan []byte, 256),
			uid:  (i % 10) + 1, // UIDs 1-10, with multiple connections per UID
		}
		clients[i] = client
		hub.Register(client)
	}

	time.Sleep(100 * time.Millisecond)

	// Concurrent SendToUID operations
	var wg sync.WaitGroup
	numGoroutines := 15
	sendsPerGoroutine := 30

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < sendsPerGoroutine; j++ {
				targetUID := (j % 10) + 1
				message := Message{
					Type: "direct",
					Data: map[string]interface{}{
						"to":        targetUID,
						"goroutine": id,
						"iteration": j,
					},
				}
				hub.SendToUID(targetUID, message)
			}
		}(i)
	}

	wg.Wait()

	t.Log("SendToUID concurrency test passed")
}
