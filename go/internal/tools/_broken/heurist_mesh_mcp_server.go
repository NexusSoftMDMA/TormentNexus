package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultMeshAPIEndpoint = "https://mesh.heurist.xyz"

func getMeshAPIKey() string {
	key := os.Getenv("HEURIST_API_KEY")
	if key == "" {
		key = os.Getenv("X_HEURIST_API_KEY")

	return key
}

}

func getMeshAPIEndpoint() string {
	ep := os.Getenv("MESH_API_ENDPOINT")
	if ep == "" {
		return defaultMeshAPIEndpoint
	}
	return ep
}

func callMeshAPI(agentID string, toolName string, toolParams map[string]interface{}) (map[string]interface{}, error) {
	apiKey := getMeshAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("API key required. Set HEURIST_API_KEY environment variable")
}

	endpoint := getMeshAPIEndpoint()
	reqURL := endpoint + "/mesh_request"

	requestData := map[string]interface{}{
		"agent_id": agentID,
		"input": map[string]interface{}{
			"tool":           toolName,
			"tool_arguments": toolParams,
			"raw_data_only":  true,
		},
		"api_key": apiKey,
	}

	bodyBytes, marshalErr := json.Marshal(requestData)
	if marshalErr != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", marshalErr)
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(context.Background(), "POST", reqURL, strings.NewReader(string(bodyBytes)))
	if reqErr != nil {
		return nil, fmt.Errorf("failed to create request: %v", reqErr)
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-HEURIST-API-KEY", apiKey)

	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return nil, fmt.Errorf("mesh API request failed: %v", httpErr)
}

	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response: %v", readErr)
}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("mesh API error (HTTP %d): %s", resp.StatusCode, string(respBody))
}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(respBody, &result); parseErr != nil {
		return nil, fmt.Errorf("failed to parse response: %v", parseErr)
}

	return result, nil
}

func callMeshAPIRaw(agentID string, toolName string, toolParams map[string]interface{}) (string, error) {
	result, apiErr := callMeshAPI(agentID, toolName, toolParams)
	if apiErr != nil {
		return "", apiErr
	}
	pretty, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return "", fmt.Errorf("failed to format response: %v", marshalErr)
}

	return string(pretty), nil
}

