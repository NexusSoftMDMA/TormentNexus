package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
)

type ExtractedRelation struct {
	SourceID string                 `json:"source_id"`
	TargetID string                 `json:"target_id"`
	RelType  string                 `json:"rel_type"`
	Weight   float64                `json:"weight"`
	Metadata map[string]interface{} `json:"metadata"`
}

// HandleMemoryExtractRelations parses entities and relations from text using LLM and registers them.
func HandleMemoryExtractRelations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if GlobalVectorStore == nil || GlobalVectorStore.RelationStore() == nil {
		return err("Graph RelationStore is not initialized")
	}

	text, hasText := getString(args, "text")
	if !hasText || text == "" {
		return err("missing or invalid required parameter 'text'")
	}

	prompt := fmt.Sprintf(`Extract key entities and their semantic/structural relationships from the following text.
Format the output as a valid JSON array of objects. Do not include markdown code block styling or any conversational text around the JSON.
Each object in the array MUST have the following structure:
{
  "source_id": "string (name of the source entity, lowercase, underscores for spaces)",
  "target_id": "string (name of the target entity, lowercase, underscores for spaces)",
  "rel_type": "string (type of relationship, e.g. works_at, located_in, depends_on, instance_of)",
  "weight": 1.0, // float value between 0.0 and 1.0 representing connection strength/relevance
  "metadata": {
    "description": "Short explanation of the relationship context"
  }
}

Text to analyze:
---
%s
---
`, text)

	messages := []ai.Message{
		{Role: "system", Content: "You are a specialized Cognee-style GraphRAG relation extraction system. Output JSON only."},
		{Role: "user", Content: prompt},
	}

	resp, errVal := ai.AutoRoute(ctx, messages)
	if errVal != nil {
		return err(fmt.Sprintf("LLM relation extraction failed: %v", errVal))
	}

	rawJSON := resp.Content
	// Strip potential markdown wrappers
	rawJSON = strings.TrimSpace(rawJSON)
	if strings.HasPrefix(rawJSON, "```json") {
		rawJSON = strings.TrimPrefix(rawJSON, "```json")
		rawJSON = strings.TrimSuffix(rawJSON, "```")
	} else if strings.HasPrefix(rawJSON, "```") {
		rawJSON = strings.TrimPrefix(rawJSON, "```")
		rawJSON = strings.TrimSuffix(rawJSON, "```")
	}
	rawJSON = strings.TrimSpace(rawJSON)

	var extracted []ExtractedRelation
	if errJSON := json.Unmarshal([]byte(rawJSON), &extracted); errJSON != nil {
		return err(fmt.Sprintf("failed to parse extracted relations JSON (LLM response was: %s): %v", rawJSON, errJSON))
	}

	store := GlobalVectorStore.RelationStore()
	var added []string
	for _, rel := range extracted {
		if rel.SourceID == "" || rel.TargetID == "" || rel.RelType == "" {
			continue
		}
		if rel.Weight <= 0 {
			rel.Weight = 1.0
		}
		if rel.Metadata == nil {
			rel.Metadata = map[string]interface{}{}
		}

		errAdd := store.AddRelation(ctx, rel.SourceID, rel.TargetID, rel.RelType, rel.Weight, rel.Metadata)
		if errAdd != nil {
			return err(fmt.Sprintf("failed to add relation (%s -> %s): %v", rel.SourceID, rel.TargetID, errAdd))
		}
		added = append(added, fmt.Sprintf("(%s) -[%s]-> (%s)", rel.SourceID, rel.RelType, rel.TargetID))
	}

	if len(added) == 0 {
		return ok("No valid relations were extracted from the text.")
	}

	return ok(fmt.Sprintf("Successfully extracted and inserted %d relations:\n%s", len(added), strings.Join(added, "\n")))
}
