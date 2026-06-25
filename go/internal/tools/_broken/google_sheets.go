package tools with two handlers: HandleReadSheet and HandleWriteSheet. Also must follow the rules: no redeclaration of types/functions from parity.go, only stdlib imports, etc.

We'll write the full code with proper variable names, complete all strings, and ensure no other issues.

Also, note: the original code had a missing closing quote in HandleWriteSheet after `fmt.Sprintf("`. In our fixed version, we'll complete that error message.")

Let's plan:

- Rename `range` to `sheetRange` in both handlers.
- Use `getString` for accessToken, spreadsheetId, sheetRange.
- In HandleWriteSheet, check values, convert to [][]interface{}.
- Build URL, make request, handle response.

We'll also handle the case where reading the response body in error status.

Make sure all imports are used. We'll import: context, encoding/json, fmt, io, net/http, net/url, strings, time.

We'll not use any other imports like os/exec etc. So it's fine.

Now produce the code.

Important: The original code had a comment about manifest but we don't need to output that. Only Go code.

Thus, output:

package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "time"
)

func HandleReadSheet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    spreadsheetId, _ :=getString(args, "spreadsheetId")
    sheetRange, _ :=getString(args, "range")
    accessToken, _ :=getString(args, "accessToken")

    if spreadsheetId == "" || sheetRange == "" || accessToken == "" {
        return err("missing required parameters: spreadsheetId, range, accessToken")
}

    u := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s", url.PathEscape(spreadsheetId), url.PathEscape(sheetRange))

    req, reqErr := http.NewRequestWithContext(ctx, "GET", u, nil)
    if reqErr != nil {
        return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

    req.Header.Set("Authorization", "Bearer "+accessToken)

    client := http.DefaultClient
    resp, fetchErr := client.Do(req)
    if fetchErr != nil {
        return err(fmt.Sprintf("failed to fetch: %v", fetchErr))
}

    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, readErr := io.ReadAll(resp.Body)
        if readErr != nil {
            return err(fmt.Sprintf("API error: status %d, and failed to read body: %v", resp.StatusCode, readErr))
}

        return err(fmt.Sprintf("API error: status %d, body: %s", resp.StatusCode, string(body)))
}

    var result struct {
        Values [][]interface{} `json:"values"`
    }
    if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
        return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

    data, marshalErr := json.Marshal(result.Values)
    if marshalErr != nil {
        return err(fmt.Sprintf("failed to marshal values: %v", marshalErr))
}

    return ok(string(data))
}

func HandleWriteSheet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    spreadsheetId, _ :=getString(args, "spreadsheetId")
    sheetRange, _ :=getString(args, "range")
    accessToken, _ :=getString(args, "accessToken")
    valuesInterface, exists := args["values"]
    if !exists {
        return err("missing required parameter: values")
}

    valuesSlice, found := valuesInterface.([]interface{})
    if !found {
        return err("values must be an array of arrays")
}

    var rows [][]interface{}
    for i, row := range valuesSlice {
        r, found := row.([]interface{})
        if !found {
            return err(fmt.Sprintf("row at index %d must be an array", i))
}

        rows = append(rows, r)

    u := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s:append", url.PathEscape(spreadsheetId), url.PathEscape(sheetRange))

    requestBody := map[string]interface{}{
        "range":         sheetRange,
        "majorDimension": "ROWS",
        "values":        rows,
    }
    bodyBytes, marshalErr := json.Marshal(requestBody)
    if marshalErr != nil {
        return err(fmt.Sprintf("failed to marshal request body: %v", marshalErr))
}

    req, reqErr := http.NewRequestWithContext(ctx, "POST", u, strings.NewReader(string(bodyBytes)))
    if reqErr != nil {
        return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

    req.Header.Set("Authorization", "Bearer "+accessToken)
    req.Header.Set("Content-Type", "application/json")

    client := http.DefaultClient
    resp, fetchErr := client.Do(req)
    if fetchErr != nil {
        return err(fmt.Sprintf("failed to fetch: %v", fetchErr))
}

    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, readErr := io.ReadAll(resp.Body)
        if readErr != nil {
            return err(fmt.Sprintf("API error: status %d, and failed to read body: %v", resp.StatusCode, readErr))
}

        return err(fmt.Sprintf("API error: status %d, body: %s", resp.StatusCode, string(body)))
}

    var result struct {
        UpdatedRange string `json:"updatedRange"`
    }
    if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
        return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

    return ok(fmt.Sprintf("successfully updated range: %s", result.UpdatedRange))
}

That's it. Now output this as final.package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "time"
)

}
