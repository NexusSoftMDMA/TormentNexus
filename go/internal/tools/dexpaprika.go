//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Network synonyms mapping
var networkSynonyms = map[string][]string{
	"ethereum":  {"ethereum", "eth", "mainnet", "eth_mainnet", "ethereum_mainnet"},
	"solana":    {"solana", "sol"},
	"bsc":       {"bsc", "binance-smart-chain", "bnb", "binance", "bnb_chain", "bnb-chain"},
	"polygon":   {"polygon", "matic", "pol", "polygon_pos"},
	"arbitrum":  {"arbitrum", "arb", "arbitrum_one", "arbitrum-one"},
	"base":      {"base", "base_mainnet"},
	"optimism":  {"optimism", "op", "optimism_mainnet", "op_mainnet"},
	"avalanche": {"avalanche", "avalanche-c", "avax", "avalanche_c"},
	"sui":       {"sui"},
	"mantle":    {"mantle", "mnt"},
	"flow_evm":  {"flow_evm", "flow-evm", "flow"},
	"katana":    {"katana"},
	"unichain":  {"unichain", "uni"},
	"ronin":     {"ronin", "ron"},
	"x_layer":   {"x_layer", "x-layer", "xlayer", "okx_xlayer"},
	"linea":     {"linea"},
	"sonic":     {"sonic", "s"},
	"cronos":    {"cronos", "cro"},
	"sei":       {"sei"},
	"blast":     {"blast"},
	"tempo":     {"tempo"},
	"aptos":     {"aptos", "apt"},
	"zksync":    {"zksync", "zksync_era", "zksync-era"},
	"scroll":    {"scroll"},
	"tron":      {"tron", "trx"},
	"ton":       {"ton"},
	"plasma":    {"plasma"},
	"bob_network": {"bob_network", "bob", "bob-network"},
	"botanix":   {"botanix"},
	"fantom":    {"fantom", "ftm"},
	"celo":      {"celo"},
	"monad":     {"monad"},
	"megaeth":   {"megaeth", "mega-eth", "mega_eth"},
	"berachain": {"berachain", "bera"},
	"hyperevm":  {"hyperevm", "hyper-evm", "hyper_evm"},
}

var reverseSynonymMap = func() map[string]string {
	m := make(map[string]string)
	for canonical, alternates := range networkSynonyms {
		for _, alt := range alternates {
			m[strings.ToLower(alt)] = canonical
		}
	}
	return m
}()

func normalizeNetwork(input string) string {
	if canonical, ok := reverseSynonymMap[strings.ToLower(input)]; ok {
		return canonical
	}
	return input
}

func getStringSlice(args map[string]interface{}, key string) []string {
	if val, ok := args[key]; ok {
		if slice, okSlice := val.([]interface{}); okSlice {
			var res []string
			for _, item := range slice {
				if s, okS := item.(string); okS {
					res = append(res, s)
				}
			}
			return res
		}
		if slice, okSlice := val.([]string); okSlice {
			return slice
		}
	}
	return nil
}

