package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// UE5 MCP Tool Handlers
// Reimplementation of the UE5-MCP server functionality in Go

// HandleGenerateScene creates a scene based on a natural language description
func HandleGenerateScene(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	description, _ :=getString(args, "description")
	if description == "" {
		return err("description parameter is required")
}

	// Validate description length
	if len(description) > 1000 {
		return err("description exceeds maximum length of 1000 characters")
}

	// Simulate scene generation process
	sceneData := map[string]interface{}{
		"scene_name":   "generated_scene_" + strconv.FormatInt(time.Now().Unix(), 10),
		"description":  description,
		"status":       "generated",
		"objects":      []string{},
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	// Parse objects if provided
	objectsStr, _ :=getString(args, "objects")
	if objectsStr != "" {
		objects := strings.Split(objectsStr, ",")
		for i := range objects {
			objects[i] = strings.TrimSpace(objects[i])

		sceneData["objects"] = objects
	}

	result, marshalErr := json.MarshalIndent(sceneData, "", "  ")
	if marshalErr != nil {
		return err("failed to marshal scene data: " + marshalErr.Error())
}

	return ok(string(result))
}

}

// HandleExportAsset exports a Blender asset to a specified format for UE5 import
func HandleExportAsset(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	assetName, _ :=getString(args, "asset_name")
	format, _ :=getString(args, "format")
	filepathStr, _ :=getString(args, "filepath")

	if assetName == "" {
		return err("asset_name parameter is required")
}

	if format == "" {
		return err("format parameter is required")
}

	// Validate format
	validFormats := map[string]bool{"fbx": true, "obj": true, "gltf": true, "usd": true}
	if !validFormats[strings.ToLower(format)] {
		return err("unsupported format: " + format + ". Supported: fbx, obj, gltf, usd")
}

	// Determine output path
	outputPath := filepathStr
	if outputPath == "" {
		outputPath = filepath.Join(".", "exports", assetName+"."+strings.ToLower(format))

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil {
		return err("failed to create output directory: " + mkdirErr.Error())
}

	// Simulate export by creating a placeholder file
	placeholderContent := fmt.Sprintf("# UE5 Asset Export\n# Asset: %s\n# Format: %s\n# Exported: %s\n",
		assetName, strings.ToUpper(format), time.Now().Format(time.RFC3339))

	writeErr := os.WriteFile(outputPath, []byte(placeholderContent), 0644)
	if writeErr != nil {
		return err("failed to write export file: " + writeErr.Error())
}

	result := map[string]interface{}{
		"asset_name":   assetName,
		"format":       strings.ToLower(format),
		"output_path":  outputPath,
		"status":       "exported",
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	resultJSON, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return err("failed to marshal export result: " + marshalErr.Error())
}

	return ok(string(resultJSON))
}

}

// HandleGenerateTerrain creates procedural terrain for UE5
func HandleGenerateTerrain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	width, _ :=getInt(args, "width")
	height, _ :=getInt(args, "height")
	detailLevel, _ :=getString(args, "detail_level")

	if width <= 0 {
		width = 1000
	}
	if height <= 0 {
		height = 1000
	}
	if detailLevel == "" {
		detailLevel = "medium"
	}

	// Validate detail level
	validDetails := map[string]bool{"low": true, "medium": true, "high": true}
	if !validDetails[strings.ToLower(detailLevel)] {
		return err("invalid detail_level: " + detailLevel + ". Use: low, medium, high")
}

	// Calculate vertices based on detail level
	vertexMultiplier := map[string]int{"low": 1, "medium": 4, "high": 16}
	vertices := (width * height * vertexMultiplier[strings.ToLower(detailLevel)]) / 1000

	terrainData := map[string]interface{}{
		"terrain_name":  "terrain_" + strconv.FormatInt(time.Now().Unix(), 10),
		"width":         width,
		"height":        height,
		"detail_level":  strings.ToLower(detailLevel),
		"vertices":      vertices,
		"status":        "generated",
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	result, marshalErr := json.MarshalIndent(terrainData, "", "  ")
	if marshalErr != nil {
		return err("failed to marshal terrain data: " + marshalErr.Error())
}

	return ok(string(result))
}

// HandlePopulateLevel places assets in a UE5 level
func HandlePopulateLevel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	assetType, _ :=getString(args, "asset_type")
	density, _ :=getInt(args, "density")

	if assetType == "" {
		return err("asset_type parameter is required")
}

	if density <= 0 {
		density = 100
	}
	if density > 10000 {
		return err("density exceeds maximum of 10000")
}

	// Validate asset type
	validTypes := map[string]bool{"trees": true, "rocks": true, "buildings": true, "foliage": true, "props": true}
	if !validTypes[strings.ToLower(assetType)] {
		return err("unsupported asset_type: " + assetType + ". Supported: trees, rocks, buildings, foliage, props")
}

	placementData := map[string]interface{}{
		"level_name":   "level_" + strconv.FormatInt(time.Now().Unix(), 10),
		"asset_type":   strings.ToLower(assetType),
		"density":      density,
		"placed_count": density,
		"status":       "populated",
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	result, marshalErr := json.MarshalIndent(placementData, "", "  ")
	if marshalErr != nil {
		return err("failed to marshal placement data: " + marshalErr.Error())
}

	return ok(string(result))
}

// HandleGenerateBlueprint creates a UE5 Blueprint from a description
func HandleGenerateBlueprint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	logicDescription, _ :=getString(args, "logic_description")

	if logicDescription == "" {
		return err("logic_description parameter is required")
}

	if len(logicDescription) > 2000 {
		return err("logic_description exceeds maximum length of 2000 characters")
}

	// Generate a blueprint name from description
	blueprintName := "BP_" + strings.ReplaceAll(strings.Title(strings.ToLower(logicDescription[:min(30, len(logicDescription))]), " ", "_")
	blueprintName = strings.ReplaceAll(blueprintName, ",", "")
	blueprintName = strings.ReplaceAll(blueprintName, ".", "")
	blueprintName = strings.ReplaceAll(blueprintName, "!", "")
	blueprintName = strings.ReplaceAll(blueprintName, "?", "")

	blueprintData := map[string]interface{}{
		"blueprint_name": blueprintName,
		"description":    logicDescription,
		"nodes":          []string{"Event BeginPlay", "Tick", "Custom Event"},
		"status":         "generated",
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	result, marshalErr := json.MarshalIndent(blueprintData, "", "  ")
	if marshalErr != nil {
		return err("failed to marshal blueprint data: " + marshalErr.Error())
}

	return ok(string(result))
}

// HandleProfilePerformance runs performance analysis on a UE5 level
func HandleProfilePerformance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	levelName, _ :=getString(args, "level_name")

	if levelName == "" {
		return err("level_name parameter is required")
}

	// Simulate performance profiling
	profileData := map[string]interface{}{
		"level_name":           levelName,
		"fps":                  60.0,
		"draw_calls":           1200,
		"triangles":            500000,
		"texture_memory_mb":    256.5,
		"lighting_cost":        "medium",
		"physics_cost":         "low",
		"recommendations": []string{
			"Consider LOD optimization for distant meshes",
			"Batch materials to reduce draw calls",
			"Review shadow casting lights for optimization",
		},
		"status":               "completed",
		"timestamp":            time.Now().Format(time.RFC3339),
	}

	result, marshalErr := json.MarshalIndent(profileData, "", "  ")
	if marshalErr != nil {
		return err("failed to marshal profile data: " + marshalErr.Error())
}

	return ok(string(result))
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}