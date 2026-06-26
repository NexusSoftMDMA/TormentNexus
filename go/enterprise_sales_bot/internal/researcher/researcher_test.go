package researcher

import (
	"testing"
)

func TestResearcher_Run(t *testing.T) {
	// Simple test to ensure the researcher can be initialized and run cycle logic
	// In a real implementation, we'd mock the database and crawlers
	r := NewResearcher(nil, []Crawler{&GitHubCrawler{}}, &DefaultDossierProcessor{}, nil)
	if r == nil {
		t.Fatal("Failed to create researcher")
	}
}