// HandleTokenSearch finds tokens by address, symbol, name, or CoinGecko ID
func HandleTokenSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	params := map[string]interface{}{
		"query": query,
	}

	result, apiErr := callMeshAPIRaw("TokenResolverAgent", "token_search", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleTokenProfile gets comprehensive token profile with market data, socials, and top pools
func HandleTokenProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address parameter is required")
}

	chain, _ :=getString(args, "chain")

	params := map[string]interface{}{
		"address": address,
	}
	if chain != "" {
		params["chain"] = chain
	}

	result, apiErr := callMeshAPIRaw("TokenResolverAgent", "token_profile", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleGetTrendingTokens gets aggregated trending tokens from multiple sources (GMGN, CoinGecko, Pump.fun, Dexscreener, Zora, Twitter)
func HandleGetTrendingTokens(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ :=getString(args, "source")
	timeframe, _ :=getString(args, "timeframe")
	limit, _ :=getInt(args, "limit")

	params := map[string]interface{}{}
	if source != "" {
		params["source"] = source
	}
	if timeframe != "" {
		params["timeframe"] = timeframe
	}
	if limit > 0 {
		params["limit"] = limit
	}

	result, apiErr := callMeshAPIRaw("TrendingTokenAgent", "get_trending_tokens", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleGetMarketSummary gets AI-generated market summary across all trending sources
func HandleGetMarketSummary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ :=getString(args, "source")
	timeframe, _ :=getString(args, "timeframe")

	params := map[string]interface{}{}
	if source != "" {
		params["source"] = source
	}
	if timeframe != "" {
		params["timeframe"] = timeframe
	}

	result, apiErr := callMeshAPIRaw("TrendingTokenAgent", "get_market_summary", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleTwitterSearch performs smart Twitter search for crypto topics
func HandleTwitterSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	limit, _ :=getInt(args, "limit")

	params := map[string]interface{}{
		"query": query,
	}
	if limit > 0 {
		params["limit"] = limit
	}

	result, apiErr := callMeshAPIRaw("TwitterIntelligenceAgent", "twitter_search", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleUserTimeline gets recent tweets from a Twitter user
func HandleUserTimeline(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username parameter is required")
}

	limit, _ :=getInt(args, "limit")

	params := map[string]interface{}{
		"username": username,
	}
	if limit > 0 {
		params["limit"] = limit
	}

	result, apiErr := callMeshAPIRaw("TwitterIntelligenceAgent", "user_timeline", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleTweetDetail gets detailed info about a specific tweet
func HandleTweetDetail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tweetID, _ :=getString(args, "tweet_id")
	if tweetID == "" {
		return err("tweet_id parameter is required")
}

	params := map[string]interface{}{
		"tweet_id": tweetID,
	}

	result, apiErr := callMeshAPIRaw("TwitterIntelligenceAgent", "tweet_detail", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleExaWebSearch performs web search with AI summarization
func HandleExaWebSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	limit, _ :=getInt(args, "limit")

	params := map[string]interface{}{
		"query": query,
	}
	if limit > 0 {
		params["limit"] = limit
	}

	result, apiErr := callMeshAPIRaw("ExaSearchDigestAgent", "exa_web_search", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleExaScrapeURL scrapes and summarizes webpage content
func HandleExaScrapeURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url parameter is required")
}

	params := map[string]interface{}{
		"url": targetURL,
	}

	result, apiErr := callMeshAPIRaw("ExaSearchDigestAgent", "exa_scrape_url", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleGetAllFundingRates gets funding rates for all Binance perpetual contracts
func HandleGetAllFundingRates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	params := map[string]interface{}{}

	result, apiErr := callMeshAPIRaw("FundingRateAgent", "get_all_funding_rates", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleGetSymbolFundingRates gets funding rates for a specific symbol
func HandleGetSymbolFundingRates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol parameter is required")
}

	params := map[string]interface{}{
		"symbol": symbol,
	}

	result, apiErr := callMeshAPIRaw("FundingRateAgent", "get_symbol_funding_rates", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleGetSymbolOIAndFunding gets open interest and funding for a specific symbol
func HandleGetSymbolOIAndFunding(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol parameter is required")
}

	params := map[string]interface{}{
		"symbol": symbol,
	}

	result, apiErr := callMeshAPIRaw("FundingRateAgent", "get_symbol_oi_and_funding", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleFindSpotFuturesOpportunities finds arbitrage opportunities between spot and futures
func HandleFindSpotFuturesOpportunities(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	params := map[string]interface{}{}

	result, apiErr := callMeshAPIRaw("FundingRateAgent", "find_spot_futures_opportunities", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleSearchProjects searches trending projects with fundamental analysis
func HandleSearchProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	params := map[string]interface{}{
		"query": query,
	}

	result, apiErr := callMeshAPIRaw("AIXBTProjectInfoAgent", "search_projects", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleFetchWalletTokens gets EVM wallet token holdings
func HandleFetchWalletTokens(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	walletAddress, _ :=getString(args, "wallet_address")
	if walletAddress == "" {
		return err("wallet_address parameter is required")
}

	params := map[string]interface{}{
		"wallet_address": walletAddress,
	}

	result, apiErr := callMeshAPIRaw("ZerionWalletAnalysisAgent", "fetch_wallet_tokens", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// HandleFetchWalletNFTs gets EVM wallet NFT holdings
func HandleFetchWalletNFTs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	walletAddress, _ :=getString(args, "wallet_address")
	if walletAddress == "" {
		return err("wallet_address parameter is required")
}

	params := map[string]interface{}{
		"wallet_address": walletAddress,
	}

	result, apiErr := callMeshAPIRaw("ZerionWalletAnalysisAgent", "fetch_wallet_nfts", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// sanitizeName converts a name to a valid identifier (used internally for tool ID generation)
func sanitizeName(name string) string {
	lower := strings.ToLower(name)
	var sb strings.Builder
	for _, c := range lower {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			sb.WriteRune(c)
		} else {
			sb.WriteRune('_')

	}
	parts := strings.Split(sb.String(), "_")
	var filtered []string
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)

	}
	result := strings.Join(filtered, "_")
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = "tool_" + result
	}
	return result
}

}
}

// createToolID creates a tool ID by combining agent ID and tool name
func createToolID(agentID string, toolName string, maxLength int) string {
	agentIDLower := sanitizeName(agentID)
	toolNameSanitized := sanitizeName(toolName)
	separator := "_"

	maxAgentIDLength := maxLength - len(separator) - len(toolNameSanitized)
	if maxAgentIDLength > 0 && len(agentIDLower) > maxAgentIDLength {
		agentIDLower = agentIDLower[:maxAgentIDLength]
	}

	return agentIDLower + separator + toolNameSanitized
}

// unused: suppress import warning for url
var _ = url.Values{}