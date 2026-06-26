package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleLookupWord(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	word, _ :=getString(args, "word")
	if word == "" {
		return err("missing word parameter")
}

	url := fmt.Sprintf("https://api.dictionaryapi.dev/api/v2/entries/en/%s", word)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("API call failed: %v", e))
}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var entries []struct {
		Word      string `json:"word"`
		Meanings []struct {
			PartOfSpeech string `json:"partOfSpeech"`
			Definitions []struct {
				Definition string `json:"definition"`
			} `json:"definitions"`
		} `json:"meanings"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&entries); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	if len(entries) == 0 {
		return err("no definition found")
}

	result := fmt.Sprintf("Word: %s\n", entries[0].Word)
	for _, m := range entries[0].Meanings {
		for _, d := range m.Definitions {
			result += fmt.Sprintf("- [%s] %s\n", m.PartOfSpeech, d.Definition)

	}
	return success(result)
}
}