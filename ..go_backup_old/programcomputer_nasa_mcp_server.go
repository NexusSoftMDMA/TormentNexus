package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleApod(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	date, _ :=getString(args, "date")
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		apiKey = "DEMO_KEY"
	}
	url := fmt.Sprintf("https://api.nasa.gov/planetary/apod?api_key=%s", apiKey)
	if date != "" {
		url += "&date=" + date
	}
	resp, e := http.Get(url)
	if e != nil {
		return err("failed to fetch APOD: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data struct {
		Title       string `json:"title"`
		Explanation string `json:"explanation"`
		URL         string `json:"url"`
	}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Title: %s\nExplanation: %s\nURL: %s", data.Title, data.Explanation, data.URL))
}

func HandleMarsRover(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sol, _ :=getString(args, "sol")
	camera, _ :=getString(args, "camera")
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		apiKey = "DEMO_KEY"
	}
	url := fmt.Sprintf("https://api.nasa.gov/mars-photos/api/v1/rovers/curiosity/photos?sol=%s&api_key=%s", sol, apiKey)
	if camera != "" {
		url += "&camera=" + camera
	}
	resp, e := http.Get(url)
	if e != nil {
		return err("failed to fetch Mars rover photos: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data struct {
		Photos []struct {
			ID   int    `json:"id"`
			ImgSrc string `json:"img_src"`
		} `json:"photos"`
	}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if len(data.Photos) == 0 {
		return ok("No photos found")
}

	photo := data.Photos[0]
	return ok(fmt.Sprintf("First photo ID: %d, URL: %s", photo.ID, photo.ImgSrc))
}