// callDexPaprikaAPI queries the free DexPaprika REST endpoints.
func callDexPaprikaAPI(ctx context.Context, urlPath string, queryParams map[string]string) ([]byte, error) {
	baseUrl := os.Getenv("DEXPAPRIKA_API_URL")
	if baseUrl == "" {
		baseUrl = "https://api.dexpaprika.com"
	}
	reqUrl := baseUrl + urlPath

	req, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}

	if len(queryParams) > 0 {
		q := req.URL.Query()
		for k, v := range queryParams {
			if v != "" {
				q.Add(k, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dexpaprika API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// HandleDexPaprikaGetNetworks lists all supported networks.
func HandleDexPaprikaGetNetworks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	respData, err := callDexPaprikaAPI(ctx, "/networks", nil)
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(respData))
}

// HandleDexPaprikaGetCapabilities returns static server capabilities metadata.
func HandleDexPaprikaGetCapabilities(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	doc := map[string]interface{}{
		"name":             "dexpaprika",
		"aliases":          []string{"dexpapika", "dexpaprica", "dex-paprika", "dex paprika"},
		"server":           map[string]string{"name": "DexPaprika MCP", "version": "2.0.0"},
		"tools_count":      17,
		"stats": map[string]interface{}{
			"networks":      35,
			"tokens_approx": 29000000,
			"pools_approx":  31000000,
			"free":          true,
			"requires_api_key": false,
		},
		"network_synonyms": networkSynonyms,
		"workflows": map[string][]string{
			"discover_networks":        {"getNetworks"},
			"find_pools_on_network":    {"getNetworks", "getNetworkPools"},
			"filter_pools_by_volume":   {"getNetworks", "getNetworkPoolsFilter"},
			"find_new_pools":           {"getNetworkPoolsFilter with created_after", "sort_by=created_at sort_dir=desc"},
			"token_details_and_pools":  {"getTokenDetails", "getTokenPools"},
			"batch_price_lookup":       {"getTokenMultiPrices (max 10 tokens per call)"},
			"top_tokens_on_network":    {"getTopTokens"},
			"filter_tokens_by_metrics": {"filterNetworkTokens"},
			"historical_price_chart":   {"getPoolOHLCV with start + interval"},
			"recent_swaps":             {"getPoolTransactions with from/to UNIX timestamps"},
			"cross_network_search":     {"search with token name/symbol/address"},
		},
		"common_pitfalls": []string{
			"/pools (global) returns 410 Gone — use /networks/{network}/pools instead",
			"getTokenMultiPrices is capped at 10 tokens per request",
			"getPoolTransactions from/to are UNIX timestamps; results always capped to last 7 days",
			"Token addresses must match the network (e.g., don't send a Solana address to ethereum queries)",
		},
		"documentation": "https://docs.dexpaprika.com",
		"agent_skills":  "https://dexpaprika.com/agents/skill.md",
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(b))
}

// HandleDexPaprikaGetStats returns ecosystem-wide statistics.
func HandleDexPaprikaGetStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	respData, err := callDexPaprikaAPI(ctx, "/stats", nil)
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(respData))
}

// HandleDexPaprikaSearch performs cross-network searches.
func HandleDexPaprikaSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query")
	if query == "" {
		return err("query parameter is required")
	}

	path := fmt.Sprintf("/search?query=%s", url.QueryEscape(query))
	respData, err := callDexPaprikaAPI(ctx, path, nil)
	if err != nil {
		return errResponseDexPaprika(err)
	}

	// Optional client side slicing to mimic TS implementation
	limit := getInt(args, "limit")
	if limit > 0 {
		var result map[string]interface{}
		if err := json.Unmarshal(respData, &result); err == nil {
			for _, key := range []string{"tokens", "pools", "dexes"} {
				if slice, ok := result[key].([]interface{}); ok && len(slice) > limit {
					result[key] = slice[:limit]
				}
			}
			if indented, err := json.MarshalIndent(result, "", "  "); err == nil {
				return ok(string(indented))
			}
		}
	}

	return ok(string(respData))
}

// HandleDexPaprikaGetNetworkDexes lists DEXes on a network.
func HandleDexPaprikaGetNetworkDexes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ := getString(args, "network")
	if network == "" {
		return err("network parameter is required")
	}
	network = normalizeNetwork(network)

	page := getInt(args, "page")
	if page <= 0 {
		page = 1
	}
	limit := getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	sortDir, _ := getString(args, "sort_dir", "sort")
	if sortDir == "" {
		sortDir = "desc"
	}
	sortBy, _ := getString(args, "sort_by", "order_by")

	path := fmt.Sprintf("/networks/%s/dexes?page=%d&limit=%d&sort=%s", network, page, limit, sortDir)
	if sortBy != "" {
		path += "&order_by=" + sortBy
	}

	respData, err := callDexPaprikaAPI(ctx, path, nil)
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(respData))
}

// HandleDexPaprikaGetNetworkPools lists top pools for a network.
func HandleDexPaprikaGetNetworkPools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ := getString(args, "network")
	if network == "" {
		return err("network parameter is required")
	}
	network = normalizeNetwork(network)

	page := getInt(args, "page")
	if page <= 0 {
		page = 1
	}
	limit := getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	sortDir, _ := getString(args, "sort_dir", "sort")
	if sortDir == "" {
		sortDir = "desc"
	}
	sortBy, _ := getString(args, "sort_by", "order_by")
	if sortBy == "" {
		sortBy = "volume_usd"
	}

	path := fmt.Sprintf("/networks/%s/pools?page=%d&limit=%d&sort=%s&order_by=%s", network, page, limit, sortDir, sortBy)
	respData, err := callDexPaprikaAPI(ctx, path, nil)
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(respData))
}

