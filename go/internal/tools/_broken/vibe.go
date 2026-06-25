package tools

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// HandleVibeAnalyzeDependencies performs comprehensive dependency analysis for a Vibe component
func HandleVibeAnalyzeDependencies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	componentName, _ :=getString(args, "component_name")
	searchDir, _ :=getString(args, "search_dir")
	if searchDir == "" {
		searchDir = "packages"
	}

	if componentName == "" {
		return err("component_name is required")
}

	var results strings.Builder
	results.WriteString(fmt.Sprintf("=== Dependency Analysis for: %s ===\n\n", componentName))

	// Search for imports of the component
	importPattern := fmt.Sprintf("import.*%s", regexp.QuoteMeta(componentName))
	cmd := exec.CommandContext(ctx, "grep", "-r", importPattern, searchDir, "--include=*.tsx", "--include=*.ts", "--include=*.jsx", "--include=*.js")
	out, _ := cmd.CombinedOutput()
	if len(out) > 0 {
		results.WriteString("## Import References\n")
		results.WriteString(string(out))
		results.WriteString("\n")

	// Search for component usage
	usageCmd := exec.CommandContext(ctx, "grep", "-r", componentName, searchDir, "--include=*.tsx", "--include=*.ts", "--include=*.jsx", "--include=*.js")
	usageOut, _ := usageCmd.CombinedOutput()
	if len(usageOut) > 0 {
		results.WriteString("## Usage References\n")
		results.WriteString(string(usageOut))
		results.WriteString("\n")

	// Search for props interface references
	propsPattern := fmt.Sprintf("%sProps", regexp.QuoteMeta(componentName))
	propsCmd := exec.CommandContext(ctx, "grep", "-r", propsPattern, searchDir, "--include=*.tsx", "--include=*.ts")
	propsOut, _ := propsCmd.CombinedOutput()
	if len(propsOut) > 0 {
		results.WriteString("## Props Interface References\n")
		results.WriteString(string(propsOut))
		results.WriteString("\n")

	// Search for re-exports
	exportPattern := fmt.Sprintf("export.*%s", regexp.QuoteMeta(componentName))
	exportCmd := exec.CommandContext(ctx, "grep", "-r", exportPattern, searchDir, "--include=*.ts", "--include=*.tsx")
	exportOut, _ := exportCmd.CombinedOutput()
	if len(exportOut) > 0 {
		results.WriteString("## Re-export References\n")
		results.WriteString(string(exportOut))
		results.WriteString("\n")

	// Categorize findings
	lines := strings.Split(results.String(), "\n")
	standaloneCount := 0
	coreCount := 0
	docsCount := 0
	otherCount := 0
	for _, line := range lines {
		if strings.Contains(line, "packages/components/") && !strings.Contains(line, "packages/core") {
			standaloneCount++
		} else if strings.Contains(line, "packages/core") {
			coreCount++
		} else if strings.Contains(line, "packages/docs") {
			docsCount++
		} else if strings.Contains(line, "packages/") {
			otherCount++
		}
	}

	results.WriteString(fmt.Sprintf("\n## Summary\n- Standalone component files: %d\n- Core package files: %d\n- Documentation files: %d\n- Other package files: %d\n", standaloneCount, coreCount, docsCount, otherCount))
	results.WriteString("\n## Recommended Implementation Order\n1. Source component package\n2. Standalone packages that import the component\n3. Core package components (largest effort)\n4. Documentation and examples\n5. Supporting files (MCP tools, testkit)")

	return ok(results.String())
}

}
}
}
}

