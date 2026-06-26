package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "sort"
    "strings"
    "sync"
    "time"
)

type TextContent string

var (
    sessionStore = struct {
        sync.RWMutex
        sessions map[string]string // sessionId -> conversationId
    }{
        sessions: make(map[string]string),
    }
)

func ok(text string) (ToolResponse, error) {
    return ToolResponse{Text: text}, nil
}

func err(e error) (ToolResponse, error) {
    return ToolResponse{}, e
}

func getString(args map[string]interface{}, key string) string {
    if val, found := args[key]; found {
        if str, found := val.(string); found {
            return str
        }
    }
    return ""
}

func getInt(args map[string]interface{}, key string) (int, bool) {
    if val, found := args[key]; found {
        if intVal, found := val.(int); found {
            return intVal, true
        }
    }
    return 0, false
}

func getBool(args map[string]interface{}, key string) (bool, bool) {
    if val, found := args[key]; found {
        if boolVal, found := val.(bool); found {
            return boolVal, true
        }
    }
    return false, false
}

func HandleCodex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    prompt, _ :=getString(args, "prompt")
    sessionId, _ :=getString(args, "sessionId")
    resetSession, _ :=getBool(args, "resetSession")
    model, _ :=getString(args, "model")
    reasoningEffort, _ :=getString(args, "reasoningEffort")
    sandbox, _ :=getString(args, "sandbox")
    fullAuto, _ :=getBool(args, "fullAuto")
    workingDirectory, _ :=getString(args, "workingDirectory")
    callbackUri, _ :=getString(args, "callbackUri")

    if resetSession && sessionId != "" {
        sessionStore.Lock()
        defer sessionStore.Unlock()
        delete(sessionStore.sessions, sessionId)

    cmdArgs := []string{"exec", "--skip-git-repo-check"}
    if prompt != "" {
        cmdArgs = append(cmdArgs, "--prompt", prompt)

    if sessionId != "" {
        cmdArgs = append(cmdArgs, "--session", sessionId)

    if resetSession {
        cmdArgs = append(cmdArgs, "--reset-session")

    if model != "" {
        cmdArgs = append(cmdArgs, "--model", model)

    if reasoningEffort != "" {
        cmdArgs = append(cmdArgs, "--reasoning-effort", reasoningEffort)

    if sandbox != "" {
        cmdArgs = append(cmdArgs, "--sandbox", sandbox)

    if fullAuto {
        cmdArgs = append(cmdArgs, "--full-auto")

    if workingDirectory != "" {
        cmdArgs = append(cmdArgs, "--working-directory", workingDirectory)

    if callbackUri != "" {
        cmdArgs = append(cmdArgs, "--callback-uri", callbackUri)

    cmd := exec.Command("codex", cmdArgs...)
    out, e := cmd.CombinedOutput()
    if e != nil {
        return err(fmt.Errorf("command failed: %w, output: %s", e, string(out)))
}

    sessionIDRegex := regexp.MustCompile(`(?i)(?:conversation|session)[\s\-:]+([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)
    match := sessionIDRegex.FindStringSubmatch(string(out))
    if len(match) > 1 && sessionId != "" {
        sessionStore.Lock()
        defer sessionStore.Unlock()
        sessionStore.sessions[sessionId] = match[1]
    }

    return ok(string(out))
}

}
}
}
}
}
}
}
}
}
}

func HandleReview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    // 実装は省略
    return ok("review output")
}

func HandleWebsearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    // 実装は省略
    return ok("web search output")
}

func HandleListSessions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    sessionStore.RLock()
    defer sessionStore.RUnlock()
    keys := make([]string, 0, len(sessionStore.sessions))
    for k := range sessionStore.sessions {
        keys = append(keys, k)

    return ok(strings.Join(keys, ", "))
}

}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("pong")
}

func HandleHelp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    cmd := exec.Command("codex", "--help")
    out, e := cmd.CombinedOutput()
    if e != nil {
        return err(fmt.Errorf("command failed: %w, output: %s", e, string(out)))
}

    return ok(string(out))
}