package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// NewsItem represents a single news item from the CryptoPanic API
type NewsItem struct {
	Title string `json:"title"`
}

// CryptoPanicResponse represents the API response structure
type CryptoPanicResponse struct {
	Results []NewsItem `json:"results"`
}

var http.DefaultClient = http.DefaultClient

// HandleGetCryptoNews fetches the latest cryptocurrency news from CryptoPanic.
func HandleGetCryptoNews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	kind, _ :=getString(args, "kind")
	if kind == "" {
		kind = "news"
	}

	numPages, _ :=getInt(args, "num_pages")
	if numPages == 0 {
		numPages = 1
	}
	if numPages > 10 {
		numPages = 10
	}

	apiKey := os.Getenv("CRYPTOPANIC_API_KEY")
	apiPlan := os.Getenv("CRYPTOPANIC_API_PLAN")
	if apiPlan == "" {
		apiPlan = "developer"
	}

	if apiKey == "" {
		return err("CRYPTOPANIC_API_KEY environment variable is not set")
}

	var allTitles []string

	for page := 1; page <= numPages; page++ {
		baseURL := fmt.Sprintf("https://cryptopanic.com/api/%s/v2/posts/", apiPlan)
		params := url.Values{}
		params.Add("auth_token", apiKey)
		params.Add("kind", kind)
		params.Add("regions", "en")
		params.Add("page", strconv.Itoa(page))

		fullURL := baseURL + "?" + params.Encode()

		req, reqErr := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
		if reqErr != nil {
			return err(reqErr.Error())
}

		resp, httpErr := http.DefaultClient.Do(req)
		if httpErr != nil {
			return err(httpErr.Error())
}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return err(fmt.Sprintf("API request failed with status code: %d", resp.StatusCode))
}

		var apiResp CryptoPanicResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&apiResp)
		if decodeErr != nil {
			return err(decodeErr.Error())
}

		if len(apiResp.Results) == 0 {
			break
		}

		for _, item := range apiResp.Results {
			if item.Title != "" {
				allTitles = append(allTitles, item.Title)

		}
	}

	var sb strings.Builder
	for _, title := range allTitles {
		sb.WriteString("- ")
		sb.WriteString(title)
		sb.WriteString("\n")

	return ok(sb.String())
}
}
}