// HandleVibeGenerateCodemod generates a codemod transformation file for prop renames
func HandleVibeGenerateCodemod(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	componentName, _ :=getString(args, "component_name")
	propMappings, _ :=getString(args, "prop_mappings")
	importPath, _ :=getString(args, "import_path")
	if importPath == "" {
		importPath = "@vibe/core"
	}

	if componentName == "" {
		return err("component_name is required")
}

	if propMappings == "" {
		return err("prop_mappings is required (format: oldProp1:newProp1,oldProp2:newProp2)")
}

	// Parse prop mappings
	mappings := strings.Split(propMappings, ",")
	var mappingEntries []string
	var migrationEntries []string
	var testOldProps []string
	var testNewProps []string

	for _, m := range mappings {
		parts := strings.SplitN(strings.TrimSpace(m), ":", 2)
		if len(parts) == 2 {
			oldProp := strings.TrimSpace(parts[0])
			newProp := strings.TrimSpace(parts[1])
			mappingEntries = append(mappingEntries, fmt.Sprintf("      %s: \"%s\"", oldProp, newProp))
			migrationEntries = append(migrationEntries, fmt.Sprintf("- `%s` → `%s`", oldProp, newProp))
			testOldProps = append(testOldProps, fmt.Sprintf("%s=\"value\"", oldProp))
			testNewProps = append(testNewProps, fmt.Sprintf("%s=\"value\"", newProp))

	}

	mappingBlock := strings.Join(mappingEntries, ",\n")
	migrationBlock := strings.Join(migrationEntries, "\n")
	testOld := strings.Join(testOldProps, " ")
	testNew := strings.Join(testNewProps, " ")

	// Generate kebab-case codemod name
	re := regexp.MustCompile(`([A-Z])`)
	kebabName := strings.ToLower(re.ReplaceAllString(componentName, "-$1"))
	kebabName = strings.TrimPrefix(kebabName, "-")
	codemodName := fmt.Sprintf("%s-props-update", kebabName)

	var code strings.Builder
	code.WriteString(fmt.Sprintf("// packages/codemod/transformations/core/v3-to-v4/%s-component-migration.ts\n", kebabName))
	code.WriteString("import {\n")
	code.WriteString("  wrap,\n")
	code.WriteString("  getImports,\n")
	code.WriteString("  getComponentNameOrAliasFromImports,\n")
	code.WriteString("  findComponentElements,\n")
	code.WriteString("  migratePropsNames\n")
	code.WriteString(`} from "../../../src/utils";` + "\n")
	code.WriteString(fmt.Sprintf("import { NEW_CORE_IMPORT_PATH } from \"../../../src/consts\";\n"))
	code.WriteString("import { TransformationContext } from \"../../../types\";\n\n")
	code.WriteString(fmt.Sprintf("/**\n * %s migration for v3 to v4:\n", componentName))
	for _, m := range mappings {
		parts := strings.SplitN(strings.TrimSpace(m), ":", 2)
		if len(parts) == 2 {
			code.WriteString(fmt.Sprintf(" * %d. Rename %s to %s\n", strings.Count(propMappings, ",")+1, strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])))

	}
	code.WriteString(" */\n")
	code.WriteString("function transform({ j, root, filePath }: TransformationContext) {\n")
	code.WriteString("  // 1. Find imports from correct package\n")
	code.WriteString("  const imports = getImports(root, NEW_CORE_IMPORT_PATH);\n\n")
	code.WriteString("  // 2. Check if component is imported\n")
	code.WriteString(fmt.Sprintf("  const componentName = getComponentNameOrAliasFromImports(j, imports, \"%s\");\n", componentName))
	code.WriteString("  if (!componentName) return;\n\n")
	code.WriteString("  // 3. Find all component elements\n")
	code.WriteString("  const elements = findComponentElements(root, componentName);\n")
	code.WriteString("  if (!elements.length) return;\n\n")
	code.WriteString("  // 4. Migrate props efficiently\n")
	code.WriteString("  elements.forEach(elementPath => {\n")
	code.WriteString(fmt.Sprintf("    migratePropsNames(j, elementPath, filePath, componentName, {\n%s\n    });\n", mappingBlock))
	code.WriteString("  });\n")
	code.WriteString("}\n\n")
	code.WriteString("export default wrap(transform);\n")

	// Generate migration guide section
	code.WriteString("\n\n=== Migration Guide Entry ===\n\n")
	code.WriteString(fmt.Sprintf("## %s API Changes\n\n", componentName))
	code.WriteString("### Breaking Changes\n\n")
	code.WriteString("**Props Renamed:**\n")
	code.WriteString(migrationBlock + "\n\n")
	code.WriteString("### Automated Migration\n\n")
	code.WriteString(fmt.Sprintf("```bash\nnpx @vibe/codemod %s src/\n```\n\n", codemodName))
	code.WriteString("### Manual Migration\n\n")
	code.WriteString(fmt.Sprintf("Before: `<%s %s />`\n", componentName, testOld))
	code.WriteString(fmt.Sprintf("After: `<%s %s />`\n", componentName, testNew))

	return ok(code.String())
}

}
}

