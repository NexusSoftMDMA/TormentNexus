package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetBlock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	blockNumber, _ :=getInt(args, "block_number")
	blockHash, _ :=getString(args, "block_hash")
	if blockNumber == 0 && blockHash == "" {
		return err("must provide block_number or block_hash")
}

	apiKey := os.Getenv("CARDANO_API_KEY")
	var url string
	if blockHash != "" {
		url = fmt.Sprintf("https://cardano-mainnet.blockfrost.io/api/v0/blocks/%s", blockHash)
	} else {
		url = fmt.Sprintf("https://cardano-mainnet.blockfrost.io/api/v0/blocks/%d", blockNumber)

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("project_id", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return success(string(body))
}
}