package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListDatabases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = os.Getenv("YUGABYTE_HOST")
		if host == "" {
			host = "localhost:13000"
		}
	}
	url := fmt.Sprintf("http://%s/databases", host)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to connect: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("response not JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Databases: %v", data))
}