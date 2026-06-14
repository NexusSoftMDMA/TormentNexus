package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func HandleUploadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	channel, _ :=getString(args, "channel")
	filePath, _ :=getString(args, "file_path")

	file, e := os.Open(filePath)
	if e != nil {
		return err("failed to open file: " + e.Error())
}

	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, e := writer.CreateFormFile("file", filePath)
	if e != nil {
		return err("failed to create form file: " + e.Error())
}

	io.Copy(part, file)
	writer.WriteField("channels", channel)
	writer.Close()

	req, e := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/files.upload", &buf)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("upload request failed: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return err("upload failed with status " + resp.Status)
}

	return ok("file uploaded successfully")
}

func HandleDownloadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	fileID, _ :=getString(args, "file_id")

	infoURL := fmt.Sprintf("https://slack.com/api/files.info?file=%s", fileID)
	req, e := http.NewRequestWithContext(ctx, "GET", infoURL, nil)
	if e != nil {
		return err("failed to create info request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("info request failed: " + e.Error())
}

	defer resp.Body.Close()

	var info map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&info); e != nil {
		return err("failed to decode info: " + e.Error())
}

	if !info["ok"].(bool) {
		return err("slack info error: " + info["error"].(string))
}

	fileData, found := info["file"].(map[string]interface{})
	if !found {
		return err("file not found in response")
}

	urlPrivateDownload, found := fileData["url_private_download"].(string)
	if !found {
		return err("download URL not found")
}

	downloadReq, e := http.NewRequestWithContext(ctx, "GET", urlPrivateDownload, nil)
	if e != nil {
		return err("failed to create download request: " + e.Error())
}

	downloadReq.Header.Set("Authorization", "Bearer "+token)

	downloadResp, e := http.DefaultClient.Do(downloadReq)
	if e != nil {
		return err("download request failed: " + e.Error())
}

	defer downloadResp.Body.Close()

	body, e := io.ReadAll(downloadResp.Body)
	if e != nil {
		return err("failed to read downloaded content: " + e.Error())
}

	return success("file downloaded, size: " + fmt.Sprint(len(body)) + " bytes")
}