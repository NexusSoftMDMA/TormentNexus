package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetJapanTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://worldtimeapi.org/api/timezone/Asia/Tokyo")
	if e != nil {
		return err("failed to fetch time: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleGetJapanPrefectures(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prefs := []string{
		"Hokkaido", "Aomori", "Iwate", "Miyagi", "Akita", "Yamagata", "Fukushima",
		"Ibaraki", "Tochigi", "Gunma", "Saitama", "Chiba", "Tokyo", "Kanagawa",
		"Niigata", "Toyama", "Ishikawa", "Fukui", "Yamanashi", "Nagano", "Gifu",
		"Shizuoka", "Aichi", "Mie", "Shiga", "Kyoto", "Osaka", "Hyogo", "Nara",
		"Wakayama", "Tottori", "Shimane", "Okayama", "Hiroshima", "Yamaguchi",
		"Tokushima", "Kagawa", "Ehime", "Kochi", "Fukuoka", "Saga", "Nagasaki",
		"Kumamoto", "Oita", "Miyazaki", "Kagoshima", "Okinawa",
	}
	b, _ := json.Marshal(prefs)
	return ok(string(b))
}