package mcp

import (
	"context"
	"fmt"
)

// NativeSessionMetaTools provides meta-tools for session management that are
// injected into MCP server sessions for introspection and control.
type NativeSessionMetaTools struct{}

// NewNativeSessionMetaTools creates a new native session meta tools handler.
func NewNativeSessionMetaTools() *NativeSessionMetaTools {
	return &NativeSessionMetaTools{}
}

// SessionInfo holds metadata about the current session.
type SessionInfo struct {
	SessionID      string `json:"sessionId"`
	ToolCount      int    `json:"toolCount"`
	ServerCount    int    `json:"serverCount"`
	Uptime         string `json:"uptime"`
	WorkingSetSize int    `json:"workingSetSize"`
}

// ListSessionsMeta returns meta-information about the current session.
func (nsmt *NativeSessionMetaTools) ListSessionsMeta(ctx context.Context, sessionID string, workingSet *SessionWorkingSet) (*SessionInfo, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	info := &SessionInfo{
		SessionID: sessionID,
	}

	if workingSet != nil {
		info.WorkingSetSize = workingSet.Size()
		info.ToolCount = workingSet.Size()
	}

	return info, nil
}

// GetToolCountByServer returns the number of tools available per server.
func (nsmt *NativeSessionMetaTools) GetToolCountByServer(ctx context.Context, inventory *CachedInventory) (map[string]int, error) {
	snapshot, err := inventory.GetSnapshot()
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, s := range snapshot.Servers {
		if count, ok := snapshot.ToolCounts[s.Name]; ok {
			counts[s.Name] = count
		}
	}
	return counts, nil
}
