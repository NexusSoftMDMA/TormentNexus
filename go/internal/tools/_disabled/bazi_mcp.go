package tools

import "context"

var heavenlyStems = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
var earthlyBranches = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

func HandleGetBazi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	year, _ :=getInt(args, "year")
	stemIdx := (year - 4) % 10
	branchIdx := (year - 4) % 12
	stem := heavenlyStems[stemIdx]
	branch := earthlyBranches[branchIdx]
	return ok("Year pillar: " + stem + branch)
}