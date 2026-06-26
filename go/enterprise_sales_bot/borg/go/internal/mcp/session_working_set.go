package mcp

import (
	"container/list"
	"sync"
	"time"
)

// SessionToolState tracks the state of a tool within a session's working set.
type SessionToolState struct {
	NamespacedName string      `json:"namespacedName"`
	ServerName     string      `json:"serverName"`
	ToolName       string      `json:"toolName"`
	Description    string      `json:"description"`
	InputSchema    interface{} `json:"inputSchema,omitempty"`
	LastUsedAt     int64       `json:"lastUsedAt"`
	UseCount       int         `json:"useCount"`
	IsPinned       bool        `json:"isPinned"`
}

// SessionWorkingSet manages the active tool set for a single session.
// It uses LRU eviction to keep the working set within a configurable size limit.
type SessionWorkingSet struct {
	mu        sync.RWMutex
	sessionID string
	maxSize   int
	tools     map[string]*list.Element // namespaced name -> list element
	order     *list.List               // LRU ordering (front = most recent)
}

// NewSessionWorkingSet creates a new working set for a session.
func NewSessionWorkingSet(sessionID string, maxSize int) *SessionWorkingSet {
	if maxSize <= 0 {
		maxSize = 50
	}
	return &SessionWorkingSet{
		sessionID: sessionID,
		maxSize:   maxSize,
		tools:     make(map[string]*list.Element),
		order:     list.New(),
	}
}

// Add adds or updates a tool in the working set. If the set exceeds maxSize,
// the least recently used unpinned tool is evicted.
func (ws *SessionWorkingSet) Add(tool SessionToolState) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// If already present, move to front and update
	if elem, ok := ws.tools[tool.NamespacedName]; ok {
		ws.order.MoveToFront(elem)
		elem.Value = &tool
		return
	}

	// Evict if over capacity
	for ws.order.Len() >= ws.maxSize {
		back := ws.order.Back()
		if back == nil {
			break
		}
		backTool := back.Value.(*SessionToolState)
		if backTool.IsPinned {
			// Move pinned tools to front instead of evicting
			ws.order.MoveToFront(back)
			break
		}
		ws.order.Remove(back)
		delete(ws.tools, backTool.NamespacedName)
	}

	elem := ws.order.PushFront(&tool)
	ws.tools[tool.NamespacedName] = elem
}

// Remove removes a tool from the working set.
func (ws *SessionWorkingSet) Remove(namespacedName string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if elem, ok := ws.tools[namespacedName]; ok {
		ws.order.Remove(elem)
		delete(ws.tools, namespacedName)
	}
}

// Get returns a tool from the working set, marking it as recently used.
func (ws *SessionWorkingSet) Get(namespacedName string) *SessionToolState {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if elem, ok := ws.tools[namespacedName]; ok {
		ws.order.MoveToFront(elem)
		tool := elem.Value.(*SessionToolState)
		tool.LastUsedAt = nowMs()
		tool.UseCount++
		return tool
	}
	return nil
}

// Contains checks if a tool is in the working set without modifying LRU order.
func (ws *SessionWorkingSet) Contains(namespacedName string) bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	_, ok := ws.tools[namespacedName]
	return ok
}

// List returns all tools in the working set, most recently used first.
func (ws *SessionWorkingSet) List() []SessionToolState {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	result := make([]SessionToolState, 0, ws.order.Len())
	for e := ws.order.Front(); e != nil; e = e.Next() {
		result = append(result, *e.Value.(*SessionToolState))
	}
	return result
}

// Pin marks a tool as pinned, preventing LRU eviction.
func (ws *SessionWorkingSet) Pin(namespacedName string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if elem, ok := ws.tools[namespacedName]; ok {
		tool := elem.Value.(*SessionToolState)
		tool.IsPinned = true
	}
}

// Unpin removes the pinned status from a tool.
func (ws *SessionWorkingSet) Unpin(namespacedName string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if elem, ok := ws.tools[namespacedName]; ok {
		tool := elem.Value.(*SessionToolState)
		tool.IsPinned = false
	}
}

// Size returns the current number of tools in the working set.
func (ws *SessionWorkingSet) Size() int {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.order.Len()
}

// Clear removes all tools from the working set.
func (ws *SessionWorkingSet) Clear() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.tools = make(map[string]*list.Element)
	ws.order = list.New()
}

// SessionID returns the session ID for this working set.
func (ws *SessionWorkingSet) SessionID() string {
	return ws.sessionID
}

func nowMs() int64 {
	return time.Now().UnixMilli()
}