// HandleDexPaprikaGetDexPools lists pools for a specific DEX.
func HandleDexPaprikaGetDexPools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ := getString(args, "network")
	if network == "" {
		return err("network parameter is required")
	}
	network = normalizeNetwork(network)

	dex, _ := getString(args, "dex")
	if dex == "" {
		return err("dex parameter is required")
	}

	page := getInt(args, "page")
	if page <= 0 {
		page = 1
	}
	limit := getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	sortDir, _ := getString(args, "sort_dir", "sort")
	if sortDir == "" {
		sortDir = "desc"
	}
	sortBy, _ := getString(args, "sort_by", "order_by")
	if sortBy == "" {
		sortBy = "volume_usd"
	}

	path := fmt.Sprintf("/networks/%s/dexes/%s/pools?page=%d&limit=%d&sort=%s&order_by=%s", network, dex, page, limit, sortDir, sortBy)
	respData, err := callDexPaprikaAPI(ctx, path, nil)
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(respData))
}

// HandleDexPaprikaGetNetworkPoolsFilter filters pools on a network.
func HandleDexPaprikaGetNetworkPoolsFilter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ := getString(args, "network")
	if network == "" {
		return err("network parameter is required")
	}
	network = normalizeNetwork(network)

	page := getInt(args, "page")
	if page <= 0 {
		page = 1
	}
	limit := getInt(args, "limit")
	if limit <= 0 {
		limit = 50
	}
	sortDir, _ := getString(args, "sort_dir", "sort")
	if sortDir == "" {
		sortDir = "desc"
	}
	sortBy, _ := getString(args, "sort_by", "order_by")
	if sortBy == "" {
		sortBy = "volume_24h"
	}

	queryParams := map[string]string{
		"page":     strconv.Itoa(page),
		"limit":    strconv.Itoa(limit),
		"sort_dir": sortDir,
		"sort_by":  sortBy,
	}

	for _, k := range []string{"volume_24h_min", "volume_24h_max", "volume_7d_min", "volume_7d_max", "liquidity_usd_min", "liquidity_usd_max", "txns_24h_min", "created_after", "created_before"} {
		if val, exists := args[k]; exists {
			switch v := val.(type) {
			case float64:
				queryParams[k] = strconv.FormatFloat(v, 'f', -1, 64)
			case int:
				queryParams[k] = strconv.Itoa(v)
			case string:
				queryParams[k] = v
			}
		}
	}

	path := fmt.Sprintf("/networks/%s/pools/filter", network)
	respData, err := callDexPaprikaAPI(ctx, path, queryParams)
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(respData))
}

// HandleDexPaprikaGetPoolDetails gets specific pool details.
func HandleDexPaprikaGetPoolDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ := getString(args, "network")
	if network == "" {
		return err("network parameter is required")
	}
	network = normalizeNetwork(network)

	poolAddress, _ := getString(args, "pool_address")
	if poolAddress == "" {
		return err("pool_address parameter is required")
	}

	queryParams := make(map[string]string)
	if inversedVal, exists := args["inversed"]; exists {
		if b, okB := inversedVal.(bool); okB {
			queryParams["inversed"] = strconv.FormatBool(b)
		}
	}

	urlPath := fmt.Sprintf("/networks/%s/pools/%s", network, poolAddress)
	respData, err := callDexPaprikaAPI(ctx, urlPath, queryParams)
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(respData))
}

// HandleDexPaprikaGetPoolOHLCV gets historical pool data.
func HandleDexPaprikaGetPoolOHLCV(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ := getString(args, "network")
	if network == "" {
		return err("network parameter is required")
	}
	network = normalizeNetwork(network)

	poolAddress, _ := getString(args, "pool_address")
	if poolAddress == "" {
		return err("pool_address parameter is required")
	}
	start, _ := getString(args, "start")
	if start == "" {
		return err("start parameter is required")
	}

	queryParams := map[string]string{
		"start": start,
	}
	if end, _ := getString(args, "end"); end != "" {
		queryParams["end"] = end
	}
	if limit := getInt(args, "limit"); limit > 0 {
		queryParams["limit"] = strconv.Itoa(limit)
	}
	if interval, _ := getString(args, "interval"); interval != "" {
		queryParams["interval"] = interval
	}
	if inversedVal, exists := args["inversed"]; exists {
		if b, okB := inversedVal.(bool); okB {
			queryParams["inversed"] = strconv.FormatBool(b)
		}
	}

	urlPath := fmt.Sprintf("/networks/%s/pools/%s/ohlcv", network, poolAddress)
	respData, err := callDexPaprikaAPI(ctx, urlPath, queryParams)
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(respData))
}

