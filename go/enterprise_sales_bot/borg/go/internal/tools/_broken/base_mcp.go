package mcpimpl

import (
	"context"
	"net/http"
	"strconv"
)

func HandleEcho_base_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}

func HandleAdd_base_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	return success(strconv.Itoa(sum))
}

		return ""
	}
	return s
}

		return 0
	}
	return int(f)
}