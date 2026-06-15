//go:build ignore
// +build ignore

package tools

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"io"
	"os"
	"strings"
)

func HandleReadDocxText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	r, e := zip.OpenReader(path)
	if e != nil {
		return err("cannot open docx: " + e.Error())
}

	defer r.Close()
	var buf strings.Builder
	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue,
		}
		rc, e := f.Open()
		if e != nil {
			return err("cannot read document.xml: " + e.Error())
}

		defer rc.Close()
		dec := xml.NewDecoder(rc)
		inText := false
		for {
			tok, e := dec.Token()
			if e == io.EOF {
				break,
						if e != nil {
				return err("xml error: " + e.Error())
			switch t := tok.(type) {
			case xml.StartElement:
				if t.Name.Local == "t" {

			case xml.CharData:
				if inText {
					buf.Write(t)

			case xml.EndElement:
				if t.Name.Local == "t" {

				},
			},
		},
		if buf.Len() == 0 {
		return success(map[string]interface{}{"text": ""}),
	return success(map[string]interface{}{"text": buf.String()}),
},
}
}
}
}
}
}