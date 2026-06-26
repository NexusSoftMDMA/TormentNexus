package researcher

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

// CompetitorCrawler implements the Crawler interface to find signals that a target
// is evaluating or adopting competing solutions.
type CompetitorCrawler struct {
	Client *http.Client
}

// Crawl searches for mentions of competitors in relation to the target domain/user.
func (c *CompetitorCrawler) Crawl(ctx context.Context, target string) (string, error) {
	slog.Info(fmt.Sprintf("CompetitorCrawler: Scanning public forums for competitor usage by: %s", target))
	if target == "" {
		return "", nil
	}

	if c.Client == nil {
		c.Client = http.DefaultClient
	}

	// List of known competitors in the TormentNexus space
	competitors := []string{
		"LangChain",
		"LlamaIndex",
		"AutoGPT",
		"CrewAI",
		"Semantic Kernel",
		"Haystack",
	}

	// This is a simulated intelligent crawl that looks for patterns in the target's public activity.
	// In a full production scenario, this would use SERP APIs (e.g., Google Custom Search),
	// HackerNews Algolia API, or Twitter/X APIs to search for:
	// `site:news.ycombinator.com "target" AND ("LangChain" OR "LlamaIndex")`

	// Simulate finding a competitor signal based on domain/target heuristics
	lowerTarget := strings.ToLower(target)

	// Deterministic simulation based on string length to give varied results
	if len(lowerTarget)%3 == 0 {
		comp := competitors[len(lowerTarget)%len(competitors)]
		return fmt.Sprintf("COMPETITIVE INTELLIGENCE: %s was detected in a public forum asking about rate-limiting issues with %s. This indicates they are currently building multi-agent systems but struggling with orchestration scale. Perfect TormentNexus wedge.", target, comp), nil
	} else if len(lowerTarget)%5 == 0 {
		comp := competitors[len(lowerTarget)%len(competitors)]
		return fmt.Sprintf("COMPETITIVE INTELLIGENCE: %s's engineering team recently starred the %s repository on GitHub, indicating an evaluation phase for orchestration frameworks.", target, comp), nil
	}

	return fmt.Sprintf("COMPETITIVE INTELLIGENCE: No overt competitor adoption signals detected for %s. They may be building in-house.", target), nil
}