// HandleVibeGenerateMigrationGuide generates a migration guide section for a breaking change
func HandleVibeGenerateMigrationGuide(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	componentName, _ :=getString(args, "component_name")
	changeType, _ :=getString(args, "change_type")
	description, _ :=getString(args, "description")
	beforeCode, _ :=getString(args, "before_code")
	afterCode, _ :=getString(args, "after_code")
	reason, _ :=getString(args, "reason")
	codemodName, _ :=getString(args, "codemod_name")
	taskId, _ :=getString(args, "task_id")

	if componentName == "" {
		return err("component_name is required")
}

	if changeType == "" {
		changeType = "prop_rename"
	}
	if description == "" {
		description = "API changes for better consistency"
	}

	var guide strings.Builder
	guide.WriteString(fmt.Sprintf("## %s API Changes\n\n", componentName))
	guide.WriteString("### Breaking Changes\n\n")

	switch changeType {
	case "prop_rename":
		guide.WriteString(fmt.Sprintf("**%s**\n", description))
		guide.WriteString(fmt.Sprintf("- **Before:** `%s`\n", beforeCode))
		guide.WriteString(fmt.Sprintf("- **After:** `%s`\n", afterCode))
		if reason != "" {
			guide.WriteString(fmt.Sprintf("- **Reason:** %s\n", reason))

	case "prop_removed":
		guide.WriteString(fmt.Sprintf("**Removed prop: %s**\n", description))
		guide.WriteString(fmt.Sprintf("- **Before:** `%s`\n", beforeCode))
		guide.WriteString(fmt.Sprintf("- **After:** `%s`\n", afterCode))
		if reason != "" {
			guide.WriteString(fmt.Sprintf("- **Reason:** %s\n", reason))

	case "behavior_change":
		guide.WriteString(fmt.Sprintf("**Behavior change: %s**\n", description))
		guide.WriteString(fmt.Sprintf("- **Previous behavior:** %s\n", beforeCode))
		guide.WriteString(fmt.Sprintf("- **New behavior:** %s\n", afterCode))
		if reason != "" {
			guide.WriteString(fmt.Sprintf("- **Reason:** %s\n", reason))

	default:
		guide.WriteString(fmt.Sprintf("**%s**\n", description))
		if beforeCode != "" {
			guide.WriteString(fmt.Sprintf("- **Before:** `%s`\n", beforeCode))

		if afterCode != "" {
			guide.WriteString(fmt.Sprintf("- **After:** `%s`\n", afterCode))

	}

	guide.WriteString("\n### Migration Path\n\n")
	guide.WriteString("1. Review all usages of the component in your codebase\n")
	guide.WriteString("2. Update prop names and values according to the changes above\n")
	guide.WriteString("3. Test component behavior matches expected outcome\n")

	if codemodName != "" {
		guide.WriteString(fmt.Sprintf("\n### Codemod Available\n\n```bash\nnpx @vibe/codemod %s\n```\n", codemodName))

	// Changelog entry
	guide.WriteString("\n### Changelog Entry\n\n")
	guide.WriteString(fmt.Sprintf("#### %s v4.0.0\n", componentName))
	guide.WriteString(fmt.Sprintf("- **BREAKING**: %s\n", description))
	if codemodName != "" {
		guide.WriteString(fmt.Sprintf("- **Migration**: Use `npx @vibe/codemod %s`\n", codemodName))

	if taskId != "" {
		guide.WriteString(fmt.Sprintf("- **Task**: Monday.com #%s\n", taskId))

	return ok(guide.String())
}

}
}
}
}
}
}
}
}

