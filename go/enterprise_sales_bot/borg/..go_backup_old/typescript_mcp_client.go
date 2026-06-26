package tools

import (
	"context"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	response, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to make request")
}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return err("non-200 response")
}

	return success("request successful")
}

func HandleY(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	age, _ :=getInt(args, "age")
	if age < 0 {
		return err("invalid age")
}

	return success("Hello " + name + ", age: " + string(age))
}