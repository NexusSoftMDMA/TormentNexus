package mcpimpl

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"strings"
)

func HandleReadEmailAttachments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	f, e := os.Open(path)
	if e != nil {
		return err("open file: " + e.Error())
}

	defer f.Close()

	msg, e := mail.ReadMessage(f)
	if e != nil {
		return err("read message: " + e.Error())
}

	mediaType, params, e := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if e != nil || !strings.HasPrefix(mediaType, "multipart/") {
		return ok("no attachments found")
}

	mr := multipart.NewReader(msg.Body, params["boundary"])
	var attachments []string
	for {
		part, e := mr.NextPart()
		if e == io.EOF {
			break
		}
		if e != nil {
			return err("read part: " + e.Error())
}

		_, params, e := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if e == nil && params["name"] != "" {
			data, e := io.ReadAll(part)
			if e != nil {
				continue
			}
			b64 := base64.StdEncoding.EncodeToString(data)
			attachments = append(attachments, fmt.Sprintf("name=%s, size=%d, b64=%s...", params["name"], len(data), trunc(b64, 20)))

	}
	if len(attachments) == 0 {
		return ok("no attachments found")
}

	return success("attachments: " + strings.Join(attachments, "; "))
}

}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}