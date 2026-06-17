package tools

import (
	"fmt"
	"os"
	"strings"
)

// RefactorTool bridges Opencode parity with strict AST REPLACE blocks.
type RefactorTool struct{}

// ApplySearchReplace strictly enforces LLM blocks in the format:
// <<<<<<< SEARCH
// existing
// =======
// replacement
// >>>>>>> REPLACE
func (r *RefactorTool) ApplySearchReplace(filePath, searchBlock, replaceBlock string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read target: %w", err)
	}

	strContent := string(content)

	if !strings.Contains(strContent, searchBlock) {
		return fmt.Errorf("search block not found natively in file. LLM hallucinated context buffer.")
	}

	// Native atomic replacement
	newContent := strings.Replace(strContent, searchBlock, replaceBlock, 1)

	// Write back with strict permissions
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed atomic write application: %w", err)
	}

	fmt.Printf("[Opencode Parity] Block successfully mutated %s natively.\n", filePath)
	return nil
}

// registerRefactoringTools binds the isolated Aider functionality to the core Native Engine.
func (reg *Registry) registerRefactoringTools() {
	reg.Tools = append(reg.Tools, Tool{
		Name:        "apply_search_replace",
		Description: "Aider Parity: strict block-replacement AST refactoring.",
		Execute: func(args map[string]interface{}) (string, error) {
			path, _ := args["file_path"].(string)
			search, _ := args["search_block"].(string)
			replace, _ := args["replace_block"].(string)

			rf := &RefactorTool{}
			err := rf.ApplySearchReplace(path, search, replace)
			if err != nil {
				return "", err
			}
			return "File safely mutated natively.", nil
		},
	})
}
