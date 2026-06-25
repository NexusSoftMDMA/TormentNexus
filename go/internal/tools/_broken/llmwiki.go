package tools. Then imports: only "context" and "fmt"? Wait wait, wait a second: do we need any other imports? Let's check the handlers: all they do is get a string from args, check if it's empty, return e if so, else return ok with a message. So yes, only context and fmt are needed? Wait wait, but let's make sure. Let's see: getString is pre-defined, so we don't need to import anything for that. ok and e are pre-defined, so no imports needed for them. So the imports are: import ( "context" "fmt" ) Wait but wait, is that all? Let's check. Then, the functions: HandleOpen, HandleMCPConfig, HandleUpload, HandleClone, HandleExtract. Let's write each of them. Let's take HandleOpen: func HandleOpen(ctx context.Context, args map[string]interface{}) (ToolResponse, error) { workspacePath := getString(args, "workspacePath") if workspacePath == "" { return err(fmt.Errorf("workspacePath is required")) } // Open workspace // Implementation details here return ok("Workspace opened successfully") } That's the same as original, but we removed the helper functions and unused imports. Wait but wait: what about the other handlers? Let's check HandleMCPConfig: same pattern, uses getString, checks if empty, returns e or ok. HandleUpload: same := getString(args, "files"), check if empty. HandleClone: repoURL := getString(args, "repoURL"), check. HandleExtract: filePath := getString(args, "filePath"), check. All of those are fine. Wait but wait: is there any other error? Let's check the rules: 1. package tools → yes. 2. func HandleXxx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) → all the handlers follow that. 3. return ok("text"), if e != nil { return err(e.Error()) } → wait, in the original code, the handlers return err(fmt.Errorf(...)), which is okay? Wait the rule says "return ok("text"), if e != nil { return err(e.Error()) }" → but in the handlers, when there's an error (like missing required param), we return e with the error, which is correct. Because e is pre-defined, so that's fine. 4. getString returns single value → yes, the pre-defined getString does that, which we are using. 5. ONLY stdlib imports → we have context and fmt, which are stdlib, no github.com imports. 6. Must compile → let's see: we removed all the redeclared functions, removed all unused imports, removed the references to Success and Message (since we don't define ok and e here, those fields are only used in the pre-defined ones which are correct). Wait but wait a second: what if the pre-defined ok and e functions are not there? No, the problem says they are pre-defined in parity.go, so we don't need to define them here. So that's okay. Wait but let's make sure: the original code's errors were redeclaration of ok, e, etc, which is because they were defined here. So removing those definitions fixes that. The unused imports are removed. The undefined fields Success and Message were in the ok and e functions we removed, so that's fixed. Wait but wait: is there any other issue? Let's check: the ToolResponse is pre-defined, so we don't need to declare it. The TextContent is pre-defined, not used here, so no problem. Let's check the imports again: we have context and fmt, both used. Context is used in all the function signatures, fmt is used for fmt.Errorf in the error returns. Correct. What about strconv? We removed it because we don't use it here, since getInt and getBool are pre-defined and not used in these handlers. Correct. What about strings? Not used, removed. All the other imports are removed. So the code should compile now. Wait but let's write it out properly. Let's see: package tools import ( "context" "fmt" ) // All the helper functions are removed because they are pre-defined in parity.go. Then the handlers: func HandleOpen(ctx context.Context, args map[string]interface{}) (ToolResponse, error) { workspacePath := getString(args, "workspacePath") if workspacePath == "" { return err(fmt.Errorf("workspacePath is required")) } // Open workspace // Implementation details here return ok("Workspace opened successfully") } func HandleMCPConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) { workspacePath := getString(args, "workspacePath") if workspacePath == "" { return err(fmt.Errorf("workspacePath is required")) } // Generate MCP config // Implementation details here return ok("MCP config generated successfully") } func HandleUpload(ctx context.Context, args map[string]interface{}) (ToolResponse, error) { files := getString(args, "files") if files == "" { return err(fmt.Errorf("files are required")) } // Upload files // Implementation details here return ok("Files uploaded successfully") } func HandleClone(ctx context.Context, args map[string]interface{}) (ToolResponse, error) { repoURL := getString(args, "repoURL") if repoURL == "" { return err(fmt.Errorf("repoURL is required")) } // Clone repository // Implementation details here return ok("Repository cloned successfully") } func HandleExtract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) { filePath := getString(args, "filePath") if filePath == "" { return err(fmt.Errorf("filePath is required")) } // Extract file // Implementation details here return ok("File extracted successfully") } Wait but wait: is there any problem here? Let's check: the pre-defined functions are ok, e, getString, getInt, getBool, ToolResponse, TextContent. All of those are used correctly here. We don't redeclare any of them. All unused imports are removed. The only imports are context and fmt, both stdlib, both used. The functions follow the required signature. The returns are ok("text") and err(e.Error()) which matches the rule. Wait the rule 3 says "return ok("text"), if e != nil { return err(e.Error()) }" → in our code, when there's an error (like missing param), we return err(fmt.Errorf(...)), which is equivalent, right? Because e takes an error, and fmt.Errorf returns an error, so that's correct. Because the pre-defined e function is func err(e error) (ToolResponse, error), so passing fmt.Errorf(...) is correct. Yes. So this should compile. Wait but let's make sure we didn't miss any imports. Let's see: do we need any other imports? Let's check: no, because all the helper functions are pre-defined, so we don't need strconv, strings, etc. All the handlers only use context, fmt, getString, ok, e. So yes, the imports are correct. Wait but wait: what if the pre-defined getString is not there? No, the problem says it's pre-defined in parity.go, so we don't need to define it here. So this code should be correct. Let's just make sure there are no redeclarations: we removed all the helper functions, so no redeclaration of ok, e, getString, getInt, getBool. All unused imports are removed. No references to Success or Message fields, because we don't define ok and e here, so those fields are only used in the pre-defined functions which are correct. So this should fix allpackage tools

import (
	"context"
	"fmt"
)

func HandleOpen(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workspacePath, _ :=getString(args, "workspacePath")
	if workspacePath == "" {
		return err(fmt.Errorf("workspacePath is required"))
}

	// Open workspace
	// Implementation details here
	return ok("Workspace opened successfully")
}

func HandleMCPConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workspacePath, _ :=getString(args, "workspacePath")
	if workspacePath == "" {
		return err(fmt.Errorf("workspacePath is required"))
}

	// Generate MCP config
	// Implementation details here
	return ok("MCP config generated successfully")
}

func HandleUpload(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	files, _ :=getString(args, "files")
	if files == "" {
		return err(fmt.Errorf("files are required"))
}

	// Upload files
	// Implementation details here
	return ok("Files uploaded successfully")
}

func HandleClone(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoURL, _ :=getString(args, "repoURL")
	if repoURL == "" {
		return err(fmt.Errorf("repoURL is required"))
}

	// Clone repository
	// Implementation details here
	return ok("Repository cloned successfully")
}

func HandleExtract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "filePath")
	if filePath == "" {
		return err(fmt.Errorf("filePath is required"))
}

	// Extract file
	// Implementation details here
	return ok("File extracted successfully")
}