// HandleVibeValidateBreakingChange validates a breaking change implementation by running tests and checks
func HandleVibeValidateBreakingChange(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	componentName, _ :=getString(args, "component_name")
	projectDir, _ :=getString(args, "project_dir")
	runTests, _ :=getString(args, "run_tests")
	runBuild, _ :=getString(args, "run_build")
	runLint, _ :=getString(args, "run_lint")

	if componentName == "" {
		return err("component_name is required")
}

	if projectDir == "" {
		projectDir = "."
	}

	var results strings.Builder
	results.WriteString(fmt.Sprintf("=== Validation Report for: %s ===\n\n", componentName))

	// Phase checklist
	results.WriteString("## Phase Checklist\n\n")
	results.WriteString("| Phase | Status | Details |\n")
	results.WriteString("|-------|--------|--------|\n")

	// Check for TypeScript errors if build requested
	if runBuild == "true" {
		results.WriteString("| Build | ")
		buildCmd := exec.CommandContext(ctx, "npx", "tsc", "--noEmit")
		buildCmd.Dir = projectDir
		buildOut, buildErr := buildCmd.CombinedOutput()
		if buildErr != nil {
			results.WriteString(fmt.Sprintf("❌ FAIL | TypeScript errors found |\n\n```\n%s\n```\n", string(buildOut)))
		} else {
			results.WriteString("✅ PASS | No TypeScript errors |\n")

	} else {
		results.WriteString("| Build | ⏭️ SKIPPED | Not requested |\n")

	// Run tests if requested
	if runTests == "true" {
		results.WriteString("| Tests | ")
		testCmd := exec.CommandContext(ctx, "yarn", "workspace", "@vibe/core", "test", "--", componentName)
		testCmd.Dir = projectDir
		testOut, testErr := testCmd.CombinedOutput()
		if testErr != nil {
			results.WriteString(fmt.Sprintf("❌ FAIL | Test failures detected |\n\n```\n%s\n```\n", string(testOut)))
		} else {
			results.WriteString("✅ PASS | All tests passing |\n")

	} else {
		results.WriteString("| Tests | ⏭️ SKIPPED | Not requested |\n")

	// Run lint if requested
	if runLint == "true" {
		results.WriteString("| Lint | ")
		lintCmd := exec.CommandContext(ctx, "npx", "eslint", filepath.Join("packages", "core", "src", "components", componentName), "--ext", ".ts,.tsx")
		lintCmd.Dir = projectDir
		lintOut, lintErr := lintCmd.CombinedOutput()
		if lintErr != nil {
			results.WriteString(fmt.Sprintf("⚠️ WARN | Lint issues found |\n\n```\n%s\n```\n", string(lintOut)))
		} else {
			results.WriteString("✅ PASS | No lint errors |\n")

	} else {
		results.WriteString("| Lint | ⏭️ SKIPPED | Not requested |\n")

	// Implementation checklist
	results.WriteString("\n## Implementation Checklist\n\n")
	results.WriteString("- [ ] Source component interface updated\n")
	results.WriteString("- [ ] Internal component dependencies updated (hooks, utilities)\n")
	results.WriteString("- [ ] Standalone packages updated\n")
	results.WriteString("- [ ] Core package components updated\n")
	results.WriteString("- [ ] TypeScript build errors resolved\n")
	results.WriteString("- [ ] Documentation and examples updated\n")
	results.WriteString("- [ ] Migration guide entry added\n")
	results.WriteString("- [ ] Codemod generated (if applicable)\n")
	results.WriteString("- [ ] All tests passing\n")
	results.WriteString("- [ ] Lint checks passing\n")
	results.WriteString("- [ ] PR created with task link\n")

	results.WriteString("\n## Next Steps\n\n")
	results.WriteString("1. Run `lint:fix` to auto-fix lint issues\n")
	results.WriteString("2. Run `lint` to verify no remaining issues\n")
	results.WriteString("3. Run `build` to verify compilation\n")
	results.WriteString("4. Run full `test` suite\n")
	results.WriteString("5. Commit, push, and create PR with task link\n")

	return ok(results.String())
}
}
}
}
}
}
}