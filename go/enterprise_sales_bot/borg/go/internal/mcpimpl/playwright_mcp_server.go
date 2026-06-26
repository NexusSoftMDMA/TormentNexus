package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGenerateTest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	test := fmt.Sprintf("const { chromium } = require('playwright');\n(async () => {\n  const browser = await chromium.launch();\n  const page = await browser.newPage();\n  await page.goto('%s');\n  console.log(await page.title());\n  await browser.close();\n})();", url)
	return ok(test)
}

func HandleValidateTest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	test, _ :=getString(args, "test")
	if test == "" {
		return err("test is required")
}

	return success("test looks valid")
}