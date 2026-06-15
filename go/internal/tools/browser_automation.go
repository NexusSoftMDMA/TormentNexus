//go:build ignore
// +build ignore

package tools

/**
 * @file browser_automation.go
 * @module go/internal/tools
 *
 * WHAT: Go-native browser automation handlers using chromedp.
 * Provides navigation, screenshots, HTML extraction, JavaScript evaluation, and form interactions.
 *
 * WHY: Replaces external puppeteer/browser-use MCP servers with a lightweight Go-native implementation.
 */

import (
	"time"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/chromedp/chromedp"
)

// HandleBrowserNavigate navigates to a URL.
// Args: url (string, required)
func HandleBrowserNavigate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ := getString(args, "url")
	if url == "" {
		return err("url is required")
	}

	timeoutMs := getInt(args, "timeout")
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	allocCtx, cancel1 := chromedp.NewExecAllocator(ctx,
		chromedp.Headless, chromedp.NoSandbox, chromedp.DisableGPU, chromedp.WindowSize(1920, 1080),
	)
	defer cancel1()

	taskCtx, cancel2 := chromedp.NewContext(allocCtx)
	defer cancel2()

	var title string
	if runErr := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Title(&title),
	); runErr != nil {
		return err(fmt.Sprintf("Navigation failed: %v", runErr))
	}

	return ok(fmt.Sprintf("Navigated to %s (title: %s)", url, title))
}

// HandleBrowserScreenshot captures a screenshot of a page.
// Args: url (string, required)
func HandleBrowserScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ := getString(args, "url")
	if url == "" {
		return err("url is required")
	}

	timeoutMs := getInt(args, "timeout")
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	fullPage := getBool(args, "fullPage")

	allocCtx, cancel1 := chromedp.NewExecAllocator(ctx,
		chromedp.Headless, chromedp.NoSandbox, chromedp.DisableGPU, chromedp.WindowSize(1920, 1080),
	)
	defer cancel1()

	taskCtx, cancel2 := chromedp.NewContext(allocCtx)
	defer cancel2()

	var buf []byte
	if fullPage {
		if runErr := chromedp.Run(taskCtx,
			chromedp.Navigate(url),
			chromedp.WaitReady("body", chromedp.ByQuery),
			chromedp.FullScreenshot(&buf, 80),
		); runErr != nil {
			return err(fmt.Sprintf("Full-page screenshot failed: %v", runErr))
		}
	} else {
		if runErr := chromedp.Run(taskCtx,
			chromedp.Navigate(url),
			chromedp.WaitReady("body", chromedp.ByQuery),
			chromedp.CaptureScreenshot(&buf),
		); runErr != nil {
			return err(fmt.Sprintf("Viewport screenshot failed: %v", runErr))
		}
	}

	base64Img := base64.StdEncoding.EncodeToString(buf)
	return ok("data:image/png;base64," + base64Img)
}

// HandleBrowserGetHTML retrieves the full HTML of a page.
// Args: url (string, required)
func HandleBrowserGetHTML(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ := getString(args, "url")
	if url == "" {
		return err("url is required")
	}

	timeoutMs := getInt(args, "timeout")
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	allocCtx, cancel1 := chromedp.NewExecAllocator(ctx,
		chromedp.Headless, chromedp.NoSandbox, chromedp.DisableGPU, chromedp.WindowSize(1920, 1080),
	)
	defer cancel1()

	taskCtx, cancel2 := chromedp.NewContext(allocCtx)
	defer cancel2()

	var html string
	if runErr := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.OuterHTML(":root", &html),
	); runErr != nil {
		return err(fmt.Sprintf("HTML retrieval failed: %v", runErr))
	}

	return ok(html)
}

// HandleBrowserEvaluate executes JavaScript on a page and returns the result.
// Args: url (string, required), script (string, required)
func HandleBrowserEvaluate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ := getString(args, "url")
	script, _ := getString(args, "script")

	if url == "" {
		return err("url is required")
	}
	if script == "" {
		return err("script is required")
	}

	timeoutMs := getInt(args, "timeout")
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	allocCtx, cancel1 := chromedp.NewExecAllocator(ctx,
		chromedp.Headless, chromedp.NoSandbox, chromedp.DisableGPU, chromedp.WindowSize(1920, 1080),
	)
	defer cancel1()

	taskCtx, cancel2 := chromedp.NewContext(allocCtx)
	defer cancel2()

	var result interface{}
	if runErr := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Evaluate(script, &result),
	); runErr != nil {
		return err(fmt.Sprintf("Evaluation failed: %v", runErr))
	}

	resultJSON, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return err(fmt.Sprintf("Failed to marshal result: %v", marshalErr))
	}

	return ok(string(resultJSON))
}

// HandleBrowserClick clicks an element on a page.
// Args: url (string, required), selector (string, required)
func HandleBrowserClick(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ := getString(args, "url")
	selector, _ := getString(args, "selector")

	if url == "" {
		return err("url is required")
	}
	if selector == "" {
		return err("selector is required")
	}

	timeoutMs := getInt(args, "timeout")
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	allocCtx, cancel1 := chromedp.NewExecAllocator(ctx,
		chromedp.Headless, chromedp.NoSandbox, chromedp.DisableGPU, chromedp.WindowSize(1920, 1080),
	)
	defer cancel1()

	taskCtx, cancel2 := chromedp.NewContext(allocCtx)
	defer cancel2()

	if runErr := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Click(selector, chromedp.ByQuery),
	); runErr != nil {
		return err(fmt.Sprintf("Click failed: %v", runErr))
	}

	return ok(fmt.Sprintf("Clicked element: %s", selector))
}

// HandleBrowserFillForm fills an input field with a value.
// Args: url (string, required), selector (string, required), value (string, required)
func HandleBrowserFillForm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ := getString(args, "url")
	selector, _ := getString(args, "selector")
	value, _ := getString(args, "value")

	if url == "" {
		return err("url is required")
	}
	if selector == "" {
		return err("selector is required")
	}

	timeoutMs := getInt(args, "timeout")
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	allocCtx, cancel1 := chromedp.NewExecAllocator(ctx,
		chromedp.Headless, chromedp.NoSandbox, chromedp.DisableGPU, chromedp.WindowSize(1920, 1080),
	)
	defer cancel1()

	taskCtx, cancel2 := chromedp.NewContext(allocCtx)
	defer cancel2()

	if runErr := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.SendKeys(selector, value, chromedp.ByQuery),
	); runErr != nil {
		return err(fmt.Sprintf("Fill form failed: %v", runErr))
	}

	return ok(fmt.Sprintf("Filled %s with value: %s", selector, value))
}
