package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListClusters_mongodb_atlas_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pubKey, _ :=getString(args, "publicKey")
	privKey, _ :=getString(args, "privateKey")
	groupID, _ :=getString(args, "groupId")
	if pubKey == "" || privKey == "" || groupID == "" {
		return err("publicKey, privateKey, and groupId are required")
}

	url := fmt.Sprintf("https://cloud.mongodb.com/api/atlas/v1.0/groups/%s/clusters", groupID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.SetBasicAuth(pubKey, privKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	return success(string(body))
}

func HandleGetCluster_mongodb_atlas_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pubKey, _ :=getString(args, "publicKey")
	privKey, _ :=getString(args, "privateKey")
	groupID, _ :=getString(args, "groupId")
	clusterName, _ :=getString(args, "clusterName")
	if pubKey == "" || privKey == "" || groupID == "" || clusterName == "" {
		return err("publicKey, privateKey, groupId, and clusterName are required")
}

	url := fmt.Sprintf("https://cloud.mongodb.com/api/atlas/v1.0/groups/%s/clusters/%s", groupID, clusterName)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.SetBasicAuth(pubKey, privKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return success(string(body))
}