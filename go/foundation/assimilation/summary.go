package assimilation

import "sort"

type CategorySummary struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

func Categories(items []SourceToolchain) []CategorySummary {
	counts := map[string]int{}
	for _, item := range items {
		counts[item.Category]++
	}
	out := make([]CategorySummary, 0, len(counts))
	for category, count := range counts {
		out = append(out, CategorySummary{Category: category, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Category < out[j].Category
		}
		return out[i].Count > out[j].Count
	})
	return out
}
