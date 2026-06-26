package interop

import (
	"errors"
	"os"
	"time"
)

type ImportedInstructions struct {
	Path       string `json:"path"`
	Available  bool   `json:"available"`
	Content    string `json:"content,omitempty"`
	ModifiedAt string `json:"modifiedAt,omitempty"`
	Size       int64  `json:"size,omitempty"`
}

func ReadImportedInstructions(docPath string) ImportedInstructions {
	result := ImportedInstructions{
		Path: docPath,
	}

	info, err := os.Stat(docPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return result
		}
		return result
	}

	content, err := os.ReadFile(docPath)
	if err != nil {
		return result
	}

	result.Available = true
	result.Content = string(content)
	result.ModifiedAt = info.ModTime().UTC().Format(time.RFC3339)
	result.Size = info.Size()
	return result
}
