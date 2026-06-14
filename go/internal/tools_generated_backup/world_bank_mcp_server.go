package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

func HandleGetCountry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    country, _ :=getString(args, "country")
    if country == "" {
        return err("country is required")
}

    url := fmt.Sprintf("https://api.worldbank.org/v2/country/%s?format=json", country)
    resp, e := http.DefaultClient.Get(url)
    if e != nil {
        return err("failed to fetch: " + e.Error())
}

    defer resp.Body.Close()
    var data []json.RawMessage
    if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
        return err("parse error: " + e.Error())
}

    if len(data) < 2 {
        return err("no data found")
}

    var countries []map[string]interface{}
    if e := json.Unmarshal(data[1], &countries); e != nil {
        return err("parse error: " + e.Error())
}

    if len(countries) == 0 {
        return err("country not found")
}

    c := countries[0]
    name, _ := c["name"].(string)
    region, found := c["region"].(map[string]interface{})
    regionName := ""
    if found {
        regionName, _ = region["value"].(string)

    return ok(fmt.Sprintf("Country: %s, Region: %s", name, regionName))
}
}