// HandleDexPaprikaGetPoolTransactions gets transactions for a pool.
func HandleDexPaprikaGetPoolTransactions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ := getString(args, "network")
	if network == "" {
		return err("network parameter is required")
	}
	network = normalizeNetwork(network)

	poolAddress, _ := getString(args, "pool_address")
	if poolAddress == "" {
		return err("pool_address parameter is required")
	}

	page := getInt(args, "page")
	if page <= 0 {
		page = 1
	}
	limit := getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}

	queryParams := map[string]string{
		"page":  strconv.Itoa(page),
		"limit": strconv.Itoa(limit),
	}

	if cursor, _ := getString(args, "cursor"); cursor != "" {
		queryParams["cursor"] = cursor
	}
	if from := getInt(args, "from"); from > 0 {
		queryParams["from"] = strconv.Itoa(from)
	}
	if to := getInt(args, "to"); to > 0 {
		queryParams["to"] = strconv.Itoa(to)
	}

	urlPath := fmt.Sprintf("/networks/%s/pools/%s/transactions", network, poolAddress)
	respData, err := callDexPaprikaAPI(ctx, urlPath, queryParams)
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(respData))
}

// HandleDexPaprikaGetTokenDetails gets detailed token information.
func HandleDexPaprikaGetTokenDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ := getString(args, "network")
	if network == "" {
		return err("network parameter is required")
	}
	network = normalizeNetwork(network)

	tokenAddress, _ := getString(args, "token_address")
	if tokenAddress == "" {
		return err("token_address parameter is required")
	}

	urlPath := fmt.Sprintf("/networks/%s/tokens/%s", network, tokenAddress)
	respData, err := callDexPaprikaAPI(ctx, urlPath, nil)
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(respData))
}

// HandleDexPaprikaGetTokenPools gets liquidity pools containing a token.
func HandleDexPaprikaGetTokenPools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ := getString(args, "network")
	if network == "" {
		return err("network parameter is required")
	}
	network = normalizeNetwork(network)

	tokenAddress, _ := getString(args, "token_address")
	if tokenAddress == "" {
		return err("token_address parameter is required")
	}

	page := getInt(args, "page")
	if page <= 0 {
		page = 1
	}
	limit := getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	sortDir, _ := getString(args, "sort_dir", "sort")
	if sortDir == "" {
		sortDir = "desc"
	}
	sortBy, _ := getString(args, "sort_by", "order_by")
	if sortBy == "" {
		sortBy = "volume_usd"
	}

	queryParams := map[string]string{
		"page":     strconv.Itoa(page),
		"limit":    strconv.Itoa(limit),
		"sort":     sortDir,
		"order_by": sortBy,
	}

	if inversed, okB := args["inversed"].(bool); okB {
		queryParams["reorder"] = strconv.FormatBool(inversed)
	} else if reorder, okR := args["reorder"].(bool); okR {
		queryParams["reorder"] = strconv.FormatBool(reorder)
	}

	if address, _ := getString(args, "paired_token_address", "address"); address != "" {
		queryParams["address"] = address
	}

	urlPath := fmt.Sprintf("/networks/%s/tokens/%s/pools", network, tokenAddress)
	respData, err := callDexPaprikaAPI(ctx, urlPath, queryParams)
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(respData))
}

