package orchestrator

import (
	"os/exec"
	"regexp"
	"strings"
)

type SubmoduleStatus struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Hash     string `json:"hash"`
	FullHash string `json:"fullHash"`
	Ref      string `json:"ref"`
	Status   string `json:"status"`
}

// ExtractSubmoduleIntelligence natively shells out to pull repository health metadata completely replacing node/simple-git parity natively.
func ExtractSubmoduleIntelligence() ([]SubmoduleStatus, error) {
	// The Go process executes from submodules/tormentnexus, so we explicitly command Git
	// to query the top-level parent monorepo.
	cmd := exec.Command("git", "-C", "../..", "submodule", "status")
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		// Log error but proceed to return empty or partial on simple execution faults
		return []SubmoduleStatus{}, err
	}

	output := string(outputBytes)
	lines := strings.Split(output, "\n")

	var submodules []SubmoduleStatus

	// Match TS pattern: /^([\s+-])([a-f0-9]+)\s+([^\s]+)(?:\s+\((.+)\))?$/
	re := regexp.MustCompile(`^([\s+\-])([a-f0-9]+)\s+([^\s]+)(?:\s+\((.+)\))?$`)

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue
		}

		statusChar := matches[1]
		fullHash := matches[2]
		path := matches[3]
		ref := "unknown"
		if len(matches) > 4 && matches[4] != "" {
			ref = matches[4]
		}

		pathParts := strings.Split(path, "/")
		name := pathParts[len(pathParts)-1]

		var status = "uninitialized"
		if statusChar == " " {
			status = "synced"
		} else if statusChar == "+" {
			status = "modified"
		}

		shortHash := fullHash
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}

		submodules = append(submodules, SubmoduleStatus{
			Name:     name,
			Path:     path,
			Hash:     shortHash,
			FullHash: fullHash,
			Ref:      ref,
			Status:   status,
		})
	}

	return submodules, nil
}
