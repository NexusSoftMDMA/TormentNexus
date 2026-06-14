package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func HandleListClusters(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("RANCHER_URL")
	if base == "" {
		return err("RANCHER_URL not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/v3/clusters", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.SetBasicAuth(os.Getenv("RANCHER_ACCESS_KEY"), os.Getenv("RANCHER_SECRET_KEY"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json decode failed: %v", e))
}

	data, found := result["data"]
	if !found {
		return err("no data in response")
}

	return ok(fmt.Sprintf("Clusters: %v", data))
}

func HandleGetCluster(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "clusterId")
	if id == "" {
		return err("clusterId required")
}

	base := os.Getenv("RANCHER_URL")
	if base == "" {
		return err("RANCHER_URL not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/v3/clusters/"+id, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.SetBasicAuth(os.Getenv("RANCHER_ACCESS_KEY"), os.Getenv("RANCHER_SECRET_KEY"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	return ok(fmt.Sprintf("Cluster info: %s", string(body)))
}