package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// HandleBinwalkScan scans firmware for embedded files, filesystems, and signatures.
func HandleBinwalkScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "filepath")
	if filePath == "" {
		return err("filepath is required")
}

	if _, checkErr := os.Stat(filePath); checkErr != nil {
		return err(fmt.Sprintf("file not found: %s", filePath))
}

	timeout, _ :=getInt(args, "timeout")
	if timeout <= 0 {
		timeout = 300
	}

	ctx2, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx2, "binwalk", filePath)
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("binwalk scan failed: %s: %s", cmdErr.Error(), string(out)))
}

	signatures := parseBinwalkOutput(string(out))
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Scan results for: %s\n", filepath.Base(filePath)))
	sb.WriteString(fmt.Sprintf("Total signatures found: %d\n\n", len(signatures)))

	if len(signatures) > 0 {
		sb.WriteString("DECIMAL       HEX         DESCRIPTION\n")
		sb.WriteString(strings.Repeat("-", 60) + "\n")
		for _, sig := range signatures {
			sb.WriteString(fmt.Sprintf("%-12d  %-10s  %s\n", sig.Offset, sig.OffsetHex, sig.Description))

	} else {
		sb.WriteString("No signatures found.\n")

	return ok(sb.String())
}

}
}

// HandleBinwalkExtract extracts embedded files and filesystems from firmware.
func HandleBinwalkExtract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "filepath")
	if filePath == "" {
		return err("filepath is required")
}

	if _, checkErr := os.Stat(filePath); checkErr != nil {
		return err(fmt.Sprintf("file not found: %s", filePath))
}

	outputDir, _ :=getString(args, "output_dir")
	if outputDir == "" {
		outputDir = "/app/output"
	}

	scanID := fmt.Sprintf("%d", time.Now().UnixNano())
	extractionDir := filepath.Join(outputDir, "extract_"+scanID)
	mkErr := os.MkdirAll(extractionDir, 0755)
	if mkErr != nil {
		return err(fmt.Sprintf("failed to create extraction directory: %s", mkErr.Error()))
}

	timeout, _ :=getInt(args, "timeout")
	if timeout <= 0 {
		timeout = 300
	}

	ctx2, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx2, "binwalk", "-e", "-C", extractionDir, filePath)
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("binwalk extract failed: %s: %s", cmdErr.Error(), string(out)))
}

	extractedFiles := listExtractedFiles(extractionDir, 100)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Extraction results for: %s\n", filepath.Base(filePath)))
	sb.WriteString(fmt.Sprintf("Extraction directory: %s\n", extractionDir))
	sb.WriteString(fmt.Sprintf("Extracted files: %d\n\n", len(extractedFiles)))

	if len(extractedFiles) > 0 {
		sb.WriteString("Extracted files:\n")
		for _, f := range extractedFiles {
			sb.WriteString(fmt.Sprintf("  - %s\n", f))

	} else {
		sb.WriteString("No files extracted.\n")

	return ok(sb.String())
}

}
}

// HandleBinwalkEntropy analyzes entropy to detect compression/encryption in firmware.
func HandleBinwalkEntropy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "filepath")
	if filePath == "" {
		return err("filepath is required")
}

	if _, checkErr := os.Stat(filePath); checkErr != nil {
		return err(fmt.Sprintf("file not found: %s", filePath))
}

	timeout, _ :=getInt(args, "timeout")
	if timeout <= 0 {
		timeout = 300
	}

	ctx2, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx2, "binwalk", "-E", filePath)
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("binwalk entropy analysis failed: %s: %s", cmdErr.Error(), string(out)))
}

	entropyBlocks := parseEntropyOutput(string(out))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Entropy analysis for: %s\n", filepath.Base(filePath)))
	sb.WriteString(fmt.Sprintf("Entropy edges detected: %d\n\n", len(entropyBlocks)))

	if len(entropyBlocks) > 0 {
		sb.WriteString("OFFSET        DESCRIPTION\n")
		sb.WriteString(strings.Repeat("-", 60) + "\n")
		for _, block := range entropyBlocks {
			sb.WriteString(fmt.Sprintf("%-12d  %s\n", block.Offset, block.Description))

	} else {
		sb.WriteString("No entropy edges detected.\n")

	sb.WriteString("\nRaw output:\n")
	sb.WriteString(string(out))

	return ok(sb.String())
}

}
}

// HandleBinwalkHexdump displays hex dump of file sections.
func HandleBinwalkHexdump(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "filepath")
	if filePath == "" {
		return err("filepath is required")
}

	if _, checkErr := os.Stat(filePath); checkErr != nil {
		return err(fmt.Sprintf("file not found: %s", filePath))
}

	offset, _ :=getInt(args, "offset")
	length, _ :=getInt(args, "length")
	if length <= 0 {
		length = 512
	}

	timeout, _ :=getInt(args, "timeout")
	if timeout <= 0 {
		timeout = 300
	}

	ctx2, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	args2 := []string{"-W", filePath}
	if offset > 0 {
		args2 = append(args2, "-o", strconv.Itoa(offset))

	args2 = append(args2, "-L", strconv.Itoa(length))

	cmd := exec.CommandContext(ctx2, "binwalk", args2...)
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("binwalk hexdump failed: %s: %s", cmdErr.Error(), string(out)))
}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Hex dump for: %s (offset: %d, length: %d)\n\n", filepath.Base(filePath), offset, length))
	sb.WriteString(string(out))

	return ok(sb.String())
}

// Internal types and helpers

type SignatureMatch struct {
	Offset      int
	OffsetHex   string
	Description string
}

type EntropyBlock struct {
	Offset      int
	Entropy     float64
	Description string
}

var binwalkLineRe = regexp.MustCompile(`^\s*(\d+)\s+(0x[0-9A-Fa-f]+)\s+(.+)$`)

}

func parseBinwalkOutput(output string) []SignatureMatch {
	var signatures []SignatureMatch
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "DECIMAL") || strings.HasPrefix(line, "-") {
			continue
		}
		matches := binwalkLineRe.FindStringSubmatch(line)
		if len(matches) >= 4 {
			decimalOffset, parseErr := strconv.Atoi(matches[1])
			if parseErr != nil {
				continue
			}
			signatures = append(signatures, SignatureMatch{
				Offset:      decimalOffset,
				OffsetHex:   matches[2],
				Description: matches[3],
			})

	}
	return signatures
}

}

func parseEntropyOutput(output string) []EntropyBlock {
	var blocks []EntropyBlock
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Rising entropy edge") || strings.Contains(line, "Falling entropy edge") {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				offset, parseErr := strconv.Atoi(parts[0])
				if parseErr == nil {
					blocks = append(blocks, EntropyBlock{
						Offset:      offset,
						Entropy:     0.0,
						Description: line,
					})

			}
		}
	}
	return blocks
}

}

func listExtractedFiles(dir string, maxFiles int) []string {
	var files []string
	filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if !info.IsDir() {
			rel, relErr := filepath.Rel(dir, path)
			if relErr == nil {
				files = append(files, rel)

			if len(files) >= maxFiles {
				return fmt.Errorf("max files reached")

		}
		return nil
	})
	return files
}
}
}