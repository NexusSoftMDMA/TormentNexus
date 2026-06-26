package httpapi

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

func localMemoryInterchangeFormats() []map[string]any {
	return []map[string]any{
		{
			"id":          "json",
			"label":       "Canonical JSON",
			"kind":        "canonical",
			"extension":   "json",
			"description": "Portable TormentNexus memory export with metadata preserved.",
		},
		{
			"id":          "csv",
			"label":       "Canonical CSV",
			"kind":        "canonical",
			"extension":   "csv",
			"description": "Spreadsheet-friendly memory export.",
		},
		{
			"id":          "jsonl",
			"label":       "Canonical JSONL",
			"kind":        "canonical",
			"extension":   "jsonl",
			"description": "Streaming-friendly newline-delimited memory export.",
		},
		{
			"id":          "json-provider",
			"label":       "TormentNexus JSON Provider",
			"kind":        "provider",
			"extension":   "json",
			"description": "Native snapshot of TormentNexus's flat-file memory provider.",
		},
		{
			"id":          "sectioned-memory-store",
			"label":       "Sectioned Memory Store",
			"kind":        "provider",
			"extension":   "json",
			"description": "Native TormentNexus sectioned memory snapshot.",
		},
	}
}

func localNormalizeMemoryImportRecord(record map[string]any, userID string, fallbackIndex int) map[string]any {
	item := cloneMap(record)
	uuid := stringValue(firstNonEmptyString(item["uuid"], item["id"]))
	if strings.TrimSpace(uuid) == "" {
		uuid = localMemoryContextID(item, fallbackIndex)
	}
	item["uuid"] = uuid

	metadata, _ := item["metadata"].(map[string]any)
	metadata = cloneMap(metadata)

	title := stringValue(firstNonEmptyString(item["title"], metadata["title"]))
	source := stringValue(firstNonEmptyString(item["source"], metadata["source"], "user"))
	url := stringValue(firstNonEmptyString(item["url"], metadata["url"]))
	if title != "" {
		item["title"] = title
		metadata["title"] = title
	}
	if source != "" {
		item["source"] = source
		metadata["source"] = source
	}
	if url != "" {
		item["url"] = url
		metadata["url"] = url
	}
	item["metadata"] = metadata

	if strings.TrimSpace(stringValue(item["userId"])) == "" {
		item["userId"] = userID
	}
	if strings.TrimSpace(stringValue(item["createdAt"])) == "" {
		item["createdAt"] = time.Now().UTC().Format(time.RFC3339)
	}
	if _, ok := item["content"]; !ok {
		item["content"] = ""
	}
	return item
}

func localNormalizeMemoryImportRecords(records []map[string]any, userID string) []map[string]any {
	normalized := make([]map[string]any, 0, len(records))
	for index, record := range records {
		normalized = append(normalized, localNormalizeMemoryImportRecord(record, userID, index+1))
	}
	return ensureLocalMemoryExportFields(normalized, userID)
}

func localParseMemoryJSONRecords(data, userID string) ([]map[string]any, error) {
	var parsed []map[string]any
	if err := json.Unmarshal([]byte(data), &parsed); err != nil {
		return nil, err
	}
	return localNormalizeMemoryImportRecords(parsed, userID), nil
}

func localParseMemoryJSONLRecords(data, userID string) ([]map[string]any, error) {
	lines := strings.Split(data, "\n")
	records := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return localNormalizeMemoryImportRecords(records, userID), nil
}

func localParseMemoryCSVLine(line string) []string {
	result := []string{}
	current := strings.Builder{}
	inQuotes := false
	for index := 0; index < len(line); index++ {
		char := line[index]
		switch {
		case char == '"':
			if inQuotes && index+1 < len(line) && line[index+1] == '"' {
				current.WriteByte('"')
				index++
			} else {
				inQuotes = !inQuotes
			}
		case char == ',' && !inQuotes:
			result = append(result, current.String())
			current.Reset()
		default:
			current.WriteByte(char)
		}
	}
	result = append(result, current.String())
	return result
}

func localParseMemoryCSVRecords(data, userID string) ([]map[string]any, error) {
	lines := strings.Split(data, "\n")
	trimmed := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			trimmed = append(trimmed, line)
		}
	}
	if len(trimmed) <= 1 {
		return []map[string]any{}, nil
	}
	records := make([]map[string]any, 0, len(trimmed)-1)
	for _, line := range trimmed[1:] {
		fields := localParseMemoryCSVLine(line)
		metadata := map[string]any{}
		if len(fields) > 5 && strings.TrimSpace(fields[5]) != "" {
			if err := json.Unmarshal([]byte(fields[5]), &metadata); err != nil {
				metadata = map[string]any{}
			}
		}
		records = append(records, map[string]any{
			"uuid":      stringValue(fieldValue(fields, 0)),
			"content":   stringValue(fieldValue(fields, 1)),
			"userId":    stringValue(firstNonEmptyString(fieldValue(fields, 2), userID)),
			"agentId":   stringValue(fieldValue(fields, 3)),
			"createdAt": stringValue(fieldValue(fields, 4)),
			"metadata":  metadata,
		})
	}
	return localNormalizeMemoryImportRecords(records, userID), nil
}

