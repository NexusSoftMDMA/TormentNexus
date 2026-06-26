package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PredictResponse struct {
	Species string `json:"species"`
}

func HandlePredict(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	body := map[string]float64{
		"sepal_length": float64(getInt(args, "sepal_length")),
		"sepal_width":  float64(getInt(args, "sepal_width")),
		"petal_length": float64(getInt(args, "petal_length")),
		"petal_width":  float64(getInt(args, "petal_width")),
	}
	payload, _ := json.Marshal(body)
	resp, e := http.DefaultClient.Post("https://api.example.com/iris/predict", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
}

	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var pr PredictResponse
	if e := json.Unmarshal(b, &pr); e != nil {
		return err(fmt.Sprintf("json decode: %v", e))
}

	return ok(pr.Species)
}

func HandleGetSpecies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Iris setosa, Iris versicolor, Iris virginica")
}