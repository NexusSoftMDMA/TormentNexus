package tools

import (
	"context"
	"fmt"
)

// HandleMemoryScratchpadGet handles the memory_scratchpad_get tool.
func HandleMemoryScratchpadGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, hasKey := getString(args, "key")
	if !hasKey || key == "" {
		return err("missing or invalid required parameter 'key'")
	}

	if GlobalVectorStore == nil {
		return err("memorystore VectorStore is not initialized")
	}

	val, errVal := GlobalVectorStore.GetScratchpadValue(ctx, key)
	if errVal != nil {
		return err(fmt.Sprintf("failed to get scratchpad value: %v", errVal))
	}

	return ok(val)
}

// HandleMemoryScratchpadSet handles the memory_scratchpad_set tool.
func HandleMemoryScratchpadSet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, hasKey := getString(args, "key")
	if !hasKey || key == "" {
		return err("missing or invalid required parameter 'key'")
	}
	value, hasVal := getString(args, "value")
	if !hasVal {
		return err("missing or invalid required parameter 'value'")
	}

	if GlobalVectorStore == nil {
		return err("memorystore VectorStore is not initialized")
	}

	errVal := GlobalVectorStore.SetScratchpadValue(ctx, key, value)
	if errVal != nil {
		return err(fmt.Sprintf("failed to set scratchpad value: %v", errVal))
	}

	return ok(fmt.Sprintf("successfully set core memory scratchpad key '%s'", key))
}

// HandleMemoryScratchpadAppend handles the memory_scratchpad_append tool.
func HandleMemoryScratchpadAppend(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, hasKey := getString(args, "key")
	if !hasKey || key == "" {
		return err("missing or invalid required parameter 'key'")
	}
	value, hasVal := getString(args, "value")
	if !hasVal {
		return err("missing or invalid required parameter 'value'")
	}

	if GlobalVectorStore == nil {
		return err("memorystore VectorStore is not initialized")
	}

	errVal := GlobalVectorStore.AppendScratchpadValue(ctx, key, value)
	if errVal != nil {
		return err(fmt.Sprintf("failed to append scratchpad value: %v", errVal))
	}

	return ok(fmt.Sprintf("successfully appended to core memory scratchpad key '%s'", key))
}
