package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type localSwarmMission struct {
	MissionID string `json:"missionId"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"createdAt"`
}

type localSwarmState struct {
	Missions  []localSwarmMission `json:"missions"`
	UpdatedAt int64               `json:"updatedAt"`
}

type localSwarmManager struct {
	mu        sync.Mutex
	state     localSwarmState
	statePath string
}

func newLocalSwarmManager(workDir string) *localSwarmManager {
	return &localSwarmManager{
		statePath: filepath.Join(workDir, "swarm_state.json"),
		state: localSwarmState{
			Missions: []localSwarmMission{},
		},
	}
}

func (m *localSwarmManager) load() {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.statePath)
	if err != nil {
		return
	}

	json.Unmarshal(data, &m.state)
}

func (m *localSwarmManager) save() {
	m.state.UpdatedAt = time.Now().UnixMilli()
	data, _ := json.MarshalIndent(m.state, "", "  ")
	os.WriteFile(m.statePath, data, 0o644)
}

func (m *localSwarmManager) ListMissions() []localSwarmMission {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state.Missions
}

func (m *localSwarmManager) StartSwarm() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := "mission-" + time.Now().Format("20060102-150405")
	m.state.Missions = append(m.state.Missions, localSwarmMission{
		MissionID: id,
		Status:    "started",
		CreatedAt: time.Now().UnixMilli(),
	})
	m.save()
	return id
}

func (m *localSwarmManager) StopMission(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, mission := range m.state.Missions {
		if mission.MissionID == id {
			m.state.Missions[i].Status = "stopped"
			m.save()
			return true
		}
	}
	return false
}

func (m *localSwarmManager) DeleteMission(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, mission := range m.state.Missions {
		if mission.MissionID == id {
			m.state.Missions = append(m.state.Missions[:i], m.state.Missions[i+1:]...)
			m.save()
			return true
		}
	}
	return false
}
