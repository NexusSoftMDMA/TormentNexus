package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	http.DefaultClient = http.DefaultClient
	ensRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+\.eth$`)
)

func HandleGetChainInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chainID, _ :=getString(args, "chain_id")
	if chainID == "" {
		chainID = "1" // default to Ethereum mainnet
	}

	apiURL := fmt.Sprintf("https://chainid.network/chains/%s", chainID)
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch chain info: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	return ok(fmt.Sprintf("Chain info: %+v", result))
}

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	chainID, _ :=getString(args, "chain_id")
	if chainID == "" {
		chainID = "1"
	}

	if ensRegex.MatchString(address) {
		// Resolve ENS name to address
		resolved, resolveErr := resolveENS(ctx, address)
		if resolveErr != nil {
			return err(fmt.Sprintf("failed to resolve ENS name: %v", resolveErr))
}

		address = resolved
	}

	apiURL := fmt.Sprintf("https://api.blockcypher.com/v1/eth/main/addrs/%s/balance", address)
	if chainID != "1" {
		apiURL = fmt.Sprintf("https://api.blockcypher.com/v1/eth/%s/addrs/%s/balance", chainID, address)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch balance: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	balance, found := result["balance"].(float64)
	if !found {
		return err("invalid balance format in response")
}

	return ok(fmt.Sprintf("Balance: %.8f ETH", balance/1e18))
}

}

func HandleTransferNative(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	amount, _ :=getString(args, "amount")
	chainID, _ :=getString(args, "chain_id")
	if chainID == "" {
		chainID = "1"
	}

	if ensRegex.MatchString(to) {
		resolved, resolveErr := resolveENS(ctx, to)
		if resolveErr != nil {
			return err(fmt.Sprintf("failed to resolve ENS name: %v", resolveErr))
}

		to = resolved
	}

	// In a real implementation, this would use a wallet private key
	// For this example, we'll just simulate the transfer
	cmd := exec.CommandContext(ctx, "echo", "simulate", "transfer", amount, "to", to, "on chain", chainID)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("failed to simulate transfer: %v", execErr))
}

	return ok(fmt.Sprintf("Transfer simulated: %s", string(output)))
}

func HandleGetTransaction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	txHash, _ :=getString(args, "tx_hash")
	chainID, _ :=getString(args, "chain_id")
	if chainID == "" {
		chainID = "1"
	}

	apiURL := fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=%s", txHash)
	if chainID != "1" {
		apiURL = fmt.Sprintf("https://api-%s.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=%s",
			getChainSubdomain(chainID), txHash)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch transaction: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	return ok(fmt.Sprintf("Transaction details: %+v", result["result"]))
}

}

func HandleGetBlock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	blockRef, _ :=getString(args, "block_ref")
	chainID, _ :=getString(args, "chain_id")
	if chainID == "" {
		chainID = "1"
	}

	apiURL := fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag=%s&boolean=true", blockRef)
	if chainID != "1" {
		apiURL = fmt.Sprintf("https://api-%s.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag=%s&boolean=true",
			getChainSubdomain(chainID), blockRef)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch block: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	return ok(fmt.Sprintf("Block details: %+v", result["result"]))
}

}

func HandleReadContract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	contractAddress, _ :=getString(args, "contract_address")
	functionName, _ :=getString(args, "function_name")
	params, _ :=getString(args, "params")
	chainID, _ :=getString(args, "chain_id")
	if chainID == "" {
		chainID = "1"
	}

	// First get the ABI
	abi, abiErr := fetchContractABI(ctx, contractAddress, chainID)
	if abiErr != nil {
		return err(fmt.Sprintf("failed to fetch contract ABI: %v", abiErr))
}

	// Then call the function
	apiURL := fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_call&to=%s&data=%s",
		contractAddress, encodeFunctionCall(functionName, params, abi))
	if chainID != "1" {
		apiURL = fmt.Sprintf("https://api-%s.etherscan.io/api?module=proxy&action=eth_call&to=%s&data=%s",
			getChainSubdomain(chainID), contractAddress, encodeFunctionCall(functionName, params, abi))

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to call contract function: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	return ok(fmt.Sprintf("Contract function result: %+v", result["result"]))
}

}

func resolveENS(ctx context.Context, ensName string) (string, error) {
	apiURL := fmt.Sprintf("https://api.ensideas.com/ens/resolve/%s", ensName)
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return "", fmt.Errorf("failed to create request: %v", reqErr)
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return "", fmt.Errorf("failed to resolve ENS name: %v", fetchErr)
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read response: %v", readErr)
}

	var result struct {
		Address string `json:"address"`
	}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return "", fmt.Errorf("failed to parse response: %v", parseErr)
}

	if result.Address == "" {
		return "", fmt.Errorf("ENS name not found")
}

	return result.Address, nil
}

func fetchContractABI(ctx context.Context, contractAddress, chainID string) (string, error) {
	apiURL := fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=getabi&address=%s", contractAddress)
	if chainID != "1" {
		apiURL = fmt.Sprintf("https://api-%s.etherscan.io/api?module=contract&action=getabi&address=%s",
			getChainSubdomain(chainID), contractAddress)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return "", fmt.Errorf("failed to create request: %v", reqErr)
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return "", fmt.Errorf("failed to fetch contract ABI: %v", fetchErr)
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read response: %v", readErr)
}

	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  string `json:"result"`
	}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return "", fmt.Errorf("failed to parse response: %v", parseErr)
}

	if result.Status != "1" || result.Message != "OK" {
		return "", fmt.Errorf("failed to get ABI: %s", result.Message)
}

	return result.Result, nil
}

}

func encodeFunctionCall(functionName, params, abiJSON string) string {
	// Simplified function call encoding
	// In a real implementation, this would properly encode based on the ABI
	return fmt.Sprintf("0x%s%s", functionName, params)
}

func getChainSubdomain(chainID string) string {
	switch chainID {
	case "1":
		return "api"
}
	case "5":
		return "api-goerli"
}
	case "11155111":
		return "api-sepolia"
	case "10":
		return "api-optimistic"
	case "420":
		return "api-goerli-optimistic"
	case "42161":
		return "api-arbitrum"
	case "421613":
		return "api-goerli-arbitrum"
	case "8453":
		return "api-base"
	case "84531":
		return "api-goerli-base"
	default:
		return "api"
	}
}