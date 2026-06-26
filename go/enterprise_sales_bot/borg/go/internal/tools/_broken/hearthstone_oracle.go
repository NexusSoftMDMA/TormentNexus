package mcpimpl

import (
    "context"
    "encoding/json"
    "net/http"
)

func HandleGetCard_hearthstone_oracle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    if name == "" {
        return err("name is required")
}

    resp, e := http.DefaultClient.Get("https://api.hearthstonejson.com/v1/25770/enUS/cards.collectible.json")
    if e != nil {
        return err("failed to fetch cards: " + e.Error())
}

    defer resp.Body.Close()
    var cards []map[string]interface{}
    if e := json.NewDecoder(resp.Body).Decode(&cards); e != nil {
        return err("failed to parse response: " + e.Error())
}

    for _, card := range cards {
        cardName, found := card["name"].(string)
        if found && cardName == name {
            data, e := json.Marshal(card)
            if e != nil {
                return err("failed to marshal card: " + e.Error())
}

            return ok(string(data))

    }
    return err("card not found: " + name)
}
}