package llm

import (
	crypto_rand "crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func cryptoRandInt() int64 {
	n, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(1<<63-1))
	if err != nil {
		return 0 // Fallback to 0 if crypto fails to avoid panic
	}
	return n.Int64()
}


// PromptVersion represents a single version of a prompt template.
type PromptVersion struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Template  string    `json:"template"`
	CreatedAt time.Time `json:"created_at"`
	Enabled   bool      `json:"enabled"`
}

// ABExperiment defines an A/B test with weighted version selection.
type ABExperiment struct {
	Name       string    `json:"name"`
	VersionIDs []string  `json:"version_ids"`
	Weights    []float64 `json:"weights"`
	// cumulative distribution for selection (runtime only)
	cdf []float64 `json:"-"`
}

// ABResult aggregates success/failure outcomes for a version within an experiment.
type ABResult struct {
	Experiment string `json:"experiment"`
	VersionID  string `json:"version_id"`
	Success    int    `json:"success"`
	Failure    int    `json:"failure"`
	Total      int    `json:"total"`
}

// PromptRegistry holds all prompt versions, experiments, and outcomes.
type PromptRegistry struct {
	Versions    map[string][]*PromptVersion          `json:"versions"`
	Experiments map[string]*ABExperiment              `json:"experiments"`
	Outcomes    map[string]map[string]*ABResult       `json:"outcomes"`
	mu          sync.RWMutex                          `json:"-"`
	filePath    string                                 `json:"-"`
}

// NewPromptRegistry creates a registry, loading persisted data if present.
func NewPromptRegistry(filePath string) *PromptRegistry {
	pr := &PromptRegistry{
		Versions:    make(map[string][]*PromptVersion),
		Experiments: make(map[string]*ABExperiment),
		Outcomes:    make(map[string]map[string]*ABResult),
		filePath:    filePath,
	}
	// Attempt to load existing JSON state.
	if data, err := os.ReadFile(filepath.Clean(filePath)); err == nil {
		_ = json.Unmarshal(data, pr)
	}
	// Seed RNG once for the entire process.
	return pr
}

// RegisterVersion adds a new prompt version for a given name.
func (pr *PromptRegistry) RegisterVersion(name, template string) *PromptVersion {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	id := fmt.Sprintf("%d-%s", time.Now().UnixNano(), name)
	v := &PromptVersion{
		ID:        id,
		Name:      name,
		Template:  template,
		CreatedAt: time.Now(),
		Enabled:   true,
	}
		pr.Versions[name] = append(pr.Versions[name], v)
	pr.save()
	return v
}

// GetActiveVersion returns the first enabled version for a prompt name.
func (pr *PromptRegistry) GetActiveVersion(name string) *PromptVersion {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	for _, v := range pr.Versions[name] {
		if v.Enabled {
			return v
		}
	}
	return nil
}

// AssignExperiment configures an A/B experiment for a prompt name.
// versionIDs must correspond to existing PromptVersion.ID values.
func (pr *PromptRegistry) AssignExperiment(name string, versionIDs []string, weights []float64) error {
	if len(versionIDs) != len(weights) || len(versionIDs) == 0 {
		return fmt.Errorf("invalid experiment: mismatched lengths or empty")
	}
	sum := 0.0
	for _, w := range weights {
		sum += w
	}
	if sum <= 0 {
		return fmt.Errorf("weights must sum > 0")
	}

	normalized := make([]float64, len(weights))
	for i, w := range weights {
		normalized[i] = w / sum
	}

	cdf := make([]float64, len(normalized))
	accum := 0.0
	for i, w := range normalized {
		accum += w
		cdf[i] = accum
	}

	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.Experiments[name] = &ABExperiment{
		Name:       name,
		VersionIDs: versionIDs,
		Weights:    normalized,
		cdf:        cdf,
	}
	pr.save()
	return nil
}

// pickVersionByExperiment selects a version ID according to weighted random.
func (pr *PromptRegistry) pickVersionByExperiment(exp *ABExperiment) string {
	r := float64(cryptoRandInt()) / float64(1<<63 - 1)
	for i, threshold := range exp.cdf {
		if r <= threshold {
			return exp.VersionIDs[i]
		}
	}
	return exp.VersionIDs[len(exp.VersionIDs)-1]
}

// ResolvePrompt selects a version and interpolates placeholders.
// Uses AB experiment if configured, otherwise the active version.
// Placeholders are ${key} format and replaced with data values.
func (pr *PromptRegistry) ResolvePrompt(name string, data map[string]string) (string, error) {
	pr.mu.RLock()
	exp, hasExp := pr.Experiments[name]
	pr.mu.RUnlock()

	if hasExp {
		versionID := pr.pickVersionByExperiment(exp)
		pr.mu.RLock()
		var version *PromptVersion
		for _, v := range pr.Versions[name] {
			if v.ID == versionID && v.Enabled {
				version = v
				break
			}
		}
		pr.mu.RUnlock()
		if version == nil {
			return "", fmt.Errorf("selected version %q not found or disabled", versionID)
		}
		return renderTemplate(version.Template, data), nil
	}

	// No experiment — use active version
	version := pr.GetActiveVersion(name)
	if version == nil {
		return "", fmt.Errorf("no enabled prompt version for %q", name)
	}
	return renderTemplate(version.Template, data), nil
}

// renderTemplate substitutes ${key} placeholders with values from data.
func renderTemplate(tpl string, data map[string]string) string {
	result := tpl
	for k, v := range data {
		result = strings.ReplaceAll(result, fmt.Sprintf("${%s}", k), v)
	}
	return result
}

// RecordOutcome logs the result of using a version in an experiment.
func (pr *PromptRegistry) RecordOutcome(experiment, versionID string, success bool) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	if _, ok := pr.Outcomes[experiment]; !ok {
		pr.Outcomes[experiment] = make(map[string]*ABResult)
	}
	res, ok := pr.Outcomes[experiment][versionID]
	if !ok {
		res = &ABResult{
			Experiment: experiment,
			VersionID:  versionID,
		}
		pr.Outcomes[experiment][versionID] = res
	}
	if success {
		res.Success++
	} else {
		res.Failure++
	}
	res.Total++
	pr.save()
}

// GetOutcomes returns a copy of all recorded outcomes.
func (pr *PromptRegistry) GetOutcomes() []ABResult {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	var results []ABResult
	for _, versionMap := range pr.Outcomes {
		for _, res := range versionMap {
			results = append(results, *res)
		}
	}
	return results
}

// save persists the registry as JSON.
func (pr *PromptRegistry) save() {
	data, err := json.MarshalIndent(pr, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(pr.filePath, data, 0600)
}

// Load reloads the registry from the JSON file.
func (pr *PromptRegistry) Load() error {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	data, err := os.ReadFile(pr.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, pr)
}
