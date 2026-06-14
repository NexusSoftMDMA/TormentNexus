package tools

import (
	"context"
	"net/http"
)

func HandleNeedleSize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	skinType, _ :=getString(args, "skin_type")
	if skinType == "" {
		return err("skin_type required")
	}
	size := "0.25mm"
	if skinType == "sensitive" {
		size = "0.20mm"
	} else if skinType == "thick" {
		size = "0.50mm"
	}
	return success("Recommended needle size: " + size)
}

func HandleGuideSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	if keyword == "" {
		return err("keyword required")
	}
	found := false
	if keyword == "acne" || keyword == "wrinkles" || keyword == "scars" {
		found = true
	}
	if !found {
		return err("no guide found")
	}
	return success("Guide found: " + keyword)
}

func HandleProductInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	productID, _ :=getString(args, "product_id")
	if productID == "" {
		return err("product_id required")
	}
	resp, e := http.DefaultClient.Get("https://example.com/products/" + productID)
	if e != nil {
		return err("failed to fetch product info")
	}
	defer resp.Body.Close()
	return success("Product: " + productID + " - Certified for microneedling")
}// touch 1781132144
