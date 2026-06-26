package memory

import "sync"

type Manager struct {
	mu       sync.Mutex
	memories []string
}

func NewManager() *Manager {
	return &Manager{
		memories: make([]string, 0),
	}
}

func (m *Manager) AddMemory(mem string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.memories = append(m.memories, mem)
}

func (m *Manager) GetMemories() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make([]string, len(m.memories))
	copy(res, m.memories)
	return res
}
