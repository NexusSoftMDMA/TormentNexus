package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type localSquadMember struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"createdAt"`
}

type localSquadState struct {
	Members       []localSquadMember `json:"members"`
	IndexerActive bool               `json:"indexerActive"`
	BrainActive   bool               `json:"brainActive"`
	UpdatedAt     int64              `json:"updatedAt"`
}

type localSquadManager struct {
	mu        sync.Mutex
	state     localSquadState
	statePath string
}

func newLocalSquadManager(workDir string) *localSquadManager {
	return &localSquadManager{
		statePath: filepath.Join(workDir, "squad_state.json"),
		state: localSquadState{
			Members: []localSquadMember{},
		},
	}
}

func (m *localSquadManager) load() {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.statePath)
	if err != nil {
		return
	}

	json.Unmarshal(data, &m.state)
}

func (m *localSquadManager) save() {
	m.state.UpdatedAt = time.Now().UnixMilli()
	data, _ := json.MarshalIndent(m.state, "", "  ")
	os.WriteFile(m.statePath, data, 0o644)
}

func (m *localSquadManager) List() []localSquadMember {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state.Members
}

func (m *localSquadManager) Spawn(role string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := "member-" + time.Now().Format("20060102-150405")
	m.state.Members = append(m.state.Members, localSquadMember{
		ID:        id,
		Role:      role,
		Status:    "active",
		CreatedAt: time.Now().UnixMilli(),
	})
	m.save()
	return id
}

func (m *localSquadManager) Kill(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, member := range m.state.Members {
		if member.ID == id {
			m.state.Members = append(m.state.Members[:i], m.state.Members[i+1:]...)
			m.save()
			return true
		}
	}
	return false
}

func (m *localSquadManager) ToggleIndexer(active bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.IndexerActive = active
	m.save()
}

func (m *localSquadManager) GetIndexerStatus() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state.IndexerActive
}
