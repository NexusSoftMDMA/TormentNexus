package sync

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/comma-compliance/arc-relay/internal/cli/config"
	"github.com/comma-compliance/arc-relay/internal/cli/project"
	"github.com/comma-compliance/arc-relay/internal/cli/relay"
	"github.com/comma-compliance/arc-relay/internal/cli/safety"
)

// Options configures a sync operation.
type Options struct {
	ConfigDir      string
	ProjectDir     string
	NonInteractive bool
	DryRun         bool
	Output         io.Writer // defaults to os.Stdout
	Input          io.Reader // defaults to os.Stdin
}

// Result holds the outcome of a sync operation.
type Result struct {
	Added   []string
	Skipped []string
	Existed []string
}

// Run executes the sync flow: fetch relay servers, compare with local config,
// prompt for additions, and write updates.
func Run(opts Options) (*Result, error) {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	if opts.Input == nil {
		opts.Input = os.Stdin
	}

	// Resolve credentials
	creds, err := config.ResolveCredentials(opts.ConfigDir)
	if err != nil {
		return nil, err
	}

	// Check config permissions
	if warning := config.CheckPermissions(opts.ConfigDir); warning != "" {
		_, _ = fmt.Fprintln(opts.Output, warning)
	}

	// Fetch relay servers
	client := relay.NewClient(creds.RelayURL, creds.APIKey)
	_, _ = fmt.Fprintf(opts.Output, "Connecting to Arc Relay at %s...\n", creds.RelayURL)

	allServers, err := client.ListServers()
	if err != nil {
		return nil, err
	}

	var running []relay.Server
	for _, s := range allServers {
		if s.Status == "running" {
			running = append(running, s)
		}
	}

	_, _ = fmt.Fprintf(opts.Output, "Found %d servers (%d running)\n\n", len(allServers), len(running))

	if len(running) == 0 {
		_, _ = fmt.Fprintln(opts.Output, "No running servers available.")
		return &Result{}, nil
	}

	// Detect project and target
	_, _ = fmt.Fprintf(opts.Output, "Current project: %s\n", opts.ProjectDir)

	targets := project.DetectedTargetsOrDefault(opts.ProjectDir)
	existing, err := project.ReadManagedServersFromTargets(opts.ProjectDir, creds.RelayURL, targets)
	if err != nil {
		return nil, fmt.Errorf("reading project config: %w", err)
	}

	existingNames := make(map[string]bool)
	for _, s := range existing {
		existingNames[s.Name] = true
	}

	if len(existing) > 0 {
		names := make([]string, 0, len(existing))
		for _, s := range existing {
			names = append(names, s.Name)
		}
		_, _ = fmt.Fprintf(opts.Output, "Already configured: %s\n", strings.Join(names, ", "))
	}

	// Load state for skip list
	state, err := config.LoadState(opts.ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("loading state: %w", err)
	}

	// Detect renamed servers: compare tracked IDs against current server names
	tracked := state.GetTrackedServers(opts.ProjectDir)
	var renames []project.ManagedServer
	for _, s := range running {
		oldName := tracked[s.ID]
		if oldName != "" && oldName != s.Name && existingNames[oldName] {
			// Server was renamed - update project config entries.
			renames = append(renames, project.ManagedServer{
				Name:    s.Name,
				URL:     client.ServerProxyURL(s.Name),
				OldName: oldName,
			})
			state.TrackServer(opts.ProjectDir, s.ID, s.Name)
			// Update skip list if the old name was skipped
			if state.IsSkipped(opts.ProjectDir, oldName) {
				state.RemoveSkipped(opts.ProjectDir, oldName)
				state.AddSkipped(opts.ProjectDir, s.Name)
			}
			existingNames[s.Name] = true
			delete(existingNames, oldName)
		}
	}

	// Apply renames to all detected targets.
	if len(renames) > 0 {
		for _, r := range renames {
			_, _ = fmt.Fprintf(opts.Output, "  Renamed: %s -> %s\n", r.OldName, r.Name)
		}
		if !opts.DryRun {
			oldNames := make([]string, 0, len(renames))
			for _, r := range renames {
				oldNames = append(oldNames, r.OldName)
			}

			for _, target := range targets {
				if _, err := target.Remove(opts.ProjectDir, oldNames); err != nil {
					slog.Warn("sync: removing renamed entries failed", "target", target.Name(), "err", err)
				}
				if err := target.Write(opts.ProjectDir, creds.RelayURL, creds.APIKey, renames); err != nil {
					slog.Warn("sync: writing renamed entries failed", "target", target.Name(), "err", err)
				}
			}
		}
	}

	// Find new servers
	var newServers []relay.Server
	for _, s := range running {
		if !existingNames[s.Name] && !state.IsSkipped(opts.ProjectDir, s.Name) {
			newServers = append(newServers, s)
		}
	}

	result := &Result{}
	for _, s := range existing {
		result.Existed = append(result.Existed, s.Name)
	}

	if len(newServers) == 0 {
		_, _ = fmt.Fprintln(opts.Output, "\nAll servers are already configured or skipped.")
		if len(renames) > 0 {
			if err := config.SaveState(opts.ConfigDir, state); err != nil {
				return nil, fmt.Errorf("saving state: %w", err)
			}
		}
		return result, nil
	}

	_, _ = fmt.Fprintln(opts.Output, "\nNew servers available:")

	// Collect servers to add
	var toAdd []project.ManagedServer
	scanner := bufio.NewScanner(opts.Input)

	for i, s := range newServers {
		displayName := s.Name
		if s.DisplayName != "" && s.DisplayName != s.Name {
			displayName = fmt.Sprintf("%s (%s)", s.Name, s.DisplayName)
		}

		// Show health warning inline
		healthNote := ""
		if s.Health == "unhealthy" {
			errMsg := s.HealthError
			if errMsg == "" {
				errMsg = "unknown error"
			}
			healthNote = fmt.Sprintf(" [unhealthy: %s]", errMsg)
		}

		if opts.NonInteractive {
			toAdd = append(toAdd, project.ManagedServer{
				Name: s.Name,
				URL:  client.ServerProxyURL(s.Name),
			})
			result.Added = append(result.Added, s.Name)
			label := "auto-added"
			if healthNote != "" {
				label += healthNote
			}
			_, _ = fmt.Fprintf(opts.Output, "  [%d] %s — %s\n", i+1, displayName, label)
			continue
		}

		_, _ = fmt.Fprintf(opts.Output, "  [%d] %s%s — Add to project? [y/n]  ", i+1, displayName, healthNote)

		if !scanner.Scan() {
			break
		}
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))

		switch answer {
		case "y", "yes":
			toAdd = append(toAdd, project.ManagedServer{
				Name: s.Name,
				URL:  client.ServerProxyURL(s.Name),
			})
			result.Added = append(result.Added, s.Name)
		default:
			// "n" or anything else — skip and don't ask again
			state.AddSkipped(opts.ProjectDir, s.Name)
			result.Skipped = append(result.Skipped, s.Name)
		}
	}

	if len(toAdd) == 0 {
		_, _ = fmt.Fprintln(opts.Output, "\nNo servers added.")
		// Still save state if there were skips or renames
		if len(result.Skipped) > 0 || len(renames) > 0 {
			if err := config.SaveState(opts.ConfigDir, state); err != nil {
				return nil, fmt.Errorf("saving state: %w", err)
			}
		}
		return result, nil
	}

	// Show change summary
	changes := make([]safety.PlannedChange, 0, len(targets)+1)
	for _, target := range targets {
		configPath := filepath.Join(opts.ProjectDir, target.ConfigFileName())
		changes = append(changes, safety.PlannedChange{
			Path:        configPath,
			Description: fmt.Sprintf("adding %d server(s)", len(toAdd)),
			Scope:       safety.ScopeProject,
		})
	}
	if len(result.Skipped) > 0 {
		changes = append(changes, safety.PlannedChange{
			Path:        config.StatePath(opts.ConfigDir),
			Description: "updating skip list",
			Scope:       safety.ScopeUser,
		})
	}

	if opts.DryRun {
		_, _ = fmt.Fprintf(opts.Output, "\nDRY RUN — no files will be modified\n\n")
		_, _ = fmt.Fprint(opts.Output, safety.FormatChangeSummary(changes, opts.ProjectDir))
		return result, nil
	}

	// Show gitignore warnings for each target file.
	var warningOutput strings.Builder
	for _, target := range targets {
		warningOutput.WriteString(safety.FormatWarnings(safety.CheckGitignore(opts.ProjectDir, target.ConfigFileName())))
	}
	if warningOutput.Len() > 0 {
		_, _ = fmt.Fprintln(opts.Output)
		_, _ = fmt.Fprint(opts.Output, warningOutput.String())
	}

	// Write changes to all targets.
	for _, target := range targets {
		if err := target.Write(opts.ProjectDir, creds.RelayURL, creds.APIKey, toAdd); err != nil {
			return nil, fmt.Errorf("writing %s config: %w", target.Name(), err)
		}
	}

	// Track server IDs for rename detection
	serversByName := make(map[string]relay.Server)
	for _, s := range running {
		serversByName[s.Name] = s
	}
	for _, ms := range toAdd {
		if srv, ok := serversByName[ms.Name]; ok {
			state.TrackServer(opts.ProjectDir, srv.ID, srv.Name)
		}
	}
	stateChanged := len(result.Skipped) > 0 || len(toAdd) > 0 || len(renames) > 0

	if stateChanged {
		if err := config.SaveState(opts.ConfigDir, state); err != nil {
			return nil, fmt.Errorf("saving state: %w", err)
		}
	}

	_, _ = fmt.Fprintf(opts.Output, "\nAdded %d server(s) to %d target(s)\n", len(toAdd), len(targets))
	if len(result.Skipped) > 0 {
		_, _ = fmt.Fprintf(opts.Output, "Skipped: %s (won't be prompted again — run 'arc-sync reset' to undo)\n",
			strings.Join(result.Skipped, ", "))
	}

	return result, nil
}
