package tools. Then the imports, make sure all required stdlib ones are there: context, encoding/json, fmt, io, net/http, net/url, strconv, strings, time. Then the four handlers, make sure each follows the signature, no e shadowing, uses the pre-defined functions from parity.go. Wait let's check the imports again: yes, all the ones we need are there. Let's make sure each handler is properly closed, no syntax errors. Let's write the full code:

First the package line, then imports, then each handler. Let's make sure the third handler's output is complete, the fourth handler too. Oh right, the third handler was cut off earlier, so finish that, then the fourth. Let's make sure all error variables are named differently, no shadowing of err("error"). Let's check:

Wait for the third handler, after parsing, format all the fields, including owners. Then the fourth handler: HandleDownloadDriveFile. Let's write that properly. Let's make sure all code compiles. Let's put it all together now.package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func HandleListDriveFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accessToken, _ :=getString(args, "access_token")
	if accessToken == "" {
		return err("missing required parameter: access_token")
}

	folderId, _ :=getString(args, "folder_id")
	pageToken, _ :=getString(args, "page_token")
	pageSize, _ :=getInt(args, "page_size")
	if pageSize <= 0 {
		pageSize = 10
	}

	params := url.Values{}
	if folderId != "" {
		params.Add("q", fmt.Sprintf("'%s' in parents", folderId))

	if pageToken != "" {
		params.Add("pageToken", pageToken)

	params.Add("pageSize", strconv.Itoa(pageSize))
	params.Add("fields", "files(id,name,mimeType,size,modifiedTime,webViewLink),nextPageToken")

	apiURL := "https://www.googleapis.com/drive/v3/files?" + params.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Google Drive API error: %s - %s", resp.Status, string(body)))
}

	var listResult struct {
		Files []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			MimeType    string `json:"mimeType"`
			Size        string `json:"size"`
			ModifiedTime string `json:"modifiedTime"`
			WebViewLink string `json:"webViewLink"`
		} `json:"files"`
		NextPageToken string `json:"nextPageToken"`
	}
	parseErr := json.Unmarshal(body, &listResult)
	if parseErr != nil {
		return err(parseErr.Error())
}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d files in Google Drive:\n", len(listResult.Files)))
	for _, file := range listResult.Files {
		output.WriteString(fmt.Sprintf("- ID: %s, Name: %s, Type: %s, Size: %s, Modified: %s, Link: %s\n", file.ID, file.Name, file.MimeType, file.Size, file.ModifiedTime, file.WebViewLink))

	if listResult.NextPageToken != "" {
		output.WriteString(fmt.Sprintf("\nNext page token: %s", listResult.NextPageToken))

	return ok(output.String())
}

}
}
}
}

func HandleSearchDriveFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accessToken, _ :=getString(args, "access_token")
	if accessToken == "" {
		return err("missing required parameter: access_token")
}

	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing required parameter: query")
}

	escapedQuery := strings.ReplaceAll(query, "'", "\\'")
	params := url.Values{}
	params.Add("q", fmt.Sprintf("name contains '%s'", escapedQuery))
	params.Add("pageSize", "10")
	params.Add("fields", "files(id,name,mimeType,size,modifiedTime,webViewLink),nextPageToken")

	apiURL := "https://www.googleapis.com/drive/v3/files?" + params.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Google Drive API error: %s - %s", resp.Status, string(body)))
}

	var searchResult struct {
		Files []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			MimeType    string `json:"mimeType"`
			Size        string `json:"size"`
			ModifiedTime string `json:"modifiedTime"`
			WebViewLink string `json:"webViewLink"`
		} `json:"files"`
		NextPageToken string `json:"nextPageToken"`
	}
	parseErr := json.Unmarshal(body, &searchResult)
	if parseErr != nil {
		return err(parseErr.Error())
}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Search results for '%s' (%d files found):\n", query, len(searchResult.Files)))
	for _, file := range searchResult.Files {
		output.WriteString(fmt.Sprintf("- ID: %s, Name: %s, Type: %s, Size: %s, Modified: %s, Link: %s\n", file.ID, file.Name, file.MimeType, file.Size, file.ModifiedTime, file.WebViewLink))

	if searchResult.NextPageToken != "" {
		output.WriteString(fmt.Sprintf("\nNext page token: %s", searchResult.NextPageToken))

	return ok(output.String())
}

}
}

func HandleGetDriveFileMetadata(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accessToken, _ :=getString(args, "access_token")
	if accessToken == "" {
		return err("missing required parameter: access_token")
}

	fileId, _ :=getString(args, "file_id")
	if fileId == "" {
		return err("missing required parameter: file_id")
}

	apiURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?fields=id,name,mimeType,size,modifiedTime,webViewLink,owners,createdTime", fileId)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Google Drive API error: %s - %s", resp.Status, string(body)))
}

	var metadata struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		MimeType    string `json:"mimeType"`
		Size        string `json:"size"`
		ModifiedTime string `json:"modifiedTime"`
		WebViewLink string `json:"webViewLink"`
		Owners      []struct {
			DisplayName string `json:"displayName"`
			Email       string `json:"email"`
		} `json:"owners"`
		CreatedTime string `json:"createdTime"`
	}
	parseErr := json.Unmarshal(body, &metadata)
	if parseErr != nil {
		return err(parseErr.Error())
}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Metadata for file %s:\n", metadata.Name))
	output.WriteString(fmt.Sprintf("- ID: %s\n", metadata.ID))
	output.WriteString(fmt.Sprintf("- MIME Type: %s\n", metadata.MimeType))
	output.WriteString(fmt.Sprintf("- Size: %s bytes\n", metadata.Size))
	output.WriteString(fmt.Sprintf("- Modified: %s\n", metadata.ModifiedTime))
	output.WriteString(fmt.Sprintf("- Created: %s\n", metadata.CreatedTime))
	output.WriteString(fmt.Sprintf("- Link: %s\n", metadata.WebViewLink))
	if len(metadata.Owners) > 0 {
		output.WriteString(fmt.Sprintf("- Owner: %s (%s)\n", metadata.Owners[0].DisplayName, metadata.Owners[0].Email))

	return ok(output.String())
}

}

func HandleDownloadDriveFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accessToken, _ :=getString(args, "access_token")
	if accessToken == "" {
		return err("missing required parameter: access_token")
}

	fileId, _ :=getString(args, "file_id")
	if fileId == "" {
		return err("missing required parameter: file_id")
}

	apiURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?alt=media", fileId)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Google Drive API error: %s - %s", resp.Status, string(body)))
}

	return ok(fmt.Sprintf("Successfully downloaded file %s, total size: %d bytes", fileId, len(body)))
}