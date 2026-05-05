package anticheat

import (
	"chemistryuno/backend/database"
	"testing"
	"time"
)

func TestNormalizeReplayEvidenceAnchorAndNavigation(t *testing.T) {
	anchor := NormalizeReplayEvidenceAnchor(database.ReplayEvidenceAnchor{
		RoomID:           "room-1",
		GameHistoryID:    77,
		EventIndex:       3,
		EventID:          "evt-3",
		PlayerUID:        42,
		EventTimestampMs: 1710000000000,
		ActionSummary:    "played H2O",
	})

	if anchor.ReplayID != "77" {
		t.Fatalf("expected replay id to fall back to game history id, got %q", anchor.ReplayID)
	}
	if anchor.EvidencePrecision != EvidencePrecisionOperation {
		t.Fatalf("expected operation precision, got %q", anchor.EvidencePrecision)
	}
	if anchor.CompatibilityLevel != EvidenceCompatibilityExact {
		t.Fatalf("expected exact compatibility, got %q", anchor.CompatibilityLevel)
	}
	if anchor.NavigationURL != "/replay/77?event_id=evt-3&event_index=3&timestamp_ms=1710000000000&uid=42" {
		t.Fatalf("unexpected navigation url: %s", anchor.NavigationURL)
	}
}

func TestEvidenceAnchorFromHistoricalReplayEvent(t *testing.T) {
	at := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	anchor := EvidenceAnchorFromReplayEvent("room-2", "88", map[string]interface{}{
		"event_index":    float64(5),
		"event":          "draw_card",
		"uid":            float64(1001),
		"timestamp":      at.Format(time.RFC3339Nano),
		"action_summary": "draw one card",
	})

	if anchor.EventIndex != 5 || anchor.EventID != "" {
		t.Fatalf("expected compatibility index anchor, got index=%d eventID=%q", anchor.EventIndex, anchor.EventID)
	}
	if anchor.EventType != "draw_card" || anchor.PlayerUID != 1001 {
		t.Fatalf("unexpected event metadata: %+v", anchor)
	}
	if anchor.EventTimestampMs != at.UnixMilli() {
		t.Fatalf("expected timestamp %d, got %d", at.UnixMilli(), anchor.EventTimestampMs)
	}
	if anchor.CompatibilityLevel != EvidenceCompatibilityIndex {
		t.Fatalf("expected index compatibility, got %q", anchor.CompatibilityLevel)
	}
}

