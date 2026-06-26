package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// MCP server endpoint
const mcpEndpoint = "http://localhost:18060/mcp"

// HandleLogin - 登录小红书账号
func HandleLogin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cookies, _ := getString(args, "cookies")
	if cookies == "" {
		return err("cookies is required")
	}

	form := url.Values{}
	form.Set("cookies", cookies)

	resp, apiErr := http.PostForm(mcpEndpoint+"/login", form)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(parseErr.Error())
	}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

// HandleCheckLoginStatus - 检查登录状态
func HandleCheckLoginStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, apiErr := http.Get(mcpEndpoint + "/check_login")
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(parseErr.Error())
	}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

// HandlePublishImageText - 发布图文内容
func HandlePublishImageText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ := getString(args, "title")
	content, _ := getString(args, "content")
	imagesJSON, _ := getString(args, "images")
	tags, _ := getString(args, "tags")

	if title == "" || content == "" {
		return err("title and content are required")
	}

	form := url.Values{}
	form.Set("title", title)
	form.Set("content", content)
	form.Set("images", imagesJSON)
	form.Set("tags", tags)

	resp, apiErr := http.PostForm(mcpEndpoint+"/publish/image_text", form)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(parseErr.Error())
	}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

// HandleSearch - 搜索内容
func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ := getString(args, "keyword")
	page, _ := getInt(args, "page")
	pageSize, _ := getInt(args, "page_size")

	if keyword == "" {
		return err("keyword is required")
	}

	form := url.Values{}
	form.Set("keyword", keyword)
	form.Set("page", fmt.Sprintf("%d", page))
	form.Set("page_size", fmt.Sprintf("%d", pageSize))

	resp, apiErr := http.PostForm(mcpEndpoint+"/search", form)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(parseErr.Error())
	}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

// HandleGetRecommendations - 获取推荐列表
func HandleGetRecommendations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	page, _ := getInt(args, "page")
	pageSize, _ := getInt(args, "page_size")

	form := url.Values{}
	form.Set("page", fmt.Sprintf("%d", page))
	form.Set("page_size", fmt.Sprintf("%d", pageSize))

	resp, apiErr := http.PostForm(mcpEndpoint+"/recommendations", form)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(parseErr.Error())
	}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

// HandleGetPostDetails - 获取帖子详情
func HandleGetPostDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	feedID, _ := getString(args, "feed_id")
	xsecToken, _ := getString(args, "xsec_token")

	if feedID == "" || xsecToken == "" {
		return err("feed_id and xsec_token are required")
	}

	form := url.Values{}
	form.Set("feed_id", feedID)
	form.Set("xsec_token", xsecToken)

	resp, apiErr := http.PostForm(mcpEndpoint+"/post/details", form)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(parseErr.Error())
	}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

// HandlePostComment - 发表评论
func HandlePostComment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	feedID, _ := getString(args, "feed_id")
	xsecToken, _ := getString(args, "xsec_token")
	comment, _ := getString(args, "comment")

	if feedID == "" || xsecToken == "" || comment == "" {
		return err("feed_id, xsec_token, and comment are required")
	}

	form := url.Values{}
	form.Set("feed_id", feedID)
	form.Set("xsec_token", xsecToken)
	form.Set("comment", comment)

	resp, apiErr := http.PostForm(mcpEndpoint+"/comment/post", form)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(parseErr.Error())
	}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

// HandleGetUserProfile - 获取用户主页
func HandleGetUserProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ := getString(args, "user_id")
	xsecToken, _ := getString(args, "xsec_token")

	if userID == "" || xsecToken == "" {
		return err("user_id and xsec_token are required")
	}

	form := url.Values{}
	form.Set("user_id", userID)
	form.Set("xsec_token", xsecToken)

	resp, apiErr := http.PostForm(mcpEndpoint+"/user/profile", form)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(parseErr.Error())
	}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

// HandleLikePost - 点赞或取消点赞
func HandleLikePost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	feedID, _ := getString(args, "feed_id")
	xsecToken, _ := getString(args, "xsec_token")
	unlike, _ := getBool(args, "unlike")

	form := url.Values{}
	form.Set("feed_id", feedID)
	form.Set("xsec_token", xsecToken)
	form.Set("unlike", fmt.Sprintf("%t", unlike))

	resp, apiErr := http.PostForm(mcpEndpoint+"/post/like", form)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(parseErr.Error())
	}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

// HandleFavoritePost - 收藏或取消收藏
func HandleFavoritePost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	feedID, _ := getString(args, "feed_id")
	xsecToken, _ := getString(args, "xsec_token")
	unfavorite, _ := getBool(args, "unfavorite")

	form := url.Values{}
	form.Set("feed_id", feedID)
	form.Set("xsec_token", xsecToken)
	form.Set("unfavorite", fmt.Sprintf("%t", unfavorite))

	resp, apiErr := http.PostForm(mcpEndpoint+"/post/favorite", form)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(parseErr.Error())
	}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

// HandlePublishVideo - 发布视频内容
func HandlePublishVideo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ := getString(args, "title")
	content, _ := getString(args, "content")
	videoPath, _ := getString(args, "video_path")
	tags, _ := getString(args, "tags")

	if title == "" || content == "" || videoPath == "" {
		return err("title, content, and video_path are required")
	}

	form := url.Values{}
	form.Set("title", title)
	form.Set("content", content)
	form.Set("video_path", videoPath)
	form.Set("tags", tags)

	resp, apiErr := http.PostForm(mcpEndpoint+"/publish/video", form)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(parseErr.Error())
	}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

// HandleReplyComment - 回复评论
func HandleReplyComment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	feedID, _ := getString(args, "feed_id")
	xsecToken, _ := getString(args, "xsec_token")
	commentID, _ := getString(args, "comment_id")
	userID, _ := getString(args, "user_id")
	replyContent, _ := getString(args, "reply_content")

	if feedID == "" || xsecToken == "" || replyContent == "" {
		return err("feed_id, xsec_token, and reply_content are required")
	}

	form := url.Values{}
	form.Set("feed_id", feedID)
	form.Set("xsec_token", xsecToken)
	form.Set("reply_content", replyContent)
	form.Set("comment_id", commentID)
	form.Set("user_id", userID)

	resp, apiErr := http.PostForm(mcpEndpoint+"/comment/reply", form)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(parseErr.Error())
	}

	data, _ := json.Marshal(result)
	return ok(string(data))
}