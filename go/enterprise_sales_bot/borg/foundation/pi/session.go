package pi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
)

type SessionMetadata struct {
	SessionID  string `json:"sessionId"`
	Name       string `json:"name,omitempty"`
	WorkingDir string `json:"workingDir"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}

type SessionEntry struct {
	ID        string          `json:"id"`
	ParentID  string          `json:"parentId,omitempty"`
	Kind      string          `json:"kind"`
	Role      string          `json:"role,omitempty"`
	Text      string          `json:"text,omitempty"`
	ToolName  string          `json:"toolName,omitempty"`
	ToolInput json.RawMessage `json:"toolInput,omitempty"`
	Result    *ToolResult     `json:"result,omitempty"`
	CreatedAt int64           `json:"createdAt"`
}

type SessionFile struct {
	Metadata SessionMetadata `json:"metadata"`
	Entries  []SessionEntry  `json:"entries"`
}

type sessionRecord struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type SessionStore struct {
	baseDir string
}

func NewSessionStore(baseDir string) *SessionStore {
	return &SessionStore{baseDir: baseDir}
}

func DefaultSessionStore(cwd string) *SessionStore {
	return NewSessionStore(filepath.Join(cwd, ".tormentnexus", "foundation", "sessions"))
}

func (s *SessionStore) BaseDir() string {
	return s.baseDir
}

func (s *SessionStore) Create(name, workingDir string) (*SessionFile, error) {
	if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create session directory: %w", err)
	}
	now := time.Now().UnixMilli()
	session := &SessionFile{
		Metadata: SessionMetadata{
			SessionID:  uuid.NewString(),
			Name:       name,
			WorkingDir: workingDir,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}
	if err := s.Save(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *SessionStore) Save(session *SessionFile) error {
	if session == nil {
		return fmt.Errorf("session is nil")
	}
	if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
		return fmt.Errorf("create session directory: %w", err)
	}
	if session.Metadata.SessionID == "" {
		session.Metadata.SessionID = uuid.NewString()
	}
	session.Metadata.UpdatedAt = time.Now().UnixMilli()
	path := s.Path(session.Metadata.SessionID)
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create session file: %w", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	writeRecord := func(recordType string, value any) error {
		payload, err := json.Marshal(value)
		if err != nil {
			return err
		}
		line, err := json.Marshal(sessionRecord{Type: recordType, Data: payload})
		if err != nil {
			return err
		}
		if _, err := writer.Write(append(line, '\n')); err != nil {
			return err
		}
		return nil
	}
	if err := writeRecord("session", session.Metadata); err != nil {
		return fmt.Errorf("write session metadata: %w", err)
	}
	for _, entry := range session.Entries {
		if err := writeRecord("entry", entry); err != nil {
			return fmt.Errorf("write session entry: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush session file: %w", err)
	}
	return nil
}

func (s *SessionStore) Load(sessionID string) (*SessionFile, error) {
	file, err := os.Open(s.Path(sessionID))
	if err != nil {
		return nil, fmt.Errorf("open session file: %w", err)
	}
	defer file.Close()

	session := &SessionFile{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record sessionRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, fmt.Errorf("decode session record: %w", err)
		}
		switch record.Type {
		case "session":
			if err := json.Unmarshal(record.Data, &session.Metadata); err != nil {
				return nil, fmt.Errorf("decode session metadata: %w", err)
			}
		case "entry":
			var entry SessionEntry
			if err := json.Unmarshal(record.Data, &entry); err != nil {
				return nil, fmt.Errorf("decode session entry: %w", err)
			}
			session.Entries = append(session.Entries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan session file: %w", err)
	}
	return session, nil
}

func (s *SessionStore) AppendEntry(sessionID string, entry SessionEntry) (*SessionFile, error) {
	session, err := s.Load(sessionID)
	if err != nil {
		return nil, err
	}
	if entry.ID == "" {
		entry.ID = uuid.NewString()
	}
	if entry.CreatedAt == 0 {
		entry.CreatedAt = time.Now().UnixMilli()
	}
	session.Entries = append(session.Entries, entry)
	if err := s.Save(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *SessionStore) Fork(sessionID, fromEntryID, name string) (*SessionFile, error) {
	session, err := s.Load(sessionID)
	if err != nil {
		return nil, err
	}
	forked, err := s.Create(name, session.Metadata.WorkingDir)
	if err != nil {
		return nil, err
	}
	if fromEntryID == "" && len(session.Entries) > 0 {
		fromEntryID = session.Entries[len(session.Entries)-1].ID
	}
	for _, entry := range session.Entries {
		forked.Entries = append(forked.Entries, entry)
		if entry.ID == fromEntryID {
			break
		}
	}
	if err := s.Save(forked); err != nil {
		return nil, err
	}
	return forked, nil
}

func (s *SessionStore) Path(sessionID string) string {
	return filepath.Join(s.baseDir, sessionID+".jsonl")
}

func (s *SessionStore) List() ([]SessionMetadata, error) {
	if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create session directory: %w", err)
	}
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("read session directory: %w", err)
	}
	result := make([]SessionMetadata, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}
		sessionID := entry.Name()[:len(entry.Name())-len(filepath.Ext(entry.Name()))]
		session, err := s.Load(sessionID)
		if err != nil {
			continue
		}
		result = append(result, session.Metadata)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].UpdatedAt > result[j].UpdatedAt
	})
	return result, nil
}
