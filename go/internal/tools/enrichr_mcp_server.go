//go:build ignore
// +build ignore

package tools

import (
    "context"
    "encoding/json"
    "net/http"
    "net/url"
    "strconv"
)

func HandleEnrichr(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    genes, _ :=getString(args, "genes")
    lib, _ :=getString(args, "library")
    if genes == "" || lib == "" {
        return err("genes and library are required")
}

    resp1, e := http.DefaultClient.PostForm("https://maayanlab.cloud/Enrichr/enrich", url.Values{
        "geneList": {genes},
        "library":  {lib},
    })
    if e != nil {
        return err("submit failed: " + e.Error())
}

    defer resp1.Body.Close()

    var submitResp struct {
        UserListID int `json:"userListId"`,
        if e := json.NewDecoder(resp1.Body).Decode(&submitResp); e != nil {
        return err("decode failed: " + e.Error())
    if submitResp.UserListID == 0 {
        return err("enrichment failed")
}

    resp2, e := http.DefaultClient.Get("https://maayanlab.cloud/Enrichr/enrich?userListId=" + strconv.Itoa(submitResp.UserListID))
    if e != nil {
        return err("fetch failed: " + e.Error())
}

    defer resp2.Body.Close()

    var results interface{    if e := json.NewDecoder(resp2.Body).Decode(&results); e != nil {
        return,
}


-reasoner (deepseek)*,
}
}
}
}