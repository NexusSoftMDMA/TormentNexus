package tools

import (
	"context"
)

func HandleWeiboHotSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(`{"hot_searches":[{"title":"埃隆·马斯克","hot":9876543},{"title":"小米SU7","hot":8765432}]}`)
}

func HandleWeiboDetail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	return ok(`{"id":"` + id + `","title":"测试热搜","comments":[{"user":"用户A","text":"评论内容"}]}`)
}