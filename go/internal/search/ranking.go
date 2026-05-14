package search

import (
	"sort"
	"strings"
	"unicode"
)

// Scorable is an interface for items that can be ranked by the BM25-style engine.
type Scorable interface {
	GetName() string
	GetDescription() string
	GetTags() []string
}

// ScoredResult wraps a Scorable item with its search score.
type ScoredResult struct {
	Item           Scorable
	Score          float64
	ScoreBreakdown map[string]float64
	MatchReason    string
	Rank           int
}

// Tokenize converts a string into a slice of lowercased, alphanumeric tokens.
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

// CalculateBM25Score provides a lightweight, heuristic-based scoring model.
func CalculateBM25Score(queryTokens []string, item Scorable) (float64, map[string]float64, string) {
	if len(queryTokens) == 0 {
		return 0.0, nil, ""
	}

	breakdown := make(map[string]float64)
	totalScore := 0.0
	matchReasons := []string{}

	nameTokens := Tokenize(item.GetName())
	descTokens := Tokenize(item.GetDescription())

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
		matchReasons = append(matchReasons, "Matched name")
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

	// Tags Scoring
	tagScore := 0.0
	for _, q := range queryTokens {
		for _, tag := range item.GetTags() {
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

	reason := "No match"
	if len(matchReasons) > 0 {
		reason = strings.Join(matchReasons, "; ")
	}

	return totalScore, breakdown, reason
}

// RankItems filters and sorts a slice of Scorable items based on the query string.
func RankItems(query string, items []Scorable, limit int) []ScoredResult {
	if query == "" {
		var results []ScoredResult
		for i, it := range items {
			if limit > 0 && i >= limit {
				break
			}
			results = append(results, ScoredResult{
				Item:        it,
				Score:       0,
				MatchReason: "Default listing",
				Rank:        i + 1,
			})
		}
		return results
	}

	queryTokens := Tokenize(query)
	var ranked []ScoredResult

	for _, it := range items {
		score, breakdown, reason := CalculateBM25Score(queryTokens, it)
		if score > 0 {
			ranked = append(ranked, ScoredResult{
				Item:           it,
				Score:          score,
				ScoreBreakdown: breakdown,
				MatchReason:    reason,
			})
		}
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return ranked[i].Item.GetName() < ranked[j].Item.GetName()
		}
		return ranked[i].Score > ranked[j].Score
	})

	var finalResults []ScoredResult
	for i := range ranked {
		if limit > 0 && i >= limit {
			break
		}
		ranked[i].Rank = i + 1
		finalResults = append(finalResults, ranked[i])
	}

	return finalResults
}
