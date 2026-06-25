package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// HTTP client with a 30 second timeout, reused across handlers.
var http.DefaultClient = http.DefaultClient

// Blockscout API base URL. Adjust if needed.
const blockscoutBase = "https://blockscout.com/eth/mainnet/api"

// apiResponse mirrors the generic Blockscout JSON envelope.
type apiResponse struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Result  json.RawMessage
}

// HandleAddressBalance returns the Ether balance of a given address.
// Expected args: {"address": "<hex address>"}
func HandleAddressBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if strings.TrimSpace(address) == "" {
		return err("address argument is required")
}

	// Build request URL with query parameters.
	values := url.Values{}
	values.Set("module", "account")
	values.Set("action", "balance")
	values.Set("address", address)
	values.Set("tag", "latest")
	values.Set("apikey", "") // optional: leave empty for public endpoints

	fullURL := fmt.Sprintf("%s?%s", blockscoutBase, values.Encode())

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var apiResp apiResponse
	if parseErr := json.Unmarshal(body, &apiResp); parseErr != nil {
		return err(parseErr.Error())
}

	if apiResp.Status != "1" {
		return err(fmt.Sprintf("API error: %s", apiResp.Message))
}

	// Result is a string representing wei balance.
	var balanceWeiStr string
	if balErr := json.Unmarshal(apiResp.Result, &balanceWeiStr); balErr != nil {
		return err(balErr.Error())
}

	balanceWei, convErr := strconv.ParseInt(balanceWeiStr, 10, 64)
	if convErr != nil {
		return err(convErr.Error())
}

	// Convert wei to ether (1 ether = 1e18 wei). Use float for readability.
	balanceEther := float64(balanceWei) / 1e18

	return ok(fmt.Sprintf("Address %s balance: %.6f ETH", address, balanceEther))
}

// HandleTxStatus returns the status of a transaction given its hash.
// Expected args: {"txhash": "<tx hash>"}
func HandleTxStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	txHash, _ :=getString(args, "txhash")
	if strings.TrimSpace(txHash) == "" {
		return err("txhash argument is required")
}

	values := url.Values{}
	values.Set("module", "transaction")
	values.Set("action", "getstatus")
	values.Set("txhash", txHash)

	fullURL := fmt.Sprintf("%s?%s", blockscoutBase, values.Encode())

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var apiResp apiResponse
	if parseErr := json.Unmarshal(body, &apiResp); parseErr != nil {
		return err(parseErr.Error())
}

	if apiResp.Status != "1" {
		return err(fmt.Sprintf("API error: %s", apiResp.Message))
}

	// Result contains a JSON object with "isError" field.
	type statusResult struct {
		IsError string `json:"isError"`
	}
	var status statusResult
	if statusErr := json.Unmarshal(apiResp.Result, &status); statusErr != nil {
		return err(statusErr.Error())
}

	statusText := "Success"
	if status.IsError != "0" {
		statusText = "Failed"
	}

	return ok(fmt.Sprintf("Transaction %s status: %s", txHash, statusText))
}

// HandleLatestBlock returns the latest block number on the chain.
// No arguments required.
func HandleLatestBlock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	values := url.Values{}
	values.Set("module", "proxy")
	values.Set("action", "eth_blockNumber")

	fullURL := fmt.Sprintf("%s?%s", blockscoutBase, values.Encode())

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var apiResp apiResponse
	if parseErr := json.Unmarshal(body, &apiResp); parseErr != nil {
		return err(parseErr.Error())
}

	if apiResp.Status != "1" {
		return err(fmt.Sprintf("API error: %s", apiResp.Message))
}

	// Result is a hex string like "0x10d4f".
	var blockHex string
	if blockErr := json.Unmarshal(apiResp.Result, &blockHex); blockErr != nil {
		return err(blockErr.Error())
}

	blockNum, convErr := strconv.ParseInt(strings.TrimPrefix(blockHex, "0x"), 16, 64)
	if convErr != nil {
		return err(convErr.Error())
}

	return ok(fmt.Sprintf("Latest block number: %d", blockNum))
}

// HandleSearch performs a simple search on Blockscout for an address or transaction hash.
// Expected args: {"query": "<address or txhash>"}
func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if strings.TrimSpace(query) == "" {
		return err("query argument is required")
}

	// Blockscout provides a generic search endpoint via its UI; we mimic it by checking format.
	// If query looks like a tx hash (64 hex chars), treat as transaction; otherwise as address.
	isTx := false
	if len(query) == 66 && strings.HasPrefix(query, "0x") {
		isTx = true
	} else if len(query) == 64 && regexp.MustCompile(`^[0-9a-fA-F]+$`).MatchString(query) {
		isTx = true
		query = "0x" + query
	}

	if isTx {
		// Reuse transaction status handler logic.
		return HandleTxStatus(ctx, map[string]interface{}{"txhash": query})
}

	// Otherwise treat as address and reuse balance handler.
	return HandleAddressBalance(ctx, map[string]interface{}{"address": query})
}