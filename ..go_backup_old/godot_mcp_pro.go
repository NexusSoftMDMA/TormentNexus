package tools

import (
	"context"
	"net/http"
)

func HandleCreateNode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	parent, _ :=getString(args, "parent")
	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:18080/tools/create_node?name="+name+"&parent="+parent, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	resp.Body.Close()
	return success("created node " + name + " under " + parent)
}

func HandlePlayAnimation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	anim, _ :=getString(args, "animation")
	loop, _ :=getBool(args, "loop")
	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:18080/tools/play_animation?animation="+anim+"&loop="+fmt.Sprintf("%t", loop), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	resp.Body.Close()
	return success("playing animation " + anim)
}