package tools

import (
    "context"
    "errors"
)

// ToolResponse, TextContent, ok, e, getString, getInt, getBool は parity.go で定義されていると仮定

func ok(text string) (ToolResponse, error) {
    return ToolResponse{Content: []TextContent{{Type: "text", Text: text}}}, nil
}

func err(msg string) (ToolResponse, error) {
    return ToolResponse{}, errors.New(msg)
}

func getString(args map[string]interface{}, key string) string {
    if v, found := args[key]; found {
        if s, found := v.(string); found {
            return s
        }
    }
    return ""
}

func getInt(args map[string]interface{}, key string) (int, error) {
    if v, found := args[key]; found {
        if i, found := v.(int); found {
            return i, nil
        }
    }
    return 0, errors.New("int value not found")
}

func getBool(args map[string]interface{}, key string) (bool, error) {
    if v, found := args[key]; found {
        if b, found := v.(bool); found {
            return b, nil
        }
    }
    return false, errors.New("bool value not found")
}

func HandleGetFileTree(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    // プラグインへのリクエスト処理
    // ここでは仮の実装
    return ok("File Tree Data")
}

func HandleSearchFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    // プラグインへのリクエスト処理
    // ここでは仮の実装
    return ok("Search Files Data")
}

func HandleGetPlaceInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    // プラグインへのリクエスト処理
    // ここでは仮の実装
    return ok("Place Info Data")
}

func HandleGetServices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    // プラグインへのリクエスト処理
    // ここでは仮の実装
    return ok("Services Data")
}

func HandleSearchObjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    // プラグインへのリクエスト処理
    // ここでは仮の実装
    return ok("Search Objects Data")
}

func HandleGetInstanceProperties(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    // プラグインへのリクエスト処理
    // ここでは仮の実装
    return ok("Instance Properties Data")
}