package lockfile

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type StartupProvenance struct {
	RequestedRuntime string `json:"requestedRuntime,omitempty"`
	ActiveRuntime    string `json:"activeRuntime,omitempty"`
	RequestedPort    int    `json:"requestedPort,omitempty"`
	ActivePort       int    `json:"activePort,omitempty"`
	PortDecision     string `json:"portDecision,omitempty"`
	PortReason       string `json:"portReason,omitempty"`
	LaunchMode       string `json:"launchMode,omitempty"`
	DashboardMode    string `json:"dashboardMode,omitempty"`
	InstallDecision  string `json:"installDecision,omitempty"`
	InstallReason    string `json:"installReason,omitempty"`
	BuildDecision    string `json:"buildDecision,omitempty"`
	BuildReason      string `json:"buildReason,omitempty"`
	UpdatedAt        string `json:"updatedAt,omitempty"`
}

type Record struct {
	Host      string             `json:"host"`
	Port      int                `json:"port"`
	Version   string             `json:"version"`
	StartedAt string             `json:"startedAt"`
	Startup   *StartupProvenance `json:"startup,omitempty"`
}

func Write(lockPath string, record Record) error {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return err
	}

	payload, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(lockPath, payload, 0o644)
}

func Read(lockPath string) (Record, error) {
	payload, err := os.ReadFile(lockPath)
	if err != nil {
		return Record{}, err
	}

	var record Record
	if err := json.Unmarshal(payload, &record); err != nil {
		return Record{}, err
	}

	return record, nil
}
