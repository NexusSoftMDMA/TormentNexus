package memorystore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

var defaultSections = []string{
	"project_context",
	"user_facts",
	"style_preferences",
	"commands",
	"general",
}

type RuntimePipelineStatus struct {
	ConfiguredMode        string   `json:"configuredMode"`
	ProviderNames         []string `json:"providerNames"`
	ProviderCount         int      `json:"providerCount"`
	SectionedStoreEnabled bool     `json:"sectionedStoreEnabled"`
}

type SectionStatus struct {
	Section    string `json:"section"`
	EntryCount int    `json:"entryCount"`
}

type StoreStatus struct {
	Exists                     bool                  `json:"exists"`
	StorePath                  string                `json:"storePath"`
	TotalEntries               int                   `json:"totalEntries"`
	SectionCount               int                   `json:"sectionCount"`
	DefaultSectionCount        int                   `json:"defaultSectionCount"`
	PresentDefaultSectionCount int                   `json:"presentDefaultSectionCount"`
	PopulatedSectionCount      int                   `json:"populatedSectionCount"`
	MissingSections            []string              `json:"missingSections"`
	RuntimePipeline            RuntimePipelineStatus `json:"runtimePipeline"`
	Sections                   []SectionStatus       `json:"sections"`
	LastUpdatedAt              string                `json:"lastUpdatedAt,omitempty"`
}

type rawStore struct {
	Sections []rawSection `json:"sections"`
}

type rawSection struct {
	Section string     `json:"section"`
	Entries []rawEntry `json:"entries"`
}

type rawEntry struct {
	CreatedAt string `json:"createdAt"`
}

func ReadStatus(workspaceRoot string) (StoreStatus, error) {
	storePath := filepath.Join(workspaceRoot, ".tormentnexus", "sectioned_memory.json")
	legacyPath := filepath.Join(workspaceRoot, ".tormentnexus", "claude_mem.json")

	for _, candidate := range []string{storePath, legacyPath} {
		raw, err := os.ReadFile(candidate)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return StoreStatus{}, err
		}

		var parsed rawStore
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return StoreStatus{}, err
		}

		return summarize(storePath, &parsed), nil
	}

	return summarize(storePath, nil), nil
}

func summarize(storePath string, raw *rawStore) StoreStatus {
	normalizedSections := make([]SectionStatus, 0)
	presentSectionNames := make(map[string]struct{})
	populatedSectionCount := 0
	totalEntries := 0
	var lastUpdatedAt time.Time

	if raw != nil {
		for index, section := range raw.Sections {
			name := section.Section
			if name == "" {
				name = "section_" + itoa(index+1)
			}

			entryCount := len(section.Entries)
			if entryCount > 0 {
				populatedSectionCount++
			}
			totalEntries += entryCount
			normalizedSections = append(normalizedSections, SectionStatus{
				Section:    name,
				EntryCount: entryCount,
			})
			presentSectionNames[name] = struct{}{}

			for _, entry := range section.Entries {
				if entry.CreatedAt == "" {
					continue
				}
				parsed, err := time.Parse(time.RFC3339, entry.CreatedAt)
				if err != nil {
					continue
				}
				if parsed.After(lastUpdatedAt) {
					lastUpdatedAt = parsed
				}
			}
		}
	}

	missingSections := make([]string, 0)
	presentDefaultSectionCount := 0
	for _, section := range defaultSections {
		if _, ok := presentSectionNames[section]; ok {
			presentDefaultSectionCount++
			continue
		}
		missingSections = append(missingSections, section)
	}

	status := StoreStatus{
		Exists:                     raw != nil,
		StorePath:                  storePath,
		TotalEntries:               totalEntries,
		SectionCount:               len(normalizedSections),
		DefaultSectionCount:        len(defaultSections),
		PresentDefaultSectionCount: presentDefaultSectionCount,
		PopulatedSectionCount:      populatedSectionCount,
		MissingSections:            missingSections,
		RuntimePipeline: RuntimePipelineStatus{
			ConfiguredMode:        "unknown",
			ProviderNames:         []string{},
			ProviderCount:         0,
			SectionedStoreEnabled: false,
		},
		Sections: normalizedSections,
	}
	if !lastUpdatedAt.IsZero() {
		status.LastUpdatedAt = lastUpdatedAt.UTC().Format(time.RFC3339)
	}
	return status
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}

	buf := [20]byte{}
	index := len(buf)
	for value > 0 {
		index--
		buf[index] = byte('0' + (value % 10))
		value /= 10
	}
	return string(buf[index:])
}
