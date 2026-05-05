package anticheat

import (
	"chemistryuno/backend/database"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	EvidencePrecisionOperation = "operation"
	EvidencePrecisionRoom      = "room"
	EvidenceCompatibilityExact = "exact"
	EvidenceCompatibilityIndex = "compatibility_index"
)

type ReplayEvidenceAnchor = database.ReplayEvidenceAnchor

func BuildReplayNavigationURL(anchor ReplayEvidenceAnchor) string {
	if anchor.GameHistoryID == 0 && anchor.ReplayID == "" {
		return ""
	}
	id := anchor.ReplayID
	if anchor.GameHistoryID > 0 {
		id = strconv.FormatUint(uint64(anchor.GameHistoryID), 10)
	}
	params := []string{}
	if anchor.EventID != "" {
		params = append(params, "event_id="+anchor.EventID)
	}
	if anchor.EventIndex > 0 {
		params = append(params, "event_index="+strconv.Itoa(anchor.EventIndex))
	}
	if anchor.EventTimestampMs > 0 {
		params = append(params, "timestamp_ms="+strconv.FormatInt(anchor.EventTimestampMs, 10))
	}
	if anchor.PlayerUID > 0 {
		params = append(params, "uid="+strconv.FormatUint(uint64(anchor.PlayerUID), 10))
	}
	if len(params) == 0 {
		return "/replay/" + id
	}
	return "/replay/" + id + "?" + strings.Join(params, "&")
}

func NormalizeReplayEvidenceAnchor(anchor ReplayEvidenceAnchor) ReplayEvidenceAnchor {
	if anchor.EvidencePrecision == "" {
		if anchor.EventID != "" || anchor.EventIndex > 0 || anchor.EventTimestampMs > 0 {
			anchor.EvidencePrecision = EvidencePrecisionOperation
		} else {
			anchor.EvidencePrecision = EvidencePrecisionRoom
		}
	}
	if anchor.CompatibilityLevel == "" {
		if anchor.EventID != "" {
			anchor.CompatibilityLevel = EvidenceCompatibilityExact
		} else {
			anchor.CompatibilityLevel = EvidenceCompatibilityIndex
		}
	}
	if anchor.ReplayID == "" && anchor.GameHistoryID > 0 {
		anchor.ReplayID = strconv.FormatUint(uint64(anchor.GameHistoryID), 10)
	}
	if anchor.NavigationURL == "" {
		anchor.NavigationURL = BuildReplayNavigationURL(anchor)
	}
	return anchor
}

func MarshalReplayEvidenceAnchor(anchor ReplayEvidenceAnchor) database.JSON {
	anchor = NormalizeReplayEvidenceAnchor(anchor)
	if anchor.RoomID == "" && anchor.ReplayID == "" && anchor.GameHistoryID == 0 {
		return nil
	}
	data, _ := json.Marshal(anchor)
	return data
}

func MarshalReplayEvidenceAnchors(anchors []ReplayEvidenceAnchor) database.JSON {
	if len(anchors) == 0 {
		return nil
	}
	normalized := make([]ReplayEvidenceAnchor, 0, len(anchors))
	for _, anchor := range anchors {
		normalized = append(normalized, NormalizeReplayEvidenceAnchor(anchor))
	}
	data, _ := json.Marshal(normalized)
	return data
}

func UnmarshalReplayEvidenceAnchor(raw database.JSON) (ReplayEvidenceAnchor, bool) {
	if len(raw) == 0 {
		return ReplayEvidenceAnchor{}, false
	}
	var anchor ReplayEvidenceAnchor
	if err := json.Unmarshal(raw, &anchor); err != nil {
		return ReplayEvidenceAnchor{}, false
	}
	return NormalizeReplayEvidenceAnchor(anchor), true
}

func EvidenceAnchorFromContext(context *DetectionContext) ReplayEvidenceAnchor {
	if context == nil {
		return ReplayEvidenceAnchor{}
	}
	anchor := context.PrimaryEvidence
	if anchor.RoomID == "" {
		anchor.RoomID = context.RoomID
	}
	if anchor.ReplayID == "" {
		anchor.ReplayID = firstEvidenceNonEmpty(context.ReplayID, context.RoomID)
	}
	if anchor.EventIndex == 0 {
		anchor.EventIndex = context.OperationIndex
	}
	if anchor.PlayerUID == 0 && context.PlayerUID > 0 {
		anchor.PlayerUID = uint(context.PlayerUID)
	}
	if anchor.EventTimestampMs == 0 && len(context.OperationTimes) > 0 {
		anchor.EventTimestampMs = context.OperationTimes[len(context.OperationTimes)-1].UnixMilli()
	}
	if anchor.ActionSummary == "" {
		anchor.ActionSummary = fmt.Sprintf("anticheat evidence for player %d", context.PlayerUID)
	}
	return NormalizeReplayEvidenceAnchor(anchor)
}

func EvidenceAnchorFromReplayEvent(roomID string, replayID string, event map[string]interface{}) ReplayEvidenceAnchor {
	anchor := ReplayEvidenceAnchor{
		RoomID:   roomID,
		ReplayID: firstEvidenceNonEmpty(replayID, roomID),
	}
	if v, ok := event["event_index"].(int); ok {
		anchor.EventIndex = v
	} else if f, ok := event["event_index"].(float64); ok {
		anchor.EventIndex = int(f)
	}
	if v, ok := event["event_id"].(string); ok {
		anchor.EventID = v
	}
	if v, ok := event["event"].(string); ok {
		anchor.EventType = v
	}
	if v, ok := event["uid"].(int); ok && v > 0 {
		anchor.PlayerUID = uint(v)
	} else if f, ok := event["uid"].(float64); ok && f > 0 {
		anchor.PlayerUID = uint(f)
	}
	if v, ok := event["unix_ms"].(int64); ok {
		anchor.EventTimestampMs = v
	} else if f, ok := event["unix_ms"].(float64); ok {
		anchor.EventTimestampMs = int64(f)
	}
	if v, ok := event["action_summary"].(string); ok {
		anchor.ActionSummary = v
	}
	if anchor.EventTimestampMs == 0 {
		if raw, ok := event["timestamp"].(string); ok {
			if parsed, err := time.Parse(time.RFC3339Nano, raw); err == nil {
				anchor.EventTimestampMs = parsed.UnixMilli()
			}
		}
	}
	return NormalizeReplayEvidenceAnchor(anchor)
}

func firstEvidenceNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
