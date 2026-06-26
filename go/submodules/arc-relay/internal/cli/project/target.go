package project

// ManagedServer represents an MCP server entry managed by the relay.
type ManagedServer struct {
	Name    string
	URL     string
	OldName string // set when this server was renamed from a different slug
}

// Target is the interface for an AI tool's MCP configuration format.
// Each target knows how to detect, read, and write its own config file.
type Target interface {
	// Name returns the human-readable name of this target (e.g., "claude-code").
	Name() string

	// ConfigFileName returns the config file name (e.g., ".mcp.json").
	ConfigFileName() string

	// Detect returns true if this target's config file exists in the project dir,
	// or if the project dir is a suitable location for this target.
	Detect(projectDir string) bool

	// Read reads the target's config file and returns the list of relay-managed
	// servers currently configured. The relayBaseURL is used to identify which
	// entries belong to the relay.
	Read(projectDir, relayBaseURL string) ([]ManagedServer, error)

	// Write adds the given servers to the target's config file, preserving all
	// existing entries. The relayBaseURL and apiKey are used to construct the
	// server entries.
	Write(projectDir, relayBaseURL, apiKey string, servers []ManagedServer) error

	// Remove removes the named servers from the target's config file, preserving
	// all other entries. Returns the list of names that were actually removed.
	Remove(projectDir string, names []string) ([]string, error)
}
