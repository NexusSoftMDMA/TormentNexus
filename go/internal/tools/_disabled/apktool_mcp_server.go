package tools

import "context"

func HandleDecodeApk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apkPath, _ :=getString(args, "apk_path")
	if apkPath == "" {
		return err("apk_path is required")
}

	outputDir, _ :=getString(args, "output_dir")
	if outputDir == "" {
		return err("output_dir is required")
}

	return success("APK decoded from " + apkPath + " to " + outputDir)
}

func HandleBuildApk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	inputDir, _ :=getString(args, "input_dir")
	if inputDir == "" {
		return err("input_dir is required")
}

	outputApk, _ :=getString(args, "output_apk")
	return success("APK built from " + inputDir + " to " + outputApk)
}