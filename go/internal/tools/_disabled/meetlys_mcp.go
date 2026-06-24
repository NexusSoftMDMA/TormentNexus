package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func HandleMeetlys(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    meetingID, _ :=getString(args, "meeting_id")
    url := "https://api.meetlys.com/meetings"
    if meetingID != "" {
        url = fmt.Sprintf("%s/%s", url, meetingID)

    resp, e := http.DefaultClient.Get(url)
    if e != nil {
        return err("failed to fetch meetings: " + e.Error())
}

    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return err("unexpected status: " + resp.Status)
}

    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response: " + e.Error())
}

    var result interface{}
    if e = json.Unmarshal(body, &result); e != nil {
        return err("failed to parse JSON: " + e.Error())
}

    return ok(fmt.Sprintf("%v", result))
}
}