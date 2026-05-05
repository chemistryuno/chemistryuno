package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chemistryuno/backend/anticheat"
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAdminEnforcementTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.FuelCompensationRecord{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	if err := database.MigrateCheatTables(db); err != nil {
		t.Fatalf("migrate cheat tables: %v", err)
	}
	database.DB = db

	if err := db.Create(&database.User{UID: 1001, Username: "player", Role: "user"}).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}

	repo := repository.NewCheatRepository(db)
	handler := NewAnticheatHandler(&anticheat.System{Repository: repo})
	router := gin.New()
	router.POST("/api/admin/anticheat/ban", func(c *gin.Context) {
		c.Set("uid", 9001)
		handler.BanFromAnticheatPanel(c)
	})
	router.POST("/api/admin/anticheat/unban", func(c *gin.Context) {
		c.Set("uid", 9001)
		handler.UnbanFromAnticheatPanel(c)
	})
	return router, db
}

func TestAnticheatPanelBanWritesAuditLog(t *testing.T) {
	router, db := setupAdminEnforcementTest(t)
	body := map[string]any{
		"player_uid":   1001,
		"banned_until": time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		"reason":       "manual anticheat enforcement",
		"room_id":      "room-a",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/anticheat/ban", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected ban success, got %d: %s", w.Code, w.Body.String())
	}

	var audit database.CheatAuditLog
	if err := db.Where("player_uid = ? AND event_type = ?", 1001, "ban").First(&audit).Error; err != nil {
		t.Fatalf("load ban audit: %v", err)
	}
	if audit.OperatorUID == nil || *audit.OperatorUID != 9001 {
		t.Fatalf("expected operator uid 9001, got %#v", audit.OperatorUID)
	}
	if audit.Remark != "manual anticheat enforcement" || audit.RoomID != "room-a" {
		t.Fatalf("unexpected audit context: remark=%q room=%q", audit.Remark, audit.RoomID)
	}
}

func TestAnticheatPanelUnbanWritesAuditLog(t *testing.T) {
	router, db := setupAdminEnforcementTest(t)
	body := map[string]any{
		"player_uid": 1001,
		"reason":     "appeal accepted manually",
		"room_id":    "room-b",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/anticheat/unban", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected unban success, got %d: %s", w.Code, w.Body.String())
	}

	var audit database.CheatAuditLog
	if err := db.Where("player_uid = ? AND event_type = ?", 1001, "unban").First(&audit).Error; err != nil {
		t.Fatalf("load unban audit: %v", err)
	}
	if audit.NewStatus != "revoked" || audit.Remark != "appeal accepted manually" {
		t.Fatalf("unexpected unban audit: status=%q remark=%q", audit.NewStatus, audit.Remark)
	}
}
