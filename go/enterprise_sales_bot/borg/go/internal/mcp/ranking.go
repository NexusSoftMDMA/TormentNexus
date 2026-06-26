package mcp

import (
	"sort"
	"strings"
	"unicode"
)

// RankedTool wraps a ToolEntry with an assigned relevance score
type RankedTool struct {
	ToolEntry
	Score          float64            `json:"score"`
	ScoreBreakdown map[string]float64 `json:"scoreBreakdown"`
	MatchReason    string             `json:"matchReason"`
	Rank           int                `json:"rank"`
}

// Tokenize converts a string into a slice of lowercased, alphanumeric tokens
func Tokenize(text string) []string {
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	rawTokens := strings.FieldsFunc(text, f)
	var tokens []string
	for _, t := range rawTokens {
		lower := strings.ToLower(t)
		if len(lower) > 2 { // Filter out extremely short words
			tokens = append(tokens, lower)
		}
	}
	return tokens
}

// CalculateBM25Score provides a lightweight, heuristic-based scoring model
// simulating the Node.js TypeScript TF-IDF and semantic weighting algorithms
// used by the TormentNexus Control Plane.
func CalculateBM25Score(queryTokens []string, tool ToolEntry) (float64, map[string]float64, string) {
	if len(queryTokens) == 0 {
		return 0.0, nil, ""
	}

	breakdown := make(map[string]float64)
	totalScore := 0.0
	matchReasons := []string{}

	nameTokens := Tokenize(tool.Name)
	descTokens := Tokenize(tool.Description)

	// Weights
	const weightName = 10.0
	const weightDesc = 3.0
	const weightTags = 5.0

	// Name Scoring
	nameScore := 0.0
	for _, q := range queryTokens {
		for _, nt := range nameTokens {
			if strings.Contains(nt, q) {
				nameScore += weightName
			}
		}
	}
	if nameScore > 0 {
		breakdown["name"] = nameScore
		totalScore += nameScore
		matchReasons = append(matchReasons, "Matched tool name")
	}

	// Description Scoring
	descScore := 0.0
	for _, q := range queryTokens {
		for _, dt := range descTokens {
			if strings.Contains(dt, q) {
				descScore += weightDesc
			}
		}
	}
	if descScore > 0 {
		breakdown["description"] = descScore
		totalScore += descScore
		matchReasons = append(matchReasons, "Matched description")
	}

	// Tags & Keywords Scoring
	tagScore := 0.0
	allTags := append(append(tool.ToolTags, tool.ServerTags...), tool.Keywords...)
	for _, q := range queryTokens {
		for _, tag := range allTags {
			if strings.Contains(strings.ToLower(tag), q) {
				tagScore += weightTags
			}
		}
	}
	if tagScore > 0 {
		breakdown["tags"] = tagScore
		totalScore += tagScore
		matchReasons = append(matchReasons, "Matched keywords/tags")
	}

	// Semantic Group Bonus
	semanticScore := 0.0
	for _, q := range queryTokens {
		if strings.Contains(strings.ToLower(tool.SemanticGroupLabel), q) || strings.Contains(strings.ToLower(tool.SemanticGroup), q) {
			semanticScore += weightTags
		}
	}
	if semanticScore > 0 {
		breakdown["semantic_group"] = semanticScore
		totalScore += semanticScore
		matchReasons = append(matchReasons, "Matched semantic category")
	}

	reason := "No match"
	if len(matchReasons) > 0 {
		reason = strings.Join(matchReasons, "; ")
	}

	return totalScore, breakdown, reason
}

// RankTools filters and sorts a slice of ToolEntry based on the query string.
func RankTools(query string, tools []ToolEntry, limit int) []RankedTool {
	if query == "" {
		// Return all tools unranked up to the limit
		var results []RankedTool
		for i, t := range tools {
			if limit > 0 && i >= limit {
				break
			}
			results = append(results, RankedTool{
				ToolEntry:   t,
				Score:       0,
				MatchReason: "Default listing",
				Rank:        i + 1,
			})
		}
		return results
	}

	queryTokens := Tokenize(query)
	var ranked []RankedTool

	for _, tool := range tools {
		score, breakdown, reason := CalculateBM25Score(queryTokens, tool)
		if score > 0 {
			ranked = append(ranked, RankedTool{
				ToolEntry:      tool,
				Score:          score,
				ScoreBreakdown: breakdown,
				MatchReason:    reason,
			})
		}
	}

	// Sort descending by score
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	// Apply rank index and truncate to limit
	var finalResults []RankedTool
	for i, r := range ranked {
		if limit > 0 && i >= limit {
			break
		}
		r.Rank = i + 1
		finalResults = append(finalResults, r)
	}

	return finalResults
}
