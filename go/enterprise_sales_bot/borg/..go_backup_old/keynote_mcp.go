package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleListSlides(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	slides := []string{"Slide 1", "Slide 2", "Slide 3"}
	data, e := json.Marshal(slides)
	if e != nil {
		return err("failed to marshal slides: " + e.Error())
}

	return ok(string(data))
}

func HandleGetSlideContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	index, _ :=getInt(args, "index")
	contents := map[int]string{0: "Title slide", 1: "Content slide", 2: "Conclusion"}
	content, found := contents[index]
	if !found {
		return err(fmt.Sprintf("slide index %d not found", index))
}

	return ok(content)
}