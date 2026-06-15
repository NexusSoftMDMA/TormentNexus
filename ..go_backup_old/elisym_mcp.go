package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
)

func HandleDiscoverAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    limit, _ :=getInt(args, "limit")
    u := fmt.Sprintf("https://api.elisym.com/v1/agents?q=%s&limit=%d", url.QueryEscape(query), limit)
    resp, e := http.DefaultClient.Get(u)
    if e != nil {
        return err("failed to fetch agents")
}

    defer resp.Body.Close()
    body, e := ioutil.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response")
}

    var data map[string]interface{}
    if e := json.Unmarshal(body, &data); e != nil {
        return err("failed to parse response")
}

    _ = data
    return success("agents discovered")
}

func HandleCreateJob(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    agentID, _ :=getString(args, "agentId")
    desc, _ :=getString(args, "description")
    payment, _ :=getInt(args, "payment")
    payload := map[string]interface{}{
        "agentId":     agentID,
        "description": desc,
        "payment":     payment,
    }
    body, e := json.Marshal(payload)
    if e != nil {
        return err("failed to create job payload")
}

    resp, e := http.DefaultClient.Post("https://api.elisym.com/v1/jobs", "application/json", body)
    if e != nil {
        return err("failed to create job")
}

    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return err("failed to create job: " + resp.Status)
}

    return success("job created")
}