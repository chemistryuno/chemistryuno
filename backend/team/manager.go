package team

import (
	"chemistryuno/backend/database"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"gorm.io/gorm"
)

const MaxTeamSize = 4

// TeamState holds the in-memory state of a team.
type TeamState struct {
	Team    database.Team
	Members []uint // UIDs
}

// Manager is the in-memory + DB team manager.
type Manager struct {
	mu     sync.RWMutex
	db     *gorm.DB
	// teamID -> TeamState
	teams  map[uint]*TeamState
	// uid -> teamID  (fast lookup)
	byUID  map[uint]uint
}

var GlobalManager *Manager

func NewManager(db *gorm.DB) *Manager {
	m := &Manager{
		db:    db,
		teams: make(map[uint]*TeamState),
		byUID: make(map[uint]uint),
	}
	GlobalManager = m
	return m
}

func generateInviteCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
}

// CreateTeam creates a new team with the caller as leader.
func (m *Manager) CreateTeam(leaderUID uint, name string) (*database.Team, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, inTeam := m.byUID[leaderUID]; inTeam {
		return nil, errors.New("请先退出当前队伍")
	}

	// Generate a unique invite code.
	var code string
	for {
		code = generateInviteCode()
		var count int64
		m.db.Model(&database.Team{}).Where("invite_code = ?", code).Count(&count)
		if count == 0 {
			break
		}
	}

	team := database.Team{
		Name:       name,
		InviteCode: code,
		LeaderUID:  leaderUID,
	}
	if err := m.db.Create(&team).Error; err != nil {
		return nil, err
	}
	member := database.TeamMember{TeamID: team.ID, UID: leaderUID}
	if err := m.db.Create(&member).Error; err != nil {
		return nil, err
	}

	state := &TeamState{
		Team:    team,
		Members: []uint{leaderUID},
	}
	m.teams[team.ID] = state
	m.byUID[leaderUID] = team.ID

	return &team, nil
}

// JoinTeam adds a user to a team via invite code.
func (m *Manager) JoinTeam(uid uint, inviteCode string) (*database.Team, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, inTeam := m.byUID[uid]; inTeam {
		return nil, errors.New("请先退出当前队伍")
	}

	// Look up team by invite code.
	var team database.Team
	if err := m.db.Where("invite_code = ?", inviteCode).First(&team).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("邀请码无效")
		}
		return nil, err
	}

	// Load or init state.
	state := m.ensureStateUnlocked(team)
	if len(state.Members) >= MaxTeamSize {
		return nil, errors.New("队伍已满")
	}
	for _, mid := range state.Members {
		if mid == uid {
			return nil, errors.New("已在队伍中")
		}
	}

	member := database.TeamMember{TeamID: team.ID, UID: uid}
	if err := m.db.Create(&member).Error; err != nil {
		return nil, err
	}
	state.Members = append(state.Members, uid)
	m.byUID[uid] = team.ID

	return &team, nil
}

// LeaveTeam removes a user from their current team. If the leader leaves, the team is disbanded.
func (m *Manager) LeaveTeam(uid uint) (disbanded bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	teamID, ok := m.byUID[uid]
	if !ok {
		return false, errors.New("不在任何队伍中")
	}
	state := m.teams[teamID]
	if state == nil {
		delete(m.byUID, uid)
		return false, errors.New("队伍不存在")
	}

	if state.Team.LeaderUID == uid {
		// Disband the whole team.
		return true, m.disbandUnlocked(state)
	}

	// Remove just this member.
	m.db.Where("team_id = ? AND uid = ?", teamID, uid).Delete(&database.TeamMember{})
	newMembers := make([]uint, 0, len(state.Members)-1)
	for _, mid := range state.Members {
		if mid != uid {
			newMembers = append(newMembers, mid)
		}
	}
	state.Members = newMembers
	delete(m.byUID, uid)
	return false, nil
}

// DisbandTeam disbands a team (only the leader can call this).
func (m *Manager) DisbandTeam(uid uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	teamID, ok := m.byUID[uid]
	if !ok {
		return errors.New("不在任何队伍中")
	}
	state := m.teams[teamID]
	if state == nil {
		return errors.New("队伍不存在")
	}
	if state.Team.LeaderUID != uid {
		return errors.New("只有队长可以解散队伍")
	}
	return m.disbandUnlocked(state)
}

func (m *Manager) disbandUnlocked(state *TeamState) error {
	teamID := state.Team.ID
	for _, mid := range state.Members {
		delete(m.byUID, mid)
	}
	delete(m.teams, teamID)
	m.db.Where("team_id = ?", teamID).Delete(&database.TeamMember{})
	return nil
}

// GetTeamByUID returns the team a user is in (nil if not in a team).
func (m *Manager) GetTeamByUID(uid uint) *TeamState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	teamID, ok := m.byUID[uid]
	if !ok {
		return nil
	}
	return m.teams[teamID]
}

// GetTeamByID returns the team by ID.
func (m *Manager) GetTeamByID(teamID uint) *TeamState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.teams[teamID]
}

// AreTeammates returns true if both UIDs are in the same team.
func (m *Manager) AreTeammates(uid1, uid2 uint) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t1, ok1 := m.byUID[uid1]
	t2, ok2 := m.byUID[uid2]
	return ok1 && ok2 && t1 == t2
}

// GetTeamChatRoomID returns the WebSocket room ID for a team's chat channel.
func GetTeamChatRoomID(teamID uint) string {
	return fmt.Sprintf("team_chat_%d", teamID)
}

// ensureStateUnlocked loads team state into memory if not already there. Must be called under lock.
func (m *Manager) ensureStateUnlocked(team database.Team) *TeamState {
	if state, ok := m.teams[team.ID]; ok {
		return state
	}
	var members []database.TeamMember
	m.db.Where("team_id = ?", team.ID).Find(&members)
	uids := make([]uint, 0, len(members))
	for _, mem := range members {
		uids = append(uids, mem.UID)
		m.byUID[mem.UID] = team.ID
	}
	state := &TeamState{Team: team, Members: uids}
	m.teams[team.ID] = state
	return state
}
