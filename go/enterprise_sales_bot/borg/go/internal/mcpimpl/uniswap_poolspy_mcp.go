package mcpimpl

import "context"

func HandleGetPools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(`{"pools":[{"id":"0x...","name":"USDC/ETH"}]}`)
}

func HandleGetPoolInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	poolId, _ :=getString(args, "poolId")
	if poolId == "" {
		return err("poolId is required")
}

	return ok(`{"id":"` + poolId + `","name":"USDC/ETH","tvl":"1000000"}`)
}