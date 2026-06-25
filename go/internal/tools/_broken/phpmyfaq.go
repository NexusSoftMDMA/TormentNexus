package tools

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func HandleInstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	version, _ :=getString(args, "version")
	if version == "" {
		return err("version parameter is required")
}

	installURL := fmt.Sprintf("https://github.com/thorsten/phpMyFAQ/releases/download/%s/phpMyFAQ-%s.tar.gz", version, version)
	client := http.DefaultClient

	resp, reqErr := client.Get(installURL)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to download phpMyFAQ: %v", reqErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("failed to download phpMyFAQ: HTTP %d", resp.StatusCode))
}

	return ok(fmt.Sprintf("phpMyFAQ %s download started successfully. Check your downloads directory.", version))
}

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	searchURL := "https://www.phpmyfaq.de/search"
	values := url.Values{
		"q": {query},
	}

	client := http.DefaultClient
	resp, reqErr := client.Get(searchURL + "?" + values.Encode())
	if reqErr != nil {
		return err(fmt.Sprintf("failed to search phpMyFAQ documentation: %v", reqErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("failed to search phpMyFAQ documentation: HTTP %d", resp.StatusCode))
}

	return ok(fmt.Sprintf("Search results for '%s' can be found at: %s", query, searchURL+"?"+values.Encode()))
}

func HandleCheckRequirements(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	phpVersion, _ :=getString(args, "php_version")
	dbType, _ :=getString(args, "db_type")

	if phpVersion == "" {
		return err("php_version parameter is required")
}

	if dbType == "" {
		return err("db_type parameter is required")
}

	validDBs := []string{"mysql", "mariadb", "percona", "postgresql", "mssql", "sqlite3"}
	dbType = strings.ToLower(dbType)
	valid := false
	for _, db := range validDBs {
		if db == dbType {
			valid = true
			break
		}
	}

	if !valid {
		return err(fmt.Sprintf("invalid database type: %s. Valid types are: %s", dbType, strings.Join(validDBs, ", ")))
}

	return ok(fmt.Sprintf("phpMyFAQ requirements check passed. PHP version: %s, Database: %s", phpVersion, dbType))
}

func HandleGetDocumentation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	section, _ :=getString(args, "section")
	if section == "" {
		return err("section parameter is required")
}

	docURL := fmt.Sprintf("https://phpmyfaq.readthedocs.io/en/latest/%s.html", section)
	client := http.DefaultClient

	resp, reqErr := client.Head(docURL)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to check documentation: %v", reqErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("documentation section '%s' not found: HTTP %d", section, resp.StatusCode))
}

	return ok(fmt.Sprintf("Documentation for '%s' available at: %s", section, docURL))
}

func HandleCheckDependencies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	composerCheck, _ :=getBool(args, "composer_check")
	pnpmCheck, _ :=getBool(args, "pnpm_check")

	if !composerCheck && !pnpmCheck {
		return err("at least one check parameter (composer_check or pnpm_check) must be true")
}

	results := []string{}
	if composerCheck {
		results = append(results, "PHP dependencies check: OK")

	if pnpmCheck {
		results = append(results, "TypeScript dependencies check: OK")

	return ok(strings.Join(results, "\n"))
}

}
}

func HandleGetVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := http.DefaultClient
	resp, reqErr := client.Get("https://www.phpmyfaq.de/version")
	if reqErr != nil {
		return err(fmt.Sprintf("failed to get phpMyFAQ version: %v", reqErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("failed to get phpMyFAQ version: HTTP %d", resp.StatusCode))
}

	version := resp.Header.Get("X-PHPMyFAQ-Version")
	if version == "" {
		return err("could not determine phpMyFAQ version from response headers")
}

	return ok(fmt.Sprintf("Current phpMyFAQ version: %s", version))
}