package memorystore

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/borghq/borg-go/internal/ai"
	"github.com/borghq/borg-go/internal/controlplane"
)

type ArchivedSessionMetadata struct {
	OriginalID     string `json:"originalId"`
	SourceTool     string `json:"sourceTool"`
	Title          string `json:"title"`
	Timestamp      int64  `json:"timestamp"`
	CompressedSize int64  `json:"compressedSize"`
}

type MemoryArchiver struct {
	workspaceRoot string
	archivePath   string
	vectorStore   *VectorStore
}

func NewMemoryArchiver(workspaceRoot string, vs *VectorStore) *MemoryArchiver {
	return &MemoryArchiver{
		workspaceRoot: workspaceRoot,
		archivePath:   filepath.Join(workspaceRoot, "data", "archives", "sessions.zip"),
		vectorStore:   vs,
	}
}

func (a *MemoryArchiver) ArchiveAndExtract(ctx context.Context, sessionData map[string]interface{}) (*ArchivedSessionMetadata, error) {
	sessionID, _ := sessionData["id"].(string)
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%d", time.Now().UnixMilli())
	}

	title, _ := sessionData["title"].(string)
	if title == "" {
		title = sessionID
	}

	sourceTool, _ := sessionData["sourceTool"].(string)
	if sourceTool == "" {
		sourceTool = "unknown"
	}

	transcript := a.formatToPlaintext(sessionData)
	if transcript == "" {
		return nil, fmt.Errorf("empty transcript or unrecognized format")
	}

	// 1. Extract Valuable Memories
	err := a.extractValuableMemories(ctx, sessionID, transcript, title)
	if err != nil {
		fmt.Printf("[Go Archiver] Memory extraction failed: %v\n", err)
	}

	// 2. Add to Compressed Archive
	compressedSize, err := a.writeToZip(sessionID, transcript, sourceTool)
	if err != nil {
		return nil, err
	}

	return &ArchivedSessionMetadata{
		OriginalID:     sessionID,
		SourceTool:     sourceTool,
		Title:          title,
		Timestamp:      time.Now().UnixMilli(),
		CompressedSize: compressedSize,
	}, nil
}

func (a *MemoryArchiver) formatToPlaintext(sessionData map[string]interface{}) string {
	messages, ok := sessionData["messages"].([]interface{})
	if !ok {
		messages, ok = sessionData["conversation"].([]interface{})
	}
	if !ok {
		return ""
	}

	var sb strings.Builder
	for i, m := range messages {
		msgMap, ok := m.(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := msgMap["role"].(string)
		if role == "" {
			role = "unknown"
		}

		content := ""
		if c, ok := msgMap["content"].(string); ok {
			content = c
		} else {
			jsonData, _ := json.Marshal(msgMap["content"])
			content = string(jsonData)
		}

		if i > 0 {
			sb.WriteString("\n\n---\n\n")
		}
		sb.WriteString(strings.ToUpper(role))
		sb.WriteString(": ")
		sb.WriteString(content)
	}

	return sb.String()
}

func (a *MemoryArchiver) extractValuableMemories(ctx context.Context, sessionID string, transcript string, title string) error {
	prompt := fmt.Sprintf(`
		Analyze the following session transcript titled "%s".
		Identify the MOST VALUABLE pieces of knowledge, decisions made, or technical discoveries.
		Ignore small talk, errors, or redundant steps.
		
		Return ONLY a JSON array of strings.
		
		TRANSCRIPT:
		%s
	`, title, transcript[:min(8000, len(transcript))])

	messages := []ai.Message{
		{Role: "user", Content: prompt},
	}

	resp, err := ai.AutoRoute(ctx, messages)
	if err != nil {
		return err
	}

	start := strings.Index(resp.Content, "[")
	end := strings.LastIndex(resp.Content, "]")
	if start == -1 || end == -1 {
		return fmt.Errorf("no JSON array found in LLM response")
	}

	var memories []string
	if err := json.Unmarshal([]byte(resp.Content[start:end+1]), &memories); err != nil {
		return err
	}

	if a.vectorStore != nil {
		for _, m := range memories {
			entry := controlplane.L2VaultRecord{
				ID:             fmt.Sprintf("mem-%d", time.Now().UnixNano()),
				SessionID:      sessionID,
				Type:           controlplane.MemoryLongTerm,
				Content:        m,
				Importance:     0.7,
				HeatScore:      50.0,
				LastAccessedAt: time.Now(),
				CreatedAt:      time.Now(),
			}
			_ = a.vectorStore.Commit(ctx, entry)
		}
	}

	fmt.Printf("[Go Archiver] Extracted %d memories from %s\n", len(memories), title)
	return nil
}

func (a *MemoryArchiver) writeToZip(sessionID, transcript, sourceTool string) (int64, error) {
	err := os.MkdirAll(filepath.Dir(a.archivePath), 0755)
	if err != nil {
		return 0, err
	}

	var buf bytes.Buffer
	var zipWriter *zip.Writer

	if _, err := os.Stat(a.archivePath); err == nil {
		data, err := os.ReadFile(a.archivePath)
		if err == nil {
			buf.Write(data)
		}
	}

	zipWriter = zip.NewWriter(&buf)
	
	f, err := zipWriter.Create(sessionID + ".txt")
	if err != nil {
		return 0, err
	}
	
	_, err = f.Write([]byte(transcript))
	if err != nil {
		return 0, err
	}

	err = zipWriter.Close()
	if err != nil {
		return 0, err
	}

	err = os.WriteFile(a.archivePath, buf.Bytes(), 0644)
	if err != nil {
		return 0, err
	}

	return int64(buf.Len()), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
