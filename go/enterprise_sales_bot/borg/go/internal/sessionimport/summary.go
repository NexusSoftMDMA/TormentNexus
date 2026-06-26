package sessionimport

import (
	"sort"
	"time"
)

type SummaryBucket struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type Summary struct {
	GeneratedAt        string          `json:"generatedAt"`
	Count              int             `json:"count"`
	ValidCount         int             `json:"validCount"`
	InvalidCount       int             `json:"invalidCount"`
	TotalEstimatedSize int64           `json:"totalEstimatedSize"`
	BySourceTool       []SummaryBucket `json:"bySourceTool"`
	BySourceType       []SummaryBucket `json:"bySourceType"`
	ByFormat           []SummaryBucket `json:"byFormat"`
	ByModelHint        []SummaryBucket `json:"byModelHint"`
	ByError            []SummaryBucket `json:"byError"`
}

func BuildSummary(candidates []ValidationResult) Summary {
	bySourceTool := map[string]int{}
	bySourceType := map[string]int{}
	byFormat := map[string]int{}
	byModelHint := map[string]int{}
	byError := map[string]int{}

	validCount := 0
	invalidCount := 0
	var totalEstimatedSize int64

	for _, candidate := range candidates {
		bySourceTool[candidate.SourceTool]++
		bySourceType[candidate.SourceType]++
		byFormat[candidate.Format]++
		if candidate.Valid {
			validCount++
		} else {
			invalidCount++
		}
		totalEstimatedSize += candidate.EstimatedSize
		for _, model := range candidate.DetectedModels {
			byModelHint[model]++
		}
		for _, message := range candidate.Errors {
			byError[message]++
		}
	}

	return Summary{
		GeneratedAt:        time.Now().UTC().Format(time.RFC3339),
		Count:              len(candidates),
		ValidCount:         validCount,
		InvalidCount:       invalidCount,
		TotalEstimatedSize: totalEstimatedSize,
		BySourceTool:       bucketsFromMap(bySourceTool),
		BySourceType:       bucketsFromMap(bySourceType),
		ByFormat:           bucketsFromMap(byFormat),
		ByModelHint:        bucketsFromMap(byModelHint),
		ByError:            bucketsFromMap(byError),
	}
}

func bucketsFromMap(values map[string]int) []SummaryBucket {
	buckets := make([]SummaryBucket, 0, len(values))
	for key, count := range values {
		if key == "" {
			key = "unknown"
		}
		buckets = append(buckets, SummaryBucket{Key: key, Count: count})
	}

	sort.Slice(buckets, func(i, j int) bool {
		if buckets[i].Count == buckets[j].Count {
			return buckets[i].Key < buckets[j].Key
		}
		return buckets[i].Count > buckets[j].Count
	})
	return buckets
}
