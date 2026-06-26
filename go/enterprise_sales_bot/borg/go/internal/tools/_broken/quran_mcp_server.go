package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetSurah(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	number, _ :=getInt(args, "number")
	if number < 1 || number > 114 {
		return err("surah number must be between 1 and 114")
}

	url := fmt.Sprintf("https://api.alquran.cloud/v1/surah/%d", number)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch surah: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data struct {
		Data struct {
			Number int `json:"number"`
			Name   string `json:"name"`
			EnglishName string `json:"englishName"`
			Verses []struct {
				Number int `json:"number"`
				Text   string `json:"text"`
			} `json:"ayahs"`
		} `json:"data"`
	}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	verses := ""
	for _, v := range data.Data.Verses {
		verses += fmt.Sprintf("%d: %s\n", v.Number, v.Text)

	result := fmt.Sprintf("Surah %d: %s (%s)\n%s", data.Data.Number, data.Data.Name, data.Data.EnglishName, verses)
	return ok(result)
}
}