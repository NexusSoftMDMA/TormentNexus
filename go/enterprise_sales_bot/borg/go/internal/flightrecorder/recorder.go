package flightrecorder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Receipt struct {
	Timestamp     time.Time `json:"timestamp"`
	URL           string    `json:"url"`
	Model         string    `json:"model"`
	Provider      string    `json:"provider"`
	PromptVersion string    `json:"prompt_version"`
	Confidence    float64   `json:"confidence"`
	Success       bool      `json:"success"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

type Recorder struct {
	mu      sync.Mutex
	logPath string
	initErr error
}

func NewRecorder(workspaceRoot string) *Recorder {
	dir := filepath.Join(workspaceRoot, ".tormentnexus", "flight_records")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &Recorder{initErr: err}
	}

	filename := fmt.Sprintf("session_%d_%d.jsonl", time.Now().UnixNano(), os.Getpid())
	return &Recorder{
		logPath: filepath.Join(dir, filename),
	}
}

func (r *Recorder) Record(receipt Receipt) error {
	if r == nil {
		return nil
	}
	if r.initErr != nil {
		return r.initErr
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	b, err := json.Marshal(receipt)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(r.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}
	return nil
}
