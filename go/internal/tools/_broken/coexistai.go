package tools

import (
	"context"
)

// getStringSlice はマップから文字列のスライスを取得します。
func getStringSlice(args map[string]interface{}, key string) []string {
	val, found := args[key]
	if !found {
		return []string{}
	}
	strSlice, found := val.([]string)
	if !found {
		return []string{}
	}
	return strSlice
}

// HandleWebSearch はウェブ検索を処理します。
func HandleWebSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	rerank, _ :=getString(args, "rerank") == "true"
	numResults, _ :=getInt(args, "num_results")
	localMode, _ :=getString(args, "local_mode") == "true"
	split, _ :=getString(args, "split") == "true"
	documentPaths := getStringSlice(args, "document_paths")

	return ok("Web search results for: " + query)
}

// HandleYouTubeSearch はYouTube検索を処理します。
func HandleYouTubeSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	prompt, _ :=getString(args, "prompt")
	n, _ :=getInt(args, "n")

	return ok("YouTube search results for: " + query)
}

// HandleRedditSearch はReddit検索を処理します。
func HandleRedditSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subreddit, _ :=getString(args, "subreddit")
	urlType, _ :=getString(args, "url_type")
	n, _ :=getInt(args, "n")
	k, _ :=getInt(args, "k")
	customURL, _ :=getString(args, "custom_url")
	timeFilter, _ :=getString(args, "time_filter")
	searchQuery, _ :=getString(args, "search_query")
	sortType, _ :=getString(args, "sort_type")

	return ok("Reddit search results for: " + subreddit)
}

// HandleMapSearch は地図検索を処理します。
func HandleMapSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	startLocation, _ :=getString(args, "start_location")
	endLocation, _ :=getString(args, "end_location")

	return ok("Map search results from: " + startLocation + " to " + endLocation)
}

// HandleWebSummarize はウェブページの要約を処理します。
func HandleWebSummarize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	localMode, _ :=getString(args, "local_mode") == "true"

	return ok("Summarized content from: " + url)
}