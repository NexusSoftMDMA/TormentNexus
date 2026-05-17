package session

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

type FleetMember struct {
	SessionID string `json:"sessionId"`
	PID       int    `json:"pid"`
	Status    string `json:"status"`
	StartedAt int64  `json:"startedAt"`
}

type FleetManager struct {
	mu      sync.RWMutex
	members map[string]*FleetMember
}

func NewFleetManager() *FleetManager {
	return &FleetManager{
		members: make(map[string]*FleetMember),
	}
}

func (fm *FleetManager) Register(sessionID string, pid int) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.members[sessionID] = &FleetMember{
		SessionID: sessionID,
		PID:       pid,
		Status:    "active",
		StartedAt: time.Now().UnixMilli(),
	}
	fmt.Printf("[Fleet] Registered PID %d for session %s\n", pid, sessionID)
}

func (fm *FleetManager) Unregister(sessionID string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	delete(fm.members, sessionID)
}

func (fm *FleetManager) CheckHealth() []string {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	var crashed []string
	for id, member := range fm.members {
		if !fm.isAlive(member.PID) {
			member.Status = "crashed"
			crashed = append(crashed, id)
			fmt.Printf("[Fleet] ⚠️ Session %s (PID %d) has crashed!\n", id, member.PID)
		}
	}
	return crashed
}

func (fm *FleetManager) isAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds. Need to signal 0 to check existence.
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (fm *FleetManager) GetFleetStatus() []*FleetMember {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var list []*FleetMember
	for _, m := range fm.members {
		copy := *m
		list = append(list, &copy)
	}
	return list
}
