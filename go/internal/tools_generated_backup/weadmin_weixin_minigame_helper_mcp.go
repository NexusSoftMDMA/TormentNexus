package tools

import (
	"context"
	"net/http"
)

func HandlePreview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	gameID, _ :=getString(args, "gameID")
	env, _ :=getString(args, "env")
	if env == "" {
		env = "development"
	}
	url := "https://minigame.weixin.qq.com/preview/" + gameID + "?env=" + env
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to trigger preview: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("preview request returned status " + resp.Status)
}

	return success("preview initiated for game " + gameID)
}

func HandlePublish(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	gameID, _ :=getString(args, "gameID")
	version, _ :=getString(args, "version")
	if gameID == "" || version == "" {
		return err("gameID and version are required")
}

	url := "https://minigame.weixin.qq.com/publish/" + gameID + "?version=" + version
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to publish: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("publish request returned status " + resp.Status)
}

	return success("publish succeeded for game " + gameID + " version " + version)
}