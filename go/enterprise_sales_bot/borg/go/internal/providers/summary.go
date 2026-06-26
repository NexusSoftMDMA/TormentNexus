package providers

type SummaryBucket struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type Summary struct {
	ProviderCount      int             `json:"providerCount"`
	ConfiguredCount    int             `json:"configuredCount"`
	AuthenticatedCount int             `json:"authenticatedCount"`
	ExecutableCount    int             `json:"executableCount"`
	ByAuthMethod       []SummaryBucket `json:"byAuthMethod"`
	ByPreferredTask    []SummaryBucket `json:"byPreferredTask"`
}

func BuildSummary(statuses []Status) Summary {
	entries := Catalog(statuses)
	byAuthMethod := make(map[string]int)
	byPreferredTask := make(map[string]int)
	configuredCount := 0
	authenticatedCount := 0
	executableCount := 0

	for _, entry := range entries {
		byAuthMethod[entry.AuthMethod]++
		if entry.Configured {
			configuredCount++
		}
		if entry.Authenticated {
			authenticatedCount++
		}
		if entry.Executable {
			executableCount++
		}
		for _, task := range entry.PreferredTasks {
			byPreferredTask[task]++
		}
	}

	return Summary{
		ProviderCount:      len(entries),
		ConfiguredCount:    configuredCount,
		AuthenticatedCount: authenticatedCount,
		ExecutableCount:    executableCount,
		ByAuthMethod:       bucketsFromCounts(byAuthMethod),
		ByPreferredTask:    bucketsFromCounts(byPreferredTask),
	}
}

func bucketsFromCounts(values map[string]int) []SummaryBucket {
	buckets := make([]SummaryBucket, 0, len(values))
	for key, count := range values {
		if key == "" {
			key = "unknown"
		}
		buckets = append(buckets, SummaryBucket{Key: key, Count: count})
	}

	for i := 0; i < len(buckets); i++ {
		for j := i + 1; j < len(buckets); j++ {
			if buckets[j].Count > buckets[i].Count || (buckets[j].Count == buckets[i].Count && buckets[j].Key < buckets[i].Key) {
				buckets[i], buckets[j] = buckets[j], buckets[i]
			}
		}
	}
	return buckets
}
