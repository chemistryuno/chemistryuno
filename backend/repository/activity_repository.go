package repository

import (
	"chemistryuno/backend/database"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ActivityRepository handles CRUD for Activity and GameVersion, plus token operations.
type ActivityRepository struct {
	db *gorm.DB
}

func NewActivityRepository() *ActivityRepository {
	return &ActivityRepository{db: database.DB}
}

// ---- GameVersion ----

func (r *ActivityRepository) ListVersions() ([]database.GameVersion, error) {
	var versions []database.GameVersion
	err := r.db.Order("start_date desc").Find(&versions).Error
	return versions, err
}

func (r *ActivityRepository) CreateVersion(v *database.GameVersion) error {
	return r.db.Create(v).Error
}

func (r *ActivityRepository) UpdateVersion(id uint, updates map[string]interface{}) error {
	return r.db.Model(&database.GameVersion{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ActivityRepository) GetVersion(id uint) (*database.GameVersion, error) {
	var v database.GameVersion
	err := r.db.First(&v, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &v, err
}

// ---- Activity ----

func (r *ActivityRepository) ListActivities(adminView bool) ([]database.Activity, error) {
	var acts []database.Activity
	q := r.db.Order("start_time desc")
	if !adminView {
		now := time.Now().UTC()
		q = q.Where("is_active = ? AND start_time <= ? AND end_time >= ?", true, now, now)
	}
	err := q.Find(&acts).Error
	return acts, err
}

func (r *ActivityRepository) GetActivity(id uint) (*database.Activity, error) {
	var act database.Activity
	err := r.db.First(&act, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &act, err
}

func (r *ActivityRepository) CreateActivity(act *database.Activity) error {
	if err := r.checkConflict(act.Type, act.StartTime, act.EndTime, 0); err != nil {
		return err
	}
	return r.db.Create(act).Error
}

func (r *ActivityRepository) UpdateActivity(id uint, updates map[string]interface{}) error {
	act, err := r.GetActivity(id)
	if err != nil || act == nil {
		return errors.New("activity not found")
	}
	startTime := act.StartTime
	endTime := act.EndTime
	if v, ok := updates["start_time"]; ok {
		startTime = v.(time.Time)
	}
	if v, ok := updates["end_time"]; ok {
		endTime = v.(time.Time)
	}
	if err := r.checkConflict(act.Type, startTime, endTime, id); err != nil {
		return err
	}
	return r.db.Model(&database.Activity{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ActivityRepository) ToggleActivity(id uint, isActive bool) error {
	return r.db.Model(&database.Activity{}).Where("id = ?", id).Update("is_active", isActive).Error
}

// GetActiveActivitiesByType returns currently running activities of a given type.
func (r *ActivityRepository) GetActiveActivitiesByType(actType string) ([]database.Activity, error) {
	var acts []database.Activity
	now := time.Now().UTC()
	err := r.db.Where("type = ? AND is_active = ? AND start_time <= ? AND end_time >= ?",
		actType, true, now, now).Find(&acts).Error
	return acts, err
}

// checkConflict returns an error if a same-type activity overlaps the given time range.
// excludeID != 0 skips that record (for updates).
func (r *ActivityRepository) checkConflict(actType string, start, end time.Time, excludeID uint) error {
	q := r.db.Model(&database.Activity{}).
		Where("type = ? AND is_active = ? AND start_time < ? AND end_time > ?", actType, true, end, start)
	if excludeID != 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("conflict: another active activity of the same type overlaps this time range")
	}
	return nil
}

// ---- DailyActivityToken (double-points quota) ----

func (r *ActivityRepository) GetTokenUsage(uid, activityID uint, date string) (*database.DailyActivityToken, error) {
	var token database.DailyActivityToken
	err := r.db.Where("uid = ? AND activity_id = ? AND date = ?", uid, activityID, date).
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &token, err
}

// GetDailyLimit reads the daily_limit field from the activity's params JSON.
// Returns defaultLimit if the field is absent or unparseable.
func GetDailyLimit(act *database.Activity, defaultLimit int) int {
	if len(act.Params) == 0 {
		return defaultLimit
	}
	var params map[string]interface{}
	if err := json.Unmarshal(act.Params, &params); err != nil {
		return defaultLimit
	}
	if v, ok := params["daily_limit"]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		}
	}
	return defaultLimit
}

// IncrementToken atomically upserts and increments the usage count.
func (r *ActivityRepository) IncrementToken(uid, activityID uint, date string) error {
	token, err := r.GetTokenUsage(uid, activityID, date)
	if err != nil {
		return err
	}
	if token == nil {
		return r.db.Create(&database.DailyActivityToken{
			UID:        uid,
			ActivityID: activityID,
			Date:       date,
			UsedCount:  1,
		}).Error
	}
	return r.db.Model(&database.DailyActivityToken{}).
		Where("uid = ? AND activity_id = ? AND date = ?", uid, activityID, date).
		Update("used_count", gorm.Expr("used_count + 1")).Error
}
