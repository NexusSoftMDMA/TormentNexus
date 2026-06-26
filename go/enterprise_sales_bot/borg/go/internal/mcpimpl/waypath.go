package mcpimpl

import (
    "context"
    "fmt"
)

func HandleWaypathList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    limit, _ :=getInt(args, "limit")
    if limit == 0 {
        limit = 10
    }
    data := `[{"id":1,"name":"Home","lat":40.7128,"lng":-74.0060},{"id":2,"name":"Work","lat":40.7580,"lng":-73.9855}]`
    return success(data)
}

func HandleWaypathGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    id, _ :=getInt(args, "id")
    if id == 0 {
        return err("id parameter required")
}

    data := fmt.Sprintf(`{"id":%d,"name":"Point","lat":0,"lng":0}`, id)
    return success(data)
}