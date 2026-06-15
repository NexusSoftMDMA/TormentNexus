package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleFetchComponent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	typ, _ :=getString(args, "type")
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	base := "https://raw.githubusercontent.com/shadcn-ui/ui/main/apps/www/registry"
	var url string
	switch typ {
	case "demo":
		url = base + "/default/example/" + name + ".tsx"
	case "block":
		url = base + "/default/block/" + name + ".tsx"
	case "metadata":
		url = base + "/default/ui/" + name + ".tsx"
	default:
		url = base + "/default/ui/" + name + ".tsx"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("not found")
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return success(string(body))
}