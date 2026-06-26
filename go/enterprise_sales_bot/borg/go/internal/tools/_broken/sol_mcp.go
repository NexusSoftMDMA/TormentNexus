package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetBalance_sol_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	body := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"getBalance","params":["%s"]}`, address)
	resp, e := http.DefaultClient.Post("https://api.mainnet-beta.solana.com", "application/json", stringToReader(body))
	if e != nil {
		return err("failed to call Solana RPC")
}

	defer resp.Body.Close()
	b, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result struct {
		Result struct {
			Value int `json:"value"`
		} `json:"result"`
	}
	if e := json.Unmarshal(b, &result); e != nil {
		return err("failed to parse JSON")
}

	return ok(fmt.Sprintf("Balance: %d lamports", result.Result.Value))
}

func stringToReader_sol_mcp(s string) io.Reader {
	return &stringReader{s: s}
}

type stringReader_sol_mcp struct {
	s string
	i int
}

func (r *stringReader) Read(p []byte) (int, error) {
	if r.i >= len(r.s) {
		return 0, io.EOF
	}
	n := copy(p, r.s[r.i:])
	r.i += n
	return n, nil
}