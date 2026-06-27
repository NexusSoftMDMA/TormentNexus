package memory

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type L3Archive struct {
	baseDir string
	mu      sync.Mutex
}

func NewL3Archive(workspaceRoot string) *L3Archive {
	dir := filepath.Join(workspaceRoot, ".tormentnexus", "memory", "l3_archive")
	_ = os.MkdirAll(dir, 0755)
	return &L3Archive{baseDir: dir}
}

// Archive stores a slice of memories into a compressed cold storage file
func (l3 *L3Archive) Archive(memories []*Memory) error {
	if len(memories) == 0 {
		return nil
	}

	l3.mu.Lock()
	defer l3.mu.Unlock()

	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(l3.baseDir, fmt.Sprintf("archive_%s.json.gz", timestamp))

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	return json.NewEncoder(gz).Encode(memories)
}

// Unarchive reads and uncompresses all L3 archive files
func (l3 *L3Archive) Unarchive() ([]*Memory, error) {
	l3.mu.Lock()
	defer l3.mu.Unlock()

	var allMemories []*Memory

	files, err := os.ReadDir(l3.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".gz" {
			continue
		}

		path := filepath.Join(l3.baseDir, file.Name())
		f, err := os.Open(path)
		if err != nil {
			continue // skip broken files
		}

		gz, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			continue
		}

		data, err := io.ReadAll(gz)
		gz.Close()
		f.Close()

		if err != nil {
			continue
		}

		var batch []*Memory
		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&batch); err == nil {
			allMemories = append(allMemories, batch...)
		}
	}

	return allMemories, nil
}
