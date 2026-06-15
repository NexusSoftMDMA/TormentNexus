//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetCandidate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	candidateID, _ :=getString(args, "candidate_id")
	if candidateID == "" {
		return err("missing candidate_id")
}

	cycle, _ :=getInt(args, "cycle")
	if cycle == 0 {
		


-reasoner (deepseek)*,
},
}