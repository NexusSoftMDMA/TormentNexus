package tools

import (
    "context"
    "io/ioutil"
    "net/http"
    "net/url"
)

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    token, _ :=getString(args, "token")
    chatID, _ :=getString(args, "chat_id")
    text, _ :=getString(args, "text")
    if token == "" || chatID == "" || text == "" {
        return err("missing required parameters: token, chat_id, text")
}

    apiURL := "https://api.telegram.org/bot" + token + "/sendMessage"
    data := url.Values{}
    data.Set("chat_id", chatID)
    data.Set("text", text)
    resp, e := http.DefaultClient.PostForm(apiURL, data)
    if e != nil {
        return err("failed to send message: " + e.Error())
}

    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body)
        return err("telegram api error: " + string(body))
}

    return ok("message sent successfully")
}