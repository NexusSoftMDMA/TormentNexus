package tools

import (
	"fmt"
	"os"
)

func (r *Registry) registerFileTools() {
	r.Tools = append(r.Tools, Tool{
		Name:        "read_file",
		Description: "Reads the content of a file. Arguments: file_path (string)",
		Execute: func(args map[string]interface{}) (string, error) {
			path, ok := args["file_path"].(string)
			if !ok {
				return "", fmt.Errorf("file_path must be a string")
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			return string(content), nil
		},
	})

	r.Tools = append(r.Tools, Tool{
		Name:        "write_file",
		Description: "Writes content to a file. Arguments: file_path (string), content (string)",
		Execute: func(args map[string]interface{}) (string, error) {
			path, ok := args["file_path"].(string)
			if !ok {
				return "", fmt.Errorf("file_path must be a string")
			}
			content, ok := args["content"].(string)
			if !ok {
				return "", fmt.Errorf("content must be a string")
			}
			err := os.WriteFile(path, []byte(content), 0644)
			if err != nil {
				return "", err
			}
			return "File written successfully", nil
		},
	})
}
