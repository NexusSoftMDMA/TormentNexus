package mcpimpl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleDexPaprikaTools(t *testing.T) {
	// Start mock DexPaprika API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		path := r.URL.Path
		switch path {
		case "/networks":
			w.Write([]byte(`[{"id": "ethereum", "name": "Ethereum"}]`))
		case "/stats":
			w.Write([]byte(`{"chains": 35, "pools": 31000000}`))
		case "/search":
			w.Write([]byte(`{"tokens": [{"id": "token1", "symbol": "T1"}], "pools": [], "dexes": []}`))
		case "/networks/ethereum/dexes":
			w.Write([]byte(`{"dexes": [{"id": "uniswap_v3", "name": "Uniswap V3"}]}`))
		case "/networks/ethereum/pools":
			w.Write([]byte(`{"pools": [{"id": "pool1", "name": "USDC/WETH"}]}`))
		case "/networks/ethereum/dexes/uniswap_v3/pools":
			w.Write([]byte(`{"pools": [{"id": "pool1", "name": "USDC/WETH"}]}`))
		case "/networks/ethereum/pools/filter":
			w.Write([]byte(`{"results": [{"id": "pool1", "name": "USDC/WETH"}]}`))
		case "/networks/ethereum/pools/pool1":
			w.Write([]byte(`{"id": "pool1", "name": "USDC/WETH", "fee": 300}`))
		case "/networks/ethereum/pools/pool1/ohlcv":
			w.Write([]byte(`{"ohlcv": [{"open": 1.0, "close": 1.1}]}`))
		case "/networks/ethereum/pools/pool1/transactions":
			w.Write([]byte(`{"transactions": [{"id": "tx1", "amount_usd": 100}]}`))
		case "/networks/ethereum/tokens/token1":
			w.Write([]byte(`{"id": "token1", "symbol": "T1", "name": "Token 1"}`))
		case "/networks/ethereum/tokens/token1/pools":
			w.Write([]byte(`{"pools": [{"id": "pool1", "name": "USDC/WETH"}]}`))
		case "/networks/ethereum/multi/prices":
			w.Write([]byte(`[{"id": "token1", "price_usd": 1.5}]`))
		case "/networks/ethereum/tokens/filter":
			w.Write([]byte(`{"results": [{"id": "token1", "symbol": "T1"}]}`))
		case "/networks/ethereum/tokens/top":
			w.Write([]byte(`{"tokens": [{"id": "token1", "symbol": "T1"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	os.Setenv("DEXPAPRIKA_API_URL", server.URL)
	defer os.Unsetenv("DEXPAPRIKA_API_URL")

	ctx := context.Background()

	// 1: HandleDexPaprikaGetNetworks
	resp, err := HandleDexPaprikaGetNetworks(ctx, nil)
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetNetworks failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "ethereum") {
		t.Errorf("Expected networks list, got: %s", resp.Content[0].Text)
	}

	// 2: HandleDexPaprikaGetCapabilities
	resp, err = HandleDexPaprikaGetCapabilities(ctx, nil)
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetCapabilities failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "DexPaprika MCP") {
		t.Errorf("Expected capabilities doc, got: %s", resp.Content[0].Text)
	}

	// 3: HandleDexPaprikaGetStats
	resp, err = HandleDexPaprikaGetStats(ctx, nil)
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetStats failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "chains") {
		t.Errorf("Expected stats, got: %s", resp.Content[0].Text)
	}

	// 4: HandleDexPaprikaSearch
	resp, err = HandleDexPaprikaSearch(ctx, map[string]interface{}{"query": "uniswap", "limit": 5})
	if err != nil {
		t.Fatalf("HandleDexPaprikaSearch failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "token1") {
		t.Errorf("Expected search results, got: %s", resp.Content[0].Text)
	}

	// 5: HandleDexPaprikaGetNetworkDexes
	resp, err = HandleDexPaprikaGetNetworkDexes(ctx, map[string]interface{}{"network": "eth"})
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetNetworkDexes failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "uniswap_v3") {
		t.Errorf("Expected dexes, got: %s", resp.Content[0].Text)
	}

	// 6: HandleDexPaprikaGetNetworkPools
	resp, err = HandleDexPaprikaGetNetworkPools(ctx, map[string]interface{}{"network": "ethereum"})
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetNetworkPools failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "USDC/WETH") {
		t.Errorf("Expected pool list, got: %s", resp.Content[0].Text)
	}

	// 7: HandleDexPaprikaGetDexPools
	resp, err = HandleDexPaprikaGetDexPools(ctx, map[string]interface{}{"network": "ethereum", "dex": "uniswap_v3"})
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetDexPools failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "USDC/WETH") {
		t.Errorf("Expected dex pools, got: %s", resp.Content[0].Text)
	}

	// 8: HandleDexPaprikaGetNetworkPoolsFilter
	resp, err = HandleDexPaprikaGetNetworkPoolsFilter(ctx, map[string]interface{}{
		"network":        "ethereum",
		"volume_24h_min": 1000.0,
	})
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetNetworkPoolsFilter failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "results") {
		t.Errorf("Expected filtered pools results, got: %s", resp.Content[0].Text)
	}

	// 9: HandleDexPaprikaGetPoolDetails
	resp, err = HandleDexPaprikaGetPoolDetails(ctx, map[string]interface{}{
		"network":      "ethereum",
		"pool_address": "pool1",
	})
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetPoolDetails failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "fee") {
		t.Errorf("Expected pool details, got: %s", resp.Content[0].Text)
	}

	// 10: HandleDexPaprikaGetPoolOHLCV
	resp, err = HandleDexPaprikaGetPoolOHLCV(ctx, map[string]interface{}{
		"network":      "ethereum",
		"pool_address": "pool1",
		"start":        "2026-06-01",
	})
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetPoolOHLCV failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "ohlcv") {
		t.Errorf("Expected historical ohlcv bucket, got: %s", resp.Content[0].Text)
	}

	// 11: HandleDexPaprikaGetPoolTransactions
	resp, err = HandleDexPaprikaGetPoolTransactions(ctx, map[string]interface{}{
		"network":      "ethereum",
		"pool_address": "pool1",
	})
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetPoolTransactions failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "transactions") {
		t.Errorf("Expected transactions, got: %s", resp.Content[0].Text)
	}

	// 12: HandleDexPaprikaGetTokenDetails
	resp, err = HandleDexPaprikaGetTokenDetails(ctx, map[string]interface{}{
		"network":       "ethereum",
		"token_address": "token1",
	})
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetTokenDetails failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "T1") {
		t.Errorf("Expected token details, got: %s", resp.Content[0].Text)
	}

	// 13: HandleDexPaprikaGetTokenPools
	resp, err = HandleDexPaprikaGetTokenPools(ctx, map[string]interface{}{
		"network":       "ethereum",
		"token_address": "token1",
	})
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetTokenPools failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "USDC/WETH") {
		t.Errorf("Expected token pools, got: %s", resp.Content[0].Text)
	}

	// 14: HandleDexPaprikaGetTokenMultiPrices
	resp, err = HandleDexPaprikaGetTokenMultiPrices(ctx, map[string]interface{}{
		"network": "ethereum",
		"tokens":  []string{"token1", "token2"},
	})
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetTokenMultiPrices failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "prices") || !strings.Contains(resp.Content[0].Text, "missing_tokens") {
		t.Errorf("Expected multi price lookup response, got: %s", resp.Content[0].Text)
	}

	// 15: HandleDexPaprikaFilterNetworkTokens
	resp, err = HandleDexPaprikaFilterNetworkTokens(ctx, map[string]interface{}{
		"network":        "ethereum",
		"volume_24h_min": 1000.0,
	})
	if err != nil {
		t.Fatalf("HandleDexPaprikaFilterNetworkTokens failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "results") {
		t.Errorf("Expected filtered tokens, got: %s", resp.Content[0].Text)
	}

	// 16: HandleDexPaprikaGetTopTokens
	resp, err = HandleDexPaprikaGetTopTokens(ctx, map[string]interface{}{"network": "ethereum"})
	if err != nil {
		t.Fatalf("HandleDexPaprikaGetTopTokens failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "tokens") {
		t.Errorf("Expected top tokens, got: %s", resp.Content[0].Text)
	}

	// 17: HandleDexPaprikaSubmitFeedback
	resp, err = HandleDexPaprikaSubmitFeedback(ctx, map[string]interface{}{"goal": "Test native port implementation"})
	if err != nil {
		t.Fatalf("HandleDexPaprikaSubmitFeedback failed: %v", err)
	}
	if !strings.Contains(resp.Content[0].Text, "severity") {
		t.Errorf("Expected feedback response, got: %s", resp.Content[0].Text)
	}
}
