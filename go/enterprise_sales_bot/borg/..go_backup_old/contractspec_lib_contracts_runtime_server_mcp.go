package tools

import (
    "context"
    "io"
    "net/http"
)

func HandleQueryContract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    contractId, _ :=getString(args, "contractId")
    if contractId == "" {
        return err("contractId is required")
}

    resp, e := http.DefaultClient.Get("https://api.contractspec.net/contract/" + contractId)
    if e != nil {
        return err("failed to fetch contract: " + e.Error())
}

    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return err("unexpected status: " + resp.Status)
}

    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read body: " + e.Error())
}

    return success(string(body))
}

func HandleExecuteContract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    contractId, _ :=getString(args, "contractId")
    method, _ :=getString(args, "method")
    if contractId == "" || method == "" {
        return err("contractId and method are required")
}

    resp, e := http.DefaultClient.Post("https://api.contractspec.net/contract/"+contractId+"/method/"+method, "application/json", nil)
    if e != nil {
        return err("execution failed: " + e.Error())
}

    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return err("execution returned: " + resp.Status)
}

    return ok("executed " + method + " on " + contractId)
}