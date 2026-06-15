package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleKustoQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cluster, _ :=getString(args, "cluster")
	database, _ :=getString(args, "database")
	query, _ :=getString(args, "query")
	if cluster == "" || database == "" || query == "" {
		return err("missing cluster, database, or query")
}

	url := fmt.Sprintf("https://%s.kusto.windows.net/%s/rest/v2/query", cluster, database)
	body := map[string]string{"db": database, "csl": query}
	payload, e := json.Marshal(body)
	if e != nil {
		return err(e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(payload)))
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+getString(args, "token"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(data)))
}

	return success(string(data))
}

func HandleKustoSchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cluster, _ :=getString(args, "cluster")
	database, _ :=getString(args, "database")
	if cluster == "" || database == "" {
		return err("missing cluster or database")
}

	query := ".show database schema"
	url := fmt.Sprintf("https://%s.kusto.windows.net/%s/rest/v2/query", cluster, database)
	body := map[string]string{"db": database, "csl": query}
	payload, e := json.Marshal(body)
	if e != nil {
		return err(e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(payload)))
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+getString(args, "token"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(data)))
}

	return success(string(data))
}