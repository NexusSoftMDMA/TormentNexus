package tools

import (
	"context"
	"os"
	"strings"
)

func HandleListFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	directory, _ :=getString(args, "directory")
	files, e := listFiles(directory)
	if e != nil {
		return err(e.Error())
}

	return ok(strings.Join(files, ", "))
}

func HandleCreateFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileName, _ :=getString(args, "file_name")
	content, _ :=getString(args, "content")
	e := createFile(fileName, content)
	if e != nil {
		return err(e.Error())
}

	return ok("File created successfully")
}

func listFiles(directory string) ([]string, error) {
	fileInfos, e := os.ReadDir(directory)
	if e != nil {
		return nil, e
	}
	var files []string
	for _, fileInfo := range fileInfos {
		files = append(files, fileInfo.Name())

	return files, nil
}

}

func createFile(fileName, content string) error {
	file, e := os.Create(fileName)
	if e != nil {
		return e
	}
	defer file.Close()
	_, e = file.WriteString(content)
	return e
}