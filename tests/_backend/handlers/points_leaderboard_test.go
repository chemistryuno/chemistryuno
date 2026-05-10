package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupLeaderboardTest(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.LevelConfig{}, &database.Bounty{}); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	database.DB = db
	repository.UserRepo = repository.NewUserRepository(db)
	repository.BountyRepo = repository.NewBountyRepository()
	websocket.GlobalHub = nil
	t.Cleanup(func() {
		websocket.GlobalHub = nil
	})

	if err := db.Create(&database.LevelConfig{Level: 1, RequiredXP: 0, Tier: "bronze", TierName: "Bronze"}).Error; err != nil {
		t.Fatalf("create level config: %v", err)
	}

	lastOffline := time.Date(2026, 5, 6, 8, 30, 0, 0, time.UTC)
	users := make([]database.User, 0, 102)
	users = append(users, database.User{
		UID:           1001,
		Username:      "ranked",
		Nickname:      "Ranked",
		Points:        1500,
		MonthlyPoints: 1200,
		Level:         1,
		LastOfflineAt: &lastOffline,
	})
	for i := 0; i < 100; i++ {
		uid := uint(1100 + i)
		users = append(users, database.User{
			UID:           uid,
			Username:      "ranked-filler-" + strconv.Itoa(i),
			Nickname:      "Ranked Filler " + strconv.Itoa(i),
			Points:        1400 - float64(i),
			MonthlyPoints: 1100 - float64(i),
			Level:         1,
		})
	}
	users = append(users, database.User{UID: 2001, Username: "self", Nickname: "Self", Points: 900, MonthlyPoints: 900, Level: 1})
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}

	router := gin.New()
	router.GET("/api/points/leaderboard", func(c *gin.Context) {
		c.Set("uid", 2001)
		GetLeaderboard(c)
	})
	return router
}

func TestGetLeaderboardIncludesLastOfflineAt(t *testing.T) {
	router := setupLeaderboardTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/points/leaderboard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload struct {
		Leaderboard []map[string]interface{} `json:"leaderboard"`
		Self        map[string]interface{}   `json:"self"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Leaderboard) == 0 {
		t.Fatalf("expected leaderboard rows")
	}
	lastOffline, ok := payload.Leaderboard[0]["last_offline_at"].(string)
	if !ok || lastOffline == "" {
		t.Fatalf("expected leaderboard row last_offline_at, got %#v", payload.Leaderboard[0]["last_offline_at"])
	}
	if _, err := time.Parse(time.RFC3339, lastOffline); err != nil {
		t.Fatalf("expected RFC3339 last_offline_at, got %q: %v", lastOffline, err)
	}
	if _, exists := payload.Self["last_offline_at"]; !exists {
		t.Fatalf("expected self last_offline_at field")
	}
}
