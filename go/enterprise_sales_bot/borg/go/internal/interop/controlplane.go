package interop

import (
	"errors"
	"os"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/lockfile"
)

type ControlPlaneStatus struct {
	Name       string `json:"name"`
	LockPath   string `json:"lockPath"`
	Running    bool   `json:"running"`
	Host       string `json:"host,omitempty"`
	Port       int    `json:"port,omitempty"`
	Version    string `json:"version,omitempty"`
	StartedAt  string `json:"startedAt,omitempty"`
	ModifiedAt string `json:"modifiedAt,omitempty"`
}

func DiscoverControlPlanes(mainLockPath, goLockPath string) []ControlPlaneStatus {
	return []ControlPlaneStatus{
		readStatus("tormentnexus-node", mainLockPath),
		readStatus("tormentnexus-go", goLockPath),
	}
}

func readStatus(name, lockPath string) ControlPlaneStatus {
	status := ControlPlaneStatus{
		Name:     name,
		LockPath: lockPath,
	}

	info, err := os.Stat(lockPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return status
		}
		return status
	}

	record, err := lockfile.Read(lockPath)
	if err != nil {
		return status
	}

	status.Running = true
	status.Host = record.Host
	status.Port = record.Port
	status.Version = record.Version
	status.StartedAt = record.StartedAt
	status.ModifiedAt = info.ModTime().UTC().Format(time.RFC3339)
	return status
}
