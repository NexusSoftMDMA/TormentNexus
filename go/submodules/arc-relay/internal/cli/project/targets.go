package project

import "fmt"

// DetectedTargetsOrDefault returns existing project targets, falling back to
// Claude Code for backward compatibility when no target config exists yet.
func DetectedTargetsOrDefault(projectDir string) []Target {
	targets := DetectTargets(projectDir, AllTargets())
	if len(targets) > 0 {
		return targets
	}
	return []Target{&ClaudeCodeTarget{}}
}

// ReadManagedServersFromTargets reads relay-managed servers from all targets
// and returns the union by server name.
func ReadManagedServersFromTargets(projectDir, relayBaseURL string, targets []Target) ([]ManagedServer, error) {
	seen := make(map[string]ManagedServer)
	var ordered []ManagedServer

	for _, target := range targets {
		servers, err := target.Read(projectDir, relayBaseURL)
		if err != nil {
			return nil, fmt.Errorf("reading %s config: %w", target.Name(), err)
		}

		for _, server := range servers {
			if _, ok := seen[server.Name]; ok {
				continue
			}
			seen[server.Name] = server
			ordered = append(ordered, server)
		}
	}

	return ordered, nil
}
