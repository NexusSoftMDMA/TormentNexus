package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Notebook represents a notebook in the system.
type Notebook struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

var (
	notebooks = make(map[string]*Notebook)
	nextID    = 1
	mu        sync.Mutex
)

// HandleCreateNotebook creates a new notebook.
func HandleCreateNotebook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	if title == "" {
		return err("title is required")
}

	content, _ :=getString(args, "content")

	mu.Lock()
	defer mu.Unlock()

	id := fmt.Sprintf("%d", nextID)
	nextID++
	now := time.Now().Format(time.RFC3339)
	notebook := &Notebook{
		ID:        id,
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	notebooks[id] = notebook

	notebookJSON, marshalErr := json.Marshal(notebook)
	if marshalErr != nil {
		return err("failed to marshal notebook: " + marshalErr.Error())
}

	return ok(string(notebookJSON))
}

// HandleListNotebooks lists all notebooks.
func HandleListNotebooks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	mu.Lock()
	defer mu.Unlock()

	notebookList := make([]*Notebook, 0, len(notebooks))
	for _, notebook := range notebooks {
		notebookList = append(notebookList, notebook)

	notebookListJSON, marshalErr := json.Marshal(notebookList)
	if marshalErr != nil {
		return err("failed to marshal notebooks: " + marshalErr.Error())
}

	return ok(string(notebookListJSON))
}

}

// HandleGetNotebook gets a notebook by ID.
func HandleGetNotebook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	mu.Lock()
	defer mu.Unlock()

	notebook, exists := notebooks[id]
	if !exists {
		return err("notebook not found")
}

	notebookJSON, marshalErr := json.Marshal(notebook)
	if marshalErr != nil {
		return err("failed to marshal notebook: " + marshalErr.Error())
}

	return ok(string(notebookJSON))
}

// HandleUpdateNotebook updates a notebook by ID.
func HandleUpdateNotebook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")

	mu.Lock()
	defer mu.Unlock()

	notebook, exists := notebooks[id]
	if !exists {
		return err("notebook not found")
}

	if title != "" {
		notebook.Title = title
	}
	if content != "" {
		notebook.Content = content
	}
	notebook.UpdatedAt = time.Now().Format(time.RFC3339)

	notebookJSON, marshalErr := json.Marshal(notebook)
	if marshalErr != nil {
		return err("failed to marshal notebook: " + marshalErr.Error())
}

	return ok(string(notebookJSON))
}

// HandleDeleteNotebook deletes a notebook by ID.
func HandleDeleteNotebook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	mu.Lock()
	defer mu.Unlock()

	_, exists := notebooks[id]
	if !exists {
		return err("notebook not found")
}

	delete(notebooks, id)

	return ok("notebook deleted")
}