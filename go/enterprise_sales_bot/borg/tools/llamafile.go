package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func (r *Registry) registerLlamafileTools() {
	r.Tools = append(r.Tools, Tool{
		Name:        "download_llamafile",
		Description: "Downloads a standalone local model binary (Llamafile parity). Arguments: url (string), dest_path (string)",
		Execute: func(args map[string]interface{}) (string, error) {
			url, _ := args["url"].(string)
			dest, _ := args["dest_path"].(string)

			resp, err := http.Get(url)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			out, err := os.Create(dest)
			if err != nil {
				return "", err
			}
			defer out.Close()

			_, err = io.Copy(out, resp.Body)
			if err != nil {
				return "", err
			}

			// Make executable
			os.Chmod(dest, 0755)

			return fmt.Sprintf("Llamafile downloaded and ready at %s", dest), nil
		},
	})
}
