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

func HandleGetMarketTicker(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instId, _ :=getString(args, "instId")
	if instId == "" {
		return err("instId required")
	}
	url := fmt.Sprintf("https://www.okx.com/api/v5/market/ticker?instId=%s", instId)
	req, e := http.NewRequestWithContext


-reasoner (deepseek)*
}