package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type pronunciationResponse []struct {
	Phonetics []struct {
		Text  string `json:"text"`
		Audio string `json:"audio"`
	} `json:"phonetics"`
}

func HandlePronunciation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	word, _ :=getString(args, "word")
	if word == "" {
		return err("word is required")
}

	url := fmt.Sprintf("https://api.dictionaryapi.dev/api/v2/entries/en/%s", word)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch pronunciation: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("word not found or API error")
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data pronunciationResponse
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse pronunciation data")
}

	if len(data) == 0 || len(data[0].Phonetics) == 0 {
		return err("no pronunciation available")
}

	phonetic := data[0].Phonetics[0]
	result := fmt.Sprintf("Pronunciation: %s\nAudio: %s", phonetic.Text, phonetic.Audio)
	return ok(result)
}