package tools

import foundationrepomap "github.com/NexusSoftMDMA/TormentNexus/foundation/repomap"

// RepoMapTool wraps the native foundation repomap package for legacy callers.
type RepoMapTool struct {
	BaseDir         string
	MentionedFiles  []string
	MentionedIdents []string
	MaxFiles        int
	IncludeTests    bool
}

func NewRepoMapTool(baseDir string) *RepoMapTool {
	return &RepoMapTool{BaseDir: baseDir, MaxFiles: 40}
}

// Generate condenses the repository tree into a ranked LLM-optimized map.
func (r *RepoMapTool) Generate() (string, error) {
	result, err := foundationrepomap.Generate(foundationrepomap.Options{
		BaseDir:         r.BaseDir,
		MentionedFiles:  r.MentionedFiles,
		MentionedIdents: r.MentionedIdents,
		MaxFiles:        r.MaxFiles,
		IncludeTests:    r.IncludeTests,
	})
	if err != nil {
		return "", err
	}
	return result.Map, nil
}
