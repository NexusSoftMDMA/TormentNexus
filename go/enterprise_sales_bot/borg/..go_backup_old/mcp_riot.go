package tools

import (
    "context"
    "net/http"
)

func HandleGetSummoner(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    region, _ :=getString(args, "region")
    if region == "" {
        region = "na1"
    }
    url := "https://" + region + ".api.riotgames.com/lol/summoner/v4/summoners/by-name/" + name
    resp, e := http.DefaultClient.Get(url)
    if e != nil {
        return err("failed to call Riot API: " + e.Error())
}

    resp.Body.Close()
    return ok("Summoner " + name + " found")
}