func fieldValue(fields []string, index int) any {
	if index >= 0 && index < len(fields) {
		return fields[index]
	}
	return nil
}

func localParseMemorySectionedStoreRecords(data, userID string) ([]map[string]any, error) {
	var parsed struct {
		Sections []struct {
			Section string `json:"section"`
			Entries []struct {
				UUID      string   `json:"uuid"`
				Content   string   `json:"content"`
				Tags      []string `json:"tags"`
				CreatedAt string   `json:"createdAt"`
				Source    string   `json:"source"`
			} `json:"entries"`
		} `json:"sections"`
	}
	if err := json.Unmarshal([]byte(data), &parsed); err != nil {
		return nil, err
	}
	records := []map[string]any{}
	for _, section := range parsed.Sections {
		sectionName := strings.TrimSpace(section.Section)
		if sectionName == "" {
			sectionName = "general"
		}
		for _, entry := range section.Entries {
			records = append(records, map[string]any{
				"uuid":      entry.UUID,
				"content":   entry.Content,
				"createdAt": entry.CreatedAt,
				"source":    entry.Source,
				"metadata": map[string]any{
					"section":  sectionName,
					"tags":     entry.Tags,
					"source":   entry.Source,
					"provider": "sectioned-store",
				},
			})
		}
	}
	return localNormalizeMemoryImportRecords(records, userID), nil
}

func (s *Server) localParseMemoryInterchange(data, format, userID string) ([]map[string]any, error) {
	if strings.TrimSpace(userID) == "" {
		userID = "default"
	}
	switch strings.TrimSpace(format) {
	case "", "json", "json-provider":
		return localParseMemoryJSONRecords(data, userID)
	case "jsonl":
		return localParseMemoryJSONLRecords(data, userID)
	case "csv":
		return localParseMemoryCSVRecords(data, userID)
	case "sectioned-memory-store":
		return localParseMemorySectionedStoreRecords(data, userID)
	default:
		return nil, errors.New("unsupported memory format: " + format)
	}
}

func (s *Server) localImportMemories(data, format, userID string) (map[string]any, error) {
	records, err := s.localParseMemoryInterchange(data, format, userID)
	if err != nil {
		return nil, err
	}
	contexts, err := s.localMemoryContexts()
	if err != nil {
		return nil, err
	}
	remaining := make([]map[string]any, 0, len(contexts))
	importIDs := map[string]struct{}{}
	for _, record := range records {
		importIDs[stringValue(record["uuid"])] = struct{}{}
	}
	for index, context := range contexts {
		if _, exists := importIDs[localMemoryContextID(context, index+1)]; exists {
			continue
		}
		remaining = append(remaining, context)
	}
	importedContexts := make([]map[string]any, 0, len(records))
	for _, record := range records {
		metadata, _ := record["metadata"].(map[string]any)
		importedContexts = append(importedContexts, map[string]any{
			"id":        stringValue(record["uuid"]),
			"title":     stringValue(firstNonEmptyString(record["title"], metadata["title"], record["url"], "Untitled")),
			"source":    stringValue(firstNonEmptyString(record["source"], metadata["source"], "unknown")),
			"url":       stringValue(firstNonEmptyString(record["url"], metadata["url"])),
			"content":   stringValue(record["content"]),
			"createdAt": stringValue(record["createdAt"]),
			"chunks":    1,
			"metadata":  cloneMap(metadata),
		})
	}
	updated := append(importedContexts, remaining...)
	if err := s.localWriteMemoryContexts(updated); err != nil {
		return nil, err
	}
	return map[string]any{
		"imported":   len(records),
		"errors":     0,
		"importedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Server) localConvertMemories(data, fromFormat, toFormat, userID string) (map[string]any, error) {
	records, err := s.localParseMemoryInterchange(data, fromFormat, userID)
	if err != nil {
		return nil, err
	}
	converted, err := localSerializeMemoryInterchange(records, toFormat, userID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"data":        converted,
		"fromFormat":  fromFormat,
		"toFormat":    toFormat,
		"convertedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func localSerializeMemoryInterchange(records []map[string]any, format, userID string) (string, error) {
	normalized := ensureLocalMemoryExportFields(records, userID)
	switch strings.TrimSpace(format) {
	case "", "json", "json-provider":
		data, err := json.MarshalIndent(normalized, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	case "jsonl":
		lines := make([]string, 0, len(normalized))
		for _, record := range normalized {
			data, err := json.Marshal(record)
			if err != nil {
				return "", err
			}
			lines = append(lines, string(data))
		}
		return strings.Join(lines, "\n"), nil
	case "csv":
		return serializeLocalMemoryCSV(normalized), nil
	case "sectioned-memory-store":
		data, err := json.MarshalIndent(localMemorySectionedStore(normalized), "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	default:
		return "", errors.New("unsupported memory export format: " + format)
	}
}
