package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ToolResponse はツールからのレスポンスを表します。

// ok は成功レスポンスを返します。
func ok(text string) (ToolResponse, error) {
	return ToolResponse{TextContent: text}, nil
}

// e はエラーレスポンスを返します。
func err(e error) (ToolResponse, error) {
	return ToolResponse{Error: e}, e
}

// getString はマップから文字列を取得します。
func getString(args map[string]interface{}, key string) (string, error) {
	val, found := args[key]
	if !found {
		return "", fmt.Errorf("missing key: %s", key)
}

	strVal, found := val.(string)
	if !found {
		return "", fmt.Errorf("key %s is not a string", key)
}

	return strVal, nil
}

// getInt はマップから整数を取得します。
func getInt(args map[string]interface{}, key string) (int, error) {
	strVal, _ :=getString(args, key)
	if e != nil {
		return 0, e
	}
	intVal, e := strconv.Atoi(strVal)
	if e != nil {
		return 0, fmt.Errorf("key %s is not a valid integer: %v", key, e)
}

	return intVal, nil
}

// getBool はマップからブーリアンを取得します。
func getBool(args map[string]interface{}, key string) (bool, error) {
	strVal, _ :=getString(args, key)
	if e != nil {
		return false, e
	}
	boolVal, e := strconv.ParseBool(strVal)
	if e != nil {
		return false, fmt.Errorf("key %s is not a valid boolean: %v", key, e)
}

	return boolVal, nil
}

// HandleListItems はアイテムの一覧を取得します。
func HandleListItems(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if e != nil {
		return err(e)
}

	client := http.DefaultClient
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.context.space/v1/items", nil)
	if e != nil {
		return err(e)
}

	q := req.URL.Query()
	q.Add("limit", strconv.Itoa(limit))
	req.URL.RawQuery = q.Encode()

	resp, e := client.Do(req)
	if e != nil {
		return err(e)
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
}

	var items []Item
	if e := json.NewDecoder(resp.Body).Decode(&items); e != nil {
		return err(e)
}

	return ok(fmt.Sprintf("Listed %d items", len(items)))
}

// HandleGetItem は特定のアイテムを取得します。
func HandleGetItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	itemID, _ :=getString(args, "item_id")
	if e != nil {
		return err(e)
}

	client := http.DefaultClient
	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.context.space/v1/items/%s", itemID), nil)
	if e != nil {
		return err(e)
}

	resp, e := client.Do(req)
	if e != nil {
		return err(e)
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
}

	var item Item
	if e := json.NewDecoder(resp.Body).Decode(&item); e != nil {
		return err(e)
}

	return ok(fmt.Sprintf("Fetched item: %s", item.Name))
}

// HandleCreateItem は新しいアイテムを作成します。
func HandleCreateItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	itemData, _ :=getString(args, "item_data")
	if e != nil {
		return err(e)
}

	client := http.DefaultClient
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.context.space/v1/items", strings.NewReader(itemData))
	if e != nil {
		return err(e)
}

	req.Header.Set("Content-Type", "application/json")

	resp, e := client.Do(req)
	if e != nil {
		return err(e)
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return err(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
}

	var createdItem Item
	if e := json.NewDecoder(resp.Body).Decode(&createdItem); e != nil {
		return err(e)
}

	return ok(fmt.Sprintf("Created item: %s", createdItem.Name))
}

// HandleUpdateItem は既存のアイテムを更新します。
func HandleUpdateItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	itemID, _ :=getString(args, "item_id")
	if e != nil {
		return err(e)
}

	itemData, _ :=getString(args, "item_data")
	if e != nil {
		return err(e)
}

	client := http.DefaultClient
	req, e := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("https://api.context.space/v1/items/%s", itemID), strings.NewReader(itemData))
	if e != nil {
		return err(e)
}

	req.Header.Set("Content-Type", "application/json")

	resp, e := client.Do(req)
	if e != nil {
		return err(e)
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
}

	return ok("Item updated successfully")
}

// HandleDeleteItem はアイテムを削除します。
func HandleDeleteItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	itemID, _ :=getString(args, "item_id")
	if e != nil {
		return err(e)
}

	client := http.DefaultClient
	req, e := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("https://api.context.space/v1/items/%s", itemID), nil)
	if e != nil {
		return err(e)
}

	resp, e := client.Do(req)
	if e != nil {
		return err(e)
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
}

	return ok("Item deleted successfully")
}