// HandleDexPaprikaGetTokenMultiPrices gets prices for multiple tokens.
func HandleDexPaprikaGetTokenMultiPrices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ := getString(args, "network")
	if network == "" {
		return err("network parameter is required")
	}
	network = normalizeNetwork(network)

	tokens := getStringSlice(args, "tokens")
	if len(tokens) == 0 {
		return err("tokens parameter is required and cannot be empty")
	}
	if len(tokens) > 10 {
		b, _ := json.Marshal(map[string]interface{}{
			"error":    "Too many tokens",
			"message":  "getTokenMultiPrices accepts at most 10 tokens per call.",
			"provided": len(tokens),
			"limit":    10,
		})
		return ok(string(b))
	}

	joined := strings.Join(tokens, ",")
	urlPath := fmt.Sprintf("/networks/%s/multi/prices", network)
	queryParams := map[string]string{
		"tokens": joined,
	}

	respData, err := callDexPaprikaAPI(ctx, urlPath, queryParams)
	if err != nil {
		return errResponseDexPaprika(err)
	}

	var upstream []interface{}
	if err := json.Unmarshal(respData, &upstream); err == nil {
		returnedIds := make(map[string]bool)
		for _, item := range upstream {
			if m, okM := item.(map[string]interface{}); okM {
				if id, okI := m["id"].(string); okI {
					returnedIds[strings.ToLower(id)] = true
				}
			}
		}

		var missing []string
		for _, t := range tokens {
			if !returnedIds[strings.ToLower(t)] {
				missing = append(missing, t)
			}
		}

		enriched := map[string]interface{}{
			"prices":         upstream,
			"missing_tokens": missing,
		}
		b, _ := json.MarshalIndent(enriched, "", "  ")
		return ok(string(b))
	}

	return ok(string(respData))
}

// HandleDexPaprikaFilterNetworkTokens filters tokens on a network.
func HandleDexPaprikaFilterNetworkTokens(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ := getString(args, "network")
	if network == "" {
		return err("network parameter is required")
	}
	network = normalizeNetwork(network)

	page := getInt(args, "page")
	if page <= 0 {
		page = 1
	}
	limit := getInt(args, "limit")
	if limit <= 0 {
		limit = 50
	}
	sortDir, _ := getString(args, "sort_dir", "sort")
	if sortDir == "" {
		sortDir = "desc"
	}
	sortBy, _ := getString(args, "sort_by", "order_by")
	if sortBy == "" {
		sortBy = "volume_24h"
	}

	queryParams := map[string]string{
		"page":     strconv.Itoa(page),
		"limit":    strconv.Itoa(limit),
		"sort_dir": sortDir,
		"sort_by":  sortBy,
	}

	for _, k := range []string{"volume_24h_min", "volume_24h_max", "liquidity_usd_min", "fdv_min", "fdv_max", "txns_24h_min", "created_after", "created_before"} {
		if val, exists := args[k]; exists {
			switch v := val.(type) {
			case float64:
				queryParams[k] = strconv.FormatFloat(v, 'f', -1, 64)
			case int:
				queryParams[k] = strconv.Itoa(v)
			case string:
				queryParams[k] = v
			}
		}
	}

	path := fmt.Sprintf("/networks/%s/tokens/filter", network)
	respData, err := callDexPaprikaAPI(ctx, path, queryParams)
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(respData))
}

// HandleDexPaprikaGetTopTokens gets top tokens on a network.
func HandleDexPaprikaGetTopTokens(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ := getString(args, "network")
	if network == "" {
		return err("network parameter is required")
	}
	network = normalizeNetwork(network)

	page := getInt(args, "page")
	if page <= 0 {
		page = 1
	}
	limit := getInt(args, "limit")
	if limit <= 0 {
		limit = 50
	}
	sortDir, _ := getString(args, "sort_dir", "sort")
	if sortDir == "" {
		sortDir = "desc"
	}
	sortBy, _ := getString(args, "sort_by", "order_by")
	if sortBy == "" {
		sortBy = "volume_24h"
	}

	path := fmt.Sprintf("/networks/%s/tokens/top?page=%d&limit=%d&sort=%s&order_by=%s", network, page, limit, sortDir, sortBy)
	respData, err := callDexPaprikaAPI(ctx, path, nil)
	if err != nil {
		return errResponseDexPaprika(err)
	}
	return ok(string(respData))
}

// HandleDexPaprikaSubmitFeedback structured feedback submission.
func HandleDexPaprikaSubmitFeedback(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	severity, _ := getString(args, "severity")
	if severity == "" {
		severity = "minor"
	}
	ack := map[string]interface{}{
		"ok":          true,
		"tracking_id": nil,
		"message":     "Thanks. This self-host build does not persist feedback; please open an issue at https://github.com/coinpaprika/dexpaprika-mcp for anything actionable.",
		"severity":    severity,
	}
	b, _ := json.MarshalIndent(ack, "", "  ")
	return ok(string(b))
}

func errResponseDexPaprika(err error) (ToolResponse, error) {
	return ToolResponse{
		Content: []TextContent{{Type: "text", Text: fmt.Sprintf("DexPaprika API failed: %v", err)}},
		IsError: true,
	}, nil
}
