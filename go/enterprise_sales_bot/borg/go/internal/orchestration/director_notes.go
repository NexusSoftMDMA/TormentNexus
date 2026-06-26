package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
)

type DirectorNote struct {
	Timestamp int64    `json:"timestamp"`
	Objective string   `json:"objective"`
	Summary   string   `json:"summary"`
	NextSteps []string `json:"nextSteps"`
}

type DirectorNotesManager struct {
	mu    sync.RWMutex
	notes []DirectorNote
}

func NewDirectorNotesManager() *DirectorNotesManager {
	return &DirectorNotesManager{
		notes: make([]DirectorNote, 0),
	}
}

func (m *DirectorNotesManager) SynthesizeSessionNote(ctx context.Context, objective string, transcript string) (*DirectorNote, error) {
	prompt := fmt.Sprintf(`
		You are the TormentNexus Director.
		Summarize the following session transcript into a high-level note.
		
		OBJECTIVE: %s
		
		TRANSCRIPT:
		%s
		
		Provide a concise summary and a list of clear next steps.
		Respond with valid JSON:
		{
			"summary": "string",
			"nextSteps": ["string"]
		}
	`, objective, transcript[:min(8000, len(transcript))])

	resp, err := ai.AutoRoute(ctx, []ai.Message{
		{Role: "system", Content: "You are the TormentNexus Director."},
		{Role: "user", Content: prompt},
	})

	if err != nil {
		// Fallback for failed LLM
		note := &DirectorNote{
			Timestamp: time.Now().UnixMilli(),
			Objective: objective,
			Summary:   "Director summary generation failed.",
			NextSteps: []string{"Manual verification required."},
		}
		m.addNote(note)
		return note, nil
	}

	// Simple JSON extraction
	var result struct {
		Summary   string   `json:"summary"`
		NextSteps []string `json:"nextSteps"`
	}

	if err := extractJSON(resp.Content, &result); err != nil {
		fmt.Printf("[Go DirectorNotes] Failed to parse summary: %v\n", err)
	}

	note := &DirectorNote{
		Timestamp: time.Now().UnixMilli(),
		Objective: objective,
		Summary:   result.Summary,
		NextSteps: result.NextSteps,
	}

	m.addNote(note)
	return note, nil
}

func (m *DirectorNotesManager) GetNotes() []DirectorNote {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]DirectorNote(nil), m.notes...)
}

func (m *DirectorNotesManager) addNote(note *DirectorNote) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notes = append(m.notes, *note)
}

func extractJSON(content string, target interface{}) error {
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 {
		return fmt.Errorf("no JSON object found")
	}
	return json.Unmarshal([]byte(content[start:end+1]), target)
}

