package tools

    import (
        "context"
        "fmt"
    )

    // Removed ok, e, getString definitions

    func HandleXxx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
        return ok("処理成功")
}

    func HandleYyy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
        // Original: return err(fmt.Errorf("エラーが発生しました"))
        // Fix: e expects string.
        return err("エラーが発生しました")
}

    func HandleZzz(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
        return ok("処理成功")
}

    func HandleAaa(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
        return ok("処理成功")
}

    func HandleBbb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
        return ok("処理成功")
    }