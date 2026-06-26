package pi

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
)

type EventSink func(event RunEvent)

type Runtime struct {
	cwd          string
	handlers     map[string]ToolHandler
	sessionStore *SessionStore
	mu           sync.RWMutex
}

func NewRuntime(cwd string, store *SessionStore) *Runtime {
	if store == nil {
		store = DefaultSessionStore(cwd)
	}
	return &Runtime{
		cwd:          cwd,
		handlers:     DefaultToolHandlers(),
		sessionStore: store,
	}
}

func (r *Runtime) RegisterTool(name string, handler ToolHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[name] = handler
}

func (r *Runtime) ExecuteTool(ctx context.Context, sessionID, toolName string, input json.RawMessage, sink EventSink) (*ToolResult, error) {
	r.mu.RLock()
	handler, ok := r.handlers[toolName]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
	if sink == nil {
		sink = func(RunEvent) {}
	}
	now := func() int64 { return time.Now().UnixMilli() }
	emit := func(event RunEvent) {
		event.Timestamp = now()
		if event.SessionID == "" {
			event.SessionID = sessionID
		}
		sink(event)
	}

	emit(RunEvent{Type: EventAgentStart, ToolName: toolName, Input: append(json.RawMessage(nil), input...)})
	emit(RunEvent{Type: EventTurnStart, ToolName: toolName, Input: append(json.RawMessage(nil), input...)})
	emit(RunEvent{Type: EventMessageStart, ToolName: toolName, Input: append(json.RawMessage(nil), input...)})
	emit(RunEvent{Type: EventMessageEnd, ToolName: toolName, Input: append(json.RawMessage(nil), input...)})
	emit(RunEvent{Type: EventToolExecutionStart, ToolName: toolName, Input: append(json.RawMessage(nil), input...)})

	result, err := handler(ctx, r.cwd, input)
	if result != nil {
		emit(RunEvent{Type: EventToolExecutionEnd, ToolName: toolName, Result: result})
		emit(RunEvent{Type: EventTurnEnd, ToolName: toolName, Result: result})
	}
	if err != nil {
		emit(RunEvent{Type: EventAgentEnd, ToolName: toolName, Result: result, Error: err.Error()})
	} else {
		emit(RunEvent{Type: EventAgentEnd, ToolName: toolName, Result: result})
	}

	if sessionID != "" {
		if appendErr := r.appendToolRun(sessionID, toolName, input, result); appendErr != nil {
			if err != nil {
				return result, fmt.Errorf("%w; session append failed: %v", err, appendErr)
			}
			return result, fmt.Errorf("session append failed: %w", appendErr)
		}
	}
	return result, err
}

func (r *Runtime) CreateSession(name string) (*SessionFile, error) {
	return r.sessionStore.Create(name, r.cwd)
}

func (r *Runtime) LoadSession(sessionID string) (*SessionFile, error) {
	return r.sessionStore.Load(sessionID)
}

func (r *Runtime) ListSessions() ([]SessionMetadata, error) {
	return r.sessionStore.List()
}

func (r *Runtime) ForkSession(sessionID, fromEntryID, name string) (*SessionFile, error) {
	return r.sessionStore.Fork(sessionID, fromEntryID, name)
}

func (r *Runtime) AppendUserText(sessionID, text string) (*SessionFile, error) {
	return r.sessionStore.AppendEntry(sessionID, SessionEntry{
		ID:        uuid.NewString(),
		Kind:      "message",
		Role:      "user",
		Text:      text,
		CreatedAt: time.Now().UnixMilli(),
	})
}

func (r *Runtime) appendToolRun(sessionID, toolName string, input json.RawMessage, result *ToolResult) error {
	session, err := r.sessionStore.Load(sessionID)
	if err != nil {
		return err
	}
	var parentID string
	if len(session.Entries) > 0 {
		parentID = session.Entries[len(session.Entries)-1].ID
	}
	callID := uuid.NewString()
	session.Entries = append(session.Entries,
		SessionEntry{
			ID:        callID,
			ParentID:  parentID,
			Kind:      "tool_call",
			ToolName:  toolName,
			ToolInput: append(json.RawMessage(nil), input...),
			CreatedAt: time.Now().UnixMilli(),
		},
		SessionEntry{
			ID:        uuid.NewString(),
			ParentID:  callID,
			Kind:      "tool_result",
			Role:      "toolResult",
			ToolName:  toolName,
			Result:    result,
			CreatedAt: time.Now().UnixMilli(),
		},
	)
	return r.sessionStore.Save(session)
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}
