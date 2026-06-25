package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "time"
)

// We assume the following structs for the CDP API responses.

type CreateWalletResponse struct {
    Data struct {
        ID string `json:"id"`
    } `json:"data"`
}

type TransferResponse struct {
    Data struct {
        ID string `json:"id"`
    } `json:"data"`
}

// Helper function to make HTTP requests with the CDP API key and secret.
func makeCDPRequest(ctx context.Context, method, url string, body interface{}, apiKeyName, privateKey string) ([]byte, error) {
    var reqBody io.Reader
    if body != nil {
        jsonBody, e := json.Marshal(body)
        if e != nil {
            return nil, e
        }
        reqBody = strings.NewReader(string(jsonBody))

    req, e := http.NewRequestWithContext(ctx, method, url, reqBody)
    if e != nil {
        return nil, e
    }

    // Set headers for CDP API. We assume the API expects these headers.
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-KEY", apiKeyName)
    req.Header.Set("X-API-SECRET", privateKey)

    client := http.DefaultClient
    resp, e := client.Do(req)
    if e != nil {
        return nil, e
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
}

    return io.ReadAll(resp.Body)
}

}

// HandleCreateCoinbaseMpcWallet creates a new MPC wallet and saves the wallet ID to a file.
func HandleCreateCoinbaseMpcWallet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    // Extract required arguments.
    apiKeyName, _ :=getString(args, "apiKeyName")
    if e != nil {
        return err(e.Error())
}

    privateKey, _ :=getString(args, "privateKey")
    if e != nil {
        return err(e.Error())
}

    walletFilePath, _ :=getString(args, "walletFilePath")
    if e != nil {
        return err(e.Error())
}

    // Create the wallet via CDP API.
    url := "https://api.cdp.coinbase.com/platform/v1/wallets"
    body := map[string]interface{}{
        "wallet_type": "mpc",
    }
    respBytes, e := makeCDPRequest(ctx, "POST", url, body, apiKeyName, privateKey)
    if e != nil {
        return err(e.Error())
}

    var resp CreateWalletResponse
    if e := json.Unmarshal(respBytes, &resp); e != nil {
        return err(e.Error())
}

    walletID := resp.Data.ID
    if walletID == "" {
        return err("wallet ID not found in response")
}

    // Ensure the directory for the wallet file exists.
    dir := filepath.Dir(walletFilePath)
    if e := os.MkdirAll(dir, 0755); e != nil {
        return err(e.Error())
}

    // Write the wallet ID to the file.
    if e := os.WriteFile(walletFilePath, []byte(walletID), 0644); e != nil {
        return err(e.Error())
}

    return ok(fmt.Sprintf("Wallet created with ID: %s", walletID))
}

// HandleTransferUsdc transfers USDC to an address (ENS or Ethereum address).
func HandleTransferUsdc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    // Extract required arguments.
    apiKeyName, _ :=getString(args, "apiKeyName")
    if e != nil {
        return err(e.Error())
}

    privateKey, _ :=getString(args, "privateKey")
    if e != nil {
        return err(e.Error())
}

    walletFilePath, _ :=getString(args, "walletFilePath")
    if e != nil {
        return err(e.Error())
}

    toAddress, _ :=getString(args, "toAddress")
    if e != nil {
        return err(e.Error())
}

    amount, _ :=getString(args, "amount")
    if e != nil {
        return err(e.Error())
}

    // Read the wallet ID from the file.
    walletIDBytes, e := os.ReadFile(walletFilePath)
    if e != nil {
        return err(e.Error())
}

    walletID := strings.TrimSpace(string(walletIDBytes))
    if walletID == "" {
        return err("wallet ID file is empty")
}

    // Resolve ENS if necessary.
    resolvedAddress := toAddress
    if strings.HasSuffix(toAddress, ".eth") {
        // Use a public ENS resolver.
        ensURL := fmt.Sprintf("https://api.ensideas.com/ens/resolve/%s", toAddress)
        resp, e := http.Get(ensURL)
        if e != nil {
            return err(e.Error())
}

        defer resp.Body.Close()
        if resp.StatusCode != http.StatusOK {
            return err(fmt.Sprintf("ENS resolution failed with status %d", resp.StatusCode))
}

        var ensResp struct {
            Address string `json:"address"`
        }
        if e := json.NewDecoder(resp.Body).Decode(&ensResp); e != nil {
            return err(e.Error())
}

        if ensResp.Address == "" {
            return err("ENS resolution returned empty address")
}

        resolvedAddress = ensResp.Address
    }

    // Validate the resolved address is a valid Ethereum address.
    matched, _ := regexp.MatchString(`^0x[a-fA-F0-9]{40}$`, resolvedAddress)
    if !matched {
        return err("invalid Ethereum address")
}

    // Convert amount to string (assuming it's already a string representation of the amount in USDC).
    // We'll pass the amount as a string to the API.

    // Prepare the transfer request.
    url := fmt.Sprintf("https://api.cdp.coinbase.com/platform/v1/wallets/%s/actions/transfer", walletID)
    body := map[string]interface{}{
        "amount":   amount,
        "currency": "USDC",
        "to":       resolvedAddress,
    }

    respBytes, e := makeCDPRequest(ctx, "POST", url, body, apiKeyName, privateKey)
    if e != nil {
        return err(e.Error())
}

    var resp TransferResponse
    if e := json.Unmarshal(respBytes, &resp); e != nil {
        return err(e.Error())
}

    transferID := resp.Data.ID
    if transferID == "" {
        return err("transfer ID not found in response")
}

    return ok(fmt.Sprintf("Transfer initiated with ID: %s", transferID))
}