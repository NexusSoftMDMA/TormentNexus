package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	baseURL = "https://api.dexpaprika.com/v1"
)

func HandleGetNetworks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiURL := fmt.Sprintf("%s/networks", baseURL)

	client := http.DefaultClient
	resp, reqErr := client.Get(apiURL)
	if reqErr != nil {
		return err(reqErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	var networks []map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&networks)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(fmt.Sprintf("%v", networks))
}

func HandleGetNetworkDexes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ :=getString(args, "network")
	if network == "" {
		return err("network parameter is required")
}

	apiURL := fmt.Sprintf("%s/networks/%s/dexes", baseURL, network)

	client := http.DefaultClient
	resp, reqErr := client.Get(apiURL)
	if reqErr != nil {
		return err(reqErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	var dexes []map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&dexes)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(fmt.Sprintf("%v", dexes))
}

func HandleGetPoolDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ :=getString(args, "network")
	poolAddress, _ :=getString(args, "pool_address")
	if network == "" || poolAddress == "" {
		return err("network and pool_address parameters are required")
}

	apiURL := fmt.Sprintf("%s/networks/%s/pools/%s", baseURL, network, poolAddress)

	client := http.DefaultClient
	resp, reqErr := client.Get(apiURL)
	if reqErr != nil {
		return err(reqErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	var poolDetails map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&poolDetails)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(fmt.Sprintf("%v", poolDetails))
}

func HandleGetTokenDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ :=getString(args, "network")
	tokenAddress, _ :=getString(args, "token_address")
	if network == "" || tokenAddress == "" {
		return err("network and token_address parameters are required")
}

	apiURL := fmt.Sprintf("%s/networks/%s/tokens/%s", baseURL, network, tokenAddress)

	client := http.DefaultClient
	resp, reqErr := client.Get(apiURL)
	if reqErr != nil {
		return err(reqErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	var tokenDetails map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&tokenDetails)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(fmt.Sprintf("%v", tokenDetails))
}

func HandleGetTokenPools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ :=getString(args, "network")
	tokenAddress, _ :=getString(args, "token_address")
	if network == "" || tokenAddress == "" {
		return err("network and token_address parameters are required")
}

	apiURL := fmt.Sprintf("%s/networks/%s/tokens/%s/pools", baseURL, network, tokenAddress)

	client := http.DefaultClient
	resp, reqErr := client.Get(apiURL)
	if reqErr != nil {
		return err(reqErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	var tokenPools []map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&tokenPools)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(fmt.Sprintf("%v", tokenPools))
}

func HandleGetTokenMultiPrices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ :=getString(args, "network")
	tokens, found := args["tokens"].([]interface{})
	if !ok || network == "" || len(tokens) == 0 {
		return err("network and tokens parameters are required")
}

	// Convert tokens to comma-separated string
	tokenList := make([]string, len(tokens))
	for i, token := range tokens {
		tokenList[i] = fmt.Sprintf("%v", token)

	tokensStr := strings.Join(tokenList, ",")

	apiURL := fmt.Sprintf("%s/networks/%s/tokens/prices?tokens=%s", baseURL, network, tokensStr)

	client := http.DefaultClient
	resp, reqErr := client.Get(apiURL)
	if reqErr != nil {
		return err(reqErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	var prices map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&prices)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(fmt.Sprintf("%v", prices))
}
}