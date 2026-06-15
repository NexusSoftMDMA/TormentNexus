//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tools := []map[string]interface{}{
		{
			"name":        "openapi",
			"description": "Fetch and parse an OpenAPI specification from a URL",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "The URL of the OpenAPI")


-reasoner (deepseek)*,
},
},
},
},
},
}