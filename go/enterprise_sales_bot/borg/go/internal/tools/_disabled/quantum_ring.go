package tools

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
)

func HandleGenerateRandom(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bits, _ :=getInt(args, "bits")
	if bits <= 0 {
		bits = 16
	}
	max := new(big.Int).Lsh(big.NewInt(1), uint(bits))
	n, e := rand.Int(rand.Reader, max)
	if e != nil {
		return err("failed to generate random number")
}

	return ok(fmt.Sprintf("Quantum random number (%d bits): %s", bits, n.String()))
}

func HandleRingInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = args
	return ok("Quantum Ring MCP server – provides quantum random number generation and ring information.")
}