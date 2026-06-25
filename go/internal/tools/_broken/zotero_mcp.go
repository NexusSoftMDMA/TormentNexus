package tools.

Then, first, maybe a helper function to make HTTP requests to Zotero's local API? Wait but we can write it inline, but maybe a helper is okay, but let's make sure it's simple.

Wait no, let's write each handler step by step.

First, HandleSearchLibrary:

func HandleSearchLibrary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query parameter is required")
}

    limit, _ :=getInt(args, "limit")
    if limit == 0 {
        limit = 10 // default limit
    }
    // Build Zotero API URL
    apiURL := "http://localhost:23119/api/users/0/items"
    u, parseErr := url.Parse(apiURL)
    if parseErr != nil {
        return err(fmt.Sprintf("failed to parse API URL: %v", parseErr))
}

    q := u.Query()
    q.Set("q", query)
    q.Set("limit", strconv.Itoa(limit))
    u.RawQuery = q.Encode()
    // Make HTTP request
    client := http.DefaultClient
    req, reqErr := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
    if reqErr != nil {
        return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

    resp, fetchErr := client.Do(req)
    if fetchErr != nil {
        return err(fmt.Sprintf("failed to fetch from Zotero API: %v", fetchErr))
}

    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return err(fmt.Sprintf("Zotero API returned status %d", resp.StatusCode))
}

    // Parse response
    var items []map[string]interface{}
    if parseErr := json.NewDecoder(resp.Body).Decode(&items); parseErr != nil {
        return err(fmt.Sprintf("failed to parse API response: %v", parseErr))
}

    // Format results
    var result strings.Builder
    result.WriteString(fmt.Sprintf("Found %d items for query '%s':\n", len(items), query))
    for i, item := range items {
        title, _ :=getString(item, "title")
        creators, _ :=getString(item, "creators")
        year, _ :=getString(item, "year")
        itemType, _ :=getString(item, "itemType")
        key, _ :=getString(item, "key")
        result.WriteString(fmt.Sprintf("%d. [%s] %s (%s) - %s (Key: %s)\n", i+1, itemType, title, year, creators, key))

    return ok(result.String())
}

Wait that's good. No e variable, uses parseErr, reqErr, fetchErr. Imports: context, encoding/json, fmt, net/http, net/url, strconv, strings, time. All stdlib.

Next, HandleGetItemContent:

}

func HandleGetItemContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    itemKey, _ :=getString(args, "item_key")
    if itemKey == "" {
        return err("item_key parameter is required")
}

    mode, _ :=getString(args, "mode")
    if mode == "" {
        mode = "standard" // default mode
    }
    // Build API URL with expand for notes/attachments
    apiURL := fmt.Sprintf("http://localhost:23119/api/users/0/items/%s?expand=notes,attachments", itemKey)
    client := http.DefaultClient
    req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
    if reqErr != nil {
        return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

    resp, fetchErr := client.Do(req)
    if fetchErr != nil {
        return err(fmt.Sprintf("failed to fetch item content: %v", fetchErr))
}

    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return err(fmt.Sprintf("Zotero API returned status %d for item %s", resp.StatusCode, itemKey))
}

    var item map[string]interface{}
    if parseErr := json.NewDecoder(resp.Body).Decode(&item); parseErr != nil {
        return err(fmt.Sprintf("failed to parse item data: %v", parseErr))
}

    // Format based on mode
    var result strings.Builder
    title, _ :=getString(item, "title")
    result.WriteString(fmt.Sprintf("Item: %s (Key: %s)\n", title, itemKey))
    result.WriteString(fmt.Sprintf("Mode: %s\n", mode))
    // Add notes if present
    if notes, found := item["notes"].([]interface{}); ok && len(notes) > 0 {
        result.WriteString("\nNotes:\n")
        for _, n := range notes {
            note := n.(map[string]interface{})
            noteTitle, _ :=getString(note, "title")
            noteContent, _ :=getString(note, "note")
            result.WriteString(fmt.Sprintf("- %s: %s\n", noteTitle, noteContent))

    }
    // Add attachments if present
    if attachments, found := item["attachments"].([]interface{}); ok && len(attachments) > 0 {
        result.WriteString("\nAttachments:\n")
        for _, a := range attachments {
            att := a.(map[string]interface{})
            attTitle, _ :=getString(att, "title")
            attKey, _ :=getString(att, "key")
            result.WriteString(fmt.Sprintf("- %s (Key: %s)\n", attTitle, attKey))

    }
    return ok(result.String())
}

Good, that's the second handler.

Third, HandleGenerateClientConfig: this is super simple, just returns the config JSON for the client, based on the port.

}
}

func HandleGenerateClientConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    port, _ :=getInt(args, "port")
    if port == 0 {
        port = 23120 // default port
    }
    config := map[string]interface{}{
        "mcpServers": map[string]interface{}{
            "zotero": map[string]interface{}{
                "transport": "streamable_http",
                "url":       fmt.Sprintf("http://127.0.0.1:%d/mcp", port),
            },
        },
    }
    configJSON, marshalErr := json.MarshalIndent(config, "", "  ")
    if marshalErr != nil {
        return err(fmt.Sprintf("failed to generate config: %v", marshalErr))
}

    return ok(string(configJSON))
}

Perfect, that's third.

Fourth, HandleCheckPluginUpdates: checks GitHub's latest release for the plugin, compares version.

func HandleCheckPluginUpdates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    currentVersion, _ :=getString(args, "current_version")
    if currentVersion == "" {
        currentVersion = "1.5.0" // default current version as per README
    }
    updateURL := "https://github.com/cookjohn/zotero-mcp/releases/latest/download/update.json"
    client := http.DefaultClient
    req, reqErr := http.NewRequestWithContext(ctx, "GET", updateURL, nil)
    if reqErr != nil {
        return err(fmt.Sprintf("failed to create update check request: %v", reqErr))
}

    resp, fetchErr := client.Do(req)
    if fetchErr != nil {
        return err(fmt.Sprintf("failed to check for updates: %v", fetchErr))
}

    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return err(fmt.Sprintf("update check failed with status %d", resp.StatusCode))
}

    var updateManifest map[string]interface{}
    if parseErr := json.NewDecoder(resp.Body).Decode(&updateManifest); parseErr != nil {
        return err(fmt.Sprintf("failed to parse update manifest: %v", parseErr))
}

    // Extract latest version from manifest
    addons, found := updateManifest["addons"].(map[string]interface{})
    if !found {
        return err("invalid update manifest format")
}

    pluginInfo, found := addons["zotero-mcp-plugin@autoagent.my"].(map[string]interface{})
    if !found {
        return err("plugin info not found in update manifest")
}

    updates, found := pluginInfo["updates"].([]interface{})
    if !ok || len(updates) == 0 {
        return err("no updates found in manifest")
}

    latestUpdate := updates[0].(map[string]interface{})
    latestVersion, _ :=getString(latestUpdate, "version")
    updateLink, _ :=getString(latestUpdate, "update_link")
    // Compare versions (simple semver compare)
    if latestVersion > currentVersion {
        return ok(fmt.Sprintf("Update available! Current version: %s, Latest version: %s\nDownload link: %s", currentVersion, latestVersion, updateLink))
    }
    return ok(fmt.Sprintf("You are using the latest version: %s", currentVersion))
}

Great, that's the fourth handler.

Fifth, HandleBrowseCollections: browses Zotero collections.

func HandleBrowseCollections(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    parentID, _ :=getString(args, "parent_id")
    apiURL := "http://localhost:23119/api/users/0/collections"
    if parentID != "" {
        apiURL = fmt.Sprintf("http://localhost:23119/api/users/0/collections/%s/collections", parentID)

    client := http.DefaultClient
    req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
    if reqErr != nil {
        return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

    resp, fetchErr := client.Do(req)
    if fetchErr != nil {
        return err(fmt.Sprintf("failed to fetch collections: %v", fetchErr))
}

    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return err(fmt.Sprintf("Zotero API returned status %d for collections", resp.StatusCode))
}

    var collections []map[string]interface{}
    if parseErr := json.NewDecoder(resp.Body).Decode(&collections); parseErr != nil {
        return err(fmt.Sprintf("failed to parse collections data: %v", parseErr))
}

    var result strings.Builder
    if parentID == "" {
        result.WriteString("Top-level Collections:\n")
    } else {
        result.WriteString(fmt.Sprintf("Sub-collections of %s:\n", parentID))

    for i, col := range collections {
        colName, _ :=getString(col, "name")
        colKey, _ :=getString(col, "key")
        numItems, _ :=getInt(col, "numItems")
        result.WriteString(fmt.Sprintf("%d. %s (Key: %s, Items: %d)\n", i+1, colName, colKey, numItems))

    if len(collections) == 0 {
        result.WriteString("No collections found.\n")

    return ok(result.String())
}
}
}
}
}