package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type chainInfo struct {
	Name    string `json:"name"`
	ChainID int    `json:"chainId"`
	RPC     []string `json:"rpc"`
}

func HandleGetChainInfo_chainlist_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chainID, _ :=getInt(args, "chainId")
	if chainID == 0 {
		return err("missing or invalid chainId")
}

	url := fmt.Sprintf("https://chainid.network/chains.json")
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch chain list")
}

	defer resp.Body.Close()

	var chains []chainInfo
	if e = json.NewDecoder(resp.Body).Decode(&chains); e != nil {
		return err("failed to parse chain list")
}

	for _, c := range chains {
		if c.ChainID == chainID {
			return ok(fmt.Sprintf("Chain: %s (ID: %d), RPC: %v", c.Name, c.ChainID, c.RPC))

	}
	return err("chain not found")
}
}