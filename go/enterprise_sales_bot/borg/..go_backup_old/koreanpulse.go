package tools

import (
	"context"
	"math/rand"
	"time"
)

func HandleGetCurrentTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	loc, e := time.LoadLocation("Asia/Seoul")
	if e != nil {
		return err("failed to load location")
}

	now := time.Now().In(loc)
	return ok("Current time in Seoul: " + now.Format(time.RFC1123))
}

func HandleGetKoreanPhrase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	phrases := []string{"안녕하세요", "감사합니다", "사랑해요", "좋아요", "대단해요"}
	idx := rand.Intn(len(phrases))
	return ok("Random Korean phrase: " + phrases